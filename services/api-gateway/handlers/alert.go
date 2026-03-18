package handlers

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"

	pb "github.com/skni-kod/iot-monitor-backend/internal/proto/alert_service"
	"github.com/skni-kod/iot-monitor-backend/internal/types"
	authMiddleware "github.com/skni-kod/iot-monitor-backend/services/api-gateway/middleware"
)

type AlertHandler struct {
	client pb.AlertServiceClient
}

func NewAlertHandler(client pb.AlertServiceClient) *AlertHandler {
	return &AlertHandler{client: client}
}

// @Summary ListAlerts retrieves a list of all alerts for the authenticated user.
// @Description Fetches all alerts from the Alert Service with pagination support.
// @Tags Alerts
// @Produce json
// @Security ApiKeyAuth
// @Param page query int false "Page number (default 1)"
// @Param limit query int false "Items per page (default 10)"
// @Success 200 {object} Response{data=types.PaginatedAlertResponse}
// @Failure 401 {object} Response{error=string}
// @Failure 500 {object} Response{error=string}
// @Router /api/alerts [get]
func (h *AlertHandler) ListAlerts(w http.ResponseWriter, r *http.Request) {
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

	res, err := h.client.ListAlerts(ctx, &pb.ListAlertsRequest{
		UserId: int64(claims.UserId),
		Limit:  int32(limit),
		Offset: int32((page - 1) * limit),
	})
	if err != nil {
		Error(w, http.StatusInternalServerError, "Failed to fetch alerts")
		return
	}

	alertResponses := make([]types.AlertResponse, 0, len(res.Alerts))
	for _, a := range res.Alerts {
		alertResponses = append(alertResponses, types.MapAlertFromProto(a))
	}

	JSON(w, http.StatusOK, types.PaginatedAlertResponse{
		Alerts:     alertResponses,
		TotalCount: res.TotalCount,
		Page:       page,
		Limit:      limit,
	})
}

// @Summary MarkAlertAsRead marks an alert as read.
// @Description Updates the alert status to read in the Alert Service.
// @Tags Alerts
// @Produce json
// @Param id path int true "Alert ID"
// @Security ApiKeyAuth
// @Success 200 {object} Response{data=bool}
// @Failure 400 {object} Response{error=string}
// @Failure 401 {object} Response{error=string}
// @Failure 500 {object} Response{error=string}
// @Router /api/alerts/{id}/read [post]
func (h *AlertHandler) MarkAlertAsRead(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		Error(w, http.StatusBadRequest, "Invalid alert ID")
		return
	}

	res, err := h.client.MarkAlertAsRead(ctx, &pb.MarkAlertAsReadRequest{Id: int64(id)})
	if err != nil {
		Error(w, http.StatusInternalServerError, "Failed to mark alert as read")
		return
	}

	JSON(w, http.StatusOK, res.Success)
}
