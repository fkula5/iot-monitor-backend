package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	pb "github.com/skni-kod/iot-monitor-backend/internal/proto/sensor_service"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type SensorTypeHandler struct {
	client pb.SensorServiceClient
}

func NewSensorTypeHandler(client pb.SensorServiceClient) *SensorTypeHandler {
	return &SensorTypeHandler{client: client}
}

// @Summary List Sensor Types
// @Description Get a list of all sensor types
// @Tags SensorTypes
// @Produce json
// @Success 200 {array} string
// @Failure 500 {object} map[string]string
// @Router /sensortypes [get]
func (h *SensorTypeHandler) ListSensorTypes(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	res, err := h.client.ListSensorTypes(ctx, &pb.ListSensorTypesRequest{})
	if err != nil {
		http.Error(w, "Failed to list sensor types: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(res.SensorTypes); err != nil {
		http.Error(w, "Failed to encode response: "+err.Error(), http.StatusInternalServerError)
	}
}

// @Summary Get Sensor Type
// @Description Get details of a specific sensor type by ID
// @Tags SensorTypes
// @Produce json
// @Param id path int true "Sensor Type ID"
// @Success 200 {object} string
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /sensortypes/{id} [get]
func (h *SensorTypeHandler) GetSensorType(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid sensor type ID", http.StatusBadRequest)
		return
	}

	res, err := h.client.GetSensorType(ctx, &pb.GetSensorTypeRequest{Id: int32(id)})
	if err != nil {
		st, ok := status.FromError(err)
		if ok && st.Code() == codes.NotFound {
			http.Error(w, "Sensor type not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Failed to get sensor type: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(res.SensorType); err != nil {
		http.Error(w, "Failed to encode response: "+err.Error(), http.StatusInternalServerError)
	}
}

// @Summary Create Sensor Type
// @Description Create a new sensor type
// @Tags SensorTypes
// @Accept json
// @Produce json
// @Param sensorType body string true "Sensor Type Data"
// @Success 201 {object} string
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /sensortypes [post]
func (h *SensorTypeHandler) CreateSensorType(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	var req pb.CreateSensorTypeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	if req.Name == "" || req.Model == "" {
		http.Error(w, "Name and Model are required", http.StatusBadRequest)
		return
	}

	res, err := h.client.CreateSensorType(ctx, &req)
	if err != nil {
		http.Error(w, "Failed to create sensor type: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(res.SensorType); err != nil {
		http.Error(w, "Failed to encode response: "+err.Error(), http.StatusInternalServerError)
	}
}

// @Summary Update Sensor Type
// @Description Update an existing sensor type by ID
// @Tags SensorTypes
// @Accept json
// @Produce json
// @Param id path int true "Sensor Type ID"
// @Param sensorType body string true "Sensor Type Data"
// @Success 200 {object} string
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /sensortypes/{id} [put]
func (h *SensorTypeHandler) UpdateSensorType(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid sensor type ID", http.StatusBadRequest)
		return
	}

	var req pb.UpdateSensorTypeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	req.Id = int32(id)

	if req.Name == "" || req.Model == "" {
		http.Error(w, "Name and Model are required", http.StatusBadRequest)
		return
	}

	res, err := h.client.UpdateSensorType(ctx, &req)
	if err != nil {
		st, ok := status.FromError(err)
		if ok && st.Code() == codes.NotFound {
			http.Error(w, "Sensor type not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Failed to update sensor type: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(res.SensorType); err != nil {
		http.Error(w, "Failed to encode response: "+err.Error(), http.StatusInternalServerError)
	}
}

// @Summary Delete Sensor Type
// @Description Delete a sensor type by ID
// @Tags SensorTypes
// @Param id path int true "Sensor Type ID"
// @Success 204
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /sensortypes/{id} [delete]
func (h *SensorTypeHandler) DeleteSensorType(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid sensor type ID", http.StatusBadRequest)
		return
	}

	_, err = h.client.DeleteSensorType(ctx, &pb.DeleteSensorTypeRequest{Id: int32(id)})
	if err != nil {
		st, ok := status.FromError(err)
		if ok && st.Code() == codes.NotFound {
			http.Error(w, "Sensor type not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Failed to delete sensor type: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
