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
	"regexp"
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
	mu             sync.Mutex
	lastDataTime   *sync.Map
}

func NewTelnetController() *TelnetController {
	controller := &TelnetController{
		DataMap:        &sync.Map{},
		ConnMap:        &sync.Map{},
		KapalDataMap:   &sync.Map{},
		UpdateInterval: 5 * time.Second,
		lastDataTime:   &sync.Map{},
	}

	go controller.checkStaleConnections()

	return controller
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
	Latitude            string              `gorm:"varchar(255)" json:"latitude" binding:"required"`
	Longitude           string              `gorm:"varchar(255)" json:"longitude" binding:"required"`
	HeadingDegree       float64             `gorm:"varchar(255)" json:"heading_degree" binding:"required"`
	SpeedInKnots        float64             `gorm:"" json:"speed_in_knots" binding:"required"`
	GpsQualityIndicator GpsQuality          `gorm:"type:enum('Fix not valid','GPS fix','Differential GPS fix','Not applicable','RTK Fixed','RTK Float','INS Dead reckoning');" json:"gps_quality_indicator"`
	WaterDepth          float64             `gorm:"" json:"water_depth" binding:"required"`
	Status              string              `json:"status"`
	TelnetStatus        models.TelnetStatus `json:"telnet_status"`
}

type KapalData struct {
	Kapal models.Kapal `json:"kapal"`
	NMEA  NMEAData     `json:"nmea"`
}

type ConnectionError struct {
	Err error
}

func (e ConnectionError) Error() string {
	return fmt.Sprintf("connection error: %v", e.Err)
}

func (r *TelnetController) StartTelnetConnections() {
	var wg sync.WaitGroup
	var servers []models.IPKapal

	err := database.DB.Preload("Kapal").Find(&servers).Error
	if err != nil {
		log.Println("Error fetching Telnet servers:", err)
		return
	}

	for _, server := range servers {
		var vesselRecord models.VesselRecord

		if err := database.DB.Preload("Kapal").Where("call_sign = ?", server.CallSign).Last(&vesselRecord).Error; err != nil {
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
			time.Sleep(15 * time.Second)

			var updatedServers []models.IPKapal
			err := database.DB.Preload("Kapal").Find(&updatedServers).Error
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
				currentServer, exists := currentServerMap[id]
				if !exists || r.hasServerChanged(currentServer, updatedServer) {
					log.Printf("Updating connection for server ID: %d, IP: %s, Port: %s", id, updatedServer.IP, updatedServer.Port)
					r.stopAndRemoveConnection(id)

					wg.Add(1)
					stopChan := make(chan struct{})
					go r.handleTelnetConnection(updatedServer, &wg, stopChan)
				}
			}

			for id, currentServer := range currentServerMap {
				if _, exists := updatedServerMap[id]; !exists {
					log.Printf("Removing connection for server ID: %d, CallSign: %s", id, currentServer.CallSign)
					r.stopAndRemoveConnection(id)
					r.DataMap.Delete(currentServer.CallSign)
				}
			}

			servers = updatedServers
		}
	}()

	wg.Wait()
}

func (r *TelnetController) hasServerChanged(current, updated models.IPKapal) bool {
	return current.IP != updated.IP || current.Port != updated.Port
}

func (r *TelnetController) stopAndRemoveConnection(id uint) {
	if stopChan, ok := r.ConnMap.Load(id); ok {
		log.Printf("Stopping telnet connection for server ID: %d", id)
		close(stopChan.(chan struct{}))
		r.ConnMap.Delete(id)
	}
}

