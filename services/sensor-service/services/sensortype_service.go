package services

import (
	"context"
	"fmt"

	"github.com/skni-kod/iot-monitor-backend/services/sensor-service/ent"
	"github.com/skni-kod/iot-monitor-backend/services/sensor-service/storage"
)

type ISensorTypeService interface {
	GetSensorType(ctx context.Context, id int) (*ent.SensorType, error)
	CreateSensorType(ctx context.Context, sensorType *ent.SensorType) (*ent.SensorType, error)
	ListSensorTypes(ctx context.Context) ([]*ent.SensorType, error)
	UpdateSensorType(ctx context.Context, id int, sensorType *ent.SensorType) (*ent.SensorType, error)
	DeleteSensorType(ctx context.Context, id int) error
}

type SensorTypeService struct {
	store storage.ISensorTypeStorage
}

func NewSensorTypeService(store storage.ISensorTypeStorage) ISensorTypeService {
	return &SensorTypeService{store: store}
}

// DeleteSensorType implements ISensorTypeService.
func (s *SensorTypeService) DeleteSensorType(ctx context.Context, id int) error {
	return s.store.Delete(ctx, id)
}

// UpdateSensorType implements ISensorTypeService.
func (s *SensorTypeService) UpdateSensorType(ctx context.Context, id int, sensorType *ent.SensorType) (*ent.SensorType, error) {
	return s.store.Update(ctx, id, sensorType)
}

// CreateSensorType implements ISensorTypeService.
func (s *SensorTypeService) CreateSensorType(ctx context.Context, sensorType *ent.SensorType) (*ent.SensorType, error) {
	if sensorType.Name == "" {
		return nil, fmt.Errorf("sensor type name cannot be empty")
	}

	if sensorType.Model == "" {
		return nil, fmt.Errorf("sensor type model cannot be empty")
	}

	return s.store.Create(ctx, sensorType)
}

// ListSensorTypes implements ISensorTypeService.
func (s *SensorTypeService) ListSensorTypes(ctx context.Context) ([]*ent.SensorType, error) {
	return s.store.List(ctx)
}

func (s *SensorTypeService) GetSensorType(ctx context.Context, id int) (*ent.SensorType, error) {
	return s.store.Get(ctx, id)
}
