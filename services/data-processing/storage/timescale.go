package storage

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	pb_data "github.com/skni-kod/iot-monitor-backend/internal/proto/data_service"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type TimescaleStorage struct {
	db *sql.DB
}

type ITimeScaleStorage interface {
	StoreReading(ctx context.Context, sensorID int64, value float32, timestamp time.Time) error
	QueryReadings(ctx context.Context, sensorID int64, startTime, endTime time.Time) ([]*pb_data.DataPoint, error)
}

func NewTimescaleStorage(db *sql.DB) ITimeScaleStorage {
	return &TimescaleStorage{db: db}
}

func (s *TimescaleStorage) StoreReading(ctx context.Context, sensorID int64, value float32, timestamp time.Time) error {
	_, err := s.db.ExecContext(ctx,
		"INSERT INTO sensor_readings (time, sensor_id, value) VALUES ($1, $2, $3)",
		timestamp, sensorID, value)
	return err
}

func (s *TimescaleStorage) QueryReadings(ctx context.Context, sensorID int64, startTime, endTime time.Time) ([]*pb_data.DataPoint, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT time, value FROM sensor_readings 
         WHERE sensor_id = $1 AND time >= $2 AND time <= $3 
         ORDER BY time ASC`,
		sensorID, startTime, endTime)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var dataPoints []*pb_data.DataPoint
	for rows.Next() {
		var t time.Time
		var v float32
		if err := rows.Scan(&t, &v); err != nil {
			return nil, err
		}
		dataPoints = append(dataPoints, &pb_data.DataPoint{
			Time:  timestamppb.New(t),
			Value: v,
		})
	}
	return dataPoints, nil
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
