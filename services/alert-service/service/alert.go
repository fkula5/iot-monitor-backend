package service

import (
	"context"

	"github.com/skni-kod/iot-monitor-backend/services/alert-service/ent"
	"github.com/skni-kod/iot-monitor-backend/services/alert-service/storage"
)

type AlertService struct {
	storage storage.IAlertStorage
}

func NewAlertService(s storage.IAlertStorage) *AlertService {
	return &AlertService{storage: s}
}

func (s *AlertService) GetAlert(ctx context.Context, id int) (*ent.Alert, error) {
	return s.storage.Get(ctx, id)
}

func (s *AlertService) ListAlerts(ctx context.Context, userID int64) ([]*ent.Alert, error) {
	return s.storage.List(ctx, userID)
}

func (s *AlertService) MarkAsRead(ctx context.Context, id int) (bool, error) {
	return s.storage.MarkAsRead(ctx, id)
}
