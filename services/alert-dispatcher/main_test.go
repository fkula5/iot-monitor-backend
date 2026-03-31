package main

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

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
