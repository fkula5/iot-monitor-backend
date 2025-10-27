package storage

import (
	"context"
	"fmt"

	"github.com/skni-kod/iot-monitor-backend/services/sensor-service/ent"
	"github.com/skni-kod/iot-monitor-backend/services/sensor-service/ent/sensor"
)

type ISensorStorage interface {
	Get(ctx context.Context, id int) (*ent.Sensor, error)
	List(ctx context.Context) ([]*ent.Sensor, error)
	Create(ctx context.Context, sensor *ent.Sensor) (*ent.Sensor, error)
	Update(ctx context.Context, sensor *ent.Sensor) (*ent.Sensor, error)
	Delete(ctx context.Context, id int) error
	SetActive(ctx context.Context, id int) (*ent.Sensor, error)
	ListActive(ctx context.Context) ([]*ent.Sensor, error)
}

type SensorStorage struct {
	client *ent.Client
}

func NewSensorStorage(client *ent.Client) ISensorStorage {
	return &SensorStorage{client: client}
}

// Create implements ISensorStorage.
func (s *SensorStorage) Create(ctx context.Context, sensorData *ent.Sensor) (*ent.Sensor, error) {
	query := s.client.Sensor.Create().
		SetName(sensorData.Name).
		SetNillableLocation(&sensorData.Location).
		SetNillableDescription(&sensorData.Description).
		SetActive(sensorData.Active)

	if sensorData.Edges.Type != nil && sensorData.Edges.Type.ID != 0 {
		query = query.SetTypeID(sensorData.Edges.Type.ID)
	}

	newSensor, err := query.Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create sensor: %w", err)
	}

	sensorWithType, err := s.client.Sensor.Query().
		Where(sensor.ID(newSensor.ID)).
		WithType().
		Only(ctx)
	if err != nil {
		return newSensor, nil
	}

	return sensorWithType, nil
}

// Delete implements ISensorStorage.
func (s *SensorStorage) Delete(ctx context.Context, id int) error {
	exists, err := s.client.Sensor.Query().
		Where(sensor.ID(id)).
		Exist(ctx)
	if err != nil {
		return fmt.Errorf("failed to check if sensor exists: %w", err)
	}

	if !exists {
		return fmt.Errorf("sensor with ID %d does not exist", id)
	}

	err = s.client.Sensor.DeleteOneID(id).Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to delete sensor: %w", err)
	}

	return nil
}

// Get implements ISensorStorage.
func (s *SensorStorage) Get(ctx context.Context, id int) (*ent.Sensor, error) {
	return s.client.Sensor.Query().Where(sensor.ID(id)).WithType().Only(ctx)
}

// List implements ISensorStorage.
func (s *SensorStorage) List(ctx context.Context) ([]*ent.Sensor, error) {
	return s.client.Sensor.Query().All(ctx)
}

// Update implements ISensorStorage.
func (s *SensorStorage) Update(ctx context.Context, sensorData *ent.Sensor) (*ent.Sensor, error) {
	query := s.client.Sensor.UpdateOneID(sensorData.ID).
		SetName(sensorData.Name).
		SetNillableLocation(&sensorData.Location).
		SetNillableDescription(&sensorData.Description).
		SetActive(sensorData.Active)

	if sensorData.Edges.Type != nil && sensorData.Edges.Type.ID != 0 {
		currentSensor, err := s.client.Sensor.Query().
			Where(sensor.ID(sensorData.ID)).
			WithType().
			Only(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch current sensor: %w", err)
		}

		if currentSensor.Edges.Type == nil || currentSensor.Edges.Type.ID != sensorData.Edges.Type.ID {
			query = query.SetTypeID(sensorData.Edges.Type.ID)
		}
	}

	updatedSensor, err := query.Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to update sensor: %w", err)
	}

	sensorWithType, err := s.client.Sensor.Query().
		Where(sensor.ID(updatedSensor.ID)).
		WithType().
		Only(ctx)
	if err != nil {
		return updatedSensor, nil
	}

	return sensorWithType, nil
}

func (s *SensorStorage) SetActive(ctx context.Context, id int) (*ent.Sensor, error) {
	return s.client.Sensor.UpdateOneID(id).
		SetActive(true).
		Save(ctx)
}

func (s *SensorStorage) ListActive(ctx context.Context) ([]*ent.Sensor, error) {
	return s.client.Sensor.Query().
		Where(sensor.Active(true)).WithType().
		All(ctx)
}
