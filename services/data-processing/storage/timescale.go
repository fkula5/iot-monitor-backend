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
	GetLatestReadings(ctx context.Context, sensorID int64, entries int64) ([]*pb_data.ReadingUpdate, error)
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

func (s *TimescaleStorage) GetLatestReadings(ctx context.Context, sensorID int64, entries int64) ([]*pb_data.ReadingUpdate, error) {
	query := `SELECT time, value FROM sensor_readings 
              WHERE sensor_id = $1 
              ORDER BY time DESC 
              LIMIT $2`

	rows, err := s.db.QueryContext(ctx, query, sensorID, entries)
	if err != nil {
		return nil, fmt.Errorf("query error: %w", err)
	}

	defer rows.Close()

	var readings []*pb_data.ReadingUpdate

	for rows.Next() {
		var t time.Time
		var v float32

		if err := rows.Scan(&t, &v); err != nil {
			return nil, fmt.Errorf("scan error: %w", err)
		}

		readings = append(readings, &pb_data.ReadingUpdate{
			SensorId:  sensorID,
			Value:     v,
			Timestamp: timestamppb.New(t),
		})
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return readings, nil
}
