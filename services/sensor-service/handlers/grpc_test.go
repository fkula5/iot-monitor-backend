package handlers

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/skni-kod/iot-monitor-backend/internal/proto/sensor_service"
	"github.com/skni-kod/iot-monitor-backend/services/sensor-service/ent"
	"github.com/skni-kod/iot-monitor-backend/services/sensor-service/services"
)

// fakeSensorService stores sensors in memory; only the methods used by the
// ownership checks are implemented.
type fakeSensorService struct {
	services.ISensorService
	sensors map[int]*ent.Sensor
	deleted []int
}

func (f *fakeSensorService) GetSensor(_ context.Context, id int) (*ent.Sensor, error) {
	if s, ok := f.sensors[id]; ok {
		return s, nil
	}
	return nil, errors.New("not found")
}

func (f *fakeSensorService) ListSensors(_ context.Context, userID int64) ([]*ent.Sensor, error) {
	var result []*ent.Sensor
	for _, s := range f.sensors {
		if s.UserID == userID {
			result = append(result, s)
		}
	}
	return result, nil
}

func (f *fakeSensorService) DeleteSensor(_ context.Context, id int) error {
	f.deleted = append(f.deleted, id)
	return nil
}

type fakeGroupService struct {
	services.ISensorGroupService
	groups map[int]*ent.SensorGroup
}

func (f *fakeGroupService) Get(_ context.Context, id int) (*ent.SensorGroup, error) {
	if g, ok := f.groups[id]; ok {
		return g, nil
	}
	return nil, errors.New("not found")
}

func newTestHandler() (*SensorsGrpcHandler, *fakeSensorService) {
	sensors := &fakeSensorService{sensors: map[int]*ent.Sensor{
		1: {ID: 1, UserID: 10},
		2: {ID: 2, UserID: 20},
	}}
	groups := &fakeGroupService{groups: map[int]*ent.SensorGroup{
		5: {ID: 5, UserID: 10},
	}}
	return &SensorsGrpcHandler{sensorsService: sensors, sensorsGroupService: groups}, sensors
}

func TestGetSensor_Ownership(t *testing.T) {
	h, _ := newTestHandler()

	res, err := h.GetSensor(context.Background(), &pb.GetSensorRequest{Id: 1, UserId: 10})
	assert.NoError(t, err)
	assert.Equal(t, int64(10), res.Sensor.UserId)

	_, err = h.GetSensor(context.Background(), &pb.GetSensorRequest{Id: 2, UserId: 10})
	assert.Equal(t, codes.NotFound, status.Code(err))

	// Internal callers (user_id 0) are not scoped.
	_, err = h.GetSensor(context.Background(), &pb.GetSensorRequest{Id: 2})
	assert.NoError(t, err)
}

func TestDeleteSensor_Ownership(t *testing.T) {
	h, sensors := newTestHandler()

	_, err := h.DeleteSensor(context.Background(), &pb.DeleteSensorRequest{Id: 2, UserId: 10})
	assert.Equal(t, codes.NotFound, status.Code(err))

	_, err = h.DeleteSensor(context.Background(), &pb.DeleteSensorRequest{Id: 1})
	assert.Equal(t, codes.InvalidArgument, status.Code(err))

	assert.Empty(t, sensors.deleted)

	_, err = h.DeleteSensor(context.Background(), &pb.DeleteSensorRequest{Id: 1, UserId: 10})
	assert.NoError(t, err)
	assert.Equal(t, []int{1}, sensors.deleted)
}

func TestGetSensorGroup_Ownership(t *testing.T) {
	h, _ := newTestHandler()

	_, err := h.GetSensorGroup(context.Background(), &pb.GetSensorGroupRequest{Id: 5, UserId: 10})
	assert.NoError(t, err)

	_, err = h.GetSensorGroup(context.Background(), &pb.GetSensorGroupRequest{Id: 5, UserId: 20})
	assert.Equal(t, codes.NotFound, status.Code(err))
}

func TestAddSensorsToGroup_RejectsForeignSensor(t *testing.T) {
	h, _ := newTestHandler()

	_, err := h.AddSensorsToGroup(context.Background(), &pb.AddSensorsToGroupRequest{
		GroupId:   5,
		SensorIds: []int64{1, 2},
		UserId:    10,
	})
	assert.Equal(t, codes.NotFound, status.Code(err))
}
