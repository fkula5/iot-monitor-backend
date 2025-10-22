package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/skni-kod/iot-monitor-backend/internal/proto/sensor_service"
)

type SensorHandler struct {
	client sensor_service.SensorServiceClient
}

type CreateSensorRequest struct {
	Name         string `json:"name"`
	Location     string `json:"location"`
	Description  string `json:"description"`
	SensorTypeId int32  `json:"sensor_type_id"`
}

type UpdateSensorRequest struct {
	Name         *string `json:"name,omitempty"`
	Location     *string `json:"location,omitempty"`
	Description  *string `json:"description,omitempty"`
	SensorTypeId *int32  `json:"sensor_type_id,omitempty"`
	Active       *bool   `json:"active,omitempty"`
}

func NewSensorHandler(client sensor_service.SensorServiceClient) *SensorHandler {
	return &SensorHandler{client: client}
}

func (h *SensorHandler) ListSensors(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	res, err := h.client.ListSensors(ctx, &sensor_service.ListSensorsRequest{})
	if err != nil {
		http.Error(w, "Failed to fetch sensors: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(res.Sensors)
	if err != nil {
		http.Error(w, "Failed to encode response: "+err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *SensorHandler) GetSensor(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid sensor ID", http.StatusBadRequest)
		return
	}

	res, err := h.client.GetSensor(ctx, &sensor_service.GetSensorRequest{Id: int32(id)})
	if err != nil {
		http.Error(w, "Failed to fetch sensor: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if res.Sensor == nil {
		http.Error(w, "Sensor not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	err = json.NewEncoder(w).Encode(res.Sensor)
	if err != nil {
		http.Error(w, "Failed to encode response: "+err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *SensorHandler) SetSensorActive(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid sensor ID", http.StatusBadRequest)
		return
	}

	var request struct {
		Active bool `json:"active"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	res, err := h.client.SetSensorActive(ctx, &sensor_service.SetSensorActiveRequest{
		Id:     int32(id),
		Active: request.Active,
	})
	if err != nil {
		http.Error(w, "Failed to update sensor: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(res.Sensor)
	if err != nil {
		http.Error(w, "Failed to encode response: "+err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *SensorHandler) CreateSensor(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	var req CreateSensorRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	if req.Name == "" {
		http.Error(w, "Name is required", http.StatusBadRequest)
		return
	}
	if req.SensorTypeId <= 0 {
		http.Error(w, "Valid sensor_type_id is required", http.StatusBadRequest)
		return
	}

	grpcReq := &sensor_service.CreateSensorRequest{
		Name:         req.Name,
		Location:     req.Location,
		Description:  req.Description,
		SensorTypeId: req.SensorTypeId,
		Active:       true,
	}

	res, err := h.client.CreateSensor(ctx, grpcReq)
	if err != nil {
		http.Error(w, "Failed to create sensor: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(res.Sensor)
	if err != nil {
		http.Error(w, "Failed to encode response: "+err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *SensorHandler) UpdateSensor(w http.ResponseWriter, r *http.Request) {
	panic("unimplemented")
}

func (h *SensorHandler) DeleteSensor(w http.ResponseWriter, r *http.Request) {
	panic("unimplemented")
}
