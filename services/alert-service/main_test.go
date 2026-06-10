package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"testing"
	"time"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/skni-kod/iot-monitor-backend/pkg/logger"
	"github.com/skni-kod/iot-monitor-backend/services/alert-service/ent"
	"github.com/skni-kod/iot-monitor-backend/services/alert-service/ent/enttest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	_ "modernc.org/sqlite"
)

type MockPublisher struct {
	mock.Mock
}

func (m *MockPublisher) PublishWithContext(ctx context.Context, exchange, key string, mandatory, immediate bool, msg amqp.Publishing) error {
	args := m.Called(ctx, exchange, key, mandatory, immediate, msg)
	return args.Error(0)
}

func TestMain(m *testing.M) {
	logger.Init(logger.Config{
		Level:       "info",
		Environment: "development",
		ServiceName: "alert-service-test",
		OutputPaths: []string{"stdout"},
	})
	m.Run()
}

func TestIsTriggered(t *testing.T) {
	tests := []struct {
		name          string
		conditionType string
		threshold     float64
		value         float64
		expected      bool
	}{
		{"GT Triggered", "GT", 25.0, 26.0, true},
		{"GT Not Triggered", "GT", 25.0, 24.0, false},
		{"LT Triggered", "LT", 10.0, 5.0, true},
		{"LT Not Triggered", "LT", 10.0, 15.0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rule := &ent.AlertRule{
				ConditionType: tt.conditionType,
				Threshold:     tt.threshold,
			}
			result := isTriggered(rule, tt.value)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestProcessMessage(t *testing.T) {
	db, err := sql.Open("sqlite", "file:ent?mode=memory&cache=shared&_pragma=foreign_keys(1)")
	if err != nil {
		t.Fatalf("failed opening connection to sqlite: %v", err)
	}
	drv := entsql.OpenDB(dialect.SQLite, db)
	client := enttest.NewClient(t, enttest.WithOptions(ent.Driver(drv)))
	defer client.Close()

	ctx := context.Background()

	_, err = client.AlertRule.Create().
		SetName("Temp Alert").
		SetSensorID(1).
		SetConditionType("GT").
		SetThreshold(30.0).
		SetUserID(100).
		SetIsEnabled(true).
		Save(ctx)
	assert.NoError(t, err)

	t.Run("Triggers and Saves Alert", func(t *testing.T) {
		mockPub := new(MockPublisher)
		
		data := SensorData{
			SensorID:  1,
			Value:     35.0,
			Timestamp: time.Now(),
		}
		body, _ := json.Marshal(data)

		mockPub.On("PublishWithContext", mock.Anything, "alerts_exchange", "", false, false, mock.MatchedBy(func(p amqp.Publishing) bool {
			var event AlertEvent
			json.Unmarshal(p.Body, &event)
			return event.Value == 35.0 && event.SensorID == 1
		})).Return(nil)

		processMessage(client, mockPub, body)

		alerts, err := client.Alert.Query().All(ctx)
		assert.NoError(t, err)
		assert.Len(t, alerts, 1)
		assert.Equal(t, 35.0, alerts[0].Value)
		assert.Contains(t, alerts[0].Message, "Temp Alert")

		mockPub.AssertExpectations(t)
	})

	t.Run("Does Not Trigger Below Threshold", func(t *testing.T) {
		mockPub := new(MockPublisher)
		
		data := SensorData{
			SensorID:  1,
			Value:     25.0,
			Timestamp: time.Now(),
		}
		body, _ := json.Marshal(data)

		processMessage(client, mockPub, body)

		count, _ := client.Alert.Query().Count(ctx)
		assert.Equal(t, 1, count)

		mockPub.AssertNotCalled(t, "PublishWithContext", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
	})
}
