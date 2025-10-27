package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/joho/godotenv"
	"github.com/skni-kod/iot-monitor-backend/internal/proto/auth"
	"github.com/skni-kod/iot-monitor-backend/internal/proto/sensor_service"
	_ "github.com/skni-kod/iot-monitor-backend/services/api-gateway/docs"
	"github.com/skni-kod/iot-monitor-backend/services/api-gateway/handlers"
	"github.com/skni-kod/iot-monitor-backend/services/api-gateway/routes"
	httpSwagger "github.com/swaggo/http-swagger/v2"
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

// @title           IoT Monitor API
// @version         1.0
// @description     API Gateway dla systemu monitorowania IoT.
//
// @contact.name   API Support
// @contact.url    https://github.com/skni-kod/iot-monitor-backend
// @contact.email  twoj.email@example.com
//
// @license.name  MIT
// @license.url   https://opensource.org/licenses/MIT
//
// @host            localhost:8080
// @BasePath        /api
//
// @securityDefinitions.apikey  BearerAuth
// @in                          header
// @name                        Authorization
// @description                 "Wpisz 'Bearer ' (ze spacją), a następnie swój token JWT. Przykład: 'Bearer eyJhbGciOi...'"
func main() {
	if err := godotenv.Load("../../.env"); err != nil {
		log.Printf("Warning: Error loading .env file: %v", err)
	}

	sensorGrpcPort := strings.TrimSpace(os.Getenv("SENSOR_SERVICE_GRPC_PORT"))
	if sensorGrpcPort == "" {
		sensorGrpcPort = "50052"
	}
	sensorGrpcAddr := fmt.Sprintf(":%s", sensorGrpcPort)

	authGrpcPort := strings.TrimSpace(os.Getenv("AUTH_SERVICE_GRPC_PORT"))
	if authGrpcPort == "" {
		authGrpcPort = "50051"
	}
	authGrpcAddr := fmt.Sprintf(":%s", authGrpcPort)

	sensorService, err := NewGrpcClient(sensorGrpcAddr)
	if err != nil {
		log.Fatalf("Failed to connect to sensor service: %v", err)
	}
	defer sensorService.Close()
	sensorClient := sensor_service.NewSensorServiceClient(sensorService)

	authService, err := NewGrpcClient(authGrpcAddr)
	if err != nil {
		log.Fatalf("Failed to connect to auth service: %v", err)
	}
	defer authService.Close()
	authClient := auth.NewAuthServiceClient(authService)

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Timeout(60 * time.Second))
	r.Use(middleware.Recoverer)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3001"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
	})

	r.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL(""), // The url pointing to API definition
	))

	sensorHandler := handlers.NewSensorHandler(sensorClient)
	sensorTypeHandler := handlers.NewSensorTypeHandler(sensorClient)
	authHandler := handlers.NewAuthHandler(authClient)

	apiRouter := chi.NewRouter()
	apiRouter.Use(middleware.RequestID)
	apiRouter.Use(middleware.RealIP)

	routes.SetupSensorRoutes(apiRouter, sensorHandler)
	routes.SetupSensorTypeRoutes(apiRouter, sensorTypeHandler)

	r.Mount("/api", apiRouter)

	authRouter := chi.NewRouter()
	routes.SetupAuthRoutes(authRouter, authHandler)
	r.Mount("/auth", authRouter)

	chi.Walk(r, func(method, route string, handler http.Handler, middlewares ...func(http.Handler) http.Handler) error {
		log.Printf("Registered route: %s %s", method, route)
		return nil
	})

	log.Println("Starting API gateway server on :3000")

	err = http.ListenAndServe(":3000", r)
	if err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
