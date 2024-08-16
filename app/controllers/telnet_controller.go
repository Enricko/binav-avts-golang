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

	Status       string              `json:"status"`
	TelnetStatus models.TelnetStatus `json:"telnet_status"`
}

type KapalData struct {
	Kapal models.Kapal `json:"kapal"`
	NMEA  NMEAData     `json:"nmea"`
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
		// wg.Add(1)
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
			time.Sleep(3 * time.Second)

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

			// Handle updated or new servers
			for id, updatedServer := range updatedServerMap {
				currentServer, exists := currentServerMap[id]
				if !exists || r.hasServerChanged(currentServer, updatedServer) {
					log.Printf("Updating connection for server ID: %d, IP: %s, Port: %s", id, updatedServer.IP, updatedServer.Port)
					r.stopAndRemoveConnection(id) // Properly stop the existing connection

					wg.Add(1)
					stopChan := make(chan struct{})
					go r.handleTelnetConnection(updatedServer, &wg, stopChan)
				}
			}

			// Handle removed servers
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

	stopTelnet := make(chan struct{})
	defer close(stopTelnet)

	nmeaDataChan := make(chan NMEAData)
	waterDepthChan := make(chan float64)
	go r.updateDataMap(server.CallSign, nmeaDataChan, waterDepthChan)

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
				if err := r.connectAndRead(server, nmeaDataChan, waterDepthChan); err != nil {
					// log.Printf("Error handling connection for %s: %v", server.CallSign, err)
					r.updateStatus(server.CallSign, models.Disconnected)
					// Immediately try to reconnect
				}
			}
		}
	}()

	for {
		select {
		case <-stopChan:
			r.DataMap.Delete(server.CallSign)
			log.Printf("asdasdTelnet connection stopped for server ID: %d, CallSign: %s, Port:%a", server.IdIpKapal, server.CallSign, server.Port)
			close(stopTelnet)
			return
		}
	}
}

func (r *TelnetController) connectAndRead(server models.IPKapal, nmeaDataChan chan<- NMEAData, waterDepthChan chan<- float64) error {
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", server.IP, server.Port))
	if err != nil {
		return fmt.Errorf("error connecting: %w", err)
	}
	defer conn.Close()

	reader := bufio.NewReader(conn)

	lastCoordinateInsertTime := time.Now()

	for {

		if err := conn.SetReadDeadline(time.Now().Add(10 * time.Second)); err != nil {
			return fmt.Errorf("error setting read deadline: %w", err)
		}

		line, err := r.readLine(reader, server.TypeIP)
		if err != nil {
			return fmt.Errorf("error reading: %w", err)
		}

		sentences := strings.Split(strings.TrimSpace(line), "\n")

		// Process each sentence
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

		// Convert to string
		line = string(data[:n])

		// Print the data
		// fmt.Println(nmeaData)
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
	}

	return nil
}

func (r *TelnetController) convertGGAToNMEAData(gga *GGASentence) NMEAData {
	return NMEAData{
		Latitude:            gga.LatMinute,
		Longitude:           gga.LongMinute,
		GpsQualityIndicator: GpsQuality(gga.GPSQuality),
		// Other fields remain default values
	}
}

func (r *TelnetController) convertHDTToNMEAData(hdt *HDTSentence) NMEAData {
	return NMEAData{
		HeadingDegree: hdt.Heading,
		// Other fields remain default values
	}
}

func (r *TelnetController) convertVTGToNMEAData(vtg *VTGSentence) NMEAData {
	return NMEAData{
		SpeedInKnots: vtg.SpeedKnots,
		// Other fields remain default values
	}
}

func (r *TelnetController) updateStatus(callSign string, status models.TelnetStatus) {
	// Update in-memory data
	if data, ok := r.DataMap.Load(callSign); ok {
		nmeaData := data.(NMEAData)
		nmeaData.TelnetStatus = status
		nmeaData.Status = string(status) // For backwards compatibility
		r.DataMap.Store(callSign, nmeaData)
	}

	// Update in database
	var vesselRecord models.VesselRecord
	result := database.DB.Where("call_sign = ?", callSign).Order("created_at DESC").First(&vesselRecord)
	if result.Error == nil {
		vesselRecord.TelnetStatus = status
		if err := database.DB.Save(&vesselRecord).Error; err != nil {
			log.Printf("Error updating telnet status in database for %s: %v", callSign, err)
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
		return true // continue iteration
	})
	return count
}

func extractNumber(message string) (int, error) {
	// Define a regular expression to match any number in the message
	re := regexp.MustCompile(`\d+`)
	matches := re.FindString(message)

	if matches == "" {
		return 0, fmt.Errorf("no number found in message")
	}

	// Return the extracted number
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

	// Create a map of all current call signs in the database
	currentCallSigns := make(map[string]bool)
	for _, kapal := range allKapal {
		currentCallSigns[kapal.CallSign] = true

		// Update or add entry in KapalDataMap
		r.KapalDataMap.Store(kapal.CallSign, KapalData{
			Kapal: kapal,
		})
	}

	// Remove entries from KapalDataMap that no longer exist in the database
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
	result := database.DB.Where("call_sign = ?", callSign).Order("created_at desc").First(&lastRecord)

	data, _ := r.DataMap.Load(callSign)
	nmeaData := data.(NMEAData)

	kapal, ok := r.KapalDataMap.Load(callSign)
	if !ok {
		return fmt.Errorf("kapal not found in dataMap")
	}

	if result.Error != nil && !errors.Is(result.Error, gorm.ErrRecordNotFound) {
		fmt.Errorf("error fetching last coordinate: %w", result.Error)
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
		// if err != nil {
		// 	return fmt.Errorf("error creating new coordinate: %w", err)
		// }

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
