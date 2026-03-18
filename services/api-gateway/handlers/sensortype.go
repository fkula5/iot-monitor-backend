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
// @Security ApiKeyAuth
// @Success 200 {object} Response{data=[]types.SensorTypeResponse}
// @Failure 401 {object} Response{error=string}
// @Failure 500 {object} Response{error=string}
// @Router /api/sensor-types [get]
func (h *SensorTypeHandler) ListSensorTypes(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	res, err := h.client.ListSensorTypes(ctx, &pb.ListSensorTypesRequest{})
	if err != nil {
		Error(w, http.StatusInternalServerError, "Failed to list sensor types")
		return
	}

	sensorTypeResponses := make([]types.SensorTypeResponse, 0, len(res.SensorTypes))
	for _, st := range res.SensorTypes {
		sensorTypeResponses = append(sensorTypeResponses, types.SensorTypeResponse{
			ID:           st.Id,
			Name:         st.Name,
			Model:        st.Model,
			Manufacturer: st.Manufacturer,
			Description:  st.Description,
			Unit:         st.Unit,
			MinValue:     st.MinValue,
			MaxValue:     st.MaxValue,
			CreatedAt:    st.CreatedAt.AsTime(),
		})
	}

	JSON(w, http.StatusOK, sensorTypeResponses)
}

// @Summary Get Sensor Type
// @Description Get details of a specific sensor type by ID
// @Tags SensorTypes
// @Produce json
// @Param id path int true "Sensor Type ID"
// @Security ApiKeyAuth
// @Success 200 {object} Response{data=types.SensorTypeResponse}
// @Failure 400 {object} Response{error=string}
// @Failure 401 {object} Response{error=string}
// @Failure 404 {object} Response{error=string}
// @Failure 500 {object} Response{error=string}
// @Router /api/sensor-types/{id} [get]
func (h *SensorTypeHandler) GetSensorType(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		Error(w, http.StatusBadRequest, "Invalid sensor type ID")
		return
	}

	res, err := h.client.GetSensorType(ctx, &pb.GetSensorTypeRequest{Id: int64(id)})
	if err != nil {
		st, ok := status.FromError(err)
		if ok && st.Code() == codes.NotFound {
			Error(w, http.StatusNotFound, "Sensor type not found")
			return
		}
		Error(w, http.StatusInternalServerError, "Failed to get sensor type")
		return
	}

	JSON(w, http.StatusOK, types.SensorTypeResponse{
		ID:           res.SensorType.Id,
		Name:         res.SensorType.Name,
		Model:        res.SensorType.Model,
		Manufacturer: res.SensorType.Manufacturer,
		Description:  res.SensorType.Description,
		Unit:         res.SensorType.Unit,
		MinValue:     res.SensorType.MinValue,
		MaxValue:     res.SensorType.MaxValue,
		CreatedAt:    res.SensorType.CreatedAt.AsTime(),
	})
}

// @Summary Create Sensor Type
// @Description Create a new sensor type
// @Tags SensorTypes
// @Accept json
// @Produce json
// @Param sensorType body string true "Sensor Type Data"
// @Security ApiKeyAuth
// @Success 201 {object} Response{data=object}
// @Failure 400 {object} Response{error=string}
// @Failure 401 {object} Response{error=string}
// @Failure 500 {object} Response{error=string}
// @Router /api/sensor-types [post]
func (h *SensorTypeHandler) CreateSensorType(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	var req pb.CreateSensorTypeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		Error(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Name == "" || req.Model == "" {
		Error(w, http.StatusBadRequest, "Name and Model are required")
		return
	}

	res, err := h.client.CreateSensorType(ctx, &req)
	if err != nil {
		Error(w, http.StatusInternalServerError, "Failed to create sensor type")
		return
	}

	Created(w, res.SensorType)
}

// @Summary Update Sensor Type
// @Description Update an existing sensor type by ID
// @Tags SensorTypes
// @Accept json
// @Produce json
// @Param id path int true "Sensor Type ID"
// @Param sensorType body string true "Sensor Type Data"
// @Security ApiKeyAuth
// @Success 200 {object} Response{data=object}
// @Failure 400 {object} Response{error=string}
// @Failure 401 {object} Response{error=string}
// @Failure 404 {object} Response{error=string}
// @Failure 500 {object} Response{error=string}
// @Router /api/sensor-types/{id} [put]
func (h *SensorTypeHandler) UpdateSensorType(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		Error(w, http.StatusBadRequest, "Invalid sensor type ID")
		return
	}

	var req pb.UpdateSensorTypeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		Error(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	req.Id = int64(id)
	if req.Name == "" || req.Model == "" {
		Error(w, http.StatusBadRequest, "Name and Model are required")
		return
	}

	res, err := h.client.UpdateSensorType(ctx, &req)
	if err != nil {
		st, ok := status.FromError(err)
		if ok && st.Code() == codes.NotFound {
			Error(w, http.StatusNotFound, "Sensor type not found")
			return
		}
		Error(w, http.StatusInternalServerError, "Failed to update sensor type")
		return
	}

	JSON(w, http.StatusOK, res.SensorType)
}

// @Summary Delete Sensor Type
// @Description Delete a sensor type by ID
// @Tags SensorTypes
// @Param id path int true "Sensor Type ID"
// @Security ApiKeyAuth
// @Success 204 {object} Response
// @Failure 400 {object} Response{error=string}
// @Failure 401 {object} Response{error=string}
// @Failure 404 {object} Response{error=string}
// @Failure 500 {object} Response{error=string}
// @Router /api/sensor-types/{id} [delete]
func (h *SensorTypeHandler) DeleteSensorType(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		Error(w, http.StatusBadRequest, "Invalid sensor type ID")
		return
	}

	if _, err = h.client.DeleteSensorType(ctx, &pb.DeleteSensorTypeRequest{Id: int64(id)}); err != nil {
		st, ok := status.FromError(err)
		if ok && st.Code() == codes.NotFound {
			Error(w, http.StatusNotFound, "Sensor type not found")
			return
		}
		Error(w, http.StatusInternalServerError, "Failed to delete sensor type")
		return
	}

	NoContent(w)
}
