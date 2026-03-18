package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/skni-kod/iot-monitor-backend/internal/proto/sensor_service"
	"github.com/skni-kod/iot-monitor-backend/internal/types"
	authMiddleware "github.com/skni-kod/iot-monitor-backend/services/api-gateway/middleware"
)

type SensorHandler struct {
	client pb.SensorServiceClient
}

func NewSensorHandler(client pb.SensorServiceClient) *SensorHandler {
	return &SensorHandler{client: client}
}

// @Summary ListSensors retrieves a list of all sensors.
// @Description Fetches all sensors from the Sensor Service.
// @Tags Sensors
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} Response{data=[]types.SensorResponse} "List of sensors"
// @Failure 401 {object} Response{error=string} "Unauthorized"
// @Failure 500 {object} Response{error=string} "Internal Server Error"
// @Router /api/sensors [get]
func (h *SensorHandler) ListSensors(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	claims, ok := authMiddleware.GetUserFromContext(r.Context())
	if !ok {
		Error(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	res, err := h.client.ListSensors(ctx, &pb.ListSensorsRequest{UserId: int64(claims.UserId)})
	if err != nil {
		Error(w, http.StatusInternalServerError, "Failed to fetch sensors")
		return
	}

	sensorResponses := make([]types.SensorResponse, 0, len(res.Sensors))
	for _, s := range res.Sensors {
		sensorResponses = append(sensorResponses, types.MapSensorFromProto(s))
	}

	JSON(w, http.StatusOK, sensorResponses)
}

// @Summary GetSensor retrieves a sensor by ID.
// @Description Fetches a sensor from the Sensor Service by its ID.
// @Tags Sensors
// @Produce json
// @Param id path int true "Sensor ID"
// @Security ApiKeyAuth
// @Success 200 {object} Response{data=types.SensorResponse} "Sensor details"
// @Failure 400 {object} Response{error=string} "Bad Request"
// @Failure 401 {object} Response{error=string} "Unauthorized"
// @Failure 404 {object} Response{error=string} "Not Found"
// @Failure 500 {object} Response{error=string} "Internal Server Error"
// @Router /api/sensors/{id} [get]
func (h *SensorHandler) GetSensor(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		Error(w, http.StatusBadRequest, "Invalid sensor ID")
		return
	}

	res, err := h.client.GetSensor(ctx, &pb.GetSensorRequest{Id: int64(id)})
	if err != nil {
		Error(w, http.StatusInternalServerError, "Failed to fetch sensor")
		return
	}

	if res.Sensor == nil {
		Error(w, http.StatusNotFound, "Sensor not found")
		return
	}

	JSON(w, http.StatusOK, types.MapSensorFromProto(res.Sensor))
}

// @Summary SetSensorActive sets a sensor as active.
// @Description Marks a sensor as active in the Sensor Service.
// @Tags Sensors
// @Produce json
// @Security ApiKeyAuth
// @Param id path int true "Sensor ID"
// @Success 200 {object} Response{data=object} "Updated sensor details"
// @Failure 400 {object} Response{error=string} "Bad Request"
// @Failure 401 {object} Response{error=string} "Unauthorized"
// @Failure 500 {object} Response{error=string} "Internal Server Error"
// @Router /api/sensors/{id}/activate [post]
func (h *SensorHandler) SetSensorActive(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		Error(w, http.StatusBadRequest, "Invalid sensor ID")
		return
	}

	res, err := h.client.SetSensorActive(ctx, &pb.SetSensorActiveRequest{
		Id: int64(id),
	})
	if err != nil {
		Error(w, http.StatusInternalServerError, "Failed to update sensor: "+err.Error())
		return
	}

	JSON(w, http.StatusOK, res.Sensor)
}

