package main

import (
	"context"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/skni-kod/iot-monitor-backend/internal/proto/data_service"
	"github.com/skni-kod/iot-monitor-backend/internal/proto/sensor_service"
	"github.com/skni-kod/iot-monitor-backend/pkg/logger"
	"github.com/skni-kod/iot-monitor-backend/services/data-generation-service/services"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func NewGrpcClient(addr string) (*grpc.ClientConn, error) {
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.Fatal("did not connect: %v", zap.Error(err))
	}

	return conn, err
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	environment := os.Getenv("ENVIRONMENT")
	logLevel := os.Getenv("LOG_LEVEL")

	err := logger.Init(logger.Config{
		Level:       logLevel,
		Environment: environment,
		OutputPaths: []string{"stdout"},
	})
	if err != nil {
		logger.Fatal("Failed to initialize logger: %v", zap.Error(err))
	}
	defer logger.Sync()

	dataProcAddr := strings.TrimSpace(os.Getenv("DATA_SERVICE_GRPC_ADDR"))
	if dataProcAddr == "" {
		logger.Info("⚠️  WARNING: DATA_SERVICE_GRPC_ADDR is empty, using default localhost:50053")
		dataProcAddr = "localhost:50053"
	}
	logger.Info("✅ Connecting to data processing service at: %s", zap.String("address", dataProcAddr))

	sensorGrpcAddr := strings.TrimSpace(os.Getenv("SENSOR_SERVICE_GRPC_ADDR"))
	if sensorGrpcAddr == "" {
		logger.Info("⚠️  WARNING: SENSOR_SERVICE_GRPC_ADDR is empty, using default localhost:50052")
		sensorGrpcAddr = "localhost:50052"
	}
	logger.Info("✅ Connecting to sensor service at: %s", zap.String("address", sensorGrpcAddr))

	sensorService, err := NewGrpcClient(sensorGrpcAddr)
	if err != nil {
		logger.Fatal("Failed to connect to sensor service: %v", zap.Error(err))
	}
	defer sensorService.Close()
	sensorClient := sensor_service.NewSensorServiceClient(sensorService)

	dataProcessingService, err := NewGrpcClient(dataProcAddr)
	if err != nil {
		logger.Fatal("Failed to connect to data processing service: %v", zap.Error(err))
	}
	defer dataProcessingService.Close()
	dataProcessingClient := data_service.NewDataServiceClient(dataProcessingService)

	generatorService := services.NewGeneratorService(sensorClient, dataProcessingClient)
	err = generatorService.Start(ctx)
	if err != nil {
		logger.Fatal("Failed to start generator service: %v", zap.Error(err))
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	sig := <-sigChan

	logger.Info("Received signal, initiating shutdown", zap.String("signal", sig.String()))

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := generatorService.Stop(); err != nil {
		logger.Info("Error stopping generator service: %v", zap.Error(err))
	}

	select {
	case <-shutdownCtx.Done():
		if shutdownCtx.Err() == context.DeadlineExceeded {
			logger.Info("Shutdown timed out, forcing exit")
		}
	case <-time.After(100 * time.Millisecond):
		logger.Info("Shutdown completed successfully")
	}

	logger.Info("Data generator service shutdown complete")
}
