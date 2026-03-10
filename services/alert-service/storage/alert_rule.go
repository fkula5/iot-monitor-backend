package storage

import (
	"context"

	"github.com/skni-kod/iot-monitor-backend/services/alert-service/ent"
	"github.com/skni-kod/iot-monitor-backend/services/alert-service/ent/alertrule"
)

type IAlertRuleStorage interface {
	Create(ctx context.Context, userID int64, rule *ent.AlertRule) (*ent.AlertRule, error)
	Get(ctx context.Context, id int64) (*ent.AlertRule, error)
	List(ctx context.Context, userID int64) ([]*ent.AlertRule, error)
	Update(ctx context.Context, id int64, rule *ent.AlertRule) (*ent.AlertRule, error)
	Delete(ctx context.Context, id int64) (bool, error)
}

type AlertRuleStorage struct {
	client *ent.Client
}

func NewAlertRuleStorage(client *ent.Client) IAlertRuleStorage {
	return &AlertRuleStorage{client: client}
}

func (s *AlertRuleStorage) Create(ctx context.Context, userID int64, rule *ent.AlertRule) (*ent.AlertRule, error) {
	return s.client.AlertRule.Create().
		SetUserID(userID).
		SetSensorID(rule.SensorID).
		SetConditionType(rule.ConditionType).
		SetThreshold(rule.Threshold).
		Save(ctx)
}

func (s *AlertRuleStorage) Get(ctx context.Context, id int64) (*ent.AlertRule, error) {
	return s.client.AlertRule.Query().Where(alertrule.ID(int(id))).Only(ctx)
}

func (s *AlertRuleStorage) List(ctx context.Context, userID int64) ([]*ent.AlertRule, error) {
	return s.client.AlertRule.Query().Where(alertrule.UserID(userID)).All(ctx)
}

func (s *AlertRuleStorage) Update(ctx context.Context, id int64, rule *ent.AlertRule) (*ent.AlertRule, error) {
	return s.client.AlertRule.UpdateOneID(int(id)).
		SetSensorID(rule.SensorID).
		SetConditionType(rule.ConditionType).
		SetThreshold(rule.Threshold).
		Save(ctx)
}

func (s *AlertRuleStorage) Delete(ctx context.Context, id int64) (bool, error) {
	err := s.client.AlertRule.DeleteOneID(int(id)).Exec(ctx)
	if err != nil {
		return false, err
	}
	return true, nil
}
