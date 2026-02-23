package storage

import (
	"context"
	"fmt"

	"github.com/skni-kod/iot-monitor-backend/services/sensor-service/ent"
	"github.com/skni-kod/iot-monitor-backend/services/sensor-service/ent/sensor"
	"github.com/skni-kod/iot-monitor-backend/services/sensor-service/ent/sensorgroup"
)

type ISensorGroupStorage interface {
	Create(ctx context.Context, group *ent.SensorGroup, sensorIDs []int64) (*ent.SensorGroup, error)
	Get(ctx context.Context, id int) (*ent.SensorGroup, error)
	List(ctx context.Context, userID int64) ([]*ent.SensorGroup, error)
	Update(ctx context.Context, id int, group *ent.SensorGroup) (*ent.SensorGroup, error)
	Delete(ctx context.Context, id int) error
	AddSensors(ctx context.Context, groupID int, sensorIDs []int64) (*ent.SensorGroup, error)
	RemoveSensors(ctx context.Context, groupID int, sensorIDs []int64) (*ent.SensorGroup, error)
	GetGroupsForSensor(ctx context.Context, sensorID int) ([]*ent.SensorGroup, error)
}

type SensorGroupStorage struct {
	client *ent.Client
}

func NewSensorGroupStorage(client *ent.Client) ISensorGroupStorage {
	return &SensorGroupStorage{client: client}
}

func (s *SensorGroupStorage) Create(ctx context.Context, groupData *ent.SensorGroup, sensorIDs []int64) (*ent.SensorGroup, error) {
	tx, err := s.client.Tx(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to start transaction: %w", err)
	}

	group, err := tx.SensorGroup.Create().
		SetName(groupData.Name).
		SetNillableDescription(&groupData.Description).
		SetNillableColor(&groupData.Color).
		SetNillableIcon(&groupData.Icon).
		SetUserID(groupData.UserID).
		Save(ctx)
	if err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to create sensor group: %w", err)
	}

	if len(sensorIDs) > 0 {
		intIDs := make([]int, len(sensorIDs))
		for i, id := range sensorIDs {
			intIDs[i] = int(id)
		}

		err = tx.SensorGroup.UpdateOneID(group.ID).
			AddSensorIDs(intIDs...).
			Exec(ctx)
		if err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("failed to add sensors to group: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return s.Get(ctx, group.ID)
}

func (s *SensorGroupStorage) Get(ctx context.Context, id int) (*ent.SensorGroup, error) {
	return s.client.SensorGroup.Query().
		Where(sensorgroup.ID(id)).
		WithSensors(func(q *ent.SensorQuery) {
			q.WithType()
		}).
		Only(ctx)
}

func (s *SensorGroupStorage) List(ctx context.Context, userID int64) ([]*ent.SensorGroup, error) {
	return s.client.SensorGroup.Query().
		Where(sensorgroup.UserID(userID)).
		WithSensors(func(q *ent.SensorQuery) {
			q.Select(sensor.FieldID)
		}).
		All(ctx)
}

func (s *SensorGroupStorage) Update(ctx context.Context, id int, groupData *ent.SensorGroup) (*ent.SensorGroup, error) {
	update := s.client.SensorGroup.UpdateOneID(id).
		SetName(groupData.Name).
		SetNillableDescription(&groupData.Description).
		SetNillableColor(&groupData.Color).
		SetNillableIcon(&groupData.Icon)

	if _, err := update.Save(ctx); err != nil {
		return nil, fmt.Errorf("failed to update sensor group: %w", err)
	}

	return s.Get(ctx, id)
}

func (s *SensorGroupStorage) Delete(ctx context.Context, id int) error {
	exists, err := s.client.SensorGroup.Query().
		Where(sensorgroup.ID(id)).
		Exist(ctx)
	if err != nil {
		return fmt.Errorf("failed to check if group exists: %w", err)
	}
	if !exists {
		return fmt.Errorf("sensor group with ID %d does not exist", id)
	}

	return s.client.SensorGroup.DeleteOneID(id).Exec(ctx)
}

func (s *SensorGroupStorage) AddSensors(ctx context.Context, groupID int, sensorIDs []int64) (*ent.SensorGroup, error) {
	intIDs := make([]int, len(sensorIDs))
	for i, id := range sensorIDs {
		intIDs[i] = int(id)
	}

	err := s.client.SensorGroup.UpdateOneID(groupID).
		AddSensorIDs(intIDs...).
		Exec(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to add sensors to group: %w", err)
	}

	return s.Get(ctx, groupID)
}

func (s *SensorGroupStorage) RemoveSensors(ctx context.Context, groupID int, sensorIDs []int64) (*ent.SensorGroup, error) {
	intIDs := make([]int, len(sensorIDs))
	for i, id := range sensorIDs {
		intIDs[i] = int(id)
	}

	err := s.client.SensorGroup.UpdateOneID(groupID).
		RemoveSensorIDs(intIDs...).
		Exec(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to remove sensors from group: %w", err)
	}

	return s.Get(ctx, groupID)
}

func (s *SensorGroupStorage) GetGroupsForSensor(ctx context.Context, sensorID int) ([]*ent.SensorGroup, error) {
	sensor, err := s.client.Sensor.Query().
		Where(sensor.ID(sensorID)).
		WithGroups().
		Only(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get groups for sensor: %w", err)
	}

	return sensor.Edges.Groups, nil
}
