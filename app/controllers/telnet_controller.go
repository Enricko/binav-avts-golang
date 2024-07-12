package controllers

import (
	"bufio"
	"errors"
	"fmt"
	"golang-app/app/models"
	"golang-app/database"
	"log"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"gorm.io/gorm"

)

type TelnetController struct {
	DataMap        *sync.Map
	ConnMap        *sync.Map
	KapalDataMap   *sync.Map
	UpdateInterval time.Duration
}

func NewTelnetController() *TelnetController {
	return &TelnetController{
		DataMap:        &sync.Map{},
		ConnMap:        &sync.Map{},
		KapalDataMap:   &sync.Map{},
		UpdateInterval: 5 * time.Second,
	}
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type NMEAData struct {
	GGA    string `json:"gga,omitempty"`
	HDT    string `json:"hdt,omitempty"`
	VTG    string `json:"vtg,omitempty"`
	Status string `json:"status"`
}

type KapalData struct {
	Kapal models.Kapal `json:"kapal"`
	NMEA  NMEAData     `json:"nmea"`
}

func (r *TelnetController) StartTelnetConnections() {
	var wg sync.WaitGroup
	var servers []models.IPKapal

	err := database.DB.Find(&servers).Error
	if err != nil {
		log.Println("Error fetching Telnet servers:", err)
		return
	}

	for _, server := range servers {
		wg.Add(1)
		var coordinate models.Coordinate

		if err := database.DB.Preload("Kapal").Preload("CoordinateGga").Preload("CoordinateHdt").Preload("CoordinateVtg").Where("call_sign = ?", server.CallSign).Order("series_id desc").First(&coordinate).Error; err != nil {
			log.Printf("Error reading from %s: %v", server.CallSign, err)
			return
		}
		fmt.Println(coordinate.Kapal)
		nmea := NMEAData{
			Status: "Disconnected",
		}
		if coordinate.IdCoorGGA != nil {
			gga := unparseGGA(*coordinate.CoordinateGga)
			nmea.GGA = gga
		}
		if coordinate.IdCoorHDT != nil {
			hdt := unparseHDT(*coordinate.CoordinateHdt)
			nmea.HDT = hdt
		}
		if coordinate.IdCoorVTG != nil {
			vtg := unparseVTG(*coordinate.CoordinateVtg)
			nmea.VTG = vtg
		}
		r.KapalDataMap.Store(coordinate.CallSign, KapalData{
			Kapal: *coordinate.Kapal,
			NMEA: nmea,
		})
		
		r.DataMap.Store(server.CallSign, nmea)

		stopChan := make(chan struct{})
		go r.handleTelnetConnection(server, &wg, stopChan)
	}

	go func() {
		for {
			r.updateKapalDataMap()
			time.Sleep(r.UpdateInterval)
		}
	}()

	go func() {
		for {
			time.Sleep(30 * time.Second)
			var updatedServers []models.IPKapal
			err := database.DB.Find(&updatedServers).Error
			if err != nil {
				log.Println("Error fetching updated Telnet servers:", err)
				continue
			}

			currentServerMap := make(map[uint]models.IPKapal)
			for _, server := range servers {
				currentServerMap[server.IdIpKapal] = server
			}

			updatedServerMap := make(map[uint]models.IPKapal)
			for _, server := range updatedServers {
				updatedServerMap[server.IdIpKapal] = server
			}

			for id, updatedServer := range updatedServerMap {
				if currentServer, exists := currentServerMap[id]; !exists || currentServer != updatedServer {
					if stopChan, ok := r.ConnMap.Load(id); ok {
						close(stopChan.(chan struct{}))
						r.ConnMap.Delete(id)
					}
					wg.Add(1)
					stopChan := make(chan struct{})
					go r.handleTelnetConnection(updatedServer, &wg, stopChan)
				}
			}

			for id := range currentServerMap {
				if _, exists := updatedServerMap[id]; !exists {
					if stopChan, ok := r.ConnMap.Load(id); ok {
						close(stopChan.(chan struct{}))
						r.ConnMap.Delete(id)
					}
					r.DataMap.Delete(currentServerMap[id].CallSign)
				}
			}

			servers = updatedServers
		}
	}()

	wg.Wait()
}

func (r *TelnetController) handleTelnetConnection(server models.IPKapal, wg *sync.WaitGroup, stopChan chan struct{}) {
	defer wg.Done()

	r.ConnMap.Store(server.IdIpKapal, stopChan)
	retryDelay := 5 * time.Second
	stopTelnet := make(chan struct{})
	defer close(stopTelnet)

	data, _ := r.DataMap.LoadOrStore(server.CallSign, NMEAData{})
	nmeaData := data.(NMEAData)

	var lastActivity time.Time
	lastCoordinateInsertTime := time.Now()

	go func() {
		defer func() {
			r.ConnMap.Delete(server.IdIpKapal)
			log.Printf("Stopped telnet connection for %s", server.CallSign)
		}()

		for {
			select {
			case <-stopTelnet:
				return
			default:
				conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", server.IP, server.Port))
				if err != nil {
					log.Printf("Error connecting to %s: %v", server.CallSign, err)
					nmeaData.Status = "Disconnected"
					r.DataMap.Store(server.CallSign, nmeaData)

					time.Sleep(retryDelay)
					retryDelay = min(2*retryDelay, 5*time.Minute)
					continue
				}

				retryDelay = 5 * time.Second
				reader := bufio.NewReader(conn)
				connClosed := make(chan struct{})
				lastActivity = time.Now()

				go func() {
					defer func() {
						conn.Close()
						close(connClosed)
					}()

					nmeaData.Status = "Connected"
					r.DataMap.Store(server.CallSign, nmeaData)

					for {
						select {
						case <-stopTelnet:
							return
						default:
							conn.SetReadDeadline(time.Now().Add(10 * time.Second))
							line, err := reader.ReadString('\n')
							line = strings.TrimSpace(line)

							if err != nil {
								log.Printf("Error reading from %s: %v", server.CallSign, err)
								nmeaData.Status = "Disconnected"
								return
							}

							lastActivity = time.Now()

							if strings.HasPrefix(line, "$GPGGA") || strings.HasPrefix(line, "$GNGGA") {
								nmeaData.GGA = line
							} else if strings.HasPrefix(line, "$GPHDT") || strings.HasPrefix(line, "$GNHDT") {
								nmeaData.HDT = line
							} else if strings.HasPrefix(line, "$GPVTG") || strings.HasPrefix(line, "$GNVTG") {
								nmeaData.VTG = line
							}
							r.DataMap.Store(server.CallSign, nmeaData)

							if time.Since(lastCoordinateInsertTime) >= 15*time.Second {
								lastCoordinateInsertTime = time.Now()
								if err := r.createOrUpdateCoordinate(server.CallSign, &lastCoordinateInsertTime, nmeaData); err != nil {
									log.Printf("Error creating or updating coordinate: %v", err)
								}
							}
						}
					}
				}()

				<-connClosed
				log.Printf("Disconnected from %s. Reconnecting...", server.CallSign)
			}
		}
	}()

	sqlTicker := time.NewTicker(5 * time.Second)
	defer sqlTicker.Stop()

	for {
		select {
		case <-stopChan:
			close(stopTelnet)
			return
		case <-sqlTicker.C:
			if time.Since(lastActivity) > 10*time.Second {
				nmeaData.Status = "Disconnected"
				r.DataMap.Store(server.CallSign, nmeaData)
			}
		}
	}
}

func (r *TelnetController) updateKapalDataMap() {
	var activeKapal []models.Kapal
	err := database.DB.Where("status = ?", true).Find(&activeKapal).Error
	if err != nil {
		log.Println("Error fetching active kapal:", err)
		return
	}

	for _, kapal := range activeKapal {
		nmeaData, ok := r.DataMap.Load(kapal.CallSign)
		if !ok {
			nmeaData = NMEAData{Status: "Disconnected"}
		}

		r.KapalDataMap.Store(kapal.CallSign, KapalData{
			Kapal: kapal,
			NMEA:  nmeaData.(NMEAData),
		})
	}
}

func (r *TelnetController) createOrUpdateCoordinate(callSign string, lastCoordinateInsertTime *time.Time, nmeaData NMEAData) error {
	var lastCoordinate models.Coordinate
	result := database.DB.Where("call_sign = ?", callSign).Order("created_at desc").First(&lastCoordinate)

	kapal, ok := r.KapalDataMap.Load(callSign)
	if !ok {
		return fmt.Errorf("kapal not found in dataMap")
	}

	if result.Error != nil && !errors.Is(result.Error, gorm.ErrRecordNotFound) {
		fmt.Errorf("error fetching last coordinate: %w", result.Error)
	}
	if result.RowsAffected <= 0 {
		if err := database.DB.Create(&models.Coordinate{
			CallSign: callSign,
			SeriesID: 1,
		}).Error; err != nil {
			return fmt.Errorf("error creating new coordinate: %w", err)
		}
	} else if time.Since(lastCoordinate.CreatedAt) >= 5*time.Minute {
		if err := database.DB.Create(&models.Coordinate{
			CallSign: callSign,
			SeriesID: lastCoordinate.SeriesID + 1,
		}).Error; err != nil {
			return fmt.Errorf("error creating new coordinate: %w", err)
		}
	}

	if lastCoordinate.IdCoorGGA == nil && nmeaData.Status == "Connected" && nmeaData.GGA != "" {
		gga, err := parseGGA(nmeaData.GGA)
		if err != nil {
			return fmt.Errorf("error parsing GGA: %w", err)
		}

		coorGGA := models.CoordinateGga{
			CallSign:            callSign,
			MessageID:           gga.MessageID,
			UtcPosition:         gga.UtcPosition,
			Latitude:            gga.Latitude,
			DirectionLatitude:   gga.DirectionLatitude,
			Longitude:           gga.Longitude,
			DirectionLongitude:  gga.DirectionLongitude,
			GpsQualityIndicator: gga.GpsQualityIndicator,
			NumberSv:            gga.NumberSv,
			Hdop:                gga.Hdop,
			OrthometricHeight:   gga.OrthometricHeight,
			UnitMeasure:         gga.UnitMeasure,
			GeoidSeparation:     gga.GeoidSeparation,
			GeoidMeasure:        gga.GeoidMeasure,
		}

		if err := database.DB.Create(&coorGGA).Error; err != nil {
			return fmt.Errorf("error creating new GGA: %w", err)
		}

		idCoorGGA := coorGGA.IdCoorGGA
		lastCoordinate.IdCoorGGA = &idCoorGGA

		if err := database.DB.Save(&lastCoordinate).Error; err != nil {
			return fmt.Errorf("failed to update coordinate: %w", err)
		}
	}
	if lastCoordinate.IdCoorHDT == nil && nmeaData.Status == "Connected" && nmeaData.HDT != "" {
		hdt, err := parseHDT(nmeaData.HDT)
		if err != nil {
			return fmt.Errorf("error parsing HDT: %w", err)
		}
		coorHDT := models.CoordinateHdt{
			CallSign:      callSign,
			MessageID:     hdt.MessageID,
			HeadingDegree: hdt.HeadingDegree + float32(kapal.(KapalData).Kapal.Calibration),
			Checksum:      hdt.Checksum,
		}

		if err := database.DB.Create(&coorHDT).Error; err != nil {
			return fmt.Errorf("error creating new GGA: %w", err)
		}
		idCoorHDT := coorHDT.IdCoorHDT
		lastCoordinate.IdCoorHDT = &idCoorHDT

		if err := database.DB.Save(&lastCoordinate).Error; err != nil {
			return fmt.Errorf("failed to update coordinate: %w", err)
		}
	}

	if lastCoordinate.IdCoorVTG == nil && nmeaData.Status == "Connected" && nmeaData.VTG != "" {
		vtg, err := parseVTG(nmeaData.VTG)
		if err != nil {
			return fmt.Errorf("error parsing VTG: %w", err)
		}
		coorVTG := models.CoordinateVtg{
			CallSign:            callSign,
			MessageID:           vtg.MessageID,
			TrackDegreeTrue:     vtg.TrackDegreeTrue,
			TrueNorth:           vtg.TrueNorth,
			TrackDegreeMagnetic: vtg.TrackDegreeMagnetic,
			MagneticNorth:       vtg.MagneticNorth,
			SpeedInKnots:        vtg.SpeedInKnots,
			MeasuredKnots:       vtg.MeasuredKnots,
			Kph:                 vtg.Kph,
			MeasuredKph:         vtg.MeasuredKph,
			ModeIndicator:       vtg.ModeIndicator,
			Checksum:            vtg.Checksum,
		}
		if err := database.DB.Create(&coorVTG).Error; err != nil {
			return fmt.Errorf("error creating new VTG: %w", err)
		}
		idCoorVTG := coorVTG.IdCoorVTG
		lastCoordinate.IdCoorVTG = &idCoorVTG
		if err := database.DB.Save(&lastCoordinate).Error; err != nil {
			return fmt.Errorf("failed to update coordinate: %w", err)
		}
	}

	r.DataMap.Store(callSign, nmeaData)
	return nil
}

func (r *TelnetController) KapalTelnetWebsocketHandler(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		http.NotFound(c.Writer, c.Request)
		return
	}
	defer conn.Close()

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			data := make(map[string]KapalData)
			r.KapalDataMap.Range(func(key, value interface{}) bool {
				kapalData := value.(KapalData)
				nmeaData, ok := r.DataMap.Load(kapalData.Kapal.CallSign)
				if ok {
					kapalData.NMEA = nmeaData.(NMEAData)
				}
				data[key.(string)] = kapalData
				return true
			})

			if err := conn.WriteJSON(data); err != nil {
				log.Println("Error writing to WebSocket client:", err)
				return
			}
		}
	}
}

