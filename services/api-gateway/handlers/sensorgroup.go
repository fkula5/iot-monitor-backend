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

type SensorGroupHandler struct {
	client pb.SensorServiceClient
}

func NewSensorGroupHandler(client pb.SensorServiceClient) *SensorGroupHandler {
	return &SensorGroupHandler{client: client}
}

// @Summary List sensor groups
// @Description Get all sensor groups for the authenticated user
// @Tags SensorGroups
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} Response{data=[]types.SensorGroupResponse}
// @Failure 401 {object} Response{error=string}
// @Failure 500 {object} Response{error=string}
// @Router /api/sensor-groups [get]
func (h *SensorGroupHandler) ListGroups(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	claims, ok := authMiddleware.GetUserFromContext(r.Context())
	if !ok {
		Error(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	res, err := h.client.ListSensorGroups(ctx, &pb.ListSensorGroupsRequest{
		UserId: int64(claims.UserId),
	})
	if err != nil {
		Error(w, http.StatusInternalServerError, "Failed to list groups")
		return
	}

	groupResponses := make([]types.SensorGroupResponse, 0, len(res.Groups))
	for _, item := range res.Groups {
		mappedSensors := make([]types.SensorResponse, 0, len(item.Sensors))
		for _, s := range item.Sensors {
			mappedSensors = append(mappedSensors, types.MapSensorFromProto(s))
		}
		groupResponses = append(groupResponses, types.MapSensorGroupFromProto(item.Group, mappedSensors))
	}

	JSON(w, http.StatusOK, groupResponses)
}

// @Summary Get sensor group
// @Description Get a specific sensor group with its sensors
// @Tags SensorGroups
// @Produce json
// @Param id path int true "Group ID"
// @Security ApiKeyAuth
// @Success 200 {object} Response{data=types.SensorGroupResponse}
// @Failure 400 {object} Response{error=string}
// @Failure 401 {object} Response{error=string}
// @Failure 404 {object} Response{error=string}
// @Failure 500 {object} Response{error=string}
// @Router /api/sensor-groups/{id} [get]
func (h *SensorGroupHandler) GetGroup(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		Error(w, http.StatusBadRequest, "Invalid group ID")
		return
	}

	res, err := h.client.GetSensorGroup(ctx, &pb.GetSensorGroupRequest{Id: id})
	if err != nil {
		st, ok := status.FromError(err)
		if ok && st.Code() == codes.NotFound {
			Error(w, http.StatusNotFound, "Group not found")
			return
		}
		Error(w, http.StatusInternalServerError, "Failed to get group")
		return
	}

	mappedSensors := make([]types.SensorResponse, 0, len(res.Sensors))
	for _, s := range res.Sensors {
		mappedSensors = append(mappedSensors, types.MapSensorFromProto(s))
	}

	JSON(w, http.StatusOK, types.MapSensorGroupFromProto(res.Group, mappedSensors))
}

// @Summary Create sensor group
// @Description Create a new sensor group
// @Tags SensorGroups
// @Accept json
// @Produce json
// @Param group body types.CreateGroupRequest true "Group data"
// @Security ApiKeyAuth
// @Success 201 {object} Response{data=types.SensorGroupResponse}
// @Failure 400 {object} Response{error=string}
// @Failure 401 {object} Response{error=string}
// @Failure 500 {object} Response{error=string}
// @Router /api/sensor-groups [post]
func (h *SensorGroupHandler) CreateGroup(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	claims, ok := authMiddleware.GetUserFromContext(r.Context())
	if !ok {
		Error(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req types.CreateGroupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		Error(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Name == "" {
		Error(w, http.StatusBadRequest, "Name is required")
		return
	}

	res, err := h.client.CreateSensorGroup(ctx, &pb.CreateSensorGroupRequest{
		Name:        req.Name,
		Description: req.Description,
		Color:       req.Color,
		Icon:        req.Icon,
		UserId:      int64(claims.UserId),
		SensorIds:   req.SensorIDs,
	})
	if err != nil {
		Error(w, http.StatusInternalServerError, "Failed to create group")
		return
	}

	Created(w, types.MapSensorGroupFromProto(res.Group, nil))
}

// @Summary Update sensor group
// @Description Update an existing sensor group
// @Tags SensorGroups
// @Accept json
// @Produce json
// @Param id path int true "Group ID"
// @Param group body types.UpdateGroupRequest true "Group data"
// @Security ApiKeyAuth
// @Success 200 {object} Response{data=object}
// @Failure 400 {object} Response{error=string}
// @Failure 401 {object} Response{error=string}
// @Failure 404 {object} Response{error=string}
// @Failure 500 {object} Response{error=string}
// @Router /api/sensor-groups/{id} [put]
func (h *SensorGroupHandler) UpdateGroup(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		Error(w, http.StatusBadRequest, "Invalid group ID")
		return
	}

	var req types.UpdateGroupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		Error(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Name == "" {
		Error(w, http.StatusBadRequest, "Name is required")
		return
	}

	res, err := h.client.UpdateSensorGroup(ctx, &pb.UpdateSensorGroupRequest{
		Id:          id,
		Name:        req.Name,
		Description: req.Description,
		Color:       req.Color,
		Icon:        req.Icon,
		SensorIds:   req.SensorIDs,
	})
	if err != nil {
		st, ok := status.FromError(err)
		if ok && st.Code() == codes.NotFound {
			Error(w, http.StatusNotFound, "Group not found")
			return
		}
		Error(w, http.StatusInternalServerError, "Failed to update group")
		return
	}

	JSON(w, http.StatusOK, res.Group)
}

// @Summary Delete sensor group
// @Description Delete a sensor group
// @Tags SensorGroups
// @Param id path int true "Group ID"
// @Security ApiKeyAuth
// @Success 204 {object} Response
// @Failure 400 {object} Response{error=string}
// @Failure 401 {object} Response{error=string}
// @Failure 404 {object} Response{error=string}
// @Failure 500 {object} Response{error=string}
// @Router /api/sensor-groups/{id} [delete]
func (h *SensorGroupHandler) DeleteGroup(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		Error(w, http.StatusBadRequest, "Invalid group ID")
		return
	}

	if _, err = h.client.DeleteSensorGroup(ctx, &pb.DeleteSensorGroupRequest{Id: id}); err != nil {
		st, ok := status.FromError(err)
		if ok && st.Code() == codes.NotFound {
			Error(w, http.StatusNotFound, "Group not found")
			return
		}
		Error(w, http.StatusInternalServerError, "Failed to delete group")
		return
	}

	NoContent(w)
}

// @Summary Add sensors to group
// @Description Add sensors to an existing group
// @Tags SensorGroups
// @Accept json
// @Produce json
// @Param id path int true "Group ID"
// @Param sensors body types.AddSensorsRequest true "Sensor IDs to add"
// @Security ApiKeyAuth
// @Success 200 {object} Response{data=object}
// @Failure 400 {object} Response{error=string}
// @Failure 401 {object} Response{error=string}
// @Failure 404 {object} Response{error=string}
// @Failure 500 {object} Response{error=string}
// @Router /api/sensor-groups/{id}/sensors [post]
func (h *SensorGroupHandler) AddSensorsToGroup(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		Error(w, http.StatusBadRequest, "Invalid group ID")
		return
	}

	var req types.AddSensorsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		Error(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if len(req.SensorIDs) == 0 {
		Error(w, http.StatusBadRequest, "At least one sensor ID is required")
		return
	}

	res, err := h.client.AddSensorsToGroup(ctx, &pb.AddSensorsToGroupRequest{
		GroupId:   id,
		SensorIds: req.SensorIDs,
	})
	if err != nil {
		Error(w, http.StatusInternalServerError, "Failed to add sensors")
		return
	}

	JSON(w, http.StatusOK, res.Group)
}

// @Summary Remove sensors from group
// @Description Remove sensors from an existing group
// @Tags SensorGroups
// @Accept json
// @Produce json
// @Param id path int true "Group ID"
// @Param sensors body types.AddSensorsRequest true "Sensor IDs"
// @Security ApiKeyAuth
// @Success 200 {object} Response{data=object}
// @Failure 400 {object} Response{error=string}
// @Failure 401 {object} Response{error=string}
// @Failure 404 {object} Response{error=string}
// @Failure 500 {object} Response{error=string}
// @Router /api/sensor-groups/{id}/sensors [delete]
func (h *SensorGroupHandler) RemoveSensorsFromGroup(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		Error(w, http.StatusBadRequest, "Invalid group ID")
		return
	}

	var req types.AddSensorsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		Error(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if len(req.SensorIDs) == 0 {
		Error(w, http.StatusBadRequest, "At least one sensor ID is required")
		return
	}

	res, err := h.client.RemoveSensorsFromGroup(ctx, &pb.RemoveSensorsFromGroupRequest{
		GroupId:   id,
		SensorIds: req.SensorIDs,
	})
	if err != nil {
		Error(w, http.StatusInternalServerError, "Failed to remove sensors")
		return
	}

	JSON(w, http.StatusOK, res.Group)
}
