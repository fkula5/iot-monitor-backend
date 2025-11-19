package main

import (
	"database/sql"
	"fmt"
	"log"
	"net"
	"os"

	pb_sensor "github.com/skni-kod/iot-monitor-backend/internal/proto/sensor_service"
	"github.com/skni-kod/iot-monitor-backend/services/data-processing/handlers"
	"github.com/skni-kod/iot-monitor-backend/services/data-processing/storage"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	grpcPort := os.Getenv("DATA_SERVICE_GRPC_PORT")
	if grpcPort == "" {
		grpcPort = "50053"
	}

	sensorServiceAddr := os.Getenv("SENSOR_SERVICE_GRPC_ADDR")
	if sensorServiceAddr == "" {
		sensorServiceAddr = "localhost:50052"
	}

	dbName := os.Getenv("DATA_SERVICE_DB_NAME")
	if dbName == "" {
		dbName = "sensor_readings_db"
	}

	dbUser := os.Getenv("DATA_SERVICE_DB_USER")
	if dbUser == "" {
		dbUser = "data_user"
	}

	dbPass := os.Getenv("DATA_SERVICE_DB_PASSWORD")
	if dbPass == "" {
		dbPass = "datapassword"
	}

	connStr := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		"localhost", "5432", dbUser, dbPass, dbName)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	sensorConn, err := grpc.NewClient(sensorServiceAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect to sensor service: %v", err)
	}
	defer sensorConn.Close()

	sensorClient := pb_sensor.NewSensorServiceClient(sensorConn)

	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", grpcPort))
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	dataStore := storage.NewTimescaleStorage(db)
	handlers.NewDataGrpcHandler(grpcServer, dataStore, sensorClient)

	log.Printf("Starting Data Service gRPC server on port %s...", grpcPort)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
