package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/gorilla/websocket"
	amqp "github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	pb_data "github.com/skni-kod/iot-monitor-backend/internal/proto/data_service"
	pb_sensor "github.com/skni-kod/iot-monitor-backend/internal/proto/sensor_service"
	"github.com/skni-kod/iot-monitor-backend/internal/types"
	"github.com/skni-kod/iot-monitor-backend/pkg/logger"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type WebSocketHandler struct {
	dataClient   pb_data.DataServiceClient
	sensorClient pb_sensor.SensorServiceClient
	clients      map[*websocket.Conn]bool
	clientsMu    sync.RWMutex
}

func NewWebSocketHandler(dataClient pb_data.DataServiceClient, sensorClient pb_sensor.SensorServiceClient, alertMsgs <-chan amqp.Delivery) *WebSocketHandler {
	h := &WebSocketHandler{
		dataClient:   dataClient,
		sensorClient: sensorClient,
		clients:      make(map[*websocket.Conn]bool),
	}

	if alertMsgs != nil {
		go h.broadcastAlerts(alertMsgs)
	}

	return h
}

// @Summary Stream sensor readings via WebSocket
// @Description Establishes a WebSocket connection for real-time sensor data streaming
// @Tags Data
// @Param sensor_ids query string false "Comma-separated sensor IDs"
// @Router /api/data/ws/readings [get]
func (h *WebSocketHandler) HandleReadings(w http.ResponseWriter, r *http.Request) {
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

	if len(sensorIDs) == 0 {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		for userId := int64(1); userId <= 10; userId++ {
			sensors, err := h.sensorClient.ListSensors(ctx, &pb_sensor.ListSensorsRequest{UserId: userId})
			if err == nil && sensors.Sensors != nil {
				for _, sensor := range sensors.Sensors {
					if sensor.Active {
						sensorIDs = append(sensorIDs, sensor.Id)
					}
				}
			}
		}
	}

	logger.Info("WebSocket connection request for sensors", zap.Int64s("sensor_ids", sensorIDs))

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logger.Error("Failed to upgrade to WebSocket", zap.Error(err))
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
		logger.Info("WebSocket client disconnected")
	}()

	logger.Info("WebSocket client connected for sensors", zap.Int64s("sensors_ids", sensorIDs))

	if len(sensorIDs) > 0 {
		go h.streamToClient(conn, sensorIDs)
	} else {
		logger.Info("No active sensors found to stream")
	}

	for {
		var msg types.SubscribeMessage
		err := conn.ReadJSON(&msg)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				logger.Error("WebSocket error", zap.Error(err))
			}
			break
		}

		if msg.Type == "subscribe" && len(msg.SensorIDs) > 0 {
			logger.Info("Client subscribing to sensors", zap.Int64s("sensor_ids", msg.SensorIDs))
			go h.streamToClient(conn, msg.SensorIDs)
		}
	}
}

func (h *WebSocketHandler) streamToClient(conn *websocket.Conn, sensorIDs []int64) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	logger.Info("Starting stream for sensors", zap.Int64s("sensor_ids", sensorIDs))

	stream, err := h.dataClient.StreamReadings(ctx, &pb_data.StreamReadingsRequest{
		SensorIds: sensorIDs,
	})
	if err != nil {
		logger.Error("Failed to start stream", zap.Error(err))
		return
	}

	logger.Info("Stream established successfully")

	for {
		update, err := stream.Recv()
		if err != nil {
			if st, ok := status.FromError(err); ok {
				if st.Code() == codes.Canceled {
					logger.Info("Stream cancelled")
					return
				}
			}
			logger.Error("Stream receive error", zap.Error(err))
			return
		}

		logger.Info("Received update",
			zap.Int64("sensor_id", update.SensorId),
			zap.Float32("value", update.Value),
		)

		msg := types.ReadingMessage{
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
			logger.Info("Client no longer exists")
			return
		}

		if err := conn.WriteJSON(msg); err != nil {
			logger.Error("Failed to write to WebSocket", zap.Error(err))
			return
		}

		logger.Info("Sent update to client",
			zap.Int64("sensor_id", update.SensorId),
			zap.Float32("value", update.Value),
		)
	}
}

