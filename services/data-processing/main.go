package main

import (
	"database/sql"
	"fmt"
	"log"
	"net"
	"os"

	_ "github.com/lib/pq"

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

	dbHost := os.Getenv("DB_HOST")
	if dbHost == "" {
		dbHost = "localhost"
	}

	dbPort := os.Getenv("DB_PORT")
	if dbPort == "" {
		dbPort = "5432"
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

	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		dbHost, dbPort, dbUser, dbPass, dbName)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("❌ Failed to open database connection: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("❌ Failed to ping database: %v", err)
	}
	log.Println("✅ Database connection established")

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
