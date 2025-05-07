package storage

import (
	"context"
	"fmt"

	"github.com/skni-kod/iot-monitor-backend/services/sensor-service/ent"
	"github.com/skni-kod/iot-monitor-backend/services/sensor-service/ent/sensortype"
)

type ISensorTypeStorage interface {
	Get(ctx context.Context, id int) (*ent.SensorType, error)
	List(ctx context.Context) ([]*ent.SensorType, error)
	Create(ctx context.Context, sensorType *ent.SensorType) (*ent.SensorType, error)
}

type SensorTypeStorage struct {
	client *ent.Client
}

func NewSensorTypeStorage(client *ent.Client) ISensorTypeStorage {
	return &SensorTypeStorage{client: client}
}

// Create implements ISensorTypeStorage.
func (s *SensorTypeStorage) Create(ctx context.Context, sensorType *ent.SensorType) (*ent.SensorType, error) {
	exists, err := s.client.SensorType.Query().
		Where(sensortype.Name(sensorType.Name)).
		Exist(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to check if sensor type exists: %w", err)
	}

	if exists {
		return nil, fmt.Errorf("sensor type with name '%s' already exists", sensorType.Name)
	}

	createdSensorType, err := s.client.SensorType.Create().
		SetName(sensorType.Name).
		SetModel(sensorType.Model).
		SetNillableManufacturer(&sensorType.Manufacturer).
		SetNillableDescription(&sensorType.Description).
		SetNillableUnit(&sensorType.Unit).
		SetMinValue(sensorType.MinValue).
		SetMaxValue(sensorType.MaxValue).
		Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create sensor type: %w", err)
	}

	return createdSensorType, nil
}

// List implements ISensorTypeStorage.
func (s *SensorTypeStorage) List(ctx context.Context) ([]*ent.SensorType, error) {
	return s.client.SensorType.Query().All(ctx)
}

func (s *SensorTypeStorage) Get(ctx context.Context, id int) (*ent.SensorType, error) {
	return s.client.SensorType.Query().Where(sensortype.ID(id)).Only(ctx)
}
