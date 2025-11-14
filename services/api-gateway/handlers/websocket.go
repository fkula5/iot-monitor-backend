package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	pb_data "github.com/skni-kod/iot-monitor-backend/internal/proto/data_service"
	authMiddleware "github.com/skni-kod/iot-monitor-backend/services/api-gateway/middleware"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Configure this properly in production
	},
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type WebSocketHandler struct {
	dataClient pb_data.DataServiceClient
	clients    map[*websocket.Conn]bool
	clientsMu  sync.RWMutex
}

func NewWebSocketHandler(dataClient pb_data.DataServiceClient) *WebSocketHandler {
	return &WebSocketHandler{
		dataClient: dataClient,
		clients:    make(map[*websocket.Conn]bool),
	}
}

type ReadingMessage struct {
	SensorID   int64     `json:"sensor_id"`
	Value      float32   `json:"value"`
	Timestamp  time.Time `json:"timestamp"`
	SensorName string    `json:"sensor_name"`
	Location   string    `json:"location"`
	Unit       string    `json:"unit"`
}

type SubscribeMessage struct {
	Type      string  `json:"type"` // "subscribe" or "unsubscribe"
	SensorIDs []int64 `json:"sensor_ids"`
}

// @Summary Stream sensor readings via WebSocket
// @Description Establishes a WebSocket connection for real-time sensor data streaming
// @Tags Data
// @Security ApiKeyAuth
// @Param sensor_ids query string false "Comma-separated sensor IDs"
// @Router /api/data/ws/readings [get]
func (h *WebSocketHandler) HandleReadings(w http.ResponseWriter, r *http.Request) {
	// Authenticate the user
	claims, ok := authMiddleware.GetUserFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get sensor IDs from query parameter
	sensorIDsParam := r.URL.Query().Get("sensor_ids")
	var sensorIDs []int64
	if sensorIDsParam != "" {
		for _, idStr := range strings.Split(sensorIDsParam, ",") {
			id, err := strconv.ParseInt(strings.TrimSpace(idStr), 10, 64)
			if err == nil {
				sensorIDs = append(sensorIDs, id)
			}
		}
	}

	// Upgrade to WebSocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Failed to upgrade to WebSocket: %v", err)
		return
	}

	h.clientsMu.Lock()
	h.clients[conn] = true
	h.clientsMu.Unlock()

	defer func() {
		h.clientsMu.Lock()
		delete(h.clients, conn)
		h.clientsMu.Unlock()
		conn.Close()
	}()

	log.Printf("WebSocket client connected for user %d with sensors: %v", claims.UserId, sensorIDs)

	// Start streaming
	if len(sensorIDs) > 0 {
		go h.streamToClient(conn, sensorIDs)
	}

	// Handle incoming messages (for dynamic subscription changes)
	for {
		var msg SubscribeMessage
		err := conn.ReadJSON(&msg)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		if msg.Type == "subscribe" && len(msg.SensorIDs) > 0 {
			go h.streamToClient(conn, msg.SensorIDs)
		}
	}
}

func (h *WebSocketHandler) streamToClient(conn *websocket.Conn, sensorIDs []int64) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	stream, err := h.dataClient.StreamReadings(ctx, &pb_data.StreamReadingsRequest{
		SensorIds: sensorIDs,
	})
	if err != nil {
		log.Printf("Failed to start stream: %v", err)
		return
	}

	for {
		update, err := stream.Recv()
		if err != nil {
			if st, ok := status.FromError(err); ok {
				if st.Code() == codes.Canceled {
					return
				}
			}
			log.Printf("Stream receive error: %v", err)
			return
		}

		msg := ReadingMessage{
			SensorID:   update.SensorId,
			Value:      update.Value,
			Timestamp:  update.Timestamp.AsTime(),
			SensorName: update.SensorName,
			Location:   update.Location,
			Unit:       update.Unit,
		}

		h.clientsMu.RLock()
		clientExists := h.clients[conn]
		h.clientsMu.RUnlock()

		if !clientExists {
			return
		}

		if err := conn.WriteJSON(msg); err != nil {
			log.Printf("Failed to write to WebSocket: %v", err)
			return
		}
	}
}