func (r *TelnetController) handleTelnetConnection(server models.IPKapal, wg *sync.WaitGroup, stopChan chan struct{}) {
	defer wg.Done()

	r.ConnMap.Store(server.IdIpKapal, stopChan)
	defer r.ConnMap.Delete(server.IdIpKapal)

	backoff := time.Second
	maxBackoff := 5 * time.Minute
	connectionEstablished := false

	nmeaDataChan := make(chan NMEAData)
	waterDepthChan := make(chan float64)
	go r.updateDataMap(server.CallSign, nmeaDataChan, waterDepthChan)

	for {
		select {
		case <-stopChan:
			r.DataMap.Delete(server.CallSign)
			r.handleDisconnection(server.CallSign)
			log.Printf("Telnet connection stopped for server ID: %d, CallSign: %s, Port:%d", server.IdIpKapal, server.CallSign, server.Port)
			return
		default:
			if !connectionEstablished {
				// Mark the latest record as disconnected before attempting to reconnect
				if err := r.markLatestRecordAsDisconnected(server.CallSign); err != nil {
					log.Printf("Error marking latest record as disconnected for %s: %v", server.CallSign, err)
				}
			}

			err := r.connectAndRead(server, nmeaDataChan, waterDepthChan)
			if err != nil {
				if _, isConnErr := err.(ConnectionError); isConnErr {
					log.Printf("Failed to establish connection for %s: %v", server.CallSign, err)
					r.handleDisconnection(server.CallSign)
				} else if connectionEstablished && models.TypeIP(server.TypeIP) != "depth" {
					r.updateStatus(server.CallSign, models.Disconnected)
				}

				time.Sleep(backoff)
				backoff *= 2
				if backoff > maxBackoff {
					backoff = maxBackoff
				}
				connectionEstablished = false
			} else {
				if !connectionEstablished {
					log.Printf("Connection established for %s", server.CallSign)
				}
				connectionEstablished = true
				backoff = time.Second
				r.updateStatus(server.CallSign, models.Connected)
			}
		}
	}
}

func (r *TelnetController) markLatestRecordAsDisconnected(callSign string) error {
	var lastRecord models.VesselRecord
	if err := database.DB.Where("call_sign = ?", callSign).Last(&lastRecord).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// No records found, nothing to update
			return nil
		}
		return fmt.Errorf("error fetching last record: %w", err)
	}

	if lastRecord.TelnetStatus != models.Disconnected {
		lastRecord.TelnetStatus = models.Disconnected
		if err := database.DB.Save(&lastRecord).Error; err != nil {
			return fmt.Errorf("error updating last record status: %w", err)
		}
	}

	return nil
}

func (r *TelnetController) connectAndRead(server models.IPKapal, nmeaDataChan chan<- NMEAData, waterDepthChan chan<- float64) error {
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", server.IP, server.Port), 10*time.Second)
	if err != nil {
		return ConnectionError{Err: fmt.Errorf("failed to establish connection: %w", err)}
	}
	defer conn.Close()

	if err := conn.SetDeadline(time.Now().Add(30 * time.Second)); err != nil {
		return ConnectionError{Err: fmt.Errorf("error setting connection deadline: %w", err)}
	}

	reader := bufio.NewReader(conn)

	lastCoordinateInsertTime := time.Now()

	for {
		if err := conn.SetReadDeadline(time.Now().Add(10 * time.Second)); err != nil {
			return fmt.Errorf("error setting read deadline: %w", err)
		}

		line, err := r.readLine(reader, server.TypeIP)
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				return fmt.Errorf("read timeout: %w", err)
			}
			return fmt.Errorf("error reading: %w", err)
		}

		sentences := strings.Split(strings.TrimSpace(line), "\n")

		fmt.Println(sentences)

		for _, sentence := range sentences {
			if err := r.processLine(sentence, server, nmeaDataChan, waterDepthChan); err != nil {
				log.Printf("Error processing line for %s: %v", server.CallSign, err)
				continue
			}
		}

		if time.Since(lastCoordinateInsertTime) >= time.Second {
			lastCoordinateInsertTime = time.Now()
			kapal, ok := r.KapalDataMap.Load(server.CallSign)
			if !ok {
				return fmt.Errorf("kapal not found in dataMap")
			}
			if models.TypeIP(server.TypeIP) != "depth" && kapal.(KapalData).Kapal.RecordStatus {
				if err := r.createOrUpdateCoordinate(server.CallSign, &lastCoordinateInsertTime); err != nil {
					log.Printf("Error creating or updating coordinate: %v", err)
				}
			}
		}
	}
}

