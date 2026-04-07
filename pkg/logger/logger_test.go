package logger

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestLoggerProduction(t *testing.T) {
	cfg := Config{
		Level:       "info",
		Environment: "production",
		ServiceName: "test-service",
	}
	err := Init(cfg)
	assert.NoError(t, err)
	assert.NotNil(t, globalLogger)
}
