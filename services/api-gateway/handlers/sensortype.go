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
