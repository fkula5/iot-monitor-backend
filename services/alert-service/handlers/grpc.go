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
	a, err := h.alertService.GetAlert(ctx, int(req.Id))
	if err != nil {
		return nil, err
	}
	return &pb.GetAlertResponse{Alert: h.mapAlert(a)}, nil
}

func (h *AlertGrpcHandler) ListAlerts(ctx context.Context, req *pb.ListAlertsRequest) (*pb.ListAlertsResponse, error) {
	alerts, err := h.alertService.ListAlerts(ctx, req.UserId)
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
	success, err := h.alertService.MarkAsRead(ctx, int(req.Id))
	if err != nil {
		return nil, err
	}
	return &pb.MarkAlertAsReadResponse{Success: success}, nil
}

func (h *AlertGrpcHandler) ListAlertRules(ctx context.Context, req *pb.ListAlertRulesRequest) (*pb.ListAlertRulesResponse, error) {
	rules, err := h.alertRuleService.ListAlertRules(ctx, req.UserId)
	if err != nil {
		return nil, err
	}
	res := make([]*pb.AlertRule, len(rules))
	for i, r := range rules {
		res[i] = &pb.AlertRule{
			Id:            int64(r.ID),
			SensorId:      r.SensorID,
			ConditionType: r.ConditionType,
			Threshold:     r.Threshold,
		}
	}
	return &pb.ListAlertRulesResponse{AlertRules: res}, nil
}

func (h *AlertGrpcHandler) CreateAlertRule(ctx context.Context, req *pb.CreateAlertRuleRequest) (*pb.CreateAlertRuleResponse, error) {
	rule, err := h.alertRuleService.CreateAlertRule(ctx, &ent.AlertRule{
		SensorID:      req.SensorId,
		ConditionType: req.ConditionType,
		Threshold:     req.Threshold,
		Description:   req.Description,
		UserID:        req.UserId,
	})
	if err != nil {
		return nil, err
	}
	return &pb.CreateAlertRuleResponse{
		AlertRule: &pb.AlertRule{
			Id:            int64(rule.ID),
			SensorId:      rule.SensorID,
			ConditionType: rule.ConditionType,
			Threshold:     rule.Threshold,
		},
	}, nil
}

func (h *AlertGrpcHandler) DeleteAlertRule(ctx context.Context, req *pb.DeleteAlertRuleRequest) (*pb.DeleteAlertRuleResponse, error) {
	err := h.alertRuleService.DeleteAlertRule(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	return &pb.DeleteAlertRuleResponse{}, nil
}

func (h *AlertGrpcHandler) GetAlertRule(ctx context.Context, req *pb.GetAlertRuleRequest) (*pb.GetAlertRuleResponse, error) {
	rule, err := h.alertRuleService.GetAlertRule(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	return &pb.GetAlertRuleResponse{
		AlertRule: &pb.AlertRule{
			Id:            int64(rule.ID),
			SensorId:      rule.SensorID,
			ConditionType: rule.ConditionType,
			Threshold:     rule.Threshold,
		},
	}, nil
}

func (h *AlertGrpcHandler) UpdateAlertRule(ctx context.Context, req *pb.UpdateAlertRuleRequest) (*pb.UpdateAlertRuleResponse, error) {
	rule, err := h.alertRuleService.UpdateAlertRule(ctx, &ent.AlertRule{
		ID:            int(req.Id),
		SensorID:      req.SensorId,
		ConditionType: req.ConditionType,
		Threshold:     req.Threshold,
		Description:   req.Description,
		IsEnabled:     req.IsEnabled,
	})

	if err != nil {
		return nil, err
	}
	return &pb.UpdateAlertRuleResponse{
		AlertRule: &pb.AlertRule{
			Id:            int64(rule.ID),
			SensorId:      rule.SensorID,
			ConditionType: rule.ConditionType,
			Threshold:     rule.Threshold,
		},
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