func (r *TelnetController) readLine(reader *bufio.Reader, typeIP models.TypeIP) (string, error) {
	var line string
	var err error

	if typeIP == "depth" {
		line, err = reader.ReadString(' ')
		if err != nil {
			return "", err
		}
		line = strings.TrimSpace(line)
		number, err := extractNumber(line)
		if err != nil {
			return "", nil
		}
		line = strconv.Itoa(number)
	} else {
		data := make([]byte, 1024)
		n, err := reader.Read(data)
		if err != nil {
			return "", err
		}
		line = string(data[:n])
	}

	return line, nil
}

func (r *TelnetController) processLine(line string, server models.IPKapal, nmeaDataChan chan<- NMEAData, waterDepthChan chan<- float64) error {
	if models.TypeIP(server.TypeIP) == "depth" {
		if line == "" {
			return nil
		}
		number, err := strconv.ParseFloat(line, 64)
		if err != nil {
			return fmt.Errorf("error parsing depth: %w", err)
		}
		waterDepthChan <- number
		return nil
	}

	var nmeaData NMEAData

	switch {
	case strings.HasPrefix(line, "$GPGGA"), strings.HasPrefix(line, "$GNGGA"):
		gga, err := parseGGASentence(line)
		if err != nil {
			return fmt.Errorf("error parsing GGA sentence: %w", err)
		}
		nmeaData = r.convertGGAToNMEAData(gga)
	case strings.HasPrefix(line, "$GPHDT"), strings.HasPrefix(line, "$GNHDT"):
		hdt, err := parseHDTSentence(line)
		if err != nil {
			return fmt.Errorf("error parsing HDT sentence: %w", err)
		}
		nmeaData = r.convertHDTToNMEAData(hdt)
	case strings.HasPrefix(line, "$GPVTG"), strings.HasPrefix(line, "$GNVTG"):
		vtg, err := parseVTGSentence(line)
		if err != nil {
			return fmt.Errorf("error parsing VTG sentence: %w", err)
		}
		nmeaData = r.convertVTGToNMEAData(vtg)
	default:
		return fmt.Errorf("unknown sentence type: %s", line)
	}

	if line != "" {
		nmeaData.Status = "Connected"
		nmeaDataChan <- nmeaData
		r.updateLastDataTime(server.CallSign)
	}

	return nil
}

func (r *TelnetController) updateLastDataTime(callSign string) {
	r.lastDataTime.Store(callSign, time.Now())
}

func (r *TelnetController) checkStaleConnections() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		r.lastDataTime.Range(func(key, value interface{}) bool {
			callSign := key.(string)
			lastTime := value.(time.Time)

			if time.Since(lastTime) > 30*time.Second {
				r.handleDisconnection(callSign)
			}
			return true
		})
	}
}
func (r *TelnetController) handleDisconnection(callSign string) {
	r.updateStatus(callSign, models.Disconnected)

	var lastRecord models.VesselRecord
	if err := database.DB.Where("call_sign = ?", callSign).Last(&lastRecord).Error; err == nil {
		lastRecord.TelnetStatus = models.Disconnected
		if err := database.DB.Save(&lastRecord).Error; err != nil {
			log.Printf("Error updating last record status for %s: %v", callSign, err)
		}
	}
}

func (r *TelnetController) convertGGAToNMEAData(gga *GGASentence) NMEAData {
	return NMEAData{
		Latitude:            gga.LatMinute,
		Longitude:           gga.LongMinute,
		GpsQualityIndicator: GpsQuality(gga.GPSQuality),
	}
}

