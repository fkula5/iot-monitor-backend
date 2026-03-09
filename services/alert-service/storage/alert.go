package storage

import (
	"context"

	"github.com/skni-kod/iot-monitor-backend/services/alert-service/ent"
	"github.com/skni-kod/iot-monitor-backend/services/alert-service/ent/alert"
)

type IAlertStorage interface {
	Get(ctx context.Context, id int) (*ent.Alert, error)
	List(ctx context.Context, userID int64) ([]*ent.Alert, error)
	MarkAsRead(ctx context.Context, id int) (bool, error)
}

type AlertStorage struct {
	client *ent.Client
}

func NewAlertStorage(client *ent.Client) IAlertStorage {
	return &AlertStorage{client: client}
}

func (s *AlertStorage) Get(ctx context.Context, id int) (*ent.Alert, error) {
	return s.client.Alert.Query().Where(alert.ID(id)).WithRule().Only(ctx)
}

func (s *AlertStorage) List(ctx context.Context, userID int64) ([]*ent.Alert, error) {
	return s.client.Alert.Query().Where(alert.UserID(userID)).WithRule().All(ctx)
}

func (s *AlertStorage) MarkAsRead(ctx context.Context, id int) (bool, error) {
	err := s.client.Alert.UpdateOneID(id).SetIsRead(true).Exec(ctx)
	if err != nil {
		return false, err
	}
	return true, nil
}
