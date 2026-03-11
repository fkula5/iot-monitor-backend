package service

import (
	"context"

	"github.com/skni-kod/iot-monitor-backend/services/alert-service/ent"
	"github.com/skni-kod/iot-monitor-backend/services/alert-service/storage"
)

type IAlertRuleService interface {
	GetAlertRule(ctx context.Context, id int64) (*ent.AlertRule, error)
	ListAlertRules(ctx context.Context, userID int64) ([]*ent.AlertRule, error)
	CreateAlertRule(ctx context.Context, rule *ent.AlertRule) (*ent.AlertRule, error)
	UpdateAlertRule(ctx context.Context, rule *ent.AlertRule) (*ent.AlertRule, error)
	DeleteAlertRule(ctx context.Context, id int64) error
}

type AlertRuleService struct {
	storage storage.IAlertRuleStorage
}

func NewAlertRuleService(s storage.IAlertRuleStorage) *AlertRuleService {
	return &AlertRuleService{storage: s}
}

func (s *AlertRuleService) GetAlertRule(ctx context.Context, id int64) (*ent.AlertRule, error) {
	return s.storage.Get(ctx, id)
}

func (s *AlertRuleService) ListAlertRules(ctx context.Context, userID int64) ([]*ent.AlertRule, error) {
	return s.storage.List(ctx, userID)
}

func (s *AlertRuleService) CreateAlertRule(ctx context.Context, rule *ent.AlertRule) (*ent.AlertRule, error) {
	return s.storage.Create(ctx, rule)
}

func (s *AlertRuleService) UpdateAlertRule(ctx context.Context, rule *ent.AlertRule) (*ent.AlertRule, error) {
	return s.storage.Update(ctx, rule)
}

func (s *AlertRuleService) DeleteAlertRule(ctx context.Context, id int64) error {
	return s.storage.Delete(ctx, id)
}
