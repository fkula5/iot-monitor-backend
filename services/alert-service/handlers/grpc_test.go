package handlers

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/skni-kod/iot-monitor-backend/internal/proto/alert_service"
	"github.com/skni-kod/iot-monitor-backend/services/alert-service/ent"
	"github.com/skni-kod/iot-monitor-backend/services/alert-service/service"
	"github.com/skni-kod/iot-monitor-backend/services/alert-service/storage"
)

type fakeRuleStorage struct {
	storage.IAlertRuleStorage
	rules   map[int64]*ent.AlertRule
	deleted []int64
}

func (f *fakeRuleStorage) Get(_ context.Context, id int64) (*ent.AlertRule, error) {
	if r, ok := f.rules[id]; ok {
		return r, nil
	}
	return nil, errors.New("not found")
}

func (f *fakeRuleStorage) Delete(_ context.Context, id int64) error {
	f.deleted = append(f.deleted, id)
	return nil
}

type fakeAlertStorage struct {
	storage.IAlertStorage
	alerts map[int]*ent.Alert
	read   []int
}

func (f *fakeAlertStorage) Get(_ context.Context, id int) (*ent.Alert, error) {
	if a, ok := f.alerts[id]; ok {
		return a, nil
	}
	return nil, errors.New("not found")
}

func (f *fakeAlertStorage) MarkAsRead(_ context.Context, id int) (bool, error) {
	f.read = append(f.read, id)
	return true, nil
}

func newTestHandler() (*AlertGrpcHandler, *fakeRuleStorage, *fakeAlertStorage) {
	rules := &fakeRuleStorage{rules: map[int64]*ent.AlertRule{7: {ID: 7, UserID: 10}}}
	alerts := &fakeAlertStorage{alerts: map[int]*ent.Alert{3: {ID: 3, UserID: 10}}}
	return NewAlertGrpcHandler(service.NewAlertService(alerts), service.NewAlertRuleService(rules)), rules, alerts
}

func TestDeleteAlertRule_Ownership(t *testing.T) {
	h, rules, _ := newTestHandler()

	_, err := h.DeleteAlertRule(context.Background(), &pb.DeleteAlertRuleRequest{Id: 7, UserId: 20})
	assert.Equal(t, codes.NotFound, status.Code(err))

	_, err = h.DeleteAlertRule(context.Background(), &pb.DeleteAlertRuleRequest{Id: 7})
	assert.Equal(t, codes.InvalidArgument, status.Code(err))
	assert.Empty(t, rules.deleted)

	_, err = h.DeleteAlertRule(context.Background(), &pb.DeleteAlertRuleRequest{Id: 7, UserId: 10})
	assert.NoError(t, err)
	assert.Equal(t, []int64{7}, rules.deleted)
}

func TestUpdateAlertRule_RejectsOtherUser(t *testing.T) {
	h, _, _ := newTestHandler()

	_, err := h.UpdateAlertRule(context.Background(), &pb.UpdateAlertRuleRequest{Id: 7, UserId: 20, IsEnabled: false})
	assert.Equal(t, codes.NotFound, status.Code(err))
}

func TestMarkAlertAsRead_Ownership(t *testing.T) {
	h, _, alerts := newTestHandler()

	_, err := h.MarkAlertAsRead(context.Background(), &pb.MarkAlertAsReadRequest{Id: 3, UserId: 20})
	assert.Equal(t, codes.NotFound, status.Code(err))
	assert.Empty(t, alerts.read)

	res, err := h.MarkAlertAsRead(context.Background(), &pb.MarkAlertAsReadRequest{Id: 3, UserId: 10})
	assert.NoError(t, err)
	assert.True(t, res.Success)
}
