package storage

import (
	"context"

	"github.com/skni-kod/iot-monitor-backend/services/sensor-service/ent"
	"github.com/skni-kod/iot-monitor-backend/services/sensor-service/ent/sensortype"
)

type ISensorTypeStorage interface {
	Get(ctx context.Context, id int) (*ent.SensorType, error)
}

type SensorTypeStorage struct {
	client *ent.Client
}

func NewSensorTypeStorage(client *ent.Client) ISensorTypeStorage {
	return &SensorTypeStorage{client: client}
}

func (s *SensorTypeStorage) Get(ctx context.Context, id int) (*ent.SensorType, error) {
	return s.client.SensorType.Query().Where(sensortype.ID(id)).Only(ctx)
}
