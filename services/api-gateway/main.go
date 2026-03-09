package main

import (
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	amqp "github.com/rabbitmq/amqp091-go"
	httpSwagger "github.com/swaggo/http-swagger/v2"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/skni-kod/iot-monitor-backend/internal/proto/alert_service"
	"github.com/skni-kod/iot-monitor-backend/internal/proto/auth"
	"github.com/skni-kod/iot-monitor-backend/internal/proto/data_service"
	"github.com/skni-kod/iot-monitor-backend/internal/proto/sensor_service"
	"github.com/skni-kod/iot-monitor-backend/pkg/logger"
	_ "github.com/skni-kod/iot-monitor-backend/services/api-gateway/docs"
	"github.com/skni-kod/iot-monitor-backend/services/api-gateway/handlers"
	"github.com/skni-kod/iot-monitor-backend/services/api-gateway/routes"
)

func NewGrpcClient(addr string) (*grpc.ClientConn, error) {
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.Fatal("did not connect", zap.Error(err))
	}

	return conn, err
}

// @title					IoT Monitor API
// @version					1.0
// @description				API dla systemu monitorowania IoT.
//
// @contact.name				API Support
// @contact.url				https://github.com/skni-kod/iot-monitor-backend
//
// @license.name				MIT
// @license.url				https://opensource.org/licenses/MIT
//
// @host						localhost:8080
// @BasePath					/

