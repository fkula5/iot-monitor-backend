package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/skni-kod/iot-monitor-backend/internal/proto/data_service"
	"github.com/skni-kod/iot-monitor-backend/internal/proto/sensor_service"
	"github.com/skni-kod/iot-monitor-backend/services/data-generation-service/services"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func NewGrpcClient(addr string) (*grpc.ClientConn, error) {
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}

	return conn, err
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	dataProcAddr := strings.TrimSpace(os.Getenv("DATA_SERVICE_GRPC_ADDR"))
	if dataProcAddr == "" {
		log.Printf("⚠️  WARNING: DATA_SERVICE_GRPC_ADDR is empty, using default localhost:50053")
		dataProcAddr = "localhost:50053"
	}
	log.Printf("✅ Connecting to data processing service at: %s", dataProcAddr)

	sensorGrpcAddr := strings.TrimSpace(os.Getenv("SENSOR_SERVICE_GRPC_ADDR"))
	if sensorGrpcAddr == "" {
		log.Printf("⚠️  WARNING: SENSOR_SERVICE_GRPC_ADDR is empty, using default localhost:50052")
		sensorGrpcAddr = "localhost:50052"
	}
	log.Printf("✅ Connecting to sensor service at: %s", sensorGrpcAddr)

	sensorService, err := NewGrpcClient(sensorGrpcAddr)
	if err != nil {
		log.Fatalf("Failed to connect to sensor service: %v", err)
	}
	defer sensorService.Close()
	sensorClient := sensor_service.NewSensorServiceClient(sensorService)

	dataProcessingService, err := NewGrpcClient(dataProcAddr)
	if err != nil {
		log.Fatalf("Failed to connect to data processing service: %v", err)
	}
	defer dataProcessingService.Close()
	dataProcessingClient := data_service.NewDataServiceClient(dataProcessingService)

	generatorService := services.NewGeneratorService(sensorClient, dataProcessingClient)
	err = generatorService.Start(ctx)
	if err != nil {
		log.Fatalf("Failed to start generator service: %v", err)
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	sig := <-sigChan
	log.Printf("Received signal: %v, initiating shutdown", sig)

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := generatorService.Stop(); err != nil {
		log.Printf("Error stopping generator service: %v", err)
	}

	select {
	case <-shutdownCtx.Done():
		if shutdownCtx.Err() == context.DeadlineExceeded {
			log.Println("Shutdown timed out, forcing exit")
		}
	case <-time.After(100 * time.Millisecond):
		log.Println("Shutdown completed successfully")
	}

	log.Println("Data generator service shutdown complete")
}
