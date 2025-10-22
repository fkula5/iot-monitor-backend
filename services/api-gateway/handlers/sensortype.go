package handlers

import (
	"net/http"

	"github.com/skni-kod/iot-monitor-backend/internal/proto/sensor_service"
)

type SensorTypeHandler struct {
	client sensor_service.SensorServiceClient
}

func NewSensorTypeHandler(client sensor_service.SensorServiceClient) *SensorTypeHandler {
	return &SensorTypeHandler{client: client}
}

func (h *SensorTypeHandler) ListSensorTypes(w http.ResponseWriter, r *http.Request) {
	panic("unimplemented")
}

func (h *SensorTypeHandler) GetSensorType(w http.ResponseWriter, r *http.Request) {
	panic("unimplemented")
}

func (h *SensorTypeHandler) CreateSensorType(w http.ResponseWriter, r *http.Request) {
	panic("unimplemented")
}

func (h *SensorTypeHandler) UpdateSensorType(w http.ResponseWriter, r *http.Request) {
	panic("unimplemented")
}

func (h *SensorTypeHandler) DeleteSensorType(w http.ResponseWriter, r *http.Request) {
	panic("unimplemented")
}
