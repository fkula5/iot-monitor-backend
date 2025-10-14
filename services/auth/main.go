package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"

	"github.com/joho/godotenv"
	"github.com/skni-kod/iot-monitor-backend/internal/database"
	"github.com/skni-kod/iot-monitor-backend/services/auth/handlers"
	"github.com/skni-kod/iot-monitor-backend/services/auth/services"
	"github.com/skni-kod/iot-monitor-backend/services/auth/storage"
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
	dbname := "users"

	db := database.NewAuthDB(host, port, user, password, dbname)
	defer db.Close()

	ctx := context.Background()

	if err := db.Schema.Create(ctx); err != nil {
		log.Fatalf("Failed to create schema: %v", err)
	}

	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", "50052"))
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()

	userStorage := storage.NewUserStorage(db)
	authService := services.NewAuthService(userStorage)
	handlers.NewGrpcHandler(grpcServer, authService)

	log.Printf("Starting gRPC server on :50052")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to start gRPC server: %v", err)
	}
}
