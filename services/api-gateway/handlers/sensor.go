package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	pb "github.com/skni-kod/iot-monitor-backend/internal/proto/sensor_service"
	authMiddleware "github.com/skni-kod/iot-monitor-backend/services/api-gateway/middleware"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type SensorHandler struct {
	client pb.SensorServiceClient
}

type CreateSensorRequest struct {
	Name         string `json:"name"`
	Location     string `json:"location"`
	Description  string `json:"description"`
	SensorTypeId int64  `json:"sensor_type_id"`
}

type UpdateSensorRequest struct {
	Name         *string `json:"name,omitempty"`
	Location     *string `json:"location,omitempty"`
	Description  *string `json:"description,omitempty"`
	SensorTypeId *int64  `json:"sensor_type_id,omitempty"`
	Active       *bool   `json:"active,omitempty"`
}

func NewSensorHandler(client pb.SensorServiceClient) *SensorHandler {
	return &SensorHandler{client: client}
}

// @Summary ListSensors retrieves a list of all sensors.
// @Description Fetches all sensors from the Sensor Service.
// @Tags Sensors
// @Produce json
// @Success 200 {array} string "List of sensors"
// @Failure 500 {string} string "Internal Server Error"
// @Router /sensors [get]
func (h *SensorHandler) ListSensors(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	claims, ok := authMiddleware.GetUserFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized: Could not retrieve user information", http.StatusUnauthorized)
		return
	}

	res, err := h.client.ListSensors(ctx, &pb.ListSensorsRequest{UserId: int64(claims.UserId)})
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

// @Summary GetSensor retrieves a sensor by ID.
// @Description Fetches a sensor from the Sensor Service by its ID.
// @Tags Sensors
// @Produce json
// @Param id path int true "Sensor ID"
// @Success 200 {object} string "Sensor details"
// @Failure 400 {string} string "Bad Request"
// @Failure 404 {string} string "Not Found"
// @Failure 500 {string} string "Internal Server Error"
// @Router /sensors/{id} [get]
func (h *SensorHandler) GetSensor(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid sensor ID", http.StatusBadRequest)
		return
	}

	res, err := h.client.GetSensor(ctx, &pb.GetSensorRequest{Id: int64(id)})
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

// @Summary SetSensorActive sets a sensor as active.
// @Description Marks a sensor as active in the Sensor Service.
// @Tags Sensors
// @Produce json
// @Param id path int true "Sensor ID"
// @Success 200 {object} string "Updated sensor details"
// @Failure 400 {string} string "Bad Request"
// @Failure 500 {string} string "Internal Server Error"
// @Router /sensors/{id}/activate [post]
func (h *SensorHandler) SetSensorActive(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid sensor ID", http.StatusBadRequest)
		return
	}

	res, err := h.client.SetSensorActive(ctx, &pb.SetSensorActiveRequest{
		Id: int64(id),
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

// @Summary CreateSensor creates a new sensor.
// @Description Creates a new sensor in the Sensor Service.
// @Tags Sensors
// @Accept json
// @Produce json
// @Param sensor body CreateSensorRequest true "Sensor to create"
// @Success 201 {object} string "Created sensor details"
// @Failure 400 {string} string "Bad Request"
// @Failure 500 {string} string "Internal Server Error"
// @Router /sensors [post]
func (h *SensorHandler) CreateSensor(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	claims, ok := authMiddleware.GetUserFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized: Could not retrieve user information", http.StatusUnauthorized)
		return
	}

	userId := claims.UserId

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

	grpcReq := &pb.CreateSensorRequest{
		Name:         req.Name,
		Location:     req.Location,
		Description:  req.Description,
		SensorTypeId: req.SensorTypeId,
		UserId:       int64(userId),
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

// @Summary UpdateSensor updates an existing sensor.
// @Description Updates an existing sensor in the Sensor Service.
// @Tags Sensors
// @Accept json
// @Produce json
// @Param id path int true "Sensor ID"
// @Param sensor body UpdateSensorRequest true "Sensor fields to update"
// @Success 200 {object} string "Updated sensor details"
// @Failure 400 {string} string "Bad Request"
// @Failure 404 {string} string "Not Found"
// @Failure 500 {string} string "Internal Server Error"
// @Router /sensors/{id} [put]
func (h *SensorHandler) UpdateSensor(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid sensor ID", http.StatusBadRequest)
		return
	}

	var req UpdateSensorRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	currentSensorRes, err := h.client.GetSensor(ctx, &pb.GetSensorRequest{Id: int64(id)})
	if err != nil {
		st, ok := status.FromError(err)
		if ok && st.Code() == codes.NotFound {
			http.Error(w, "Sensor not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Failed to fetch current sensor state: "+err.Error(), http.StatusInternalServerError)
		return
	}
	currentSensor := currentSensorRes.Sensor

	grpcReq := &pb.UpdateSensorRequest{
		Id:           int64(id),
		Name:         currentSensor.Name,
		Location:     currentSensor.Location,
		Description:  currentSensor.Description,
		Active:       currentSensor.Active,
		SensorTypeId: currentSensor.SensorTypeId,
	}

	if req.Name != nil {
		grpcReq.Name = *req.Name
	}
	if req.Location != nil {
		grpcReq.Location = *req.Location
	}
	if req.Description != nil {
		grpcReq.Description = *req.Description
	}
	if req.Active != nil {
		grpcReq.Active = *req.Active
	}
	if req.SensorTypeId != nil {

		_, err := h.client.GetSensorType(ctx, &pb.GetSensorTypeRequest{Id: *req.SensorTypeId})
		if err != nil {
			st, ok := status.FromError(err)
			if ok && st.Code() == codes.NotFound {
				http.Error(w, "Specified SensorTypeID not found", http.StatusBadRequest)
				return
			}
			http.Error(w, "Failed to validate SensorTypeID: "+err.Error(), http.StatusInternalServerError)
			return
		}
		grpcReq.SensorTypeId = *req.SensorTypeId
	}

	if grpcReq.Name == "" {
		http.Error(w, "Name cannot be empty", http.StatusBadRequest)
		return
	}
	if grpcReq.SensorTypeId <= 0 {
		http.Error(w, "Valid SensorTypeID is required", http.StatusBadRequest)
		return
	}

	res, err := h.client.UpdateSensor(ctx, grpcReq)
	if err != nil {

		st, ok := status.FromError(err)
		if ok && st.Code() == codes.NotFound {
			http.Error(w, "Sensor not found or related entity missing", http.StatusNotFound)
			return
		}
		http.Error(w, "Failed to update sensor: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(res.Sensor); err != nil {
		http.Error(w, "Failed to encode response: "+err.Error(), http.StatusInternalServerError)
	}
}

// @Summary DeleteSensor deletes a sensor by ID.
// @Description Deletes a sensor from the Sensor Service by its ID.
// @Tags Sensors
// @Param id path int true "Sensor ID"
// @Success 204 {string} string "No Content"
// @Failure 400 {string} string "Bad Request"
// @Failure 404 {string} string "Not Found"
// @Failure 500 {string} string "Internal Server Error"
// @Router /sensors/{id} [delete]
func (h *SensorHandler) DeleteSensor(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid sensor ID", http.StatusBadRequest)
		return
	}

	_, err = h.client.DeleteSensor(ctx, &pb.DeleteSensorRequest{Id: int64(id)})
	if err != nil {
		st, ok := status.FromError(err)
		if ok && st.Code() == codes.NotFound {
			http.Error(w, "Sensor not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Failed to delete sensor: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
