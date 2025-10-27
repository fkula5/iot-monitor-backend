package services

import (
	"context"

	"github.com/skni-kod/iot-monitor-backend/services/sensor-service/ent"
	"github.com/skni-kod/iot-monitor-backend/services/sensor-service/storage"
)

type ISensorService interface {
	GetSensor(ctx context.Context, id int) (*ent.Sensor, error)
	ListSensors(ctx context.Context) ([]*ent.Sensor, error)
	CreateSensor(ctx context.Context, sensor *ent.Sensor) (*ent.Sensor, error)
	UpdateSensor(ctx context.Context, sensor *ent.Sensor) (*ent.Sensor, error)
	DeleteSensor(ctx context.Context, id int) error
	SetSensorActive(ctx context.Context, id int) (*ent.Sensor, error)
	ListActiveSensors(ctx context.Context) ([]*ent.Sensor, error)
}

type SensorService struct {
	store storage.ISensorStorage
}

func NewSensorService(store storage.ISensorStorage) ISensorService {
	return &SensorService{store: store}
}

func (s *SensorService) CreateSensor(ctx context.Context, sensor *ent.Sensor) (*ent.Sensor, error) {
	return s.store.Create(ctx, sensor)
}

func (s *SensorService) DeleteSensor(ctx context.Context, id int) error {
	return s.store.Delete(ctx, id)
}

func (s *SensorService) GetSensor(ctx context.Context, id int) (*ent.Sensor, error) {
	return s.store.Get(ctx, id)
}

func (s *SensorService) ListSensors(ctx context.Context) ([]*ent.Sensor, error) {
	return s.store.List(ctx)
}

func (s *SensorService) UpdateSensor(ctx context.Context, sensor *ent.Sensor) (*ent.Sensor, error) {
	return s.store.Update(ctx, sensor)
}

func (s *SensorService) SetSensorActive(ctx context.Context, id int) (*ent.Sensor, error) {
	return s.store.SetActive(ctx, id)
}

func (s *SensorService) ListActiveSensors(ctx context.Context) ([]*ent.Sensor, error) {
	return s.store.ListActive(ctx)
}
