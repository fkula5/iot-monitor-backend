package main

import (
	"context"

	pb "github.com/skni-kod/iot-monitor-backend/internal/proto/api"
	"google.golang.org/grpc"
)

type SensorGrpcHandler struct {
	pb.UnimplementedSensorServiceServer
}

func NewGrpcHandler(s *grpc.Server) {
	handler := &SensorGrpcHandler{}

	pb.RegisterSensorServiceServer(s, handler)
}

func (h *SensorGrpcHandler) ListSensors(ctx context.Context, req *pb.ListSensorsRequest) (*pb.ListSensorsResponse, error) {
	panic("unimplemented")
}

func (h *SensorGrpcHandler) GetSensor(ctx context.Context, req *pb.GetSensorRequest) (*pb.GetSensorResponse, error) {
	panic("unimplemented")
}
