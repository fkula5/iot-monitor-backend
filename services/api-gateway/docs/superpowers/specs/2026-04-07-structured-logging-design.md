# Design Spec: Structured JSON Logging Upgrade

**Date:** 2026-04-07
**Status:** Approved
**Topic:** Standardizing logger output for production searchability in a microservices environment.

## 1. Problem Statement
Currently, the logger in `pkg/logger/logger.go` uses `zap.NewProductionConfig()` for production and `zap.NewDevelopmentConfig()` for development. While production already outputs JSON, it lacks service-level metadata (e.g., which service generated the log) and doesn't explicitly standardize field names for optimal searchability in log aggregation tools.

## 2. Goals
- Ensure all production logs are structured JSON.
- Add a mandatory `service` field to identify the source service in the monorepo.
- Include the `environment` in every log entry.
- Maintain human-readable, colorized output for local development.
- Standardize timestamp and field names for indexing.

## 3. Architecture & Design

### 3.1. Logger Configuration
The `logger.Config` struct in `pkg/logger/logger.go` will be updated to include a `ServiceName` field.

```go
type Config struct {
    Level       string
    Environment string
    OutputPaths []string
    ServiceName string // New mandatory field for production identification
}
```

### 3.2. Structured Output (Production)
For `Environment == "production"`, the logger will be configured with:
- **Encoder:** `zapcore.NewJSONEncoder`
- **Time Format:** ISO8601 (`2006-01-02T15:04:05.000Z0700`)
- **Fixed Fields:** 
    - `service`: Value from `Config.ServiceName`
    - `environment`: Value from `Config.Environment`
- **Standard Keys:**
    - `timestamp`: The log time.
    - `level`: The log severity.
    - `caller`: File and line number.
    - `msg`: The log message.
    - `stacktrace`: Included for `Error` and `Fatal` levels.

### 3.3. Human-Readable Output (Development)
For other environments (e.g., `development`, `local`), the logger will continue to use a console encoder with:
- **Encoder:** `zapcore.NewConsoleEncoder`
- **Coloring:** Enabled for levels.
- **Time Format:** ISO8601 (or standard development time).

## 4. Implementation Details

### 4.1. `pkg/logger/logger.go` Changes
- Update `Init(cfg Config)` to use `zap.NewProductionEncoderConfig()` and customize its keys.
- Use `logger.With(zap.String("service", cfg.ServiceName), zap.String("environment", cfg.Environment))` during initialization to attach global fields.

### 4.2. Service Integration
Each service (e.g., `api-gateway`, `auth`, `sensor-service`) must update its `main.go` to pass its specific name to `logger.Init`.

Example for `api-gateway`:
```go
logger.Init(logger.Config{
    Level:       logLevel,
    Environment: environment,
    ServiceName: "api-gateway", // Explicit service identification
    OutputPaths: []string{"stdout"},
})
```

## 5. Testing & Validation
- **Unit Test:** Add a test in `pkg/logger/logger_test.go` (if it exists, or create it) to verify JSON output format and presence of mandatory fields in production mode.
- **Manual Verification:** Run a service with `ENVIRONMENT=production` and verify logs are valid JSON with the `service` and `environment` fields.
