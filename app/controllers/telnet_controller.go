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

type GpsQuality string

const (
	FixNotValid        GpsQuality = "Fix not valid"
	GpsFix             GpsQuality = "GPS fix"
	DifferentialGpsFix GpsQuality = "Differential GPS fix"
	NotApplicable      GpsQuality = "Not applicable"
	RtkFixed           GpsQuality = "RTK Fixed"
	RtkFloat           GpsQuality = "RTK Float"
	InsDeadReckoning   GpsQuality = "INS Dead reckoning"
)

type NMEAData struct {
	Latitude            string     `gorm:"varchar(255)" json:"latitude" binding:"required"`
	Longitude           string     `gorm:"varchar(255)" json:"longitude" binding:"required"`
	HeadingDegree       float64    `gorm:"varchar(255)" json:"heading_degree" binding:"required"`
	SpeedInKnots        float64    `gorm:"" json:"speed_in_knots" binding:"required"`
	GpsQualityIndicator GpsQuality `gorm:"type:enum('Fix not valid','GPS fix','Differential GPS fix','Not applicable','RTK Fixed','RTK Float','INS Dead reckoning');" json:"gps_quality_indicator"`
	WaterDepth          float64    `gorm:"" json:"water_depth" binding:"required"`

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
		var vesselRecord models.VesselRecord

		if err := database.DB.Preload("Kapal").Where("call_sign = ?", server.CallSign).Order("series_id desc").First(&vesselRecord).Error; err != nil {
			log.Printf("Error reading from %s: %v", server.CallSign, err)
		} else {
			nmea := NMEAData{
				Latitude:            vesselRecord.Latitude,
				Longitude:           vesselRecord.Longitude,
				HeadingDegree:       vesselRecord.HeadingDegree,
				SpeedInKnots:        vesselRecord.SpeedInKnots,
				GpsQualityIndicator: GpsQuality(vesselRecord.GpsQualityIndicator),
				WaterDepth:          vesselRecord.WaterDepth,
				Status:              "Disconnected",
			}

			r.KapalDataMap.Store(vesselRecord.CallSign, KapalData{
				Kapal: *vesselRecord.Kapal,
				NMEA:  nmea,
			})

			r.DataMap.Store(server.CallSign, nmea)
		}

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

					r.DataMap.Store(server.CallSign, nmeaData)

					for {
						select {
						case <-stopTelnet:
							return
						default:
							nmeaData.Status = "Connected"

							conn.SetReadDeadline(time.Now().Add(10 * time.Second))
							line, err := reader.ReadString('\n')
							line = strings.TrimSpace(line)

							if err != nil {
								log.Printf("Error reading from %s: %v", server.CallSign, err)
								nmeaData.Status = "Disconnected"
								r.DataMap.Store(server.CallSign, nmeaData)
								return
							}

							lastActivity = time.Now()

							if strings.HasPrefix(line, "$GPGGA") || strings.HasPrefix(line, "$GNGGA") {
								gga, err := parseGGASentence(line)
								if err != nil {
									log.Printf("Error parsing GGA: %v", err)
								}
								if gga != nil {
									nmeaData.Latitude = gga.LatMinute
									nmeaData.Longitude = gga.LongMinute
									nmeaData.GpsQualityIndicator = GpsQuality(gga.GPSQuality)
								}
							} else if strings.HasPrefix(line, "$GPHDT") || strings.HasPrefix(line, "$GNHDT") {
								hdt, err := parseHDTSentence(line)
								if err != nil {
									log.Printf("Error parsing GGA: %v", err)
								}
								if hdt != nil {
									nmeaData.HeadingDegree = hdt.Heading
								}
							} else if strings.HasPrefix(line, "$GPVTG") || strings.HasPrefix(line, "$GNVTG") {
								vtg, err := parseVTGSentence(line)
								if err != nil {
									log.Printf("Error parsing GGA: %v", err)
								}
								if vtg != nil {
									nmeaData.SpeedInKnots = vtg.SpeedKnots
								}
							}
							r.DataMap.Store(server.CallSign, nmeaData)

							if time.Since(lastCoordinateInsertTime) >= 1*time.Second {
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

				time.Sleep(retryDelay)
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
	var lastRecord models.VesselRecord
	result := database.DB.Where("call_sign = ?", callSign).Order("created_at desc").First(&lastRecord)

	kapal, ok := r.KapalDataMap.Load(callSign)
	if !ok {
		return fmt.Errorf("kapal not found in dataMap")
	}

	if result.Error != nil && !errors.Is(result.Error, gorm.ErrRecordNotFound) {
		fmt.Errorf("error fetching last coordinate: %w", result.Error)
	}
	if result.RowsAffected <= 0 {

		if nmeaData.Longitude != "" && nmeaData.Latitude != "" {
			gpsQuality, err := models.StringToGpsQuality(string(nmeaData.GpsQualityIndicator))
			if err != nil {
				return err
			}
			if err := database.DB.Create(&models.VesselRecord{
				CallSign:            callSign,
				SeriesID:            1,
				Latitude:            nmeaData.Latitude,
				Longitude:           nmeaData.Longitude,
				HeadingDegree:       nmeaData.HeadingDegree,
				SpeedInKnots:        nmeaData.SpeedInKnots,
				WaterDepth:          nmeaData.WaterDepth,
				GpsQualityIndicator: models.GpsQuality(gpsQuality),
			}).Error; err != nil {
				return fmt.Errorf("error creating new coordinate: %w", err)
			}
		}
	} else if time.Since(lastRecord.CreatedAt) >= (time.Duration(kapal.(KapalData).Kapal.HistoryPerSecond))*time.Second {
		gpsQuality, err := models.StringToGpsQuality(string(nmeaData.GpsQualityIndicator))
		if err != nil {
			return err
		}

		if nmeaData.Longitude != "" && nmeaData.Latitude != "" {
			if err := database.DB.Create(&models.VesselRecord{
				CallSign:            callSign,
				SeriesID:            lastRecord.SeriesID + 1,
				Latitude:            nmeaData.Latitude,
				Longitude:           nmeaData.Longitude,
				HeadingDegree:       nmeaData.HeadingDegree,
				SpeedInKnots:        nmeaData.SpeedInKnots,
				WaterDepth:          nmeaData.WaterDepth,
				GpsQualityIndicator: models.GpsQuality(gpsQuality),
			}).Error; err != nil {
				return fmt.Errorf("error creating new coordinate: %w", err)
			}
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

func calculateChecksum(sentence string) string {
	checksum := 0
	for i := 1; i < len(sentence); i++ {
		checksum ^= int(sentence[i])
	}
	return fmt.Sprintf("%02X", checksum)
}

func (r *TelnetController) GetKapalDataByCallSign(callSign string) (*KapalData, error) {
	data, ok := r.KapalDataMap.Load(callSign)
	if !ok {
		return nil, fmt.Errorf("data not found for call sign: %s", callSign)
	}
	kapalData := data.(KapalData)
	return &kapalData, nil
}

// Constants for GPS quality descriptions
var gpsQualityDescriptions = []string{
	"Fix not valid",
	"GPS fix",
	"Differential GPS fix",
	"Not applicable",
	"RTK Fixed",
	"RTK Float",
	"INS Dead reckoning",
}

// GGASentence represents a parsed GGA sentence
type GGASentence struct {
	Latitude   float64
	LatMinute  string
	Longitude  float64
	LongMinute string
	GPSQuality string
}

// HDTSentence represents a parsed HDT sentence
type HDTSentence struct {
	Heading float64
}

// VTGSentence represents a parsed VTG sentence
type VTGSentence struct {
	CourseTrue        float64
	CourseMagnetic    *float64
	SpeedKnots        float64
	SpeedKmh          float64
	ModeIndicator     string
	ModeIndicatorText string
}

// Convert DMS to decimal degrees
func convertDMSToDecimal(dms float64, direction string) (float64, error) {
	degrees := float64(int(dms / 100))
	minutes := dms - (degrees * 100)
	decimal := degrees + (minutes / 60)

	if direction == "S" || direction == "W" {
		decimal *= -1
	}

	return decimal, nil
}

// Parse GGA sentence
func parseGGASentence(gga string) (*GGASentence, error) {
	fields := strings.Split(gga, ",")
	if len(fields) < 15 {
		return nil, fmt.Errorf("invalid GGA sentence")
	}

	latitudeDMS, err := strconv.ParseFloat(fields[2], 64)
	if err != nil {
		return nil, err
	}

	longitudeDMS, err := strconv.ParseFloat(fields[4], 64)
	if err != nil {
		return nil, err
	}

	latitude, err := convertDMSToDecimal(latitudeDMS, fields[3])
	if err != nil {
		return nil, err
	}

	longitude, err := convertDMSToDecimal(longitudeDMS, fields[5])
	if err != nil {
		return nil, err
	}

	gpsQuality := gpsQualityDescriptions[atoi(fields[6])]

	return &GGASentence{
		Latitude:   latitude,
		LatMinute:  formatCoordinate(latitudeDMS, fields[3]),
		Longitude:  longitude,
		LongMinute: formatCoordinate(longitudeDMS, fields[5]),
		GPSQuality: gpsQuality,
	}, nil
}

func formatCoordinate(dms float64, direction string) string {
	degrees := int(dms / 100)
	minutes := dms - float64(degrees*100)
	if minutes < 0 {
		minutes = -minutes
	}
	return fmt.Sprintf("%d°%.4f°%s", degrees, minutes, direction)
}

// Parse HDT sentence
func parseHDTSentence(hdt string) (*HDTSentence, error) {
	fields := strings.Split(hdt, ",")
	if len(fields) < 3 {
		return nil, fmt.Errorf("invalid HDT sentence")
	}

	heading, err := strconv.ParseFloat(fields[1], 64)
	if err != nil {
		return nil, err
	}

	return &HDTSentence{
		Heading: heading,
	}, nil
}

// Parse VTG sentence
func parseVTGSentence(vtg string) (*VTGSentence, error) {
	fields := strings.Split(vtg, ",")
	if len(fields) < 10 {
		return nil, fmt.Errorf("invalid VTG sentence")
	}

	courseTrue, err := strconv.ParseFloat(fields[1], 64)
	if err != nil {
		return nil, err
	}

	var courseMagnetic *float64
	if fields[3] != "" {
		courseMagneticValue, err := strconv.ParseFloat(fields[3], 64)
		if err != nil {
			return nil, err
		}
		courseMagnetic = &courseMagneticValue
	}

	speedKnots, err := strconv.ParseFloat(fields[5], 64)
	if err != nil {
		return nil, err
	}

	speedKmh, err := strconv.ParseFloat(fields[7], 64)
	if err != nil {
		return nil, err
	}

	modeIndicator := fields[9]
	modeIndicatorText := getModeIndicatorText(modeIndicator)

	return &VTGSentence{
		CourseTrue:        courseTrue,
		CourseMagnetic:    courseMagnetic,
		SpeedKnots:        speedKnots,
		SpeedKmh:          speedKmh,
		ModeIndicator:     modeIndicator,
		ModeIndicatorText: modeIndicatorText,
	}, nil
}

func getModeIndicatorText(modeIndicator string) string {
	switch modeIndicator {
	case "A":
		return "Autonomous"
	case "D":
		return "Differential"
	case "E":
		return "Estimated"
	case "M":
		return "Manual Input"
	case "S":
		return "Simulator"
	case "N":
		return "Data Not Valid"
	default:
		return "Unknown"
	}
}

func atoi(s string) int {
	value, _ := strconv.Atoi(s)
	return value
}
