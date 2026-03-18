package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"

	pb "github.com/skni-kod/iot-monitor-backend/internal/proto/alert_service"
	"github.com/skni-kod/iot-monitor-backend/internal/types"
	authMiddleware "github.com/skni-kod/iot-monitor-backend/services/api-gateway/middleware"
)

type AlertRuleHandler struct {
	client pb.AlertServiceClient
}

func NewAlertRuleHandler(client pb.AlertServiceClient) *AlertRuleHandler {
	return &AlertRuleHandler{client: client}
}

// @Summary List Alert Rules
// @Description Get a list of alert rules for the authenticated user with pagination support
// @Tags Alert Rules
// @Accept json
// @Produce json
// @Param page query int false "Page number (default 1)"
// @Param limit query int false "Items per page (default 10)"
// @Success 200 {object} Response{data=types.PaginatedAlertRuleResponse}
// @Failure 401 {object} Response{error=string}
// @Failure 500 {object} Response{error=string}
// @Router /api/alert-rules [get]
func (h *AlertRuleHandler) ListAlertRules(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	claims, ok := authMiddleware.GetUserFromContext(r.Context())
	if !ok {
		Error(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	page := 1
	limit := 10

	if p := r.URL.Query().Get("page"); p != "" {
		if val, err := strconv.Atoi(p); err == nil && val > 0 {
			page = val
		}
	}

	if l := r.URL.Query().Get("limit"); l != "" {
		if val, err := strconv.Atoi(l); err == nil && val > 0 {
			limit = val
		}
	}

	res, err := h.client.ListAlertRules(ctx, &pb.ListAlertRulesRequest{
		UserId: int64(claims.UserId),
		Limit:  int32(limit),
		Offset: int32((page - 1) * limit),
	})
	if err != nil {
		Error(w, http.StatusInternalServerError, "Failed to list alert rules")
		return
	}

	alertRulesResponse := make([]types.AlertRuleResponse, 0, len(res.AlertRules))
	for _, r := range res.AlertRules {
		alertRulesResponse = append(alertRulesResponse, types.MapAlertRuleFromProto(r))
	}

	JSON(w, http.StatusOK, types.PaginatedAlertRuleResponse{
		AlertRules: alertRulesResponse,
		TotalCount: res.TotalCount,
		Page:       page,
		Limit:      limit,
	})
}

// @Summary Create Alert Rule
// @Description Create a new alert rule for the authenticated user
// @Tags Alert Rules
// @Accept json
// @Produce json
// @Param alertRule body types.AlertRuleRequest true "Alert Rule Request"
// @Success 200 {object} Response{data=types.AlertRuleResponse}
// @Failure 400 {object} Response{error=string}
// @Failure 401 {object} Response{error=string}
// @Failure 500 {object} Response{error=string}
// @Router /api/alert-rules [post]
func (h *AlertRuleHandler) CreateAlertRule(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	claims, ok := authMiddleware.GetUserFromContext(r.Context())
	if !ok {
		Error(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req types.AlertRuleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		Error(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	res, err := h.client.CreateAlertRule(ctx, &pb.CreateAlertRuleRequest{
		Name:          req.Name,
		UserId:        int64(claims.UserId),
		SensorId:      req.SensorID,
		ConditionType: req.Condition_Type,
		Threshold:     req.Threshold,
		Description:   req.Description,
	})
	if err != nil {
		Error(w, http.StatusInternalServerError, "Failed to create alert rule")
		return
	}

	JSON(w, http.StatusOK, types.MapAlertRuleFromProto(res.AlertRule))
}

// @Summary Delete Alert Rule
// @Description Delete an alert rule by ID
// @Tags Alert Rules
// @Param id path int true "Alert Rule ID"
// @Success 204 {object} Response
// @Failure 400 {object} Response{error=string}
// @Failure 401 {object} Response{error=string}
// @Failure 500 {object} Response{error=string}
// @Router /api/alert-rules/{id} [delete]
func (h *AlertRuleHandler) DeleteAlertRule(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	if _, ok := authMiddleware.GetUserFromContext(r.Context()); !ok {
		Error(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		Error(w, http.StatusBadRequest, "Invalid alert rule ID")
		return
	}

	if _, err = h.client.DeleteAlertRule(ctx, &pb.DeleteAlertRuleRequest{Id: id}); err != nil {
		Error(w, http.StatusInternalServerError, "Failed to delete alert rule")
		return
	}

	NoContent(w)
}

// @Summary Update Alert Rule
// @Description Update an existing alert rule
// @Tags Alert Rules
// @Accept json
// @Produce json
// @Param id path int true "Alert Rule ID"
// @Param updateRequest body types.UpdateAlertRuleRequest true "Update Alert Rule Request"
// @Success 200 {object} Response{data=types.AlertRuleResponse}
// @Failure 400 {object} Response{error=string}
// @Failure 401 {object} Response{error=string}
// @Failure 500 {object} Response{error=string}
// @Router /api/alert-rules/{id} [put]
func (h *AlertRuleHandler) UpdateAlertRule(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	if _, ok := authMiddleware.GetUserFromContext(r.Context()); !ok {
		Error(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		Error(w, http.StatusBadRequest, "Invalid alert rule ID")
		return
	}

	var req types.UpdateAlertRuleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		Error(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	res, err := h.client.UpdateAlertRule(ctx, &pb.UpdateAlertRuleRequest{
		Id:            id,
		Name:          req.Name,
		SensorId:      req.SensorID,
		ConditionType: req.Condition_Type,
		Threshold:     req.Threshold,
		Description:   req.Description,
		IsEnabled:     req.IsEnabled,
	})
	if err != nil {
		Error(w, http.StatusInternalServerError, "Failed to update alert rule")
		return
	}

	JSON(w, http.StatusOK, types.MapAlertRuleFromProto(res.AlertRule))
}

// @Summary Get Alert Rule
// @Description Get an alert rule by ID
// @Tags Alert Rules
// @Param id path int true "Alert Rule ID"
// @Success 200 {object} Response{data=types.AlertRuleResponse}
// @Failure 400 {object} Response{error=string}
// @Failure 401 {object} Response{error=string}
// @Failure 500 {object} Response{error=string}
// @Router /api/alert-rules/{id} [get]
func (h *AlertRuleHandler) GetAlertRule(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	if _, ok := authMiddleware.GetUserFromContext(r.Context()); !ok {
		Error(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		Error(w, http.StatusBadRequest, "Invalid alert rule ID")
		return
	}

	res, err := h.client.GetAlertRule(ctx, &pb.GetAlertRuleRequest{Id: id})
	if err != nil {
		Error(w, http.StatusInternalServerError, "Failed to get alert rule")
		return
	}

	JSON(w, http.StatusOK, types.MapAlertRuleFromProto(res.AlertRule))
}
