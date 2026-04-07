# Structured JSON Logging Upgrade Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Upgrade the logger to output structured JSON in production with service-specific metadata.

**Architecture:** Update `pkg/logger` to accept a `ServiceName` and `Environment`. In production, use a JSON encoder with standardized fields. In development, maintain human-readable console output.

**Tech Stack:** Go, Zap (go.uber.org/zap)

---

### Task 1: Update pkg/logger/logger.go

**Files:**
- Modify: `pkg/logger/logger.go`
- Create: `pkg/logger/logger_test.go`

- [ ] **Step 1: Update Config and Init in pkg/logger/logger.go**

```go
package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var globalLogger *zap.Logger

type Config struct {
	Level       string
	Environment string
	OutputPaths []string
	ServiceName string // New field
}

func Init(cfg Config) error {
	var zapConfig zap.Config
	var encoderConfig zapcore.EncoderConfig

	if cfg.Environment == "production" {
		zapConfig = zap.NewProductionConfig()
		encoderConfig = zap.NewProductionEncoderConfig()
		encoderConfig.TimeKey = "timestamp"
		encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
		zapConfig.EncoderConfig = encoderConfig
	} else {
		zapConfig = zap.NewDevelopmentConfig()
		encoderConfig = zap.NewDevelopmentEncoderConfig()
		encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		zapConfig.EncoderConfig = encoderConfig
	}

	level := zapcore.InfoLevel
	if cfg.Level != "" {
		if err := level.UnmarshalText([]byte(cfg.Level)); err != nil {
			return err
		}
	}
	zapConfig.Level = zap.NewAtomicLevelAt(level)

	if len(cfg.OutputPaths) > 0 {
		zapConfig.OutputPaths = cfg.OutputPaths
	}

	logger, err := zapConfig.Build(zap.AddCallerSkip(1))
	if err != nil {
		return err
	}

	// Add service and environment fields globally
	globalLogger = logger.With(
		zap.String("service", cfg.ServiceName),
		zap.String("environment", cfg.Environment),
	)
	return nil
}
// ... rest of the file remains same ...
```

- [ ] **Step 2: Create pkg/logger/logger_test.go**

```go
package logger

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoggerProduction(t *testing.T) {
	// We'll use a custom Core to capture output or just check if it initializes without error
	cfg := Config{
		Level:       "info",
		Environment: "production",
		ServiceName: "test-service",
	}
	err := Init(cfg)
	assert.NoError(t, err)
	assert.NotNil(t, globalLogger)
}
```

- [ ] **Step 3: Run tests**

Run: `go test ./pkg/logger/...`
Expected: PASS

- [ ] **Step 4: Commit**

```bash
git add pkg/logger/logger.go pkg/logger/logger_test.go
git commit -m "feat: upgrade logger to support structured JSON and service metadata"
```

### Task 2: Update All Services

**Files:**
- Modify: `services/api-gateway/main.go`
- Modify: `services/auth/main.go`
- Modify: `services/sensor-service/main.go`
- Modify: `services/alert-service/main.go`
- Modify: `services/data-processing/main.go`
- Modify: `services/data-generation-service/main.go`
- Modify: `services/alert-dispatcher/main.go`

- [ ] **Step 1: Update api-gateway**

```go
	err := logger.Init(logger.Config{
		Level:       logLevel,
		Environment: environment,
		ServiceName: "api-gateway",
		OutputPaths: []string{"stdout"},
	})
```

- [ ] **Step 2: Update auth**

```go
	err := logger.Init(logger.Config{
		Level:       logLevel,
		Environment: environment,
		ServiceName: "auth",
		OutputPaths: []string{"stdout"},
	})
```

- [ ] **Step 3: Update sensor-service**

```go
	err := logger.Init(logger.Config{
		Level:       logLevel,
		Environment: environment,
		ServiceName: "sensor-service",
		OutputPaths: []string{"stdout"},
	})
```

- [ ] **Step 4: Update alert-service**

```go
	err := logger.Init(logger.Config{
		Level:       logLevel,
		Environment: environment,
		ServiceName: "alert-service",
		OutputPaths: []string{"stdout"},
	})
```

- [ ] **Step 5: Update data-processing**

```go
	err := logger.Init(logger.Config{
		Level:       logLevel,
		Environment: environment,
		ServiceName: "data-processing",
		OutputPaths: []string{"stdout"},
	})
```

- [ ] **Step 6: Update data-generation-service**

```go
	err := logger.Init(logger.Config{
		Level:       logLevel,
		Environment: environment,
		ServiceName: "data-generation-service",
		OutputPaths: []string{"stdout"},
	})
```

- [ ] **Step 7: Update alert-dispatcher**

```go
	err := logger.Init(logger.Config{
		Level:       logLevel,
		Environment: environment,
		ServiceName: "alert-dispatcher",
		OutputPaths: []string{"stdout"},
	})
```

- [ ] **Step 8: Verify all services build**

Run: `go build ./services/...`
Expected: PASS

- [ ] **Step 9: Commit**

```bash
git add services/*/main.go
git commit -m "feat: update all services to use structured logging with service names"
```
