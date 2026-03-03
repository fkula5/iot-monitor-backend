package main

import (
	"database/sql"
	"fmt"
	"net"
	"os"

	_ "github.com/lib/pq"
	"go.uber.org/zap"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	amqp "github.com/rabbitmq/amqp091-go"

	pb_sensor "github.com/skni-kod/iot-monitor-backend/internal/proto/sensor_service"
	"github.com/skni-kod/iot-monitor-backend/pkg/logger"
	"github.com/skni-kod/iot-monitor-backend/services/data-processing/handlers"
	"github.com/skni-kod/iot-monitor-backend/services/data-processing/storage"
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

	rabbitMQURL := os.Getenv("RABBITMQ_URL")
	if rabbitMQURL == "" {
		rabbitMQURL = "amqp://guest:guest@localhost:5672/"
	}

	environment := os.Getenv("ENVIRONMENT")
	logLevel := os.Getenv("LOG_LEVEL")

	err := logger.Init(logger.Config{
		Level:       logLevel,
		Environment: environment,
		OutputPaths: []string{"stdout"},
	})
	if err != nil {
		logger.Fatal("Failed to initialize logger", zap.Error(err))
	}
	defer logger.Sync()

	conn, err := amqp.Dial(rabbitMQURL)
	if err != nil {
		logger.Fatal("Failed to connect to RabbitMQ", zap.Error(err))
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		logger.Fatal("Failed to open a channel", zap.Error(err))
	}
	defer ch.Close()

	err = ch.ExchangeDeclare(
		"readings_exchange", // name
		"fanout",            // type
		true,                // durable
		false,               // auto-deleted
		false,               // internal
		false,               // no-wait
		nil,                 // arguments
	)
	if err != nil {
		logger.Fatal("Failed to declare exchange", zap.Error(err))
	}

	q, err := ch.QueueDeclare(
		"readings_queue", // name
		false,            // durable
		true,             // delete when unused
		true,             // exclusive
		false,            // no-wait
		nil,              // arguments
	)
	if err != nil {
		logger.Fatal("Failed to declare queue", zap.Error(err))
	}

	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		dbHost, dbPort, dbUser, dbPass, dbName)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		logger.Fatal("Failed to open database connection", zap.Error(err))
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		logger.Fatal("Failed to ping database", zap.Error(err))
	}
	logger.Info("Database connection established")

	sensorConn, err := grpc.NewClient(sensorServiceAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.Fatal("Failed to connect to sensor service", zap.Error(err))
	}
	defer sensorConn.Close()

	sensorClient := pb_sensor.NewSensorServiceClient(sensorConn)

	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", grpcPort))
	if err != nil {
		logger.Fatal("Failed to listen on port",
			zap.String("port", grpcPort),
			zap.Error(err),
		)
	}

	grpcServer := grpc.NewServer()
	dataStore := storage.NewTimescaleStorage(db)
	handlers.NewDataGrpcHandler(grpcServer, dataStore, sensorClient, ch, q)

	logger.Info("Starting Data Service gRPC server on port", zap.String("port", grpcPort))
	if err := grpcServer.Serve(lis); err != nil {
		logger.Fatal("Failed to serve", zap.Error(err))
	}
}
