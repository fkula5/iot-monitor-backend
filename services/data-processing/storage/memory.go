package storage

import (
	"context"
	"fmt"
	"sync"
	"time"

	pb_data "github.com/skni-kod/iot-monitor-backend/internal/proto/data_service"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type IDataStorage interface {
	StoreReading(ctx context.Context, sensorID int64, value float32, timestamp time.Time) error
	QueryReadings(ctx context.Context, sensorID int64, startTime, endTime time.Time) ([]*pb_data.DataPoint, error)
	GetLatestReading(ctx context.Context, sensorID int64) (*pb_data.ReadingUpdate, error)
}

type reading struct {
	value     float32
	timestamp time.Time
}

type InMemoryStorage struct {
	data   map[int64][]reading
	latest map[int64]*pb_data.ReadingUpdate
	mu     sync.RWMutex
}

func NewInMemoryStorage() IDataStorage {
	return &InMemoryStorage{
		data:   make(map[int64][]reading),
		latest: make(map[int64]*pb_data.ReadingUpdate),
	}
}

func (s *InMemoryStorage) StoreReading(ctx context.Context, sensorID int64, value float32, timestamp time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	r := reading{
		value:     value,
		timestamp: timestamp,
	}

	s.data[sensorID] = append(s.data[sensorID], r)

	// Keep only last 1000 readings per sensor
	if len(s.data[sensorID]) > 1000 {
		s.data[sensorID] = s.data[sensorID][len(s.data[sensorID])-1000:]
	}

	s.latest[sensorID] = &pb_data.ReadingUpdate{
		SensorId:  sensorID,
		Value:     value,
		Timestamp: timestamppb.New(timestamp),
	}

	return nil
}

func (s *InMemoryStorage) QueryReadings(ctx context.Context, sensorID int64, startTime, endTime time.Time) ([]*pb_data.DataPoint, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	readings, exists := s.data[sensorID]
	if !exists {
		return []*pb_data.DataPoint{}, nil
	}

	var result []*pb_data.DataPoint
	for _, r := range readings {
		if (r.timestamp.Equal(startTime) || r.timestamp.After(startTime)) &&
			(r.timestamp.Equal(endTime) || r.timestamp.Before(endTime)) {
			result = append(result, &pb_data.DataPoint{
				Time:  timestamppb.New(r.timestamp),
				Value: r.value,
			})
		}
	}

	return result, nil
}

func (s *InMemoryStorage) GetLatestReading(ctx context.Context, sensorID int64) (*pb_data.ReadingUpdate, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	update, exists := s.latest[sensorID]
	if !exists {
		return nil, fmt.Errorf("no readings found for sensor %d", sensorID)
	}

	return update, nil
}
