package services

import (
	"context"
	"fmt"

	"github.com/skni-kod/iot-monitor-backend/services/sensor-service/ent"
	"github.com/skni-kod/iot-monitor-backend/services/sensor-service/storage"
)

type ISensorGroupService interface {
	CreateGroup(ctx context.Context, group *ent.SensorGroup, sensorIDs []int64) (*ent.SensorGroup, error)
	GetGroup(ctx context.Context, id int) (*ent.SensorGroup, error)
	GetGroupWithSensors(ctx context.Context, id int) (*ent.SensorGroup, error)
	ListGroups(ctx context.Context, userID int64) ([]*ent.SensorGroup, error)
	UpdateGroup(ctx context.Context, id int, group *ent.SensorGroup) (*ent.SensorGroup, error)
	DeleteGroup(ctx context.Context, id int) error
	AddSensorsToGroup(ctx context.Context, groupID int, sensorIDs []int64) (*ent.SensorGroup, error)
	RemoveSensorsFromGroup(ctx context.Context, groupID int, sensorIDs []int64) (*ent.SensorGroup, error)
	GetGroupsForSensor(ctx context.Context, sensorID int) ([]*ent.SensorGroup, error)
}

type SensorGroupService struct {
	store storage.ISensorGroupStorage
}

func NewSensorGroupService(store storage.ISensorGroupStorage) ISensorGroupService {
	return &SensorGroupService{store: store}
}

func (s *SensorGroupService) CreateGroup(ctx context.Context, group *ent.SensorGroup, sensorIDs []int64) (*ent.SensorGroup, error) {
	if group.Name == "" {
		return nil, fmt.Errorf("group name cannot be empty")
	}

	if group.Color == "" {
		group.Color = "#3B82F6"
	}

	return s.store.Create(ctx, group, sensorIDs)
}

func (s *SensorGroupService) GetGroup(ctx context.Context, id int) (*ent.SensorGroup, error) {
	return s.store.Get(ctx, id)
}

func (s *SensorGroupService) GetGroupWithSensors(ctx context.Context, id int) (*ent.SensorGroup, error) {
	return s.store.GetWithSensors(ctx, id)
}

func (s *SensorGroupService) ListGroups(ctx context.Context, userID int64) ([]*ent.SensorGroup, error) {
	return s.store.List(ctx, userID)
}

func (s *SensorGroupService) UpdateGroup(ctx context.Context, id int, group *ent.SensorGroup) (*ent.SensorGroup, error) {
	if group.Name == "" {
		return nil, fmt.Errorf("group name cannot be empty")
	}

	return s.store.Update(ctx, id, group)
}

func (s *SensorGroupService) DeleteGroup(ctx context.Context, id int) error {
	return s.store.Delete(ctx, id)
}

func (s *SensorGroupService) AddSensorsToGroup(ctx context.Context, groupID int, sensorIDs []int64) (*ent.SensorGroup, error) {
	if len(sensorIDs) == 0 {
		return nil, fmt.Errorf("at least one sensor ID is required")
	}

	return s.store.AddSensors(ctx, groupID, sensorIDs)
}

func (s *SensorGroupService) RemoveSensorsFromGroup(ctx context.Context, groupID int, sensorIDs []int64) (*ent.SensorGroup, error) {
	if len(sensorIDs) == 0 {
		return nil, fmt.Errorf("at least one sensor ID is required")
	}

	return s.store.RemoveSensors(ctx, groupID, sensorIDs)
}

func (s *SensorGroupService) GetGroupsForSensor(ctx context.Context, sensorID int) ([]*ent.SensorGroup, error) {
	return s.store.GetGroupsForSensor(ctx, sensorID)
}