func (r *TelnetController) convertHDTToNMEAData(hdt *HDTSentence) NMEAData {
	return NMEAData{
		HeadingDegree: hdt.Heading,
	}
}

func (r *TelnetController) convertVTGToNMEAData(vtg *VTGSentence) NMEAData {
	return NMEAData{
		SpeedInKnots: vtg.SpeedKnots,
	}
}

func (r *TelnetController) updateStatus(callSign string, status models.TelnetStatus) {
	kapal, ok := r.KapalDataMap.Load(callSign)
	if !ok {
		log.Printf("kapal not found in dataMap")
		return
	}
	historyPerSecond := kapal.(KapalData).Kapal.HistoryPerSecond
	disconnectThreshold := time.Duration(float64(historyPerSecond)*1.5) * time.Second

	if data, ok := r.DataMap.Load(callSign); ok {
		nmeaData := data.(NMEAData)
		nmeaData.TelnetStatus = status
		nmeaData.Status = string(status)
		r.DataMap.Store(callSign, nmeaData)
	}

	if status == models.Connected {
		r.updateLastDataTime(callSign)
	}

	var vesselRecord models.VesselRecord
	result := database.DB.Where("call_sign = ?", callSign).Last(&vesselRecord)
	if result.Error == nil {
		vesselRecord.TelnetStatus = status
		if time.Since(vesselRecord.CreatedAt) > disconnectThreshold && string(status) == "Disconnected" {
			if err := database.DB.Save(&vesselRecord).Error; err != nil {
				log.Printf("Error updating telnet status in database for %s: %v", callSign, err)
			}
		}
		if string(status) != "Disconnected" {
			if err := database.DB.Save(&vesselRecord).Error; err != nil {
				log.Printf("Error updating telnet status in database for %s: %v", callSign, err)
			}
		}
	} else {
		log.Printf("Error fetching vessel record for %s: %v", callSign, result.Error)
	}
}

func (r *TelnetController) updateDataMap(callSign string, nmeaDataChan <-chan NMEAData, waterDepthChan <-chan float64) {
	for {
		select {
		case nmeaData := <-nmeaDataChan:
			data, _ := r.DataMap.LoadOrStore(callSign, NMEAData{})
			storedData := data.(NMEAData)
			if nmeaData.Latitude != "" {
				storedData.Latitude = nmeaData.Latitude
			}
			if nmeaData.Longitude != "" {
				storedData.Longitude = nmeaData.Longitude
			}
			if nmeaData.HeadingDegree != 0 {
				storedData.HeadingDegree = nmeaData.HeadingDegree
			}
			if nmeaData.SpeedInKnots != 0 {
				storedData.SpeedInKnots = nmeaData.SpeedInKnots
			}
			if nmeaData.GpsQualityIndicator != "" {
				storedData.GpsQualityIndicator = nmeaData.GpsQualityIndicator
			}
			storedData.Status = nmeaData.Status
			r.DataMap.Store(callSign, storedData)
		case waterDepth := <-waterDepthChan:
			data, _ := r.DataMap.LoadOrStore(callSign, NMEAData{})
			storedData := data.(NMEAData)
			storedData.WaterDepth = waterDepth
			r.DataMap.Store(callSign, storedData)
		}
	}
}

func (r *TelnetController) countEntriesInDataMap() int {
	count := 0
	r.DataMap.Range(func(key, value interface{}) bool {
		count++
		return true
	})
	return count
}

func extractNumber(message string) (int, error) {
	re := regexp.MustCompile(`\d+`)
	matches := re.FindString(message)

	if matches == "" {
		return 0, fmt.Errorf("no number found in message")
	}

	var number int
	_, err := fmt.Sscanf(matches, "%d", &number)
	if err != nil {
		return 0, fmt.Errorf("error converting number: %v", err)
	}

	return number, nil
}

