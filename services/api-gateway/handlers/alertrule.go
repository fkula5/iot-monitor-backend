package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"

	pb "github.com/skni-kod/iot-monitor-backend/internal/proto/alert_service"
	"github.com/skni-kod/iot-monitor-backend/internal/types"
	"github.com/skni-kod/iot-monitor-backend/pkg/logger"
	authMiddleware "github.com/skni-kod/iot-monitor-backend/services/api-gateway/middleware"
)

type AlertRuleHandler struct {
	client pb.AlertServiceClient
}

func NewAlertRuleHandler(client pb.AlertServiceClient) *AlertRuleHandler {
	return &AlertRuleHandler{client: client}
}

// @Summary List Alert Rules
// @Description Get a list of alert rules for the authenticated user
// @Tags Alert Rules
// @Accept json
// @Produce json
// @Success 200 {array} types.AlertRuleResponse
// @Failure 401 {string} string "Unauthorized"
// @Failure 500 {string} string "Internal Server Error"
// @Router /api/alert-rules [get]
func (h *AlertRuleHandler) ListAlertRules(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	claims, ok := authMiddleware.GetUserFromContext(r.Context())
	if !ok {
		logger.Warn("Unauthorized access attempt to ListAlertRules")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	res, err := h.client.ListAlertRules(ctx, &pb.ListAlertRulesRequest{UserId: int64(claims.UserId)})
	if err != nil {
		logger.Error("Failed to list alert rules from alert service", zap.Error(err), zap.Int("userId", claims.UserId))
		http.Error(w, "Failed to list alert rules", http.StatusInternalServerError)
		return
	}

	alerRulesResponse := make([]types.AlertRuleResponse, 0, len(res.AlertRules))
	for _, r := range res.AlertRules {
		alerRulesResponse = append(alerRulesResponse, types.MapAlertRuleFromProto(r))
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(alerRulesResponse)
}

// @Summary Create Alert Rule
// @Description Create a new alert rule for the authenticated user
// @Tags Alert Rules
// @Accept json
// @Produce json
// @Param alertRule body types.AlertRuleRequest true "Alert Rule Request"
// @Success 200 {object} types.AlertRuleResponse
// @Failure 400 {string} string "Bad Request"
// @Failure 401 {string} string "Unauthorized"
// @Failure 500 {string} string "Internal Server Error"
// @Router /api/alert-rules [post]
func (h *AlertRuleHandler) CreateAlertRule(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	claims, ok := authMiddleware.GetUserFromContext(r.Context())
	if !ok {
		logger.Warn("Unauthorized access attempt to CreateAlertRule")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req types.AlertRuleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Warn("Invalid request body in CreateAlertRule", zap.Error(err))
		http.Error(w, "Invalid request body", http.StatusBadRequest)
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
		logger.Error("Failed to create alert rule in alert service", zap.Error(err), zap.Int("userId", claims.UserId))
		http.Error(w, "Failed to create alert rule", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(types.MapAlertRuleFromProto(res.AlertRule))
}

// @Summary Delete Alert Rule
// @Description Delete an alert rule by ID
// @Tags Alert Rules
// @Param id path int true "Alert Rule ID"
// @Success 204 {string} string "No Content"
// @Failure 400 {string} string "Bad Request"
// @Failure 401 {string} string "Unauthorized"
// @Failure 500 {string} string "Internal Server Error"
// @Router /api/alert-rules/{id} [delete]
func (h *AlertRuleHandler) DeleteAlertRule(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	claims, ok := authMiddleware.GetUserFromContext(r.Context())
	if !ok {
		logger.Warn("Unauthorized access attempt to DeleteAlertRule")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		logger.Warn("Invalid alert rule ID in DeleteAlertRule request", zap.String("id", idStr))
		http.Error(w, "Invalid alert rule ID", http.StatusBadRequest)
		return
	}

	_, err = h.client.DeleteAlertRule(ctx, &pb.DeleteAlertRuleRequest{
		Id: id,
	})
	if err != nil {
		logger.Error("Failed to delete alert rule in alert service", zap.Error(err), zap.Int64("ruleId", id), zap.Int("userId", claims.UserId))
		http.Error(w, "Failed to delete alert rule", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// @Summary Update Alert Rule
// @Description Update an existing alert rule
// @Tags Alert Rules
// @Accept json
// @Produce json
// @Param id path int true "Alert Rule ID"
// @Param updateRequest body types.UpdateAlertRuleRequest true "Update Alert Rule Request"
// @Success 200 {object} types.AlertRuleResponse
// @Failure 400 {string} string "Bad Request"
// @Failure 401 {string} string "Unauthorized"
// @Failure 500 {string} string "Internal Server Error"
// @Router /api/alert-rules/{id} [put]
func (h *AlertRuleHandler) UpdateAlertRule(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	claims, ok := authMiddleware.GetUserFromContext(r.Context())
	if !ok {
		logger.Warn("Unauthorized access attempt to UpdateAlertRule")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		logger.Warn("Invalid alert rule ID in UpdateAlertRule request", zap.String("id", idStr))
		http.Error(w, "Invalid alert rule ID", http.StatusBadRequest)
		return
	}

	var req types.UpdateAlertRuleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Warn("Invalid request body in UpdateAlertRule", zap.Error(err))
		http.Error(w, "Invalid request body", http.StatusBadRequest)
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
		logger.Error("Failed to update alert rule in alert service", zap.Error(err), zap.Int64("ruleId", id), zap.Int("userId", claims.UserId))
		http.Error(w, "Failed to update alert rule", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(types.MapAlertRuleFromProto(res.AlertRule))
}

// @Summary Get Alert Rule
// @Description Get an alert rule by ID
// @Tags Alert Rules
// @Param id path int true "Alert Rule ID"
// @Success 200 {object} types.AlertRuleResponse
// @Failure 400 {string} string "Bad Request"
// @Failure 401 {string} string "Unauthorized"
// @Failure 500 {string} string "Internal Server Error"
// @Router /api/alert-rules/{id} [get]
func (h *AlertRuleHandler) GetAlertRule(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	claims, ok := authMiddleware.GetUserFromContext(r.Context())
	if !ok {
		logger.Warn("Unauthorized access attempt to GetAlertRule")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		logger.Warn("Invalid alert rule ID in GetAlertRule request", zap.String("id", idStr))
		http.Error(w, "Invalid alert rule ID", http.StatusBadRequest)
		return
	}

	res, err := h.client.GetAlertRule(ctx, &pb.GetAlertRuleRequest{
		Id: id,
	})
	if err != nil {
		logger.Error("Failed to get alert rule from alert service", zap.Error(err), zap.Int64("ruleId", id), zap.Int("userId", claims.UserId))
		http.Error(w, "Failed to get alert rule", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(types.MapAlertRuleFromProto(res.AlertRule))
}
