package controllers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sort"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"

	"golang-app/app/models"
	"golang-app/database"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type WebSocketController struct {
	TelnetController *TelnetController
}

type WebSocketMessage struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}

type SafeConnection struct {
	conn  *websocket.Conn
	mutex sync.Mutex
}

func (sc *SafeConnection) WriteJSON(v interface{}) error {
	sc.mutex.Lock()
	defer sc.mutex.Unlock()
	return sc.conn.WriteJSON(v)
}

func NewWebSocketController(telnetController *TelnetController) *WebSocketController {
	return &WebSocketController{
		TelnetController: telnetController,
	}
}

func (wsc *WebSocketController) HandleWebSocket(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println("Error upgrading to WebSocket:", err)
		return
	}
	defer conn.Close()

	safeConn := &SafeConnection{conn: conn}

	// Channels for communication
	done := make(chan struct{})
	realtimeUpdatesChan := make(chan WebSocketMessage, 100)
	vesselRecordsChan := make(chan WebSocketMessage, 100)

	defer close(done)

	// Start goroutines for handling different types of messages
	go wsc.sendRealtimeUpdates(safeConn, realtimeUpdatesChan, done)
	go wsc.processVesselRecords(safeConn, vesselRecordsChan, done)

	// Start a goroutine to handle writing messages to the WebSocket
	go wsc.writeMessages(safeConn, realtimeUpdatesChan, vesselRecordsChan, done)

	// Handle incoming messages
	for {
		_, p, err := conn.ReadMessage()
		if err != nil {
			log.Println("Error reading message:", err)
			return
		}

		var message WebSocketMessage
		if err := json.Unmarshal(p, &message); err != nil {
			log.Println("Error unmarshalling message:", err)
			continue
		}

		switch message.Type {
		case "vessel_records_request":
			go wsc.handleVesselRecordsRequest(vesselRecordsChan, message.Payload)
		default:
			log.Printf("Unknown message type: %s", message.Type)
		}
	}
}

func (wsc *WebSocketController) sendRealtimeUpdates(conn *SafeConnection, updateChan chan<- WebSocketMessage, done <-chan struct{}) {
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-done:
			return
		case <-ticker.C:
			data := make(map[string]KapalData)
			wsc.TelnetController.KapalDataMap.Range(func(key, value interface{}) bool {
				kapalData := value.(KapalData)
				nmeaData, ok := wsc.TelnetController.DataMap.Load(kapalData.Kapal.CallSign)
				if ok {
					kapalData.NMEA = nmeaData.(NMEAData)
				}
				data[key.(string)] = kapalData
				return true
			})

			updateChan <- WebSocketMessage{
				Type:    "realtime_update",
				Payload: data,
			}
		}
	}
}

func (wsc *WebSocketController) processVesselRecords(conn *SafeConnection, recordsChan <-chan WebSocketMessage, done <-chan struct{}) {
	for {
		select {
		case <-done:
			return
		case message := <-recordsChan:
			if err := conn.WriteJSON(message); err != nil {
				log.Println("Error writing vessel records message:", err)
			}
		}
	}
}

func (wsc *WebSocketController) writeMessages(conn *SafeConnection, realtimeUpdatesChan, vesselRecordsChan <-chan WebSocketMessage, done <-chan struct{}) {
	for {
		select {
		case <-done:
			return
		case update := <-realtimeUpdatesChan:
			if err := conn.WriteJSON(update); err != nil {
				log.Println("Error writing realtime update:", err)
			}
		case record := <-vesselRecordsChan:
			if err := conn.WriteJSON(record); err != nil {
				log.Println("Error writing vessel record:", err)
			}
		}
	}
}

func (wsc *WebSocketController) handleVesselRecordsRequest(recordsChan chan<- WebSocketMessage, payload interface{}) {
	var request struct {
		CallSign string `json:"call_sign"`
		Start    string `json:"start"`
		End      string `json:"end"`
	}

	payloadBytes, _ := json.Marshal(payload)
	if err := json.Unmarshal(payloadBytes, &request); err != nil {
		log.Println("Error unmarshalling vessel records request:", err)
		return
	}

	startTime := time.Now()

	var kapal models.Kapal
	if err := database.DB.Where("call_sign = ?", request.CallSign).First(&kapal).Error; err != nil {
		recordsChan <- WebSocketMessage{
			Type:    "error",
			Payload: map[string]string{"error": "Kapal not found"},
		}
		return
	}

	// Send initial message with kapal info
	recordsChan <- WebSocketMessage{
		Type: "vessel_records_start",
		Payload: map[string]interface{}{
			"call_sign": request.CallSign,
			"kapal":     kapal,
			"status":    "started",
		},
	}

	// Prepare the query
	query := database.DB.Where("call_sign = ? AND created_at BETWEEN ? AND ?", request.CallSign, request.Start, request.End)

	// Count total records
	var totalRecords int64
	if err := query.Model(&models.VesselRecord{}).Count(&totalRecords).Error; err != nil {
		recordsChan <- WebSocketMessage{
			Type:    "error",
			Payload: map[string]string{"error": err.Error()},
		}
		return
	}

	// Send total count
	recordsChan <- WebSocketMessage{
		Type:    "vessel_records_count",
		Payload: map[string]int64{"total_records": totalRecords},
	}

	const batchSize = 1000
	var offset int64 = 0

	// Create a worker pool
	workerCount := 5
	jobs := make(chan struct {
		batchNumber int
		records     []models.VesselRecord
	}, workerCount)
	results := make(chan WebSocketMessage, workerCount)
	var wg sync.WaitGroup

	// Start worker goroutines
	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for job := range jobs {
				// Sort the records by id_vessel_record
				sort.Slice(job.records, func(i, j int) bool {
					return job.records[i].IdVesselRecord < job.records[j].IdVesselRecord
				})

				results <- WebSocketMessage{
					Type: "vessel_records_batch",
					Payload: map[string]interface{}{
						"batch_number": job.batchNumber,
						"records":      job.records,
					},
				}
			}
		}()
	}

	// Start a goroutine to close the results channel when all workers are done
	go func() {
		wg.Wait()
		close(results)
	}()

	// Start a goroutine to send results to the recordsChan
	go func() {
		for result := range results {
			recordsChan <- result
			time.Sleep(100 * time.Millisecond) // Basic rate limiting
		}
	}()

	// Fetch and process records in batches
	batchNumber := 1
	for offset < totalRecords {
		var records []models.VesselRecord
		if err := query.Offset(int(offset)).Limit(batchSize).Find(&records).Error; err != nil {
			recordsChan <- WebSocketMessage{
				Type:    "error",
				Payload: map[string]string{"error": err.Error()},
			}
			close(jobs)
			return
		}
		jobs <- struct {
			batchNumber int
			records     []models.VesselRecord
		}{batchNumber, records}
		offset += int64(len(records))
		batchNumber++
	}

	close(jobs)

	processingTime := time.Since(startTime).Milliseconds()

	// Send final message
	recordsChan <- WebSocketMessage{
		Type: "vessel_records_complete",
		Payload: map[string]interface{}{
			"processing_time": processingTime,
			"status":          "completed",
		},
	}
}
