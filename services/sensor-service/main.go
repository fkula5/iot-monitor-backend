package main

import (
	"context"
	"fmt"
	"log"
	"math/rand/v2"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

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

	sensor, err := sensorsStore.Get(ctx, 1)
	if err != nil {
		log.Printf("error getting sensor: %v", err)
	}

	lastValue := sensor.Edges.Type.MinValue + rand.Float64()*(sensor.Edges.Type.MaxValue-sensor.Edges.Type.MinValue)

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		log.Printf("Starting gRPC server on port %s...", grpcPort)
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("Failed to serve: %v", err)
		}
		wg.Done()
	}()

	ctx, cancel := context.WithCancel(context.Background())
	wg.Add(1)
	go func() {
		defer wg.Done()
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:

				minValue := sensor.Edges.Type.MinValue
				maxValue := sensor.Edges.Type.MaxValue

				if minValue == maxValue {
					minValue = 0.0
					maxValue = 100.0
				}

				variation := rand.NormFloat64() * 0.02

				newValue := lastValue * (1 + variation)

				midPoint := (minValue + maxValue) / 2
				drift := (midPoint - newValue) * 0.01 * rand.Float64()
				newValue += drift

				if newValue < minValue {
					newValue = minValue
				}
				if newValue > maxValue {
					newValue = maxValue
				}

				lastValue = newValue

				log.Printf("Mock sensor data: %f", lastValue)
			case <-ctx.Done():
				log.Println("Data generator shutting down")
				return
			}
		}
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	sig := <-sigChan
	log.Printf("Received signal: %v, initiating shutdown", sig)

	cancel()
	grpcServer.Stop()

	wg.Wait()
	log.Println("Service shutdown complete")
}