func (r *TelnetController) updateKapalDataMap() {
	var allKapal []models.Kapal
	err := database.DB.Find(&allKapal).Error
	if err != nil {
		log.Println("Error fetching kapal data:", err)
		return
	}

	currentCallSigns := make(map[string]bool)
	for _, kapal := range allKapal {
		currentCallSigns[kapal.CallSign] = true

		r.KapalDataMap.Store(kapal.CallSign, KapalData{
			Kapal: kapal,
		})
	}

	r.KapalDataMap.Range(func(key, value interface{}) bool {
		callSign := key.(string)
		if !currentCallSigns[callSign] {
			r.KapalDataMap.Delete(callSign)
			log.Printf("Removed KapalDataMap entry for deleted ship with call sign: %s", callSign)
		}
		return true
	})
}

func (r *TelnetController) createOrUpdateCoordinate(callSign string, lastCoordinateInsertTime *time.Time) error {
	var lastRecord models.VesselRecord
	result := database.DB.Where("call_sign = ?", callSign).Last(&lastRecord)

	data, _ := r.DataMap.Load(callSign)
	nmeaData := data.(NMEAData)

	kapal, ok := r.KapalDataMap.Load(callSign)
	if !ok {
		return fmt.Errorf("kapal not found in dataMap")
	}

	if result.Error != nil && !errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return fmt.Errorf("error fetching last coordinate: %w", result.Error)
	}
	if result.RowsAffected <= 0 {
		if nmeaData.Longitude != "" && nmeaData.Latitude != "" {
			gpsQuality := nmeaData.GpsQualityIndicator
			if err := database.DB.Create(&models.VesselRecord{
				CallSign:            callSign,
				SeriesID:            1,
				Latitude:            nmeaData.Latitude,
				Longitude:           nmeaData.Longitude,
				HeadingDegree:       nmeaData.HeadingDegree,
				SpeedInKnots:        nmeaData.SpeedInKnots,
				WaterDepth:          nmeaData.WaterDepth,
				GpsQualityIndicator: models.GpsQuality(gpsQuality),
				TelnetStatus:        models.Connected,
			}).Error; err != nil {
				return fmt.Errorf("error creating new coordinate: %w", err)
			}
		}
	} else if time.Since(lastRecord.CreatedAt) >= (time.Duration(kapal.(KapalData).Kapal.HistoryPerSecond))*time.Second {
		gpsQuality := nmeaData.GpsQualityIndicator
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
				TelnetStatus:        models.Connected,
			}).Error; err != nil {
				return fmt.Errorf("error creating new coordinate: %w", err)
			}
		}
	}

	return nil
}

func (r *TelnetController) KapalTelnetWebsocketHandler(c *gin.Context) {
	upgrader := websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println("Error upgrading to WebSocket:", err)
		return
	}
	defer conn.Close()

	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for range ticker.C {
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

func (r *TelnetController) GetKapalDataByCallSign(callSign string) (*KapalData, error) {
	data, ok := r.KapalDataMap.Load(callSign)
	if !ok {
		return nil, fmt.Errorf("data not found for call sign: %s", callSign)
	}
	kapalData := data.(KapalData)
	return &kapalData, nil
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

func convertDMSToDecimal(dms float64, direction string) (float64, error) {
	degrees := float64(int(dms / 100))
	minutes := dms - (degrees * 100)
	decimal := degrees + (minutes / 60)

	if direction == "S" || direction == "W" {
		decimal *= -1
	}

	return decimal, nil
}

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

var gpsQualityDescriptions = []string{
	"Fix not valid",
	"GPS fix",
	"Differential GPS fix",
	"Not applicable",
	"RTK Fixed",
	"RTK Float",
	"INS Dead reckoning",
}

func calculateChecksum(sentence string) string {
	checksum := 0
	for i := 1; i < len(sentence); i++ {
		checksum ^= int(sentence[i])
	}
	return fmt.Sprintf("%02X", checksum)
}

func min(a, b time.Duration) time.Duration {
	if a < b {
		return a
	}
	return b
}