func min(a, b time.Duration) time.Duration {
	if a < b {
		return a
	}
	return b
}

// Example: $GPGGA,123456.78,4916.45,N,12311.12,W,1,08,0.9,545.4,M,46.9,M,,*47
func parseGGA(line string) (models.CoordinateGga, error) {
	fields := strings.Split(line, ",")
	if len(fields) < 15 {
		return models.CoordinateGga{}, errors.New("invalid GGA sentence")
	}

	utcPosition, _ := strconv.ParseFloat(fields[1], 32)
	latitude, _ := strconv.ParseFloat(fields[2], 32)
	longitude, _ := strconv.ParseFloat(fields[4], 32)
	numberSv, _ := strconv.Atoi(fields[7])
	hdop, _ := strconv.ParseFloat(fields[8], 32)
	orthometricHeight, _ := strconv.ParseFloat(fields[9], 32)
	geoidSeparation, _ := strconv.ParseFloat(fields[11], 32)

	return models.CoordinateGga{
		MessageID:           fields[0],
		UtcPosition:         float32(utcPosition),
		Latitude:            float32(latitude),
		DirectionLatitude:   fields[3],
		Longitude:           float32(longitude),
		DirectionLongitude:  fields[5],
		GpsQualityIndicator: models.GpsQuality(fields[6]),
		NumberSv:            numberSv,
		Hdop:                float32(hdop),
		OrthometricHeight:   float32(orthometricHeight),
		UnitMeasure:         fields[10],
		GeoidSeparation:     float32(geoidSeparation),
		GeoidMeasure:        fields[12],
	}, nil
}