// @Summary Get historical sensor readings
// @Description Fetches historical data for a specific sensor
// @Tags Data
// @Security ApiKeyAuth
// @Param sensor_id path int true "Sensor ID"
// @Param start_time query string false "Start time (RFC3339)"
// @Param end_time query string false "End time (RFC3339)"
// @Success 200 {object} []string
// @Router /api/data/sensors/{sensor_id}/readings [get]
func (h *WebSocketHandler) GetHistoricalReadings(w http.ResponseWriter, r *http.Request) {
	// Extract sensor_id from URL
	sensorIDStr := r.URL.Query().Get("sensor_id")
	sensorID, err := strconv.ParseInt(sensorIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid sensor_id", http.StatusBadRequest)
		return
	}

	// Parse time range
	startTimeStr := r.URL.Query().Get("start_time")
	endTimeStr := r.URL.Query().Get("end_time")

	var startTime, endTime time.Time
	if startTimeStr != "" {
		startTime, err = time.Parse(time.RFC3339, startTimeStr)
		if err != nil {
			http.Error(w, "Invalid start_time format", http.StatusBadRequest)
			return
		}
	} else {
		startTime = time.Now().Add(-24 * time.Hour) // Default: last 24 hours
	}

	if endTimeStr != "" {
		endTime, err = time.Parse(time.RFC3339, endTimeStr)
		if err != nil {
			http.Error(w, "Invalid end_time format", http.StatusBadRequest)
			return
		}
	} else {
		endTime = time.Now()
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	res, err := h.dataClient.QueryReadings(ctx, &pb_data.QueryReadingsRequest{
		SensorId:  sensorID,
		StartTime: timestamppb.New(startTime),
		EndTime:   timestamppb.New(endTime),
	})
	if err != nil {
		http.Error(w, "Failed to query readings: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res.DataPoints)
}

// @Summary Get latest readings for multiple sensors
// @Description Fetches the most recent reading for each specified sensor
// @Tags Data
// @Security ApiKeyAuth
// @Param sensor_ids query string true "Comma-separated sensor IDs"
// @Success 200 {object} string
// @Router /api/data/readings/latest [get]
func (h *WebSocketHandler) GetLatestReadings(w http.ResponseWriter, r *http.Request) {
	sensorIDsParam := r.URL.Query().Get("sensor_ids")
	if sensorIDsParam == "" {
		http.Error(w, "sensor_ids parameter is required", http.StatusBadRequest)
		return
	}

	var sensorIDs []int64
	for _, idStr := range strings.Split(sensorIDsParam, ",") {
		id, err := strconv.ParseInt(strings.TrimSpace(idStr), 10, 64)
		if err != nil {
			http.Error(w, "Invalid sensor_id: "+idStr, http.StatusBadRequest)
			return
		}
		sensorIDs = append(sensorIDs, id)
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	res, err := h.dataClient.GetLatestReadings(ctx, &pb_data.LatestReadingsRequest{
		SensorIds: sensorIDs,
	})
	if err != nil {
		http.Error(w, "Failed to get latest readings: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}

func (h *WebSocketHandler) wsHandler(w http.ResponseWriter, r *http.Request) {
	// Upgrade the HTTP connection to a WebSocket connection
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("Error upgrading:", err)
		return
	}
	defer conn.Close()
	// Listen for incoming messages
	for {
		// Read message from the client
		_, message, err := conn.ReadMessage()
		if err != nil {
			fmt.Println("Error reading message:", err)
			break
		}
		fmt.Printf("Received: %s\\n", message)
		// Echo the message back to the client
		if err := conn.WriteMessage(websocket.TextMessage, message); err != nil {
			fmt.Println("Error writing message:", err)
			break
		}
	}
}
