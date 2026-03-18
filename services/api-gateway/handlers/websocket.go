package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/gorilla/websocket"
	amqp "github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"
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
	}()

	if len(sensorIDs) > 0 {
		go h.streamToClient(conn, sensorIDs)
	}

	for {
		var msg types.SubscribeMessage
		if err := conn.ReadJSON(&msg); err != nil {
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
		return
	}

	for {
		update, err := stream.Recv()
		if err != nil {
			return
		}

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
			return
		}

		if err := conn.WriteJSON(msg); err != nil {
			return
		}
	}
}

// @Summary Get historical sensor readings
// @Description Fetches historical data for a specific sensor
// @Tags Data
// @Param sensor_id path int true "Sensor ID"
// @Param start_time query string false "Start time (RFC3339)"
// @Param end_time query string false "End time (RFC3339)"
// @Success 200 {object} Response{data=[]object}
// @Failure 400 {object} Response{error=string}
// @Failure 500 {object} Response{error=string}
// @Router /api/data/sensors/{sensor_id}/readings [get]
func (h *WebSocketHandler) GetHistoricalReadings(w http.ResponseWriter, r *http.Request) {
	sensorID, err := strconv.ParseInt(chi.URLParam(r, "sensor_id"), 10, 64)
	if err != nil {
		Error(w, http.StatusBadRequest, "Invalid sensor_id")
		return
	}

	startTimeStr := r.URL.Query().Get("start_time")
	endTimeStr := r.URL.Query().Get("end_time")

	var startTime, endTime time.Time
	if startTimeStr != "" {
		startTime, err = time.Parse(time.RFC3339, startTimeStr)
		if err != nil {
			Error(w, http.StatusBadRequest, "Invalid start_time format")
			return
		}
	} else {
		startTime = time.Now().Add(-24 * time.Hour)
	}

	if endTimeStr != "" {
		endTime, err = time.Parse(time.RFC3339, endTimeStr)
		if err != nil {
			Error(w, http.StatusBadRequest, "Invalid end_time format")
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
		Error(w, http.StatusInternalServerError, "Failed to query readings")
		return
	}

	JSON(w, http.StatusOK, res.DataPoints)
}

// @Summary Get latest readings for multiple sensors
// @Description Fetches the most recent reading for each specified sensor
// @Tags Data
// @Param sensor_ids query string true "Comma-separated sensor IDs"
// @Success 200 {object} Response{data=object}
// @Failure 400 {object} Response{error=string}
// @Failure 500 {object} Response{error=string}
// @Router /api/data/readings/latest [get]
func (h *WebSocketHandler) GetLatestReadings(w http.ResponseWriter, r *http.Request) {
	sensorIDsParam := r.URL.Query().Get("sensor_ids")
	if sensorIDsParam == "" {
		Error(w, http.StatusBadRequest, "sensor_ids parameter is required")
		return
	}

	var sensorIDs []int64
	for _, idStr := range strings.Split(sensorIDsParam, ",") {
		id, err := strconv.ParseInt(strings.TrimSpace(idStr), 10, 64)
		if err != nil {
			Error(w, http.StatusBadRequest, "Invalid sensor_id: "+idStr)
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
		Error(w, http.StatusInternalServerError, "Failed to get latest readings")
		return
	}

	JSON(w, http.StatusOK, res)
}

// @Summary Get latest N readings for a single sensor
// @Description Fetches the most recent N readings for a specific sensor
// @Tags Data
// @Param sensor_id path int true "Sensor ID"
// @Param limit query int false "Number of readings to fetch (default 10)"
// @Success 200 {object} Response{data=object}
// @Failure 400 {object} Response{error=string}
// @Failure 500 {object} Response{error=string}
// @Router /api/data/sensors/{sensor_id}/latest [get]
func (h *WebSocketHandler) GetSensorLatestReadings(w http.ResponseWriter, r *http.Request) {
	sensorID, err := strconv.ParseInt(chi.URLParam(r, "sensor_id"), 10, 64)
	if err != nil {
		Error(w, http.StatusBadRequest, "Invalid sensor_id")
		return
	}

	limit := int64(10)
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if val, err := strconv.ParseInt(limitStr, 10, 64); err == nil && val > 0 {
			limit = val
		}
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	res, err := h.dataClient.GetLatestReadingsBySensor(ctx, &pb_data.LatestReadingsBySensorRequest{
		SensorId: sensorID,
		Limit:    limit,
	})
	if err != nil {
		Error(w, http.StatusInternalServerError, "Failed to get sensor readings")
		return
	}

	JSON(w, http.StatusOK, res)
}

// @Summary Store a new sensor reading
// @Description Sends a sensor reading to the data processing service
// @Tags Data
// @Accept json
// @Produce json
// @Param reading body types.StoreReadingRequest true "Sensor Reading"
// @Success 200 {object} Response{data=object}
// @Failure 400 {object} Response{error=string}
// @Failure 500 {object} Response{error=string}
// @Router /api/data/readings [post]
func (h *WebSocketHandler) StoreReading(w http.ResponseWriter, r *http.Request) {
	var req types.StoreReadingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		Error(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Timestamp.IsZero() {
		req.Timestamp = time.Now()
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	_, err := h.dataClient.StoreReading(ctx, &pb_data.StoreReadingRequest{
		SensorId:  req.SensorID,
		Value:     req.Value,
		Timestamp: timestamppb.New(req.Timestamp),
	})
	if err != nil {
		Error(w, http.StatusInternalServerError, "Failed to store reading")
		return
	}

	JSON(w, http.StatusOK, map[string]string{"status": "success"})
}

func (h *WebSocketHandler) broadcastAlerts(alertMsgs <-chan amqp.Delivery) {
	for m := range alertMsgs {
		var alert map[string]interface{}
		if err := json.Unmarshal(m.Body, &alert); err != nil {
			continue
		}

		wsMsg := types.AlertMessage{
			Type:    "alert",
			Payload: alert,
		}

		h.clientsMu.RLock()
		for client := range h.clients {
			if err := client.WriteJSON(wsMsg); err != nil {
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
