package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/skni-kod/iot-monitor-backend/internal/proto/api"
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

	apiRouter := chi.NewRouter()

	apiRouter.Use(middleware.RequestID)
	apiRouter.Use(middleware.RealIP)

	setupSensorRoutes(apiRouter, sensorClient)

	r.Mount("/api", apiRouter)

	log.Println("Starting API gateway server on :3000")

	err := http.ListenAndServe(":3000", r)
	if err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

func setupSensorRoutes(r chi.Router, client api.SensorServiceClient) {
	r.Route("/sensors", func(r chi.Router) {
		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			listSensors(w, r, client)
		})

		r.Get("/{id}", func(w http.ResponseWriter, r *http.Request) {
			getSensor(w, r, client)
		})
	})
}

func listSensors(w http.ResponseWriter, r *http.Request, client api.SensorServiceClient) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	res, err := client.ListSensors(ctx, &api.ListSensorsRequest{})
	if err != nil {
		http.Error(w, "Failed to fetch sensors: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(res.Sensors)
	if err != nil {
		http.Error(w, "Failed to encode response: "+err.Error(), http.StatusInternalServerError)
		return
	}
}

func getSensor(w http.ResponseWriter, r *http.Request, client api.SensorServiceClient) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid sensor ID", http.StatusBadRequest)
		return
	}

	res, err := client.GetSensor(ctx, &api.GetSensorRequest{Id: int32(id)})
	if err != nil {
		http.Error(w, "Failed to fetch sensor: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if res.Sensor == nil {
		http.Error(w, "Sensor not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	err = json.NewEncoder(w).Encode(res.Sensor)
	if err != nil {
		http.Error(w, "Failed to encode response: "+err.Error(), http.StatusInternalServerError)
		return
	}
}
