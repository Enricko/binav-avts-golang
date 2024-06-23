package controllers

import (
	"bufio"
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
	Status string `json:"status"` // Added status field
}

type KapalData struct {
	Kapal models.Kapal `json:"kapal"`
	NMEA  NMEAData     `json:"nmea"`
}

func handleTelnetConnection(server models.IPKapal, dataMap *sync.Map, connMap *sync.Map, wg *sync.WaitGroup, stopChan chan struct{}) {
	defer wg.Done()

	connMap.Store(server.IdIpKapal, stopChan)
	retryDelay := 5 * time.Second

	// Channel to control when to stop telnet connection
	stopTelnet := make(chan struct{})
	defer close(stopTelnet)
	data, _ := dataMap.LoadOrStore(server.CallSign, NMEAData{})
	nmeaData := data.(NMEAData)

	// Start a goroutine to handle telnet connection
	go func() {
		defer func() {
			connMap.Delete(server.IdIpKapal)
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
					dataMap.Store(server.CallSign, nmeaData)

					time.Sleep(retryDelay)
					retryDelay = min(2*retryDelay, 5*time.Minute) // Exponential backoff with a max delay of 5 minutes
					continue
				}

				retryDelay = 5 * time.Second // Reset retry delay on successful connection
				reader := bufio.NewReader(conn)

				connClosed := make(chan struct{})
				go func() {
					defer func() {
						conn.Close()
						close(connClosed)
					}()

					nmeaData.Status = "Disconnected"
					dataMap.Store(server.CallSign, nmeaData)
					for {
						select {
						case <-stopTelnet:
							return
						default:
							line, err := reader.ReadString('\n')
							line = strings.TrimSpace(line)

							if err != nil {
								log.Printf("Error reading from %s: %v", server.CallSign, err)
								nmeaData.Status = "Disconnected"
								return
							}

							if strings.HasPrefix(line, "$GPGGA") || strings.HasPrefix(line, "$GNGGA") {
								nmeaData.GGA = line
							} else if strings.HasPrefix(line, "$GPHDT") || strings.HasPrefix(line, "$GNHDT") {
								nmeaData.HDT = line
							} else if strings.HasPrefix(line, "$GPVTG") || strings.HasPrefix(line, "$GNVTG") {
								nmeaData.VTG = line
							}
							nmeaData.Status = "Connected"
							// fmt.Println(line)
							dataMap.Store(server.CallSign, nmeaData)
							// fmt.Println(server.IP, server.Port, line)
						}
					}
				}()

				<-connClosed
				log.Printf("Disconnected from %s. Reconnecting...", server.CallSign)
			}
		}
	}()

	// Handle SQL operations with the sqlTicker
	sqlTicker := time.NewTicker(5 * time.Second)
	defer sqlTicker.Stop()

	for {
		select {
		case <-stopChan:
			// Stop telnet connection and exit
			close(stopTelnet)
			return
		case <-sqlTicker.C:
			fmt.Println(nmeaData)
			// Execute SQL operations every 30 seconds
			var ipKapal models.IPKapal
			if err := database.DB.Where("call_sign = ?", server.CallSign).First(&ipKapal).Error; err != nil {
				log.Printf("Error reading from %s: %v", server.CallSign, err)
				return
			}
			if ipKapal.IP != server.IP || ipKapal.Port != server.Port {
				log.Printf("Error reading from %s: %v", server.CallSign)
				return
			}
		}
	}
}

