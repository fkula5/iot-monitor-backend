package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/skni-kod/iot-monitor-backend/internal/proto/api"
	"github.com/skni-kod/iot-monitor-backend/services/data-generation-service/services"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	conn, err := grpc.NewClient(":50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect to sensor service: %v", err)
	}

	defer conn.Close()

	sensorClient := api.NewSensorServiceClient(conn)
	generatorService := services.NewGeneratorService(sensorClient)
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
