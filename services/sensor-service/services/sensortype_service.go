package services

import (
	"context"

	"github.com/skni-kod/iot-monitor-backend/services/sensor-service/ent"
	"github.com/skni-kod/iot-monitor-backend/services/sensor-service/storage"
)

type ISensorTypeService interface {
	GetSensorType(ctx context.Context, id int) (*ent.SensorType, error)
}

type SensorTypeService struct {
	store storage.ISensorTypeStorage
}

func NewSensorTypeService(store storage.ISensorTypeStorage) ISensorTypeService {
	return &SensorTypeService{store: store}
}

func (s *SensorTypeService) GetSensorType(ctx context.Context, id int) (*ent.SensorType, error) {
	return s.store.Get(ctx, id)
}
