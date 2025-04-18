package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"

	_ "github.com/joho/godotenv/autoload"

	"github.com/skni-kod/iot-monitor-backend/internal/database"
	"github.com/skni-kod/iot-monitor-backend/services/sensor-service/handlers"
	"github.com/skni-kod/iot-monitor-backend/services/sensor-service/services"
	"github.com/skni-kod/iot-monitor-backend/services/sensor-service/storage"
	"google.golang.org/grpc"
)

var (
	host     = os.Getenv("SENSOR_SERVICE_DB_HOST")
	port     = os.Getenv("SENSOR_SERVICE_DB_PORT")
	user     = os.Getenv("SENSOR_SERVICE_DB_USERNAME")
	password = os.Getenv("SENSOR_SERVICE_DB_PASSWORD")
	dbname   = os.Getenv("SENSOR_SERVICE_DB_DATABASE")
	grpcPort = os.Getenv("SENSOR_SERVICE_GRPC_PORT")
)

func main() {
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