// Example: $GPHDT,274.07,T*03
func parseHDT(line string) (models.CoordinateHdt, error) {
	fields := strings.Split(line, ",")
	if len(fields) < 3 {
		return models.CoordinateHdt{}, errors.New("invalid HDT sentence")
	}

	headingDegree, _ := strconv.ParseFloat(fields[1], 32)

	return models.CoordinateHdt{
		MessageID:     fields[0],
		HeadingDegree: float32(headingDegree),
		Checksum:      fields[2],
	}, nil
}

// Example: $GPVTG,054.7,T,034.4,M,005.5,N,010.2,K*48
func parseVTG(line string) (models.CoordinateVtg, error) {
	fields := strings.Split(line, ",")
	if len(fields) < 10 {
		return models.CoordinateVtg{}, errors.New("invalid VTG sentence")
	}

	trackDegreeTrue, _ := strconv.ParseFloat(fields[1], 32)
	trackDegreeMagnetic, _ := strconv.ParseFloat(fields[3], 32)
	speedInKnots, _ := strconv.ParseFloat(fields[5], 32)
	kph, _ := strconv.ParseFloat(fields[7], 32)

	return models.CoordinateVtg{
		MessageID:           fields[0],
		TrackDegreeTrue:     float32(trackDegreeTrue),
		TrueNorth:           fields[2],
		TrackDegreeMagnetic: float32(trackDegreeMagnetic),
		MagneticNorth:       fields[4],
		SpeedInKnots:        float32(speedInKnots),
		MeasuredKnots:       fields[6],
		Kph:                 float32(kph),
		MeasuredKph:         fields[8],
		Checksum:            fields[9],
		ModeIndicator:       determineModeIndicator(string(fields[9][0])), // Assuming mode indicator is at index 10
	}, nil
}

