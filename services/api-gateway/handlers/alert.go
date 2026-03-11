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

type AlertHandler struct {
	client pb.AlertServiceClient
}

func NewAlertHandler(client pb.AlertServiceClient) *AlertHandler {
	return &AlertHandler{client: client}
}

// @Summary ListAlerts retrieves a list of all alerts for the authenticated user.
// @Description Fetches all alerts from the Alert Service.
// @Tags Alerts
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {array} types.AlertResponse "List of alerts"
// @Failure 401 {string} string "Unauthorized"
// @Failure 500 {string} string "Internal Server Error"
// @Router /api/alerts [get]
func (h *AlertHandler) ListAlerts(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	claims, ok := authMiddleware.GetUserFromContext(r.Context())
	if !ok {
		logger.Warn("Unauthorized access attempt to ListAlerts")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	res, err := h.client.ListAlerts(ctx, &pb.ListAlertsRequest{UserId: int64(claims.UserId)})
	if err != nil {
		logger.Error("Failed to fetch alerts from alert service", zap.Error(err), zap.Int("userId", claims.UserId))
		http.Error(w, "Failed to fetch alerts", http.StatusInternalServerError)
		return
	}

	alertResponses := make([]types.AlertResponse, 0, len(res.Alerts))
	for _, a := range res.Alerts {
		alertResponses = append(alertResponses, types.MapAlertFromProto(a))
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(alertResponses)
}

// @Summary MarkAlertAsRead marks an alert as read.
// @Description Updates the alert status to read in the Alert Service.
// @Tags Alerts
// @Produce json
// @Param id path int true "Alert ID"
// @Security ApiKeyAuth
// @Success 200 {object} bool "Success status"
// @Failure 400 {string} string "Bad Request"
// @Failure 401 {string} string "Unauthorized"
// @Failure 500 {string} string "Internal Server Error"
// @Router /api/alerts/{id}/read [post]
func (h *AlertHandler) MarkAlertAsRead(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		logger.Warn("Invalid alert ID in MarkAlertAsRead", zap.String("id", idStr))
		http.Error(w, "Invalid alert ID", http.StatusBadRequest)
		return
	}

	claims, ok := authMiddleware.GetUserFromContext(r.Context())
	if !ok {
		logger.Warn("Unauthorized access attempt to MarkAlertAsRead", zap.Int("alertId", id))
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	res, err := h.client.MarkAlertAsRead(ctx, &pb.MarkAlertAsReadRequest{Id: int64(id)})
	if err != nil {
		logger.Error("Failed to mark alert as read", zap.Error(err), zap.Int("alertId", id), zap.Int("userId", claims.UserId))
		http.Error(w, "Failed to mark alert as read", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res.Success)
}
