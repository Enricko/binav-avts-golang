package controllers

import (
	"bufio"
	"fmt"
	"golang-app/app/models"
	"golang-app/database"
	"log"
	"net"
	"net/http"
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
	GGA string `json:"gga,omitempty"`
	HDT string `json:"hdt,omitempty"`
	VTG string `json:"vtg,omitempty"`
}

type KapalData struct {
	Kapal models.Kapal `json:"kapal"`
	NMEA  NMEAData     `json:"nmea"`
}

func handleTelnetConnection(server models.IPKapal, dataMap *sync.Map, connMap *sync.Map, wg *sync.WaitGroup, stopChan chan struct{}) {
	defer wg.Done()

	connMap.Store(server.IdIpKapal, stopChan)
	retryDelay := 5 * time.Second

	for {
		select {
		case <-stopChan:
			log.Printf("Stopping connection to %s", server.CallSign)
			return
		default:
			conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", server.IP, server.Port))
			if err != nil {
				log.Printf("Error connecting to %s: %v", server.CallSign, err)
				dataMap.Store(server.CallSign, NMEAData{
					// GGA: "Error connecting",
					GGA: "$GPGGA,120000.00,0116.367,S,11649.483,E,1,08,0.9,10.0,M,-34.0,M,,*47",
					HDT: "$GPHDT,90.0,T*0C",
					VTG: "Error connecting",
				})
				time.Sleep(retryDelay)
				retryDelay = min(2*retryDelay, 5*time.Minute) // Exponential backoff with a max delay of 5 minutes
				continue
			}

			retryDelay = 5 * time.Second // Reset retry delay on successful connection
			reader := bufio.NewReader(conn)

			// Start a goroutine to read from the connection
			go func() {
				defer conn.Close()
				for {
					line, err := reader.ReadString('\n')
					if err != nil {
						log.Printf("Error reading from %s: %v", server.CallSign, err)
						break
					}
					line = strings.TrimSpace(line)
					data, _ := dataMap.LoadOrStore(server.CallSign, NMEAData{})
					nmeaData := data.(NMEAData)

					if strings.HasPrefix(line, "$GPGGA") || strings.HasPrefix(line, "$GNGGA") {
						// nmeaData.GGA = line
						nmeaData.GGA = "$GPGGA,120000.00,0116.367,S,11649.483,E,1,08,0.9,10.0,M,-34.0,M,,*47"
					} else if strings.HasPrefix(line, "$GPHDT") || strings.HasPrefix(line, "$GNHDT") {
						// nmeaData.HDT = line
						nmeaData.HDT = "$GPHDT,90.0,T*0C"
					} else if strings.HasPrefix(line, "$GPVTG") || strings.HasPrefix(line, "$GNVTG") {
						nmeaData.VTG = line
					}
					dataMap.Store(server.CallSign, nmeaData)
				}
			}()

			<-stopChan // Wait for stop signal
			log.Printf("Disconnected from %s. Reconnecting...", server.CallSign)
			return
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

			// Map current servers by ID for quick lookup
			currentServerMap := make(map[uint]models.IPKapal)
			for _, server := range servers {
				currentServerMap[server.IdIpKapal] = server
			}

			// Map updated servers by ID for quick lookup
			updatedServerMap := make(map[uint]models.IPKapal)
			for _, server := range updatedServers {
				updatedServerMap[server.IdIpKapal] = server
			}

			// Check for new or updated servers
			for id, updatedServer := range updatedServerMap {
				if currentServer, exists := currentServerMap[id]; !exists || currentServer != updatedServer {
					// Stop the old connection if it exists
					if stopChan, ok := r.ConnMap.Load(id); ok {
						close(stopChan.(chan struct{}))
						r.ConnMap.Delete(id)
					}
					// Start a new connection
					wg.Add(1)
					stopChan := make(chan struct{})
					go handleTelnetConnection(updatedServer, r.DataMap, r.ConnMap, &wg, stopChan)
				}
			}

			// Check for deleted servers
			for id := range currentServerMap {
				if _, exists := updatedServerMap[id]; !exists {
					// Stop the connection
					if stopChan, ok := r.ConnMap.Load(id); ok {
						close(stopChan.(chan struct{}))
						r.ConnMap.Delete(id)
					}
					r.DataMap.Delete(currentServerMap[id].CallSign)
				}
			}

			// Update the current server list
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
				// GGA: "No data",
				GGA: "$GPGGA,120000.00,0116.367,S,11649.483,E,1,08,0.9,10.0,M,-34.0,M,,*47",
				// HDT: "No data",
				HDT: "$GPHDT,90.0,T*0C",
				VTG: "No data",
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
				data[key.(string)] = value.(KapalData)
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
