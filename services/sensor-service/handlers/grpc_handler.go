package handlers

import (
	"context"
	"log"

	pb "github.com/skni-kod/iot-monitor-backend/internal/proto/sensor_service"
	"github.com/skni-kod/iot-monitor-backend/services/sensor-service/ent"
	"github.com/skni-kod/iot-monitor-backend/services/sensor-service/services"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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
func (h *SensorsGrpcHandler) CreateSensor(ctx context.Context, req *pb.CreateSensorRequest) (*pb.CreateSensorResponse, error) {
	if req.Name == "" {
		return nil, status.Error(codes.InvalidArgument, "Name is required")
	}

	if req.SensorTypeId <= 0 {
		return nil, status.Error(codes.InvalidArgument, "sensor_type_id must be a positive integer")
	}

	_, err := h.sensorsTypeService.GetSensorType(ctx, int(req.SensorTypeId))
	if err != nil {
		return nil, status.Error(codes.NotFound, "sensor type not found")
	}

	sensor, err := h.sensorsService.CreateSensor(ctx, &ent.Sensor{
		Name:        req.Name,
		Location:    req.Location,
		Description: req.Description,
		Active:      req.Active,
		Edges: ent.SensorEdges{
			Type: &ent.SensorType{ID: int(req.SensorTypeId)},
		},
	})
	if err != nil {
		log.Printf("Failed to create sensor: %v", err)
		return nil, status.Error(codes.Internal, "failed to create sensor")
	}

	return &pb.CreateSensorResponse{
		Sensor: convertSensorToProto(sensor),
	}, nil
}

// CreateSensorType implements api.SensorServiceServer.
// Subtle: this method shadows the method (UnimplementedSensorServiceServer).CreateSensorType of SensorsGrpcHandler.UnimplementedSensorServiceServer.
func (h *SensorsGrpcHandler) CreateSensorType(ctx context.Context, req *pb.CreateSensorTypeRequest) (*pb.CreateSensorTypeResponse, error) {
	if req.Name == "" || req.Model == "" {
		return nil, status.Error(codes.InvalidArgument, "name and model are required fields")
	}

	sensorType, err := h.sensorsTypeService.CreateSensorType(ctx, &ent.SensorType{
		Name:         req.Name,
		Model:        req.Model,
		Manufacturer: req.Manufacturer,
		Description:  req.Description,
		Unit:         req.Unit,
		MinValue:     float64(req.MinValue),
		MaxValue:     float64(req.MaxValue),
	})
	if err != nil {
		log.Printf("Failed to create sensor type: %v", err)
		return nil, status.Error(codes.Internal, "failed to create sensor type")
	}

	return &pb.CreateSensorTypeResponse{
		SensorType: convertSensorTypeToProto(sensorType),
	}, nil
}

// DeleteSensor implements api.SensorServiceServer.
func (h *SensorsGrpcHandler) DeleteSensor(ctx context.Context, req *pb.DeleteSensorRequest) (*pb.DeleteSensorResponse, error) {
	if req.Id <= 0 {
		return nil, status.Error(codes.InvalidArgument, "sensor id must be a positive integer")
	}

	_, err := h.sensorsService.GetSensor(ctx, int(req.Id))
	if err != nil {
		return nil, status.Error(codes.NotFound, "sensor not found")
	}

	err = h.sensorsService.DeleteSensor(ctx, int(req.Id))
	if err != nil {
		log.Printf("Failed to delete sensor: %v", err)
		return nil, status.Error(codes.Internal, "failed to delete sensor")
	}

	return &pb.DeleteSensorResponse{}, nil
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
func (h *SensorsGrpcHandler) ListSensorTypes(ctx context.Context, req *pb.ListSensorTypesRequest) (*pb.ListSensorTypesResponse, error) {
	sensorTypes, err := h.sensorsTypeService.ListSensorTypes(ctx)
	if err != nil {
		log.Printf("Failed to list sensor types: %v", err)
		return nil, status.Error(codes.Internal, "failed to list sensor types")
	}

	var protoSensorTypes []*pb.SensorType
	for _, st := range sensorTypes {
		protoSensorTypes = append(protoSensorTypes, convertSensorTypeToProto(st))
	}

	return &pb.ListSensorTypesResponse{
		SensorTypes: protoSensorTypes,
	}, nil
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
func (h *SensorsGrpcHandler) UpdateSensor(ctx context.Context, req *pb.UpdateSensorRequest) (*pb.UpdateSensorResponse, error) {
	if req.Id <= 0 {
		return nil, status.Error(codes.InvalidArgument, "sensor id must be a positive integer")
	}

	existingSensor, err := h.sensorsService.GetSensor(ctx, int(req.Id))
	if err != nil {
		return nil, status.Error(codes.NotFound, "sensor not found")
	}

	existingSensor.Name = req.Name
	existingSensor.Location = req.Location
	existingSensor.Description = req.Description
	existingSensor.Active = req.Active

	if req.SensorTypeId > 0 {
		_, err := h.sensorsTypeService.GetSensorType(ctx, int(req.SensorTypeId))
		if err != nil {
			return nil, status.Error(codes.NotFound, "sensor type not found")
		}
		existingSensor.Edges.Type = &ent.SensorType{ID: int(req.SensorTypeId)}
	}

	updatedSensor, err := h.sensorsService.UpdateSensor(ctx, existingSensor)
	if err != nil {
		log.Printf("Failed to update sensor: %v", err)
		return nil, status.Error(codes.Internal, "failed to update sensor")
	}

	return &pb.UpdateSensorResponse{
		Sensor: convertSensorToProto(updatedSensor),
	}, nil
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
