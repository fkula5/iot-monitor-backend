package main

import (
	"context"
	"fmt"
	"net"
	"os"

	"go.uber.org/zap"
	"google.golang.org/grpc"

	"github.com/skni-kod/iot-monitor-backend/internal/database"
	"github.com/skni-kod/iot-monitor-backend/pkg/logger"
	"github.com/skni-kod/iot-monitor-backend/services/sensor-service/ent"
	"github.com/skni-kod/iot-monitor-backend/services/sensor-service/handlers"
	"github.com/skni-kod/iot-monitor-backend/services/sensor-service/services"
	"github.com/skni-kod/iot-monitor-backend/services/sensor-service/storage"
)

func getEnvOrFail(key string) string {
	value := os.Getenv(key)
	if value == "" {
		logger.Fatal("Environment variable not set", zap.String("key", key))
	}
	return value
}

func main() {
	environment := os.Getenv("ENVIRONMENT")
	if environment == "" {
		environment = "development"
	}

	logLevel := os.Getenv("LOG_LEVEL")
	if logLevel == "" {
		logLevel = "info"
	}

	err := logger.Init(logger.Config{
		Level:       logLevel,
		Environment: environment,
		OutputPaths: []string{"stdout"},
		ServiceName: "sensor-service",
	})
	if err != nil {
		logger.Fatal("Failed to initialize logger", zap.Error(err))
	}
	defer logger.Sync()

	logger.Info("Starting Sensor Service",
		zap.String("environment", environment),
		zap.String("log_level", logLevel),
	)

	host := getEnvOrFail("DB_HOST")
	port := getEnvOrFail("DB_PORT")
	user := getEnvOrFail("SENSOR_SERVICE_DB_USER")
	password := getEnvOrFail("SENSOR_SERVICE_DB_PASSWORD")
	dbname := getEnvOrFail("SENSOR_SERVICE_DB_NAME")
	grpcPort := getEnvOrFail("SENSOR_SERVICE_GRPC_PORT")

	logger.Info("Database configuration loaded",
		zap.String("host", host),
		zap.String("port", port),
		zap.String("database", dbname),
		zap.String("user", user),
	)

	drv, err := database.NewDriver(host, port, user, password, dbname)
	if err != nil {
		logger.Fatal("Failed to create database driver", zap.Error(err))
	}
	client := ent.NewClient(ent.Driver(drv))
	defer client.Close()

	ctx := context.Background()

	logger.Info("Creating database schema")
	if err := client.Schema.Create(ctx); err != nil {
		logger.Fatal("Failed to create schema", zap.Error(err))
	}
	logger.Info("Database schema created successfully")

	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", grpcPort))
	if err != nil {
		logger.Fatal("Failed to listen on port",
			zap.String("port", grpcPort),
			zap.Error(err),
		)
	}

	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(logger.UnaryServerInterceptor()),
	)

	sensorsStore := storage.NewSensorStorage(client)
	sensorsTypeStore := storage.NewSensorTypeStorage(client)
	sensorsGroupStore := storage.NewSensorGroupStorage(client)
	sensorsService := services.NewSensorService(sensorsStore)
	sensorsTypeService := services.NewSensorTypeService(sensorsTypeStore)
	sensorsGroupService := services.NewSensorGroupService(sensorsGroupStore)

	handlers.NewGrpcHandler(grpcServer, sensorsService, sensorsTypeService, sensorsGroupService)
	logger.Info("Starting Sensor Service gRPC server on port", zap.String("port", grpcPort))
	if err := grpcServer.Serve(lis); err != nil {
		logger.Fatal("Failed to serve", zap.Error(err))
	}
}
