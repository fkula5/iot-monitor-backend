package storage

import (
	"context"

	"github.com/skni-kod/iot-monitor-backend/services/sensor-service/ent"
	"github.com/skni-kod/iot-monitor-backend/services/sensor-service/ent/sensor"
)

type ISensorStorage interface {
	Get(ctx context.Context, id int) (*ent.Sensor, error)
	List(ctx context.Context) ([]*ent.Sensor, error)
	Create(ctx context.Context, sensor *ent.Sensor) (*ent.Sensor, error)
	Update(ctx context.Context, sensor *ent.Sensor) (*ent.Sensor, error)
	Delete(ctx context.Context, id int) error
	SetActive(ctx context.Context, id int, active bool) (*ent.Sensor, error)
	ListActive(ctx context.Context) ([]*ent.Sensor, error)
}

type SensorStorage struct {
	client *ent.Client
}

func NewSensorStorage(client *ent.Client) ISensorStorage {
	return &SensorStorage{client: client}
}

// Create implements ISensorStorage.
func (s *SensorStorage) Create(ctx context.Context, sensor *ent.Sensor) (*ent.Sensor, error) {
	panic("unimplemented")
}

// Delete implements ISensorStorage.
func (s *SensorStorage) Delete(ctx context.Context, id int) error {
	panic("unimplemented")
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
func (s *SensorStorage) Update(ctx context.Context, sensor *ent.Sensor) (*ent.Sensor, error) {
	panic("unimplemented")
}

func (s *SensorStorage) SetActive(ctx context.Context, id int, active bool) (*ent.Sensor, error) {
	return s.client.Sensor.UpdateOneID(id).
		SetActive(active).
		Save(ctx)
}

func (s *SensorStorage) ListActive(ctx context.Context) ([]*ent.Sensor, error) {
	return s.client.Sensor.Query().
		Where(sensor.Active(true)).WithType().
		All(ctx)
}
