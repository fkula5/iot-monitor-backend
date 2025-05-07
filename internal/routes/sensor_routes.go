package routes

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

func SetupSensorRoutes(r chi.Router, client sensor_service.SensorServiceClient) {
	handler := &SensorHandler{client: client}
	r.Route("/sensors", func(r chi.Router) {
		r.Get("/", handler.ListSensors)
		r.Get("/{id}", handler.GetSensor)
		r.Put("/{id}/active", handler.SetSensorActive)
	})
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