// @securityDefinitions.apikey	ApiKeyAuth
// @in							header
// @name						Authorization
// @description				Wprowadź token JWT w formacie 'Bearer {token}'.
func main() {
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

	rabbitURL := os.Getenv("RABBITMQ_URL")
	if rabbitURL == "" {
		rabbitURL = "amqp://guest:guest@localhost:5672/"
	}

	var alertMsgs <-chan amqp.Delivery
	conn, err := amqp.Dial(rabbitURL)
	if err != nil {
		logger.Warn("Failed to connect to RabbitMQ, alerts will be disabled", zap.Error(err))
	} else {
		ch, err := conn.Channel()
		if err != nil {
			logger.Warn("Failed to open channel to RabbitMQ, alerts will be disabled", zap.Error(err))
		} else {
			q, err := ch.QueueDeclare(
				"alerts", // name
				true,     // durable
				false,    // delete when unused
				false,    // exclusive
				false,    // no-wait
				nil,      // arguments
			)
			if err != nil {
				logger.Warn("Failed to declare alerts queue, alerts will be disabled", zap.Error(err))
			} else {
				err = ch.QueueBind(
					q.Name,            // queue name
					"",                // routing key
					"alerts_exchange", // exchange
					false,             // no-wait
					nil,               // arguments
				)
				if err != nil {
					logger.Warn("Failed to bind alerts queue, alerts will be disabled", zap.Error(err))
				}

				alertMsgs, err = ch.Consume(
					q.Name, // queue
					"",     // consumer
					true,   //
					false,  // exclusive
					false,  // no-local
					false,  // no-wait
					nil,    // args
				)
				if err != nil {
					logger.Warn("Failed to consume from alerts queue, alerts will be disabled", zap.Error(err))
				}
			}
		}
	}

	authGrpcAddr := strings.TrimSpace(os.Getenv("AUTH_SERVICE_GRPC_ADDR"))
	if authGrpcAddr == "" {
		logger.Warn("WARNING: AUTH_SERVICE_GRPC_ADDR is empty, using default localhost:50051")
		authGrpcAddr = "localhost:50051"
	}
	logger.Info("Connecting to auth service", zap.String("address", authGrpcAddr))

	sensorGrpcAddr := strings.TrimSpace(os.Getenv("SENSOR_SERVICE_GRPC_ADDR"))
	if sensorGrpcAddr == "" {
		logger.Warn("WARNING: SENSOR_SERVICE_GRPC_ADDR is empty, using default localhost:50052")
		sensorGrpcAddr = "localhost:50052"
	}
	logger.Info("Connecting to sensor service at", zap.String("address", sensorGrpcAddr))

	alertGrpcAddr := strings.TrimSpace(os.Getenv("ALERT_SERVICE_GRPC_ADDR"))
	if alertGrpcAddr == "" {
		logger.Warn("WARNING: ALERT_SERVICE_GRPC_ADDR is empty, using default localhost:50054")
		alertGrpcAddr = "localhost:50054"
	}
	logger.Info("Connecting to alert service at", zap.String("address", alertGrpcAddr))

	sensorService, err := NewGrpcClient(sensorGrpcAddr)
	if err != nil {
		logger.Fatal("Failed to connect to sensor service", zap.Error(err))
	}
	defer sensorService.Close()
	sensorClient := sensor_service.NewSensorServiceClient(sensorService)

	authService, err := NewGrpcClient(authGrpcAddr)
	if err != nil {
		logger.Fatal("Failed to connect to auth service", zap.Error(err))
	}
	defer authService.Close()
	authClient := auth.NewAuthServiceClient(authService)

	alertService, err := NewGrpcClient(alertGrpcAddr)
	if err != nil {
		logger.Fatal("Failed to connect to alert service", zap.Error(err))
	}
	defer alertService.Close()
	alertClient := alert_service.NewAlertServiceClient(alertService)

	dataProcAddr := strings.TrimSpace(os.Getenv("DATA_SERVICE_GRPC_ADDR"))
	if dataProcAddr == "" {
		logger.Warn("Warning: DATA_SERVICE_GRPC_ADDR is empty, using default localhost:50053")
		dataProcAddr = "localhost:50053"
	}
	logger.Info("Connecting to data processing service at", zap.String("address", dataProcAddr))

	dataProcService, err := NewGrpcClient(dataProcAddr)
	if err != nil {
		logger.Fatal("Failed to connect to data processing service", zap.Error(err))
	}
	defer dataProcService.Close()
	dataProcClient := data_service.NewDataServiceClient(dataProcService)

	apiGatewayPort := strings.TrimSpace(os.Getenv("API_GATEWAY_PORT"))
	if apiGatewayPort == "" {
		apiGatewayPort = "8080"
	}

	corsAllowedOrigins := strings.Split(os.Getenv("CORS_ALLOWED_ORIGINS"), ",")
	if len(corsAllowedOrigins) == 0 || (len(corsAllowedOrigins) == 1 && corsAllowedOrigins[0] == "") {
		corsAllowedOrigins = []string{"http://localhost:5173"}
	}

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Timeout(60 * time.Second))
	r.Use(middleware.Recoverer)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   corsAllowedOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
	})

	r.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL("/swagger/doc.json"),
	))

	sensorHandler := handlers.NewSensorHandler(sensorClient)
	sensorTypeHandler := handlers.NewSensorTypeHandler(sensorClient)
	sensorGroupHandler := handlers.NewSensorGroupHandler(sensorClient)
	authHandler := handlers.NewAuthHandler(authClient)
	alertHandler := handlers.NewAlertHandler(alertClient)
	dataHandler := handlers.NewWebSocketHandler(dataProcClient, sensorClient, alertMsgs)

	apiRouter := chi.NewRouter()
	apiRouter.Use(middleware.RequestID)
	apiRouter.Use(middleware.RealIP)

	routes.SetupSensorRoutes(apiRouter, sensorHandler)
	routes.SetupSensorTypeRoutes(apiRouter, sensorTypeHandler)
	routes.SetupSensorGroupRoutes(apiRouter, sensorGroupHandler)
	routes.SetupAlertRoutes(apiRouter, alertHandler)
	routes.SetupDataRoutes(apiRouter, dataHandler)

	r.Mount("/api", apiRouter)

	authRouter := chi.NewRouter()
	routes.SetupAuthRoutes(authRouter, authHandler)
	r.Mount("/auth", authRouter)

	chi.Walk(r, func(method, route string, handler http.Handler, middlewares ...func(http.Handler) http.Handler) error {
		logger.Info("Registered route", zap.String("method", method), zap.String("route", route))
		return nil
	})

	logger.Info("Starting API gateway server on port", zap.String("port", apiGatewayPort))

	err = http.ListenAndServe(":"+apiGatewayPort, r)
	if err != nil {
		logger.Fatal("Server failed to start", zap.Error(err))
	}
}
