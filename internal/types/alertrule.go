package types

import (
	"time"

	pb "github.com/skni-kod/iot-monitor-backend/internal/proto/alert_service"
)

type AlertRuleResponse struct {
	ID             int64     `json:"id"`
	Name           string    `json:"name"`
	UserID         int64     `json:"user_id"`
	SensorID       int64     `json:"sensor_id"`
	Condition_Type string    `json:"condition_type"`
	Threshold      float64   `json:"threshold"`
	Description    string    `json:"description"`
	IsEnabled      bool      `json:"is_enabled"`
	CreatedAt      time.Time `json:"created_at"`
}

type AlertRuleRequest struct {
	Name           string  `json:"name"`
	SensorID       int64   `json:"sensor_id"`
	Condition_Type string  `json:"condition_type"`
	Threshold      float64 `json:"threshold"`
	Description    string  `json:"description"`
}

type UpdateAlertRuleRequest struct {
	ID             int64   `json:"id"`
	Name           string  `json:"name"`
	SensorID       int64   `json:"sensor_id"`
	Condition_Type string  `json:"condition_type"`
	Threshold      float64 `json:"threshold"`
	Description    string  `json:"description"`
	IsEnabled      bool    `json:"is_enabled"`
}

type DeleteAlertRuleRequest struct {
	Id int64 `json:"id"`
}

type GetAlertRulesRequest struct {
	Id int64 `json:"id"`
}

func MapAlertRuleFromProto(r *pb.AlertRule) AlertRuleResponse {
	return AlertRuleResponse{
		ID:             r.Id,
		Name:           r.Name,
		UserID:         r.UserId,
		SensorID:       r.SensorId,
		Condition_Type: r.ConditionType,
		Threshold:      r.Threshold,
		Description:    r.Description,
		IsEnabled:      r.IsEnabled,
		CreatedAt:      r.CreatedAt.AsTime(),
	}
}
