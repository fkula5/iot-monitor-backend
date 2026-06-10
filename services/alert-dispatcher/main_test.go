package main

import (
	"context"
	"encoding/json"
	"os"
	"testing"
	"time"

	pb_auth "github.com/skni-kod/iot-monitor-backend/internal/proto/auth"
	"github.com/skni-kod/iot-monitor-backend/pkg/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc"
)

func TestMain(m *testing.M) {
	logger.Init(logger.Config{
		Level:       "info",
		Environment: "development",
		ServiceName: "alert-dispatcher-test",
		OutputPaths: []string{"stdout"},
	})
	os.Exit(m.Run())
}

func TestAlertEventUnmarshaling(t *testing.T) {
	jsonData := `{
		"alert_id": 1,
		"rule_id": 2,
		"user_id": 123,
		"sensor_id": 456,
		"message": "Temperature is too high",
		"value": 35.5,
		"timestamp": "2026-03-31T12:00:00Z"
	}`

	var event AlertEvent
	err := json.Unmarshal([]byte(jsonData), &event)
	assert.NoError(t, err)

	assert.Equal(t, 1, event.AlertID)
	assert.Equal(t, 2, event.RuleID)
	assert.Equal(t, int64(123), event.UserID)
	assert.Equal(t, int64(456), event.SensorID)
	assert.Equal(t, "Temperature is too high", event.Message)
	assert.Equal(t, 35.5, event.Value)
	assert.False(t, event.Timestamp.IsZero())
}

type MockMailer struct {
	mock.Mock
}

func (m *MockMailer) SendAlertEmail(to string, event AlertEvent) error {
	args := m.Called(to, event)
	return args.Error(0)
}

type MockAuthServiceClient struct {
	mock.Mock
	pb_auth.AuthServiceClient
}

func (m *MockAuthServiceClient) GetUser(ctx context.Context, in *pb_auth.GetUserRequest, opts ...grpc.CallOption) (*pb_auth.UserResponse, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*pb_auth.UserResponse), args.Error(1)
}

func TestProcessAlert_CooldownAndResolve(t *testing.T) {
	// Setup
	mockAuth := new(MockAuthServiceClient)
	mockMail := new(MockMailer)

	// Mock User Service response
	userResponse := &pb_auth.UserResponse{
		User: &pb_auth.User{
			Id:       123,
			Username: "testuser",
			Email:    "test@example.com",
		},
	}
	mockAuth.On("GetUser", mock.Anything, &pb_auth.GetUserRequest{Id: 123}).Return(userResponse, nil)

	// Set a long cooldown so we can test suppression easily
	cooldownPeriod = 1 * time.Hour
	cooldowns = make(map[int]time.Time)

	// 1. Send first alert (rule 1) - Expect email sent
	event1 := AlertEvent{
		AlertID:    1,
		RuleID:     10,
		UserID:     123,
		SensorID:   456,
		Message:    "High Temperature Alert",
		Value:      42.5,
		Timestamp:  time.Now(),
		IsResolved: false,
	}
	body1, _ := json.Marshal(event1)
	mockMail.On("SendAlertEmail", "test@example.com", mock.MatchedBy(func(e AlertEvent) bool {
		return e.RuleID == 10 && !e.IsResolved
	})).Return(nil).Once()

	processAlert(body1, mockAuth, mockMail)
	mockMail.AssertExpectations(t)

	// 2. Send second alert (rule 10) within cooldown - Expect suppressed (no email call)
	mockMail2 := new(MockMailer) // new mock to ensure no calls are made
	processAlert(body1, mockAuth, mockMail2)
	mockMail2.AssertNotCalled(t, "SendAlertEmail", mock.Anything, mock.Anything)

	// 3. Send resolved event for rule 10 - Expect email sent & cooldown cleared
	eventResolved := AlertEvent{
		AlertID:    0,
		RuleID:     10,
		UserID:     123,
		SensorID:   456,
		Message:    "High Temperature Resolved",
		Value:      25.0,
		Timestamp:  time.Now(),
		IsResolved: true,
	}
	bodyResolved, _ := json.Marshal(eventResolved)
	mockMail3 := new(MockMailer)
	mockMail3.On("SendAlertEmail", "test@example.com", mock.MatchedBy(func(e AlertEvent) bool {
		return e.RuleID == 10 && e.IsResolved
	})).Return(nil).Once()

	processAlert(bodyResolved, mockAuth, mockMail3)
	mockMail3.AssertExpectations(t)

	// 4. Send alert for rule 10 again - Expect email sent because cooldown was cleared by resolution
	mockMail4 := new(MockMailer)
	mockMail4.On("SendAlertEmail", "test@example.com", mock.MatchedBy(func(e AlertEvent) bool {
		return e.RuleID == 10 && !e.IsResolved
	})).Return(nil).Once()

	processAlert(body1, mockAuth, mockMail4)
	mockMail4.AssertExpectations(t)
}
