package handlers

import (
	"context"

	pb "github.com/skni-kod/iot-monitor-backend/internal/proto/api"
	"github.com/skni-kod/iot-monitor-backend/services/sensor-service/ent"
	"github.com/skni-kod/iot-monitor-backend/services/sensor-service/services"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type SensorsGrpcHandler struct {
	pb.UnimplementedSensorServiceServer
	sensorsService     services.ISensorService
	sensorsTypeService services.ISensorTypeService
}

func NewGrpcHandler(s *grpc.Server, sensorsService services.ISensorService, sensorsTypeService services.ISensorTypeService) {
	handler := &SensorsGrpcHandler{sensorsService: sensorsService, sensorsTypeService: sensorsTypeService}

	pb.RegisterSensorServiceServer(s, handler)
}

// CreateSensor implements api.SensorServiceServer.
func (h *SensorsGrpcHandler) CreateSensor(context.Context, *pb.CreateSensorRequest) (*pb.CreateSensorResponse, error) {
	panic("unimplemented")
}

// CreateSensorType implements api.SensorServiceServer.
// Subtle: this method shadows the method (UnimplementedSensorServiceServer).CreateSensorType of SensorsGrpcHandler.UnimplementedSensorServiceServer.
func (h *SensorsGrpcHandler) CreateSensorType(context.Context, *pb.CreateSensorTypeRequest) (*pb.CreateSensorTypeResponse, error) {
	panic("unimplemented")
}

// DeleteSensor implements api.SensorServiceServer.
func (h *SensorsGrpcHandler) DeleteSensor(context.Context, *pb.DeleteSensorRequest) (*pb.DeleteSensorResponse, error) {
	panic("unimplemented")
}

// GetSensor implements api.SensorServiceServer.
func (h *SensorsGrpcHandler) GetSensor(ctx context.Context, req *pb.GetSensorRequest) (*pb.GetSensorResponse, error) {
	sensor, err := h.sensorsService.GetSensor(ctx, int(req.Id))
	if err != nil {
		return nil, err
	}

	return &pb.GetSensorResponse{
		Sensor: convertSensorToProto(sensor),
	}, nil
}

// GetSensorType implements api.SensorServiceServer.
// Subtle: this method shadows the method (UnimplementedSensorServiceServer).GetSensorType of SensorsGrpcHandler.UnimplementedSensorServiceServer.
func (h *SensorsGrpcHandler) GetSensorType(ctx context.Context, req *pb.GetSensorTypeRequest) (*pb.GetSensorTypeResponse, error) {
	sensorType, err := h.sensorsTypeService.GetSensorType(ctx, int(req.Id))
	if err != nil {
		return nil, err
	}

	return &pb.GetSensorTypeResponse{
		SensorType: convertSensorTypeToProto(sensorType),
	}, nil
}

// ListSensorTypes implements api.SensorServiceServer.
// Subtle: this method shadows the method (UnimplementedSensorServiceServer).ListSensorTypes of SensorsGrpcHandler.UnimplementedSensorServiceServer.
func (h *SensorsGrpcHandler) ListSensorTypes(context.Context, *pb.ListSensorTypesRequest) (*pb.ListSensorTypesResponse, error) {
	panic("unimplemented")
}

// ListSensors implements api.SensorServiceServer.
func (h *SensorsGrpcHandler) ListSensors(ctx context.Context, req *pb.ListSensorsRequest) (*pb.ListSensorsResponse, error) {
	sensors, err := h.sensorsService.ListSensors(ctx)
	if err != nil {
		return nil, err
	}

	var protoSensors []*pb.Sensor
	for _, s := range sensors {
		protoSensors = append(protoSensors, convertSensorToProto(s))
	}

	return &pb.ListSensorsResponse{
		Sensors: protoSensors,
	}, nil
}

// UpdateSensor implements api.SensorServiceServer.
func (h *SensorsGrpcHandler) UpdateSensor(context.Context, *pb.UpdateSensorRequest) (*pb.UpdateSensorResponse, error) {
	panic("unimplemented")
}

func convertSensorToProto(s *ent.Sensor) *pb.Sensor {
	if s == nil {
		return nil
	}

	sensorProto := &pb.Sensor{
		Id:          int32(s.ID),
		Name:        s.Name,
		Location:    s.Location,
		Description: s.Description,
		Active:      s.Active,
		CreatedAt:   timestamppb.New(s.CreatedAt),
		UpdatedAt:   timestamppb.New(s.UpdatedAt),
	}

	if !s.LastUpdated.IsZero() {
		sensorProto.LastUpdated = timestamppb.New(s.LastUpdated)
	}

	if s.Edges.Type != nil {
		sensorProto.SensorTypeId = int32(s.Edges.Type.ID)
	}

	return sensorProto
}

func convertSensorTypeToProto(st *ent.SensorType) *pb.SensorType {
	if st == nil {
		return nil
	}

	sensorTypeProto := &pb.SensorType{
		Id:           int32(st.ID),
		Name:         st.Name,
		Model:        st.Model,
		Manufacturer: st.Manufacturer,
		Description:  st.Description,
		Unit:         st.Unit,
		MinValue:     float32(st.MinValue),
		MaxValue:     float32(st.MaxValue),
		CreatedAt:    timestamppb.New(st.CreatedAt),
	}

	return sensorTypeProto
}

func (h *SensorsGrpcHandler) SetSensorActive(ctx context.Context, req *pb.SetSensorActiveRequest) (*pb.SetSensorActiveResponse, error) {
	sensor, err := h.sensorsService.SetSensorActive(ctx, int(req.Id), req.Active)
	if err != nil {
		return nil, err
	}

	return &pb.SetSensorActiveResponse{
		Sensor: convertSensorToProto(sensor),
	}, nil
}
