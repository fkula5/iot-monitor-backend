package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

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
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	res, err := h.client.ListAlertRules(ctx, &pb.ListAlertRulesRequest{UserId: int64(claims.UserId)})
	if err != nil {
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
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req types.AlertRuleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	res, err := h.client.CreateAlertRule(ctx, &pb.CreateAlertRuleRequest{
		UserId:        int64(claims.UserId),
		SensorId:      req.SensorID,
		ConditionType: req.Condition_Type,
		Threshold:     req.Threshold,
		Description:   req.Description,
	})
	if err != nil {
		http.Error(w, "Failed to create alert rule", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(types.MapAlertRuleFromProto(res.AlertRule))
}

// @Summary Delete Alert Rule
// @Description Delete an alert rule by ID
// @Tags Alert Rules
// @Accept json
// @Produce json
// @Param deleteRequest body types.DeleteAlertRuleRequest true "Delete Alert Rule Request"
// @Success 204 {string} string "No Content"
// @Failure 400 {string} string "Bad Request"
// @Failure 401 {string} string "Unauthorized"
// @Failure 500 {string} string "Internal Server Error"
// @Router /api/alert-rules/{id} [delete]
func (h *AlertRuleHandler) DeleteAlertRule(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	_, ok := authMiddleware.GetUserFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req types.DeleteAlertRuleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	_, err := h.client.DeleteAlertRule(ctx, &pb.DeleteAlertRuleRequest{
		Id: req.AlertRuleID,
	})
	if err != nil {
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
// @Param updateRequest body types.UpdateAlertRuleRequest true "Update Alert Rule Request"
// @Success 200 {object} types.AlertRuleResponse
// @Failure 400 {string} string "Bad Request"
// @Failure 401 {string} string "Unauthorized"
// @Failure 500 {string} string "Internal Server Error"
// @Router /api/alert-rules/{id} [put]
func (h *AlertRuleHandler) UpdateAlertRule(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	_, ok := authMiddleware.GetUserFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req types.UpdateAlertRuleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	res, err := h.client.UpdateAlertRule(ctx, &pb.UpdateAlertRuleRequest{
		Id:            req.AlertRuleID,
		SensorId:      req.SensorID,
		ConditionType: req.Condition_Type,
		Threshold:     req.Threshold,
		Description:   req.Description,
	})
	if err != nil {
		http.Error(w, "Failed to update alert rule", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(types.MapAlertRuleFromProto(res.AlertRule))
}

// @Summary Get Alert Rule
// @Description Get an alert rule by ID
// @Tags Alert Rules
// @Accept json
// @Produce json
// @Param getRequest body types.GetAlertRulesRequest true "Get Alert Rule Request"
// @Success 200 {object} types.AlertRuleResponse
// @Failure 400 {string} string "Bad Request"
// @Failure 401 {string} string "Unauthorized"
// @Failure 500 {string} string "Internal Server Error"
// @Router /api/alert-rules/{id} [get]
func (h *AlertRuleHandler) GetAlertRule(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	_, ok := authMiddleware.GetUserFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req types.GetAlertRulesRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	res, err := h.client.GetAlertRule(ctx, &pb.GetAlertRuleRequest{
		Id: req.Id,
	})
	if err != nil {
		http.Error(w, "Failed to get alert rule", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(types.MapAlertRuleFromProto(res.AlertRule))
}
