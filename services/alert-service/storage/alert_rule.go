package storage

import (
	"context"
	"fmt"

	"github.com/skni-kod/iot-monitor-backend/services/alert-service/ent"
	"github.com/skni-kod/iot-monitor-backend/services/alert-service/ent/alertrule"
)

type IAlertRuleStorage interface {
	Create(ctx context.Context, rule *ent.AlertRule) (*ent.AlertRule, error)
	Get(ctx context.Context, id int64) (*ent.AlertRule, error)
	List(ctx context.Context, userID int64, limit, offset int) ([]*ent.AlertRule, int, error)
	Update(ctx context.Context, rule *ent.AlertRule) (*ent.AlertRule, error)
	Delete(ctx context.Context, id int64) error
}

type AlertRuleStorage struct {
	client *ent.Client
}

func NewAlertRuleStorage(client *ent.Client) IAlertRuleStorage {
	return &AlertRuleStorage{client: client}
}

func (s *AlertRuleStorage) Create(ctx context.Context, rule *ent.AlertRule) (*ent.AlertRule, error) {
	return s.client.AlertRule.Create().
		SetName(rule.Name).
		SetUserID(rule.UserID).
		SetSensorID(rule.SensorID).
		SetConditionType(rule.ConditionType).
		SetThreshold(rule.Threshold).
		SetDescription(rule.Description).
		Save(ctx)
}

func (s *AlertRuleStorage) Get(ctx context.Context, id int64) (*ent.AlertRule, error) {
	return s.client.AlertRule.Query().Where(alertrule.ID(int(id))).Only(ctx)
}

func (s *AlertRuleStorage) List(ctx context.Context, userID int64, limit, offset int) ([]*ent.AlertRule, int, error) {
	query := s.client.AlertRule.Query().Where(alertrule.UserID(userID))

	totalCount, err := query.Count(ctx)
	if err != nil {
		return nil, 0, err
	}

	rules, err := query.
		Limit(limit).
		Offset(offset).
		Order(ent.Desc(alertrule.FieldCreatedAt)).
		All(ctx)
	if err != nil {
		return nil, 0, err
	}

	return rules, totalCount, nil
}

func (s *AlertRuleStorage) Update(ctx context.Context, rule *ent.AlertRule) (*ent.AlertRule, error) {
	return s.client.AlertRule.UpdateOneID(rule.ID).
		SetName(rule.Name).
		SetSensorID(rule.SensorID).
		SetConditionType(rule.ConditionType).
		SetThreshold(rule.Threshold).
		SetDescription(rule.Description).
		SetIsEnabled(rule.IsEnabled).
		Save(ctx)
}

func (s *AlertRuleStorage) Delete(ctx context.Context, id int64) error {
	exists, err := s.client.AlertRule.Query().Where(alertrule.ID(int(id))).Exist(ctx)
	if err != nil {
		return fmt.Errorf("failed to check if rule exists: %w", err)
	}
	if !exists {
		return fmt.Errorf("rule with ID %d does not exist", id)
	}
	err = s.client.AlertRule.DeleteOneID(int(id)).Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to delete rule: %w", err)
	}
	return nil
}
