package handlers

import (
	"context"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	pb "github.com/skni-kod/iot-monitor-backend/internal/proto/sensor_service"
	"github.com/skni-kod/iot-monitor-backend/pkg/logger"
	"github.com/skni-kod/iot-monitor-backend/services/sensor-service/ent"
	"github.com/skni-kod/iot-monitor-backend/services/sensor-service/services"
)

type SensorsGrpcHandler struct {
	pb.UnimplementedSensorServiceServer
	sensorsService      services.ISensorService
	sensorsTypeService  services.ISensorTypeService
	sensorsGroupService services.ISensorGroupService
}

func NewGrpcHandler(s *grpc.Server, sensorsService services.ISensorService, sensorsTypeService services.ISensorTypeService, sensorsGroupService services.ISensorGroupService) {
	handler := &SensorsGrpcHandler{sensorsService: sensorsService, sensorsTypeService: sensorsTypeService, sensorsGroupService: sensorsGroupService}
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
		UserID:      req.UserId,
		Edges: ent.SensorEdges{
			Type: &ent.SensorType{ID: int(req.SensorTypeId)},
		},
	})
	if err != nil {
		logger.Error("Failed to create sensor", zap.Error(err))
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
		logger.Error("Failed to create sensor type", zap.Error(err))
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
		logger.Error("Failed to delete sensor", zap.Error(err))
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
		logger.Error("Failed to list sensor types", zap.Error(err))
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

func (h *SensorsGrpcHandler) DeleteSensorType(ctx context.Context, req *pb.DeleteSensorTypeRequest) (*pb.DeleteSensorTypeResponse, error) {
	if req.Id <= 0 {
		return nil, status.Error(codes.InvalidArgument, "sensor type id must be a positive integer")
	}

	_, err := h.sensorsTypeService.GetSensorType(ctx, int(req.Id))
	if err != nil {
		return nil, status.Error(codes.NotFound, "sensor type not found")
	}

	err = h.sensorsTypeService.DeleteSensorType(ctx, int(req.Id))
	if err != nil {
		logger.Error("Failed to delete sensor type", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to delete sensor type")
	}

	return &pb.DeleteSensorTypeResponse{}, nil
}

func (h *SensorsGrpcHandler) UpdateSensorType(ctx context.Context, req *pb.UpdateSensorTypeRequest) (*pb.UpdateSensorTypeResponse, error) {
	if req.Id <= 0 {
		return nil, status.Error(codes.InvalidArgument, "sensor type id must be a positive integer")
	}

	existingST, err := h.sensorsTypeService.GetSensorType(ctx, int(req.Id))
	if err != nil {
		return nil, status.Error(codes.NotFound, "sensor type not found")
	}

	if req.Name == "" || req.Model == "" {
		return nil, status.Error(codes.InvalidArgument, "name and model are required fields")
	}

	existingST.Name = req.Name
	existingST.Model = req.Model
	existingST.Manufacturer = req.Manufacturer
	existingST.Description = req.Description
	existingST.Unit = req.Unit
	existingST.MinValue = float64(req.MinValue)
	existingST.MaxValue = float64(req.MaxValue)

	updatedST, err := h.sensorsTypeService.UpdateSensorType(ctx, int(req.Id), existingST)
	if err != nil {
		logger.Error("Failed to update sensor type", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to update sensor type")
	}

	return &pb.UpdateSensorTypeResponse{
		SensorType: convertSensorTypeToProto(updatedST),
	}, nil
}

// ListSensors implements api.SensorServiceServer.
func (h *SensorsGrpcHandler) ListSensors(ctx context.Context, req *pb.ListSensorsRequest) (*pb.ListSensorsResponse, error) {
	sensors, err := h.sensorsService.ListSensors(ctx, req.UserId)
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
		logger.Error("Failed to update sensor", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to update sensor")
	}

	return &pb.UpdateSensorResponse{
		Sensor: convertSensorToProto(updatedSensor),
	}, nil
}

func (h *SensorsGrpcHandler) SetSensorActive(ctx context.Context, req *pb.SetSensorActiveRequest) (*pb.SetSensorActiveResponse, error) {
	sensor, err := h.sensorsService.SetSensorActive(ctx, int(req.Id))
	if err != nil {
		return nil, err
	}

	return &pb.SetSensorActiveResponse{
		Sensor: convertSensorToProto(sensor),
	}, nil
}

func (h *SensorsGrpcHandler) CreateSensorGroup(ctx context.Context, req *pb.CreateSensorGroupRequest) (*pb.CreateSensorGroupResponse, error) {
	if req.Name == "" {
		return nil, status.Error(codes.InvalidArgument, "Group name is required")
	}

	group := &ent.SensorGroup{
		Name:        req.Name,
		Description: req.Description,
		Color:       req.Color,
		Icon:        req.Icon,
		UserID:      req.UserId,
	}

	createdGroup, err := h.sensorsGroupService.CreateGroup(ctx, group, req.SensorIds)
	if err != nil {
		logger.Error("Failed to create sensor group", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to create sensor group")
	}

	return &pb.CreateSensorGroupResponse{
		Group: convertSensorGroupToProto(createdGroup),
	}, nil
}

func (h *SensorsGrpcHandler) GetSensorGroup(ctx context.Context, req *pb.GetSensorGroupRequest) (*pb.GetSensorGroupResponse, error) {
	group, err := h.sensorsGroupService.Get(ctx, int(req.Id))
	if err != nil {
		return nil, status.Error(codes.NotFound, "sensor group not found")
	}

	var protoSensors []*pb.Sensor
	if group.Edges.Sensors != nil {
		for _, s := range group.Edges.Sensors {
			protoSensors = append(protoSensors, convertSensorToProto(s))
		}
	}

	return &pb.GetSensorGroupResponse{
		Group:   convertSensorGroupToProto(group),
		Sensors: protoSensors,
	}, nil
}

func (h *SensorsGrpcHandler) ListSensorGroups(ctx context.Context, req *pb.ListSensorGroupsRequest) (*pb.ListSensorGroupsResponse, error) {
	groups, err := h.sensorsGroupService.ListGroups(ctx, req.UserId)
	if err != nil {
		logger.Error("Failed to list sensor groups", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to list sensor groups")
	}

	var protoGroups []*pb.SensorGroup
	for _, g := range groups {
		protoGroups = append(protoGroups, convertSensorGroupToProto(g))
	}

	return &pb.ListSensorGroupsResponse{
		Groups: protoGroups,
	}, nil
}

func (h *SensorsGrpcHandler) UpdateSensorGroup(ctx context.Context, req *pb.UpdateSensorGroupRequest) (*pb.UpdateSensorGroupResponse, error) {
	if req.Id <= 0 {
		return nil, status.Error(codes.InvalidArgument, "group id must be a positive integer")
	}

	groupUpdate := &ent.SensorGroup{
		Name:        req.Name,
		Description: req.Description,
		Color:       req.Color,
		Icon:        req.Icon,
	}

	updatedGroup, err := h.sensorsGroupService.UpdateGroup(ctx, int(req.Id), groupUpdate)
	if err != nil {
		logger.Error("Failed to update sensor group", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to update sensor group")
	}

	return &pb.UpdateSensorGroupResponse{
		Group: convertSensorGroupToProto(updatedGroup),
	}, nil
}

func (h *SensorsGrpcHandler) DeleteSensorGroup(ctx context.Context, req *pb.DeleteSensorGroupRequest) (*pb.DeleteSensorGroupResponse, error) {
	if req.Id <= 0 {
		return nil, status.Error(codes.InvalidArgument, "group id must be a positive integer")
	}

	err := h.sensorsGroupService.DeleteGroup(ctx, int(req.Id))
	if err != nil {
		logger.Error("Failed to delete sensor group", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to delete sensor group")
	}

	return &pb.DeleteSensorGroupResponse{}, nil
}

func (h *SensorsGrpcHandler) AddSensorsToGroup(ctx context.Context, req *pb.AddSensorsToGroupRequest) (*pb.AddSensorsToGroupResponse, error) {
	if req.GroupId <= 0 {
		return nil, status.Error(codes.InvalidArgument, "group id must be a positive integer")
	}
	if len(req.SensorIds) == 0 {
		return nil, status.Error(codes.InvalidArgument, "sensor ids list cannot be empty")
	}

	group, err := h.sensorsGroupService.AddSensorsToGroup(ctx, int(req.GroupId), req.SensorIds)
	if err != nil {
		logger.Error("Failed to add sensors to group", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to add sensors to group")
	}

	return &pb.AddSensorsToGroupResponse{
		Group: convertSensorGroupToProto(group),
	}, nil
}

func (h *SensorsGrpcHandler) RemoveSensorsFromGroup(ctx context.Context, req *pb.RemoveSensorsFromGroupRequest) (*pb.RemoveSensorsFromGroupResponse, error) {
	if req.GroupId <= 0 {
		return nil, status.Error(codes.InvalidArgument, "group id must be a positive integer")
	}
	if len(req.SensorIds) == 0 {
		return nil, status.Error(codes.InvalidArgument, "sensor ids list cannot be empty")
	}

	group, err := h.sensorsGroupService.RemoveSensorsFromGroup(ctx, int(req.GroupId), req.SensorIds)
	if err != nil {
		logger.Error("Failed to remove sensors from group", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to remove sensors from group")
	}

	return &pb.RemoveSensorsFromGroupResponse{
		Group: convertSensorGroupToProto(group),
	}, nil
}

func convertSensorGroupToProto(g *ent.SensorGroup) *pb.SensorGroup {
	if g == nil {
		return nil
	}

	groupProto := &pb.SensorGroup{
		Id:          int64(g.ID),
		Name:        g.Name,
		Description: g.Description,
		Color:       g.Color,
		Icon:        g.Icon,
		UserId:      g.UserID,
		CreatedAt:   timestamppb.New(g.CreatedAt),
		UpdatedAt:   timestamppb.New(g.UpdatedAt),
	}

	if g.Edges.Sensors != nil {
		var sensorIds []int64
		for _, s := range g.Edges.Sensors {
			sensorIds = append(sensorIds, int64(s.ID))
		}
		groupProto.SensorIds = sensorIds
	}

	return groupProto
}

func convertSensorToProto(s *ent.Sensor) *pb.Sensor {
	if s == nil {
		return nil
	}

	sensorProto := &pb.Sensor{
		Id:          int64(s.ID),
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
		sensorProto.SensorTypeId = int64(s.Edges.Type.ID)
		sensorProto.SensorType = convertSensorTypeToProto(s.Edges.Type)
	}

	return sensorProto
}

func convertSensorTypeToProto(st *ent.SensorType) *pb.SensorType {
	if st == nil {
		return nil
	}

	sensorTypeProto := &pb.SensorType{
		Id:           int64(st.ID),
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
