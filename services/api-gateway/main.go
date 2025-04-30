package main

import (
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/skni-kod/iot-monitor-backend/internal/proto/api"
	"github.com/skni-kod/iot-monitor-backend/internal/routes"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func NewGrpcClient(addr string) *grpc.ClientConn {
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}

	return conn
}

func main() {
	conn := NewGrpcClient(":50051")
	defer conn.Close()

	sensorClient := api.NewSensorServiceClient(conn)

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Timeout(60 * time.Second))
	r.Use(middleware.Recoverer)

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
	})

	r.Patch("/active/{id}", func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
		defer cancel()

		idStr := chi.URLParam(r, "id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			http.Error(w, "Invalid sensor ID", http.StatusBadRequest)
			return
		}

		res, err := h.client.GetSensor(ctx, &api.GetSensorRequest{Id: int32(id)})
	}

	apiRouter := chi.NewRouter()

	apiRouter.Use(middleware.RequestID)
	apiRouter.Use(middleware.RealIP)

	routes.SetupSensorRoutes(apiRouter, sensorClient)

	r.Mount("/api", apiRouter)

	log.Println("Starting API gateway server on :3000")

	err := http.ListenAndServe(":3000", r)
	if err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