func determineModeIndicator(value string) models.ModeIndicator {
	switch value {
	case "A":
		return models.AutonomousMode
	case "D":
		return models.DifferentialMode
	case "E":
		return models.EstimatedMode
	case "M":
		return models.ManualInputMode
	case "S":
		return models.SimulatorMode
	default:
		return models.DataNotValidMode
	}
}

func unparseGGA(gga models.CoordinateGga) string {
	gpsQuality := gpsQualityToInt(gga.GpsQualityIndicator)

	return fmt.Sprintf("%s,%.2f,%.5f,%s,%.5f,%s,%d,%02d,%.1f,%.1f,%s,%.1f,%s,,*%s",
		gga.MessageID,
		gga.UtcPosition,
		gga.Latitude,
		gga.DirectionLatitude,
		gga.Longitude,
		gga.DirectionLongitude,
		gpsQuality,
		gga.NumberSv,
		gga.Hdop,
		gga.OrthometricHeight,
		gga.UnitMeasure,
		gga.GeoidSeparation,
		gga.GeoidMeasure,
		calculateChecksum(fmt.Sprintf("%s,%.2f,%.5f,%s,%.5f,%s,%d,%02d,%.1f,%.1f,%s,%.1f,%s,,",
			gga.MessageID,
			gga.UtcPosition,
			gga.Latitude,
			gga.DirectionLatitude,
			gga.Longitude,
			gga.DirectionLongitude,
			gpsQuality,
			gga.NumberSv,
			gga.Hdop,
			gga.OrthometricHeight,
			gga.UnitMeasure,
			gga.GeoidSeparation,
			gga.GeoidMeasure)),
	)
}

