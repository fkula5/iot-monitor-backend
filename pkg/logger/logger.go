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
	ServiceName string
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

	globalLogger = logger.With(
		zap.String("service", cfg.ServiceName),
		zap.String("environment", cfg.Environment),
	)
	return nil
}

func Sync() error {
	if globalLogger != nil {
		return globalLogger.Sync()
	}
	return nil
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
