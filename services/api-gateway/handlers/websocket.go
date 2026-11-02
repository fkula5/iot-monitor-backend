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
	"github.com/skni-kod/iot-monitor-backend/pkg/validator"
	authMiddleware "github.com/skni-kod/iot-monitor-backend/services/api-gateway/middleware"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

const wsWriteTimeout = 10 * time.Second

type WebSocketHandler struct {
	dataClient   pb_data.DataServiceClient
	sensorClient pb_sensor.SensorServiceClient
	clients      map[*wsClient]struct{}
	clientsMu    sync.RWMutex
}

// wsClient wraps a WebSocket connection. gorilla/websocket allows only one
// concurrent writer, so all writes go through writeJSON. ctx is cancelled when
// the connection closes, which stops the client's reading streams.
type wsClient struct {
	conn    *websocket.Conn
	userID  int64
	ctx     context.Context
	writeMu sync.Mutex
}

func (c *wsClient) writeJSON(v interface{}) error {
	c.writeMu.Lock()
	defer c.writeMu.Unlock()
	if err := c.conn.SetWriteDeadline(time.Now().Add(wsWriteTimeout)); err != nil {
		return err
	}
	return c.conn.WriteJSON(v)
}

func NewWebSocketHandler(dataClient pb_data.DataServiceClient, sensorClient pb_sensor.SensorServiceClient, alertMsgs <-chan amqp.Delivery) *WebSocketHandler {
	h := &WebSocketHandler{
		dataClient:   dataClient,
		sensorClient: sensorClient,
		clients:      make(map[*wsClient]struct{}),
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
	claims, ok := authMiddleware.GetUserFromContext(r.Context())
	if !ok {
		Error(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	userID := int64(claims.UserId)

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	owned, err := h.userSensors(ctx, userID)
	cancel()
	if err != nil {
		Error(w, http.StatusInternalServerError, "Failed to fetch sensors")
		return
	}

	var sensorIDs []int64
	if sensorIDsParam := r.URL.Query().Get("sensor_ids"); sensorIDsParam != "" {
		for _, idStr := range strings.Split(sensorIDsParam, ",") {
			id, err := strconv.ParseInt(strings.TrimSpace(idStr), 10, 64)
			if err == nil {
				sensorIDs = append(sensorIDs, id)
			}
		}
		sensorIDs = filterOwned(sensorIDs, owned)
	} else {
		for id, active := range owned {
			if active {
				sensorIDs = append(sensorIDs, id)
			}
		}
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logger.Error("Failed to upgrade to WebSocket", zap.Error(err))
		return
	}

	ctx, cancel = context.WithCancel(context.Background())
	client := &wsClient{conn: conn, userID: userID, ctx: ctx}

	h.clientsMu.Lock()
	h.clients[client] = struct{}{}
	h.clientsMu.Unlock()

	defer func() {
		cancel()
		h.clientsMu.Lock()
		delete(h.clients, client)
		h.clientsMu.Unlock()
		conn.Close()
	}()

	if len(sensorIDs) > 0 {
		go h.streamToClient(client, sensorIDs)
	}

	for {
		var msg types.SubscribeMessage
		if err := conn.ReadJSON(&msg); err != nil {
			break
		}

		if msg.Type == "subscribe" {
			if ids := filterOwned(msg.SensorIDs, owned); len(ids) > 0 {
				go h.streamToClient(client, ids)
			}
		}
	}
}

// userSensors returns the IDs of the user's sensors mapped to their active flag.
func (h *WebSocketHandler) userSensors(ctx context.Context, userID int64) (map[int64]bool, error) {
	res, err := h.sensorClient.ListSensors(ctx, &pb_sensor.ListSensorsRequest{UserId: userID})
	if err != nil {
		return nil, err
	}
	owned := make(map[int64]bool, len(res.Sensors))
	for _, s := range res.Sensors {
		owned[s.Id] = s.Active
	}
	return owned, nil
}

func filterOwned(ids []int64, owned map[int64]bool) []int64 {
	result := make([]int64, 0, len(ids))
	for _, id := range ids {
		if _, ok := owned[id]; ok {
			result = append(result, id)
		}
	}
	return result
}

// authorizeSensors writes an error response and returns false unless every
// sensor ID belongs to the authenticated user.
func (h *WebSocketHandler) authorizeSensors(w http.ResponseWriter, r *http.Request, ids ...int64) bool {
	claims, ok := authMiddleware.GetUserFromContext(r.Context())
	if !ok {
		Error(w, http.StatusUnauthorized, "Unauthorized")
		return false
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	owned, err := h.userSensors(ctx, int64(claims.UserId))
	if err != nil {
		Error(w, http.StatusInternalServerError, "Failed to fetch sensors")
		return false
	}
	if len(filterOwned(ids, owned)) != len(ids) {
		Error(w, http.StatusNotFound, "Sensor not found")
		return false
	}
	return true
}

func (h *WebSocketHandler) streamToClient(client *wsClient, sensorIDs []int64) {
	stream, err := h.dataClient.StreamReadings(client.ctx, &pb_data.StreamReadingsRequest{
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

		if err := client.writeJSON(msg); err != nil {
			client.conn.Close()
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
	if !h.authorizeSensors(w, r, sensorID) {
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
	if !h.authorizeSensors(w, r, sensorIDs...) {
		return
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
	if !h.authorizeSensors(w, r, sensorID) {
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

	if errs := validator.ValidateStruct(req); errs != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{"error": "Validation failed", "details": errs})
		return
	}
	if !h.authorizeSensors(w, r, req.SensorID) {
		return
	}

	if req.Timestamp.IsZero() {
		req.Timestamp = time.Now()
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	_, err := h.dataClient.StoreReading(ctx, &pb_data.StoreReadingRequest{
		SensorId:  req.SensorID,
		Value:     *req.Value,
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
		var owner struct {
			UserID int64 `json:"user_id"`
		}
		if err := json.Unmarshal(m.Body, &owner); err != nil || owner.UserID == 0 {
			continue
		}

		wsMsg := types.AlertMessage{
			Type:    "alert",
			Payload: alert,
		}

		h.clientsMu.RLock()
		var recipients []*wsClient
		for client := range h.clients {
			if client.userID == owner.UserID {
				recipients = append(recipients, client)
			}
		}
		h.clientsMu.RUnlock()

		// Closing the connection ends its read loop in HandleReadings, which
		// removes the client from the map.
		for _, client := range recipients {
			if err := client.writeJSON(wsMsg); err != nil {
				client.conn.Close()
			}
		}
	}
}
