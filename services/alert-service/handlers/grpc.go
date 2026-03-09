package handlers

import (
	"context"

	"google.golang.org/protobuf/types/known/timestamppb"

	pb "github.com/skni-kod/iot-monitor-backend/internal/proto/alert_service"
	"github.com/skni-kod/iot-monitor-backend/services/alert-service/ent"
	"github.com/skni-kod/iot-monitor-backend/services/alert-service/service"
)

type AlertGrpcHandler struct {
	pb.UnimplementedAlertServiceServer
	service *service.AlertService
}

func NewAlertGrpcHandler(s *service.AlertService) *AlertGrpcHandler {
	return &AlertGrpcHandler{service: s}
}

func (h *AlertGrpcHandler) GetAlert(ctx context.Context, req *pb.GetAlertRequest) (*pb.GetAlertResponse, error) {
	a, err := h.service.GetAlert(ctx, int(req.Id))
	if err != nil {
		return nil, err
	}
	return &pb.GetAlertResponse{Alert: h.mapAlert(a)}, nil
}

func (h *AlertGrpcHandler) ListAlerts(ctx context.Context, req *pb.ListAlertsRequest) (*pb.ListAlertsResponse, error) {
	alerts, err := h.service.ListAlerts(ctx, req.UserId)
	if err != nil {
		return nil, err
	}
	res := make([]*pb.Alert, len(alerts))
	for i, a := range alerts {
		res[i] = h.mapAlert(a)
	}
	return &pb.ListAlertsResponse{Alerts: res}, nil
}

func (h *AlertGrpcHandler) MarkAlertAsRead(ctx context.Context, req *pb.MarkAlertAsReadRequest) (*pb.MarkAlertAsReadResponse, error) {
	success, err := h.service.MarkAsRead(ctx, int(req.Id))
	if err != nil {
		return nil, err
	}
	return &pb.MarkAlertAsReadResponse{Success: success}, nil
}

func (h *AlertGrpcHandler) mapAlert(a *ent.Alert) *pb.Alert {
	return &pb.Alert{
		Id:          int64(a.ID),
		RuleId:      int64(a.Edges.Rule.ID),
		SensorId:    a.Edges.Rule.SensorID,
		Message:     a.Message,
		Value:       a.Value,
		IsRead:      a.IsRead,
		TriggeredAt: timestamppb.New(a.TriggeredAt),
	}
}
