package logger

import (
	"context"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var globalLogger *zap.Logger

type Config struct {
	Level       string
	Environment string
	OutputPaths []string
	ServiceName string
}

func Init(cfg Config) error {
	var zapConfig zap.Config

	if cfg.Environment == "production" {
		zapConfig = zap.NewProductionConfig()
		zapConfig.EncoderConfig.TimeKey = "timestamp"
		zapConfig.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	} else {
		zapConfig = zap.NewDevelopmentConfig()
		zapConfig.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
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

	if cfg.ServiceName != "" {
		logger = logger.With(zap.String("service", cfg.ServiceName))
	}

	globalLogger = logger
	return nil
}

func Sync() error {
	if globalLogger != nil {
		return globalLogger.Sync()
	}
	return nil
}

// L returns the global logger instance
func L() *zap.Logger {
	return globalLogger
}

// With returns a logger with the given fields
func With(fields ...zap.Field) *zap.Logger {
	return globalLogger.With(fields...)
}

// Named returns a logger with a new name appended to the current name
func Named(name string) *zap.Logger {
	return globalLogger.Named(name)
}

func Debug(msg string, fields ...zap.Field) {
	globalLogger.Debug(msg, fields...)
}

func Info(msg string, fields ...zap.Field) {
	globalLogger.Info(msg, fields...)
}

func Warn(msg string, fields ...zap.Field) {
	globalLogger.Warn(msg, fields...)
}

func Error(msg string, fields ...zap.Field) {
	globalLogger.Error(msg, fields...)
}

func Fatal(msg string, fields ...zap.Field) {
	globalLogger.Fatal(msg, fields...)
}

// FromContext returns a logger from the context if available, otherwise returns global logger
func FromContext(ctx context.Context) *zap.Logger {
	// This can be expanded when we implement tracing/request IDs in context
	return globalLogger
}
