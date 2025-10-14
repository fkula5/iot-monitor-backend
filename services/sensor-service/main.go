package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"

	"github.com/joho/godotenv"

	"github.com/skni-kod/iot-monitor-backend/internal/database"
	"github.com/skni-kod/iot-monitor-backend/services/sensor-service/handlers"
	"github.com/skni-kod/iot-monitor-backend/services/sensor-service/services"
	"github.com/skni-kod/iot-monitor-backend/services/sensor-service/storage"
	"google.golang.org/grpc"
)

func getEnvOrFail(key string) string {
	value := os.Getenv(key)
	if value == "" {
		log.Fatalf("Environment variable %s is not set", key)
	}
	return value
}

func main() {
	if err := godotenv.Load("../../.env"); err != nil {
		log.Printf("Warning: Error loading .env file: %v", err)
	}

	host := getEnvOrFail("DB_HOST")
	port := getEnvOrFail("DB_PORT")
	user := getEnvOrFail("DB_USER")
	password := getEnvOrFail("DB_PASSWORD")
	dbname := getEnvOrFail("SENSOR_SERVICE_DB_DATABASE")
	grpcPort := getEnvOrFail("SENSOR_SERVICE_GRPC_PORT")

	db := database.NewSensorDB(host, port, user, password, dbname)
	defer db.Close()

	ctx := context.Background()

	if err := db.Schema.Create(ctx); err != nil {
		log.Fatalf("Failed to create schema: %v", err)
	}

	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", grpcPort))
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()

	sensorsStore := storage.NewSensorStorage(db)
	sensorsTypeStore := storage.NewSensorTypeStorage(db)
	sensorsService := services.NewSensorService(sensorsStore)
	sensorsTypeService := services.NewSensorTypeService(sensorsTypeStore)
	handlers.NewGrpcHandler(grpcServer, sensorsService, sensorsTypeService)

	log.Printf("Starting gRPC server on port %s...", grpcPort)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