// @Summary Get historical sensor readings
// @Description Fetches historical data for a specific sensor
// @Tags Data
// @Param sensor_id path int true "Sensor ID"
// @Param start_time query string false "Start time (RFC3339)"
// @Param end_time query string false "End time (RFC3339)"
// @Success 200 {object} []string
// @Router /api/data/sensors/{sensor_id}/readings [get]
func (h *WebSocketHandler) GetHistoricalReadings(w http.ResponseWriter, r *http.Request) {
	sensorIDStr := r.URL.Query().Get("sensor_id")
	sensorID, err := strconv.ParseInt(sensorIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid sensor_id", http.StatusBadRequest)
		return
	}

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
		startTime = time.Now().Add(-24 * time.Hour)
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

	res, err := h.dataClient.GetLatestReadingsBatch(ctx, &pb_data.LatestReadingsBatchRequest{
		SensorIds: sensorIDs,
	})
	if err != nil {
		http.Error(w, "Failed to get latest readings: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}

// REPLACE GetHistoricalReadings with this for getting last N readings of ONE sensor
// @Summary Get latest N readings for a single sensor
// @Description Fetches the most recent N readings for a specific sensor
// @Tags Data
// @Param sensor_id path int true "Sensor ID"
// @Param limit query int false "Number of readings to fetch (default 10)"
// @Success 200 {object} string
// @Router /api/data/sensors/{sensor_id}/latest [get]
func (h *WebSocketHandler) GetSensorLatestReadings(w http.ResponseWriter, r *http.Request) {
	sensorIDStr := chi.URLParam(r, "sensor_id")
	sensorID, err := strconv.ParseInt(sensorIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid sensor_id", http.StatusBadRequest)
		return
	}

	limitStr := r.URL.Query().Get("limit")
	limit := int64(10)
	if limitStr != "" {
		limit, err = strconv.ParseInt(limitStr, 10, 64)
		if err != nil || limit <= 0 {
			http.Error(w, "Invalid limit parameter", http.StatusBadRequest)
			return
		}
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	res, err := h.dataClient.GetLatestReadingsBySensor(ctx, &pb_data.LatestReadingsBySensorRequest{
		SensorId: sensorID,
		Limit:    limit,
	})
	if err != nil {
		http.Error(w, "Failed to get sensor readings: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}

func (h *WebSocketHandler) WsHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("Error upgrading:", err)
		return
	}
	defer conn.Close()
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			fmt.Println("Error reading message:", err)
			break
		}
		fmt.Printf("Received: %s\n", message)
		if err := conn.WriteMessage(websocket.TextMessage, message); err != nil {
			fmt.Println("Error writing message:", err)
			break
		}
	}
}

// @Summary Store a new sensor reading
// @Description Sends a sensor reading to the data processing service
// @Tags Data
// @Accept json
// @Produce json
// @Param reading body StoreReadingRequest true "Sensor Reading"
// @Success 200 {object} map[string]string
// @Router /api/data/readings [post]
func (h *WebSocketHandler) StoreReading(w http.ResponseWriter, r *http.Request) {
	var req types.StoreReadingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Domyślny timestamp, jeśli nie podano
	if req.Timestamp.IsZero() {
		req.Timestamp = time.Now()
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	// Wywołanie usługi gRPC data-processing
	_, err := h.dataClient.StoreReading(ctx, &pb_data.StoreReadingRequest{
		SensorId:  req.SensorID,
		Value:     req.Value,
		Timestamp: timestamppb.New(req.Timestamp),
	})

	if err != nil {
		logger.Error("Failed to store reading via gRPC", zap.Error(err))
		http.Error(w, "Failed to store reading", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

func (h *WebSocketHandler) broadcastAlerts(alertMsgs <-chan amqp.Delivery) {
	for m := range alertMsgs {
		var alert map[string]interface{}
		if err := json.Unmarshal(m.Body, &alert); err != nil {
			logger.Error("Failed to unmarshal alert message", zap.Error(err))
			continue
		}

		wsMsg := types.AlertMessage{
			Type:    "alert",
			Payload: alert,
		}

		h.clientsMu.RLock()
		for client := range h.clients {
			if err := client.WriteJSON(wsMsg); err != nil {
				logger.Error("Failed to write alert to WebSocket", zap.Error(err))
				client.Close()
				h.clientsMu.RUnlock()
				h.clientsMu.Lock()
				delete(h.clients, client)
				h.clientsMu.Unlock()
				h.clientsMu.RLock()
			}
		}
		h.clientsMu.RUnlock()
	}
}
