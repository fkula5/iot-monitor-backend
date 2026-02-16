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
	authMiddleware "github.com/skni-kod/iot-monitor-backend/services/api-gateway/middleware"
)

type SensorGroupHandler struct {
	client pb.SensorServiceClient
}

func NewSensorGroupHandler(client pb.SensorServiceClient) *SensorGroupHandler {
	return &SensorGroupHandler{client: client}
}

type CreateGroupRequest struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Color       string  `json:"color"`
	Icon        string  `json:"icon"`
	SensorIDs   []int64 `json:"sensor_ids"`
}

type UpdateGroupRequest struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Color       string  `json:"color"`
	Icon        string  `json:"icon"`
	SensorIDs   []int64 `json:"sensor_ids"`
}

type AddSensorsRequest struct {
	SensorIDs []int64 `json:"sensor_ids"`
}

// @Summary List sensor groups
// @Description Get all sensor groups for the authenticated user
// @Tags SensorGroups
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {array} string
// @Failure 401 {string} string "Unauthorized"
// @Failure 500 {object} map[string]string
// @Router /api/sensor-groups [get]
func (h *SensorGroupHandler) ListGroups(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	claims, ok := authMiddleware.GetUserFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	res, err := h.client.ListSensorGroups(ctx, &pb.ListSensorGroupsRequest{
		UserId: int64(claims.UserId),
	})
	if err != nil {
		http.Error(w, "Failed to list groups: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res.Groups)
}

// @Summary Get sensor group
// @Description Get a specific sensor group with its sensors
// @Tags SensorGroups
// @Produce json
// @Param id path int true "Group ID"
// @Security ApiKeyAuth
// @Success 200 {object} string
// @Failure 400 {object} map[string]string
// @Failure 401 {string} string "Unauthorized"
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/sensor-groups/{id} [get]
func (h *SensorGroupHandler) GetGroup(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid group ID", http.StatusBadRequest)
		return
	}

	res, err := h.client.GetSensorGroup(ctx, &pb.GetSensorGroupRequest{Id: id})
	if err != nil {
		st, ok := status.FromError(err)
		if ok && st.Code() == codes.NotFound {
			http.Error(w, "Group not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Failed to get group: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}

// @Summary Create sensor group
// @Description Create a new sensor group
// @Tags SensorGroups
// @Accept json
// @Produce json
// @Param group body CreateGroupRequest true "Group data"
// @Security ApiKeyAuth
// @Success 201 {object} string
// @Failure 400 {object} map[string]string
// @Failure 401 {string} string "Unauthorized"
// @Failure 500 {object} map[string]string
// @Router /api/sensor-groups [post]
func (h *SensorGroupHandler) CreateGroup(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	claims, ok := authMiddleware.GetUserFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req CreateGroupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Name == "" {
		http.Error(w, "Name is required", http.StatusBadRequest)
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
		http.Error(w, "Failed to create group: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(res.Group)
}

// @Summary Update sensor group
// @Description Update an existing sensor group
// @Tags SensorGroups
// @Accept json
// @Produce json
// @Param id path int true "Group ID"
// @Param group body UpdateGroupRequest true "Group data"
// @Security ApiKeyAuth
// @Success 200 {object} string
// @Failure 400 {object} map[string]string
// @Failure 401 {string} string "Unauthorized"
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/sensor-groups/{id} [put]
func (h *SensorGroupHandler) UpdateGroup(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid group ID", http.StatusBadRequest)
		return
	}

	var req UpdateGroupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Name == "" {
		http.Error(w, "Name is required", http.StatusBadRequest)
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
			http.Error(w, "Group not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Failed to update group: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res.Group)
}

// @Summary Delete sensor group
// @Description Delete a sensor group
// @Tags SensorGroups
// @Param id path int true "Group ID"
// @Security ApiKeyAuth
// @Success 204
// @Failure 400 {object} map[string]string
// @Failure 401 {string} string "Unauthorized"
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/sensor-groups/{id} [delete]
func (h *SensorGroupHandler) DeleteGroup(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid group ID", http.StatusBadRequest)
		return
	}

	_, err = h.client.DeleteSensorGroup(ctx, &pb.DeleteSensorGroupRequest{Id: id})
	if err != nil {
		st, ok := status.FromError(err)
		if ok && st.Code() == codes.NotFound {
			http.Error(w, "Group not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Failed to delete group: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// @Summary Add sensors to group
// @Description Add sensors to an existing group
// @Tags SensorGroups
// @Accept json
// @Produce json
// @Param id path int true "Group ID"
// @Param sensors body AddSensorsRequest true "Sensor IDs"
// @Security ApiKeyAuth
// @Success 200 {object} string
// @Failure 400 {object} map[string]string
// @Failure 401 {string} string "Unauthorized"
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/sensor-groups/{id}/sensors [post]
func (h *SensorGroupHandler) AddSensorsToGroup(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid group ID", http.StatusBadRequest)
		return
	}

	var req AddSensorsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if len(req.SensorIDs) == 0 {
		http.Error(w, "At least one sensor ID is required", http.StatusBadRequest)
		return
	}

	res, err := h.client.AddSensorsToGroup(ctx, &pb.AddSensorsToGroupRequest{
		GroupId:   id,
		SensorIds: req.SensorIDs,
	})
	if err != nil {
		http.Error(w, "Failed to add sensors: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res.Group)
}

// @Summary Remove sensors from group
// @Description Remove sensors from an existing group
// @Tags SensorGroups
// @Accept json
// @Produce json
// @Param id path int true "Group ID"
// @Param sensors body AddSensorsRequest true "Sensor IDs"
// @Security ApiKeyAuth
// @Success 200 {object} string
// @Failure 400 {object} map[string]string
// @Failure 401 {string} string "Unauthorized"
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/sensor-groups/{id}/sensors [delete]
func (h *SensorGroupHandler) RemoveSensorsFromGroup(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid group ID", http.StatusBadRequest)
		return
	}

	var req AddSensorsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if len(req.SensorIDs) == 0 {
		http.Error(w, "At least one sensor ID is required", http.StatusBadRequest)
		return
	}

	res, err := h.client.RemoveSensorsFromGroup(ctx, &pb.RemoveSensorsFromGroupRequest{
		GroupId:   id,
		SensorIds: req.SensorIDs,
	})
	if err != nil {
		http.Error(w, "Failed to remove sensors: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res.Group)
}
