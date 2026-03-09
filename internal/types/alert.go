package types

import (
	"time"

	pb "github.com/skni-kod/iot-monitor-backend/internal/proto/alert_service"
)

type AlertResponse struct {
	ID          int64     `json:"id"`
	RuleID      int64     `json:"rule_id"`
	SensorID    int64     `json:"sensor_id"`
	Message     string    `json:"message"`
	Value       float64   `json:"value"`
	IsRead      bool      `json:"is_read"`
	TriggeredAt time.Time `json:"triggered_at"`
}

func MapAlertFromProto(a *pb.Alert) AlertResponse {
	return AlertResponse{
		ID:          a.Id,
		RuleID:      a.RuleId,
		SensorID:    a.SensorId,
		Message:     a.Message,
		Value:       a.Value,
		IsRead:      a.IsRead,
		TriggeredAt: a.TriggeredAt.AsTime(),
	}
}
