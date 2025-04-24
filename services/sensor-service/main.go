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

	host := getEnvOrFail("SENSOR_SERVICE_DB_HOST")
	port := getEnvOrFail("SENSOR_SERVICE_DB_PORT")
	user := getEnvOrFail("SENSOR_SERVICE_DB_USERNAME")
	password := getEnvOrFail("SENSOR_SERVICE_DB_PASSWORD")
	dbname := getEnvOrFail("SENSOR_SERVICE_DB_DATABASE")
	grpcPort := getEnvOrFail("SENSOR_SERVICE_GRPC_PORT")

	db := database.New(host, port, user, password, dbname)

	defer db.Client.Close()

	ctx := context.Background()

	if err := db.Client.Schema.Create(ctx); err != nil {
		log.Fatalf("Failed to create schema: %v", err)
	}

	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", grpcPort))
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()

	sensorsStore := storage.NewSensorStorage(db.Client)
	sensorsService := services.NewSensorService(sensorsStore)
	handlers.NewGrpcHandler(grpcServer, sensorsService)

	log.Printf("Starting gRPC server on port %s...", grpcPort)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
