package handlers

import (
	"context"
	"log"
	"sync"
	"time"

	pb_data "github.com/skni-kod/iot-monitor-backend/internal/proto/data_service"
	pb_sensor "github.com/skni-kod/iot-monitor-backend/internal/proto/sensor_service"
	"github.com/skni-kod/iot-monitor-backend/services/data-processing/storage"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type DataGrpcHandler struct {
	pb_data.UnimplementedDataServiceServer
	store         storage.ITimeScaleStorage
	sensorClient  pb_sensor.SensorServiceClient
	subscribers   map[string]chan *pb_data.ReadingUpdate
	subscribersMu sync.RWMutex
}

func NewDataGrpcHandler(s *grpc.Server, store storage.ITimeScaleStorage, sensorClient pb_sensor.SensorServiceClient) {
	handler := &DataGrpcHandler{
		store:        store,
		sensorClient: sensorClient,
		subscribers:  make(map[string]chan *pb_data.ReadingUpdate),
	}
	pb_data.RegisterDataServiceServer(s, handler)
}

func (h *DataGrpcHandler) StoreReading(ctx context.Context, req *pb_data.StoreReadingRequest) (*pb_data.StoreReadingResponse, error) {
	if req.SensorId <= 0 {
		return nil, status.Error(codes.InvalidArgument, "sensor_id must be positive")
	}

	err := h.store.StoreReading(ctx, req.SensorId, req.Value, req.Timestamp.AsTime())
	if err != nil {
		log.Printf("Failed to store reading: %v", err)
		return nil, status.Error(codes.Internal, "failed to store reading")
	}

	sensor, err := h.sensorClient.GetSensor(ctx, &pb_sensor.GetSensorRequest{Id: req.SensorId})
	if err == nil && sensor.Sensor != nil {
		update := &pb_data.ReadingUpdate{
			SensorId:   req.SensorId,
			Value:      req.Value,
			Timestamp:  req.Timestamp,
			SensorName: sensor.Sensor.Name,
			Location:   sensor.Sensor.Location,
		}

		if sensor.Sensor.SensorType != nil {
			update.Unit = sensor.Sensor.SensorType.Unit
		}

		h.broadcastUpdate(update)
	}

	return &pb_data.StoreReadingResponse{}, nil
}

func (h *DataGrpcHandler) QueryReadings(ctx context.Context, req *pb_data.QueryReadingsRequest) (*pb_data.QueryReadingsResponse, error) {
	if req.SensorId <= 0 {
		return nil, status.Error(codes.InvalidArgument, "sensor_id must be positive")
	}

	readings, err := h.store.QueryReadings(ctx, req.SensorId, req.StartTime.AsTime(), req.EndTime.AsTime())
	if err != nil {
		log.Printf("Failed to query readings: %v", err)
		return nil, status.Error(codes.Internal, "failed to query readings")
	}

	return &pb_data.QueryReadingsResponse{
		DataPoints: readings,
	}, nil
}

func (h *DataGrpcHandler) StreamReadings(req *pb_data.StreamReadingsRequest, stream pb_data.DataService_StreamReadingsServer) error {
	if len(req.SensorIds) == 0 {
		return status.Error(codes.InvalidArgument, "at least one sensor_id is required")
	}

	ch := make(chan *pb_data.ReadingUpdate, 100)
	id := generateSubscriberID()

	h.subscribersMu.Lock()
	h.subscribers[id] = ch
	h.subscribersMu.Unlock()

	defer func() {
		h.subscribersMu.Lock()
		delete(h.subscribers, id)
		close(ch)
		h.subscribersMu.Unlock()
	}()

	initialReadings, err := h.store.GetLatestReadingsBatch(stream.Context(), req.SensorIds)
	if err == nil {
		for _, reading := range initialReadings {
			sensor, err := h.sensorClient.GetSensor(stream.Context(), &pb_sensor.GetSensorRequest{Id: reading.SensorId})
			if err == nil && sensor.Sensor != nil {
				reading.SensorName = sensor.Sensor.Name
				reading.Location = sensor.Sensor.Location
				if sensor.Sensor.SensorType != nil {
					reading.Unit = sensor.Sensor.SensorType.Unit
				}
			}
			if err := stream.Send(reading); err != nil {
				return err
			}
		}
	}

	sensorIDMap := make(map[int64]bool)
	for _, id := range req.SensorIds {
		sensorIDMap[id] = true
	}

	for {
		select {
		case <-stream.Context().Done():
			return nil
		case update, ok := <-ch:
			if !ok {
				return nil
			}
			if sensorIDMap[update.SensorId] {
				if err := stream.Send(update); err != nil {
					return err
				}
			}
		}
	}
}

func (h *DataGrpcHandler) GetLatestReadingsBatch(ctx context.Context, req *pb_data.LatestReadingsBatchRequest) (*pb_data.LatestReadingsBatchResponse, error) {
	if len(req.SensorIds) == 0 {
		return nil, status.Error(codes.InvalidArgument, "at least one sensor_id is required")
	}

	readings, err := h.store.GetLatestReadingsBatch(ctx, req.SensorIds)
	if err != nil {
		log.Printf("Failed to get latest readings batch: %v", err)
		return nil, status.Error(codes.Internal, "failed to get latest readings")
	}

	for _, reading := range readings {
		sensor, err := h.sensorClient.GetSensor(ctx, &pb_sensor.GetSensorRequest{Id: reading.SensorId})
		if err == nil && sensor.Sensor != nil {
			reading.SensorName = sensor.Sensor.Name
			reading.Location = sensor.Sensor.Location
			if sensor.Sensor.SensorType != nil {
				reading.Unit = sensor.Sensor.SensorType.Unit
			}
		}
	}

	return &pb_data.LatestReadingsBatchResponse{
		Readings: readings,
	}, nil
}

func (h *DataGrpcHandler) GetLatestReadingsBySensor(ctx context.Context, req *pb_data.LatestReadingsBySensorRequest) (*pb_data.LatestReadingsBySensorResponse, error) {
	if req.SensorId <= 0 {
		return nil, status.Error(codes.InvalidArgument, "sensor_id must be positive")
	}

	limit := req.Limit
	if limit <= 0 {
		limit = 1
	}

	readings, err := h.store.GetLatestReadingsBySensor(ctx, req.SensorId, limit)
	if err != nil {
		log.Printf("Failed to get latest readings by sensor: %v", err)
		return nil, status.Error(codes.Internal, "failed to get sensor readings")
	}

	sensor, err := h.sensorClient.GetSensor(ctx, &pb_sensor.GetSensorRequest{Id: req.SensorId})
	if err == nil && sensor.Sensor != nil {
		for _, reading := range readings {
			reading.SensorName = sensor.Sensor.Name
			reading.Location = sensor.Sensor.Location
			if sensor.Sensor.SensorType != nil {
				reading.Unit = sensor.Sensor.SensorType.Unit
			}
		}
	}

	return &pb_data.LatestReadingsBySensorResponse{
		Readings: readings,
	}, nil
}

func (h *DataGrpcHandler) broadcastUpdate(update *pb_data.ReadingUpdate) {
	h.subscribersMu.RLock()
	defer h.subscribersMu.RUnlock()

	for _, ch := range h.subscribers {
		select {
		case ch <- update:
		default:
		}
	}
}

func generateSubscriberID() string {
	return time.Now().Format("20060102150405.000000")
}
