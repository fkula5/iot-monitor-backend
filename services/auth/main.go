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
	"github.com/skni-kod/iot-monitor-backend/services/auth/handlers"
	"github.com/skni-kod/iot-monitor-backend/services/auth/services"
	"github.com/skni-kod/iot-monitor-backend/services/auth/storage"
)

func getEnvOrFail(key string) string {
	value := os.Getenv(key)
	if value == "" {
		logger.Fatal("Environment variable is not set", zap.String("key", key))
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
	})

	host := getEnvOrFail("DB_HOST")
	port := getEnvOrFail("DB_PORT")
	user := getEnvOrFail("AUTH_SERVICE_DB_USER")
	password := getEnvOrFail("AUTH_SERVICE_DB_PASSWORD")
	dbname := getEnvOrFail("AUTH_SERVICE_DB_NAME")
	grpcPort := getEnvOrFail("AUTH_SERVICE_GRPC_PORT")

	if err != nil {
		logger.Fatal("Failed to initialize logger", zap.Error(err))
	}
	defer logger.Sync()

	db := database.NewAuthDB(host, port, user, password, dbname)
	defer db.Close()

	ctx := context.Background()

	if err := db.Schema.Create(ctx); err != nil {
		logger.Fatal("Failed to create schema", zap.Error(err))
	}

	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", grpcPort))
	if err != nil {
		logger.Fatal("Failed to listen", zap.Error(err))
	}

	grpcServer := grpc.NewServer()

	userStorage := storage.NewUserStorage(db)
	authService := services.NewAuthService(userStorage)
	handlers.NewGrpcHandler(grpcServer, authService)

	logger.Info("gRPC server starting", zap.String("port", grpcPort))
	if err := grpcServer.Serve(lis); err != nil {
		logger.Fatal("Failed to serve", zap.Error(err))
	}
}
