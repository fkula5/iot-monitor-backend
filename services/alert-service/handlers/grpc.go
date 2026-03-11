package handlers

import (
	"context"

	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/timestamppb"

	pb "github.com/skni-kod/iot-monitor-backend/internal/proto/alert_service"
	"github.com/skni-kod/iot-monitor-backend/pkg/logger"
	"github.com/skni-kod/iot-monitor-backend/services/alert-service/ent"
	"github.com/skni-kod/iot-monitor-backend/services/alert-service/service"
)

type AlertGrpcHandler struct {
	pb.UnimplementedAlertServiceServer
	alertService     *service.AlertService
	alertRuleService *service.AlertRuleService
}

func NewAlertGrpcHandler(alertService *service.AlertService, alertRuleService *service.AlertRuleService) *AlertGrpcHandler {
	return &AlertGrpcHandler{
		alertService:     alertService,
		alertRuleService: alertRuleService,
	}
}

func (h *AlertGrpcHandler) GetAlert(ctx context.Context, req *pb.GetAlertRequest) (*pb.GetAlertResponse, error) {
	logger.Info("gRPC GetAlert", zap.Int64("id", req.Id))
	a, err := h.alertService.GetAlert(ctx, int(req.Id))
	if err != nil {
		logger.Error("Failed to get alert", zap.Error(err), zap.Int64("id", req.Id))
		return nil, err
	}
	return &pb.GetAlertResponse{Alert: h.mapAlert(a)}, nil
}

func (h *AlertGrpcHandler) ListAlerts(ctx context.Context, req *pb.ListAlertsRequest) (*pb.ListAlertsResponse, error) {
	logger.Info("gRPC ListAlerts", zap.Int64("userId", req.UserId))
	alerts, err := h.alertService.ListAlerts(ctx, req.UserId)
	if err != nil {
		logger.Error("Failed to list alerts", zap.Error(err), zap.Int64("userId", req.UserId))
		return nil, err
	}
	res := make([]*pb.Alert, len(alerts))
	for i, a := range alerts {
		res[i] = h.mapAlert(a)
	}
	return &pb.ListAlertsResponse{Alerts: res}, nil
}

func (h *AlertGrpcHandler) MarkAlertAsRead(ctx context.Context, req *pb.MarkAlertAsReadRequest) (*pb.MarkAlertAsReadResponse, error) {
	logger.Info("gRPC MarkAlertAsRead", zap.Int64("id", req.Id))
	success, err := h.alertService.MarkAsRead(ctx, int(req.Id))
	if err != nil {
		logger.Error("Failed to mark alert as read", zap.Error(err), zap.Int64("id", req.Id))
		return nil, err
	}
	return &pb.MarkAlertAsReadResponse{Success: success}, nil
}

func (h *AlertGrpcHandler) ListAlertRules(ctx context.Context, req *pb.ListAlertRulesRequest) (*pb.ListAlertRulesResponse, error) {
	logger.Info("gRPC ListAlertRules", zap.Int64("userId", req.UserId))
	rules, err := h.alertRuleService.ListAlertRules(ctx, req.UserId)
	if err != nil {
		logger.Error("Failed to list alert rules", zap.Error(err), zap.Int64("userId", req.UserId))
		return nil, err
	}
	res := make([]*pb.AlertRule, len(rules))
	for i, r := range rules {
		res[i] = h.mapAlertRule(r)
	}
	return &pb.ListAlertRulesResponse{AlertRules: res}, nil
}

func (h *AlertGrpcHandler) CreateAlertRule(ctx context.Context, req *pb.CreateAlertRuleRequest) (*pb.CreateAlertRuleResponse, error) {
	logger.Info("gRPC CreateAlertRule", zap.Int64("userId", req.UserId), zap.Int64("sensorId", req.SensorId))
	rule, err := h.alertRuleService.CreateAlertRule(ctx, &ent.AlertRule{
		Name:          req.Name,
		SensorID:      req.SensorId,
		ConditionType: req.ConditionType,
		Threshold:     req.Threshold,
		Description:   req.Description,
		UserID:        req.UserId,
	})
	if err != nil {
		logger.Error("Failed to create alert rule", zap.Error(err), zap.Int64("userId", req.UserId), zap.Int64("sensorId", req.SensorId))
		return nil, err
	}
	return &pb.CreateAlertRuleResponse{
		AlertRule: h.mapAlertRule(rule),
	}, nil
}

func (h *AlertGrpcHandler) DeleteAlertRule(ctx context.Context, req *pb.DeleteAlertRuleRequest) (*pb.DeleteAlertRuleResponse, error) {
	logger.Info("gRPC DeleteAlertRule", zap.Int64("id", req.Id))
	err := h.alertRuleService.DeleteAlertRule(ctx, req.Id)
	if err != nil {
		logger.Error("Failed to delete alert rule", zap.Error(err), zap.Int64("id", req.Id))
		return nil, err
	}
	return &pb.DeleteAlertRuleResponse{}, nil
}

func (h *AlertGrpcHandler) GetAlertRule(ctx context.Context, req *pb.GetAlertRuleRequest) (*pb.GetAlertRuleResponse, error) {
	logger.Info("gRPC GetAlertRule", zap.Int64("id", req.Id))
	rule, err := h.alertRuleService.GetAlertRule(ctx, req.Id)
	if err != nil {
		logger.Error("Failed to get alert rule", zap.Error(err), zap.Int64("id", req.Id))
		return nil, err
	}
	return &pb.GetAlertRuleResponse{
		AlertRule: h.mapAlertRule(rule),
	}, nil
}

func (h *AlertGrpcHandler) UpdateAlertRule(ctx context.Context, req *pb.UpdateAlertRuleRequest) (*pb.UpdateAlertRuleResponse, error) {
	logger.Info("gRPC UpdateAlertRule", zap.Int64("id", req.Id))
	rule, err := h.alertRuleService.UpdateAlertRule(ctx, &ent.AlertRule{
		ID:            int(req.Id),
		Name:          req.Name,
		SensorID:      req.SensorId,
		ConditionType: req.ConditionType,
		Threshold:     req.Threshold,
		Description:   req.Description,
		IsEnabled:     req.IsEnabled,
	})

	if err != nil {
		logger.Error("Failed to update alert rule", zap.Error(err), zap.Int64("id", req.Id))
		return nil, err
	}
	return &pb.UpdateAlertRuleResponse{
		AlertRule: h.mapAlertRule(rule),
	}, nil
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

func (h *AlertGrpcHandler) mapAlertRule(r *ent.AlertRule) *pb.AlertRule {
	return &pb.AlertRule{
		Id:            int64(r.ID),
		Name:          r.Name,
		SensorId:      r.SensorID,
		ConditionType: r.ConditionType,
		Threshold:     r.Threshold,
		Description:   r.Description,
		IsEnabled:     r.IsEnabled,
		CreatedAt:     timestamppb.New(r.CreatedAt),
		UserId:        r.UserID,
	}
}
