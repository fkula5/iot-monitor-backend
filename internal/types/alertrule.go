package types

import pb "github.com/skni-kod/iot-monitor-backend/internal/proto/alert_service"

type AlertRuleResponse struct {
	ID             int64   `json:"id"`
	UserID         int64   `json:"user_id"`
	SensorID       int64   `json:"sensor_id"`
	Condition_Type string  `json:"condition_type"`
	Threshold      float64 `json:"threshold"`
	Description    string  `json:"description"`
}

type AlertRuleRequest struct {
	SensorID       int64   `json:"sensor_id"`
	Condition_Type string  `json:"condition_type"`
	Threshold      float64 `json:"threshold"`
	Description    string  `json:"description"`
}

type UpdateAlertRuleRequest struct {
	AlertRuleID    int64   `json:"alert_rule_id"`
	SensorID       int64   `json:"sensor_id"`
	Condition_Type string  `json:"condition_type"`
	Threshold      float64 `json:"threshold"`
	Description    string  `json:"description"`
}

type DeleteAlertRuleRequest struct {
	AlertRuleID int64 `json:"alert_rule_id"`
}

type GetAlertRulesRequest struct {
	Id int64 `json:"id"`
}

func MapAlertRuleFromProto(r *pb.AlertRule) AlertRuleResponse {
	return AlertRuleResponse{
		ID:             r.Id,
		UserID:         r.UserId,
		SensorID:       r.SensorId,
		Condition_Type: r.ConditionType,
		Threshold:      r.Threshold,
		Description:    r.Description,
	}
}
