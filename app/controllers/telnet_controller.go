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

type TelnetController struct{}

func NewTelnetController() *TelnetController {
	return &TelnetController{}
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// NMEAData holds the parsed NMEA message types
type NMEAData struct {
	GGA string `json:"gga,omitempty"`
	HDT string `json:"hdt,omitempty"`
	VTG string `json:"vtg,omitempty"`
}

func handleTelnetConnection(server models.IPKapal, dataMap *sync.Map, connMap *sync.Map, wg *sync.WaitGroup, stopChan chan struct{}) {
	defer wg.Done()

	connMap.Store(server.IdIpKapal, stopChan)

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
					GGA: "Error connecting",
					HDT: "Error connecting",
					VTG: "Error connecting",
				})
				time.Sleep(5 * time.Second)
				continue
			}

			reader := bufio.NewReader(conn)
			for {
				select {
				case <-stopChan:
					conn.Close()
					log.Printf("Disconnected from %s", server.CallSign)
					return
				default:
					line, err := reader.ReadString('\n')
					if err != nil {
						log.Printf("Error reading from %s: %v", server.CallSign, err)
						break
					}
					line = strings.TrimSpace(line)
					data, _ := dataMap.LoadOrStore(server.CallSign, NMEAData{})
					nmeaData := data.(NMEAData)

					if strings.HasPrefix(line, "$GPGGA") || strings.HasPrefix(line, "$GNGGA") {
						nmeaData.GGA = line
					} else if strings.HasPrefix(line, "$GPHDT") || strings.HasPrefix(line, "$GNHDT") {
						nmeaData.HDT = line
					} else if strings.HasPrefix(line, "$GPVTG") || strings.HasPrefix(line, "$GNVTG") {
						nmeaData.VTG = line
					}

					dataMap.Store(server.CallSign, nmeaData)
				}
			}
			conn.Close()
			log.Printf("Disconnected from %s. Reconnecting...", server.CallSign)
			time.Sleep(5 * time.Second)
		}
	}
}

func (r *TelnetController) KapalTelnetWebsocketHandler(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		http.NotFound(c.Writer, c.Request)
		return
	}
	defer conn.Close()

	var wg sync.WaitGroup
	dataMap := &sync.Map{}
	connMap := &sync.Map{}

	// Fetch initial Telnet servers from the database
	var servers []models.IPKapal
	err = database.DB.Find(&servers).Error
	if err != nil {
		log.Println("Error fetching Telnet servers:", err)
		return
	}

	for _, server := range servers {
		wg.Add(1)
		stopChan := make(chan struct{})
		go handleTelnetConnection(server, dataMap, connMap, &wg, stopChan)
	}

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
					if stopChan, ok := connMap.Load(id); ok {
						close(stopChan.(chan struct{}))
						connMap.Delete(id)
					}
					// Start a new connection
					wg.Add(1)
					stopChan := make(chan struct{})
					go handleTelnetConnection(updatedServer, dataMap, connMap, &wg, stopChan)
				}
			}

			// Check for deleted servers
			for id := range currentServerMap {
				if _, exists := updatedServerMap[id]; !exists {
					// Stop the connection
					if stopChan, ok := connMap.Load(id); ok {
						close(stopChan.(chan struct{}))
						connMap.Delete(id)
					}
					dataMap.Delete(currentServerMap[id].CallSign)
				}
			}

			// Update the current server list
			servers = updatedServers
		}
	}()

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	done := make(chan struct{})

	go func() {
		wg.Wait()
		close(done)
	}()

	for {
		select {
		case <-done:
			return
		case <-ticker.C:
			data := make(map[string]NMEAData)
			dataMap.Range(func(key, value interface{}) bool {
				data[key.(string)] = value.(NMEAData)
				return true
			})
			if err := conn.WriteJSON(data); err != nil {
				log.Println("Error writing to WebSocket client:", err)
				return
			}
		}
	}
}