// @Summary CreateSensor creates a new sensor.
// @Description Creates a new sensor in the Sensor Service.
// @Tags Sensors
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param sensor body types.CreateSensorRequest true "Sensor to create"
// @Success 201 {object} Response{data=object} "Created sensor details"
// @Failure 400 {object} Response{error=string} "Bad Request"
// @Failure 401 {object} Response{error=string} "Unauthorized"
// @Failure 500 {object} Response{error=string} "Internal Server Error"
// @Router /api/sensors [post]
func (h *SensorHandler) CreateSensor(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	claims, ok := authMiddleware.GetUserFromContext(r.Context())
	if !ok {
		Error(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req types.CreateSensorRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		Error(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Name == "" {
		Error(w, http.StatusBadRequest, "Name is required")
		return
	}
	if req.SensorTypeId <= 0 {
		Error(w, http.StatusBadRequest, "Valid sensor_type_id is required")
		return
	}

	res, err := h.client.CreateSensor(ctx, &pb.CreateSensorRequest{
		Name:         req.Name,
		Location:     req.Location,
		Description:  req.Description,
		SensorTypeId: req.SensorTypeId,
		UserId:       int64(claims.UserId),
		Active:       true,
	})
	if err != nil {
		Error(w, http.StatusInternalServerError, "Failed to create sensor: "+err.Error())
		return
	}

	Created(w, res.Sensor)
}

// @Summary UpdateSensor updates an existing sensor.
// @Description Updates an existing sensor in the Sensor Service.
// @Tags Sensors
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path int true "Sensor ID"
// @Param sensor body types.UpdateSensorRequest true "Sensor fields to update"
// @Success 200 {object} Response{data=object} "Updated sensor details"
// @Failure 400 {object} Response{error=string} "Bad Request"
// @Failure 401 {object} Response{error=string} "Unauthorized"
// @Failure 404 {object} Response{error=string} "Not Found"
// @Failure 500 {object} Response{error=string} "Internal Server Error"
// @Router /api/sensors/{id} [put]
func (h *SensorHandler) UpdateSensor(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		Error(w, http.StatusBadRequest, "Invalid sensor ID")
		return
	}

	var req types.UpdateSensorRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		Error(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	currentSensorRes, err := h.client.GetSensor(ctx, &pb.GetSensorRequest{Id: int64(id)})
	if err != nil {
		st, ok := status.FromError(err)
		if ok && st.Code() == codes.NotFound {
			Error(w, http.StatusNotFound, "Sensor not found")
			return
		}
		Error(w, http.StatusInternalServerError, "Failed to fetch current sensor state")
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
		if _, err := h.client.GetSensorType(ctx, &pb.GetSensorTypeRequest{Id: *req.SensorTypeId}); err != nil {
			st, ok := status.FromError(err)
			if ok && st.Code() == codes.NotFound {
				Error(w, http.StatusBadRequest, "Specified SensorTypeID not found")
				return
			}
			Error(w, http.StatusInternalServerError, "Failed to validate SensorTypeID")
			return
		}
		grpcReq.SensorTypeId = *req.SensorTypeId
	}

	if grpcReq.Name == "" {
		Error(w, http.StatusBadRequest, "Name cannot be empty")
		return
	}

	res, err := h.client.UpdateSensor(ctx, grpcReq)
	if err != nil {
		st, ok := status.FromError(err)
		if ok && st.Code() == codes.NotFound {
			Error(w, http.StatusNotFound, "Sensor not found")
			return
		}
		Error(w, http.StatusInternalServerError, "Failed to update sensor")
		return
	}

	JSON(w, http.StatusOK, res.Sensor)
}

// @Summary DeleteSensor deletes a sensor by ID.
// @Description Deletes a sensor from the Sensor Service by its ID.
// @Tags Sensors
// @Param id path int true "Sensor ID"
// @Security ApiKeyAuth
// @Success 204 {object} Response "No Content"
// @Failure 400 {object} Response{error=string} "Bad Request"
// @Failure 401 {object} Response{error=string} "Unauthorized"
// @Failure 404 {object} Response{error=string} "Not Found"
// @Failure 500 {object} Response{error=string} "Internal Server Error"
// @Router /api/sensors/{id} [delete]
func (h *SensorHandler) DeleteSensor(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		Error(w, http.StatusBadRequest, "Invalid sensor ID")
		return
	}

	if _, err = h.client.DeleteSensor(ctx, &pb.DeleteSensorRequest{Id: int64(id)}); err != nil {
		st, ok := status.FromError(err)
		if ok && st.Code() == codes.NotFound {
			Error(w, http.StatusNotFound, "Sensor not found")
			return
		}
		Error(w, http.StatusInternalServerError, "Failed to delete sensor")
		return
	}

	NoContent(w)
}