func gpsQualityToInt(quality models.GpsQuality) int {
	switch quality {
	case models.FixNotValid:
		return 0
	case models.GpsFix:
		return 1
	case models.DifferentialGpsFix:
		return 2
	case models.NotApplicable:
		return 3
	case models.RtkFixed:
		return 4
	case models.RtkFloat:
		return 5
	case models.InsDeadReckoning:
		return 6
	default:
		return -1
	}
}

func unparseHDT(hdt models.CoordinateHdt) string {
	return fmt.Sprintf("%s,%.2f,T*%s",
		hdt.MessageID,
		hdt.HeadingDegree,
		calculateChecksum(fmt.Sprintf("%s,%.2f,T",
			hdt.MessageID,
			hdt.HeadingDegree)),
	)
}

func unparseVTG(vtg models.CoordinateVtg) string {
	modeIndicator := modeIndicatorToChar(vtg.ModeIndicator)

	return fmt.Sprintf("%s,%.1f,%s,%.1f,%s,%.1f,%s,%.1f,%s,%s*%s",
		vtg.MessageID,
		vtg.TrackDegreeTrue,
		"T", // True track made good
		vtg.TrackDegreeMagnetic,
		"M", // Magnetic track made good
		vtg.SpeedInKnots,
		"N", // Speed in knots
		vtg.Kph,
		"K", // Speed in kilometers per hour
		modeIndicator,
		calculateChecksum(fmt.Sprintf("%s,%.1f,%s,%.1f,%s,%.1f,%s,%.1f,%s,%s",
			vtg.MessageID,
			vtg.TrackDegreeTrue,
			"T",
			vtg.TrackDegreeMagnetic,
			"M",
			vtg.SpeedInKnots,
			"N",
			vtg.Kph,
			"K",
			modeIndicator)),
	)
}

func modeIndicatorToChar(mode models.ModeIndicator) string {
	switch mode {
	case models.AutonomousMode:
		return "A"
	case models.DifferentialMode:
		return "D"
	case models.EstimatedMode:
		return "E"
	case models.ManualInputMode:
		return "M"
	case models.SimulatorMode:
		return "S"
	case models.DataNotValidMode:
		return "N"
	default:
		return ""
	}
}

func calculateChecksum(sentence string) string {
	checksum := 0
	for i := 1; i < len(sentence); i++ {
		checksum ^= int(sentence[i])
	}
	return fmt.Sprintf("%02X", checksum)
}