func (r *TelnetController) StartTelnetConnections() {
	var wg sync.WaitGroup

	// Fetch initial Telnet servers from the database
	var servers []models.IPKapal
	err := database.DB.Find(&servers).Error
	if err != nil {
		log.Println("Error fetching Telnet servers:", err)
		return
	}

	for _, server := range servers {
		wg.Add(1)
		stopChan := make(chan struct{})
		go handleTelnetConnection(server, r.DataMap, r.ConnMap, &wg, stopChan)
	}

	// Start a goroutine to update KapalDataMap every 5 seconds
	go func() {
		for {
			r.updateKapalDataMap()
			time.Sleep(r.UpdateInterval)
		}
	}()

	// Goroutine to monitor database changes
	go func() {
		for {
			time.Sleep(30 * time.Second) // Adjust the interval as needed
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
					go handleTelnetConnection(updatedServer, r.DataMap, r.ConnMap, &wg, stopChan)
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

	// Wait for all goroutines to finish
	wg.Wait()
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
			nmeaData = NMEAData{
				// GGA: "No Data",
				// GGA: "$GPGGA,120000.00,0116.367,S,11649.483,E,1,08,0.9,10.0,M,-34.0,M,,*47",
				// HDT: "No Data",
				// HDT: "$GPHDT,90.0,T*0C",
				// VTG:    "No Data",
				Status: "Disconnected",
			}
		}

		r.KapalDataMap.Store(kapal.CallSign, KapalData{
			Kapal: kapal,
			NMEA:  nmeaData.(NMEAData),
		})
	}
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

func storeGGA(data, callSign string) {
	coordinateGga, err := parseGGA(data)
	if err != nil {
		log.Printf("Error parsing GGA data: %v", err)
		return
	}
	coordinateGga.CallSign = callSign

	err = database.DB.Save(&coordinateGga).Error
	if err != nil {
		log.Printf("Error storing GGA data: %v", err)
	}
}

func storeHDT(data, callSign string) {
	coordinateHdt, err := parseHDT(data)
	if err != nil {
		log.Printf("Error parsing HDT data: %v", err)
		return
	}
	coordinateHdt.CallSign = callSign

	err = database.DB.Save(&coordinateHdt).Error
	if err != nil {
		log.Printf("Error storing HDT data: %v", err)
	}
}

func storeVTG(data, callSign string) {
	coordinateVtg, err := parseVTG(data)
	if err != nil {
		log.Printf("Error parsing VTG data: %v", err)
		return
	}
	coordinateVtg.CallSign = callSign

	err = database.DB.Save(&coordinateVtg).Error
	if err != nil {
		log.Printf("Error storing VTG data: %v", err)
	}
}

// Function to parse GGA sentence
func parseGGA(sentence string) (models.CoordinateGga, error) {
	parts := strings.Split(sentence, ",")
	if len(parts) < 15 {
		return models.CoordinateGga{}, fmt.Errorf("invalid GGA sentence: %s", sentence)
	}

	utcPosition, err := strconv.ParseFloat(parts[1], 32)
	if err != nil {
		return models.CoordinateGga{}, err
	}

	latitude, err := strconv.ParseFloat(parts[2], 32)
	if err != nil {
		return models.CoordinateGga{}, err
	}

	longitude, err := strconv.ParseFloat(parts[4], 32)
	if err != nil {
		return models.CoordinateGga{}, err
	}

	numberSv, err := strconv.Atoi(parts[7])
	if err != nil {
		return models.CoordinateGga{}, err
	}

	hdop, err := strconv.ParseFloat(parts[8], 32)
	if err != nil {
		return models.CoordinateGga{}, err
	}

	orthometricHeight, err := strconv.ParseFloat(parts[9], 32)
	if err != nil {
		return models.CoordinateGga{}, err
	}

	geoidSeparation, err := strconv.ParseFloat(parts[11], 32)
	if err != nil {
		return models.CoordinateGga{}, err
	}

	coordinateGga := models.CoordinateGga{
		MessageID:           parts[0],
		UtcPosition:         float32(utcPosition),
		Latitude:            float32(latitude),
		DirectionLatitude:   parts[3],
		Longitude:           float32(longitude),
		DirectionLongitude:  parts[5],
		GpsQualityIndicator: models.GpsQuality(parts[6]),
		NumberSv:            numberSv,
		Hdop:                float32(hdop),
		OrthometricHeight:   float32(orthometricHeight),
		UnitMeasure:         parts[10],
		GeoidSeparation:     float32(geoidSeparation),
		GeoidMeasure:        parts[12],
		CreatedAt:           time.Now(),
		UpdatedAt:           time.Now(),
	}

	return coordinateGga, nil
}

// Function to parse HDT sentence
func parseHDT(sentence string) (models.CoordinateHdt, error) {
	parts := strings.Split(sentence, ",")
	if len(parts) < 3 {
		return models.CoordinateHdt{}, fmt.Errorf("invalid HDT sentence: %s", sentence)
	}

	headingDegree, err := strconv.ParseFloat(parts[1], 32)
	if err != nil {
		return models.CoordinateHdt{}, err
	}

	coordinateHdt := models.CoordinateHdt{
		MessageID:     parts[0],
		HeadingDegree: float32(headingDegree),
		Checksum:      parts[2],
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	return coordinateHdt, nil
}

// Function to parse VTG sentence
func parseVTG(sentence string) (models.CoordinateVtg, error) {
	parts := strings.Split(sentence, ",")
	if len(parts) < 9 {
		return models.CoordinateVtg{}, fmt.Errorf("invalid VTG sentence: %s", sentence)
	}

	var trackDegreeTrue, trackDegreeMagnetic, speedInKnots, kph float64
	var err error

	if parts[1] != "" {
		trackDegreeTrue, err = strconv.ParseFloat(parts[1], 32)
		if err != nil {
			return models.CoordinateVtg{}, err
		}
	}

	if parts[3] != "" {
		trackDegreeMagnetic, err = strconv.ParseFloat(parts[3], 32)
		if err != nil {
			return models.CoordinateVtg{}, err
		}
	}

	if parts[5] != "" {
		speedInKnots, err = strconv.ParseFloat(parts[5], 32)
		if err != nil {
			return models.CoordinateVtg{}, err
		}
	}

	if parts[7] != "" {
		kph, err = strconv.ParseFloat(parts[7], 32)
		if err != nil {
			return models.CoordinateVtg{}, err
		}
	}

	coordinateVtg := models.CoordinateVtg{
		MessageID:           parts[0],
		TrackDegreeTrue:     float32(trackDegreeTrue),
		TrueNorth:           parts[2],
		TrackDegreeMagnetic: float32(trackDegreeMagnetic),
		MagneticNorth:       parts[4],
		SpeedInKnots:        float32(speedInKnots),
		MeasuredKnots:       parts[6],
		Kph:                 float32(kph),
		MeasuredKph:         parts[8],
		Checksum:            parts[9], // Assuming the checksum is not used
		CreatedAt:           time.Now(),
		UpdatedAt:           time.Now(),
	}

	// Check for the existence of a checksum (indicated by an asterisk)
	if strings.Contains(parts[8], "*") {
		checksumParts := strings.Split(parts[8], "*")
		if len(checksumParts) == 2 {
			coordinateVtg.MeasuredKph = checksumParts[0]
			coordinateVtg.Checksum = checksumParts[1]
		} else {
			coordinateVtg.Checksum = checksumParts[0]
		}
	}

	return coordinateVtg, nil
}
