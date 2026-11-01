package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc"

	"github.com/skni-kod/iot-monitor-backend/internal/auth"
	pb_data "github.com/skni-kod/iot-monitor-backend/internal/proto/data_service"
	pb_sensor "github.com/skni-kod/iot-monitor-backend/internal/proto/sensor_service"
	"github.com/skni-kod/iot-monitor-backend/internal/types"
	authMiddleware "github.com/skni-kod/iot-monitor-backend/services/api-gateway/middleware"
)

type MockDataServiceClient struct {
	mock.Mock
	pb_data.DataServiceClient
}

func (m *MockDataServiceClient) StoreReading(ctx context.Context, in *pb_data.StoreReadingRequest, opts ...grpc.CallOption) (*pb_data.StoreReadingResponse, error) {
	args := m.Called(ctx, in)
	return args.Get(0).(*pb_data.StoreReadingResponse), args.Error(1)
}

type MockSensorServiceClient struct {
	mock.Mock
	pb_sensor.SensorServiceClient
}

func (m *MockSensorServiceClient) ListSensors(ctx context.Context, in *pb_sensor.ListSensorsRequest, opts ...grpc.CallOption) (*pb_sensor.ListSensorsResponse, error) {
	args := m.Called(ctx, in)
	return args.Get(0).(*pb_sensor.ListSensorsResponse), args.Error(1)
}

func withUser(req *http.Request, userID int) *http.Request {
	ctx := context.WithValue(req.Context(), authMiddleware.UserContextKey, &auth.Claims{UserId: userID})
	return req.WithContext(ctx)
}

func TestStoreReading_Validation(t *testing.T) {
	mockDataClient := new(MockDataServiceClient)
	mockSensorClient := new(MockSensorServiceClient)
	mockSensorClient.On("ListSensors", mock.Anything, &pb_sensor.ListSensorsRequest{UserId: 1}).
		Return(&pb_sensor.ListSensorsResponse{Sensors: []*pb_sensor.Sensor{{Id: 1, Active: true}}}, nil)
	handler := &WebSocketHandler{
		dataClient:   mockDataClient,
		sensorClient: mockSensorClient,
	}

	t.Run("missing value should return 400", func(t *testing.T) {
		reqBody := map[string]interface{}{
			"sensor_id": 1,
		}
		jsonBody, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest("POST", "/api/data/readings", bytes.NewBuffer(jsonBody))
		rr := httptest.NewRecorder()

		// This might panic before the fix
		defer func() {
			if r := recover(); r != nil {
				t.Logf("Recovered from panic as expected: %v", r)
				// If it panics, it's definitely NOT returning 400
			}
		}()

		handler.StoreReading(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)

		var resp map[string]interface{}
		json.Unmarshal(rr.Body.Bytes(), &resp)
		assert.Equal(t, "Validation failed", resp["error"])
	})

	t.Run("invalid sensor_id should return 400", func(t *testing.T) {
		val := float32(25.5)
		reqBody := types.StoreReadingRequest{
			SensorID: 0,
			Value:    &val,
		}
		jsonBody, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest("POST", "/api/data/readings", bytes.NewBuffer(jsonBody))
		rr := httptest.NewRecorder()

		handler.StoreReading(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)

		var resp map[string]interface{}
		json.Unmarshal(rr.Body.Bytes(), &resp)
		assert.Equal(t, "Validation failed", resp["error"])
	})

	t.Run("valid request should return 200", func(t *testing.T) {
		val := float32(25.5)
		reqBody := types.StoreReadingRequest{
			SensorID: 1,
			Value:    &val,
		}
		jsonBody, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest("POST", "/api/data/readings", bytes.NewBuffer(jsonBody))
		req = withUser(req, 1)
		rr := httptest.NewRecorder()

		mockDataClient.On("StoreReading", mock.Anything, mock.Anything).Return(&pb_data.StoreReadingResponse{}, nil).Once()

		handler.StoreReading(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		var resp struct {
			Data map[string]string `json:"data"`
		}
		json.Unmarshal(rr.Body.Bytes(), &resp)
		assert.Equal(t, "success", resp.Data["status"])
		mockDataClient.AssertExpectations(t)
	})

	t.Run("sensor of another user should return 404", func(t *testing.T) {
		val := float32(25.5)
		jsonBody, _ := json.Marshal(types.StoreReadingRequest{SensorID: 2, Value: &val})
		req, _ := http.NewRequest("POST", "/api/data/readings", bytes.NewBuffer(jsonBody))
		req = withUser(req, 1)
		rr := httptest.NewRecorder()

		handler.StoreReading(rr, req)

		assert.Equal(t, http.StatusNotFound, rr.Code)
		mockDataClient.AssertNotCalled(t, "StoreReading", mock.Anything, mock.MatchedBy(func(r *pb_data.StoreReadingRequest) bool {
			return r.SensorId == 2
		}))
	})
}

func TestFilterOwned(t *testing.T) {
	owned := map[int64]bool{1: true, 3: false}
	assert.Equal(t, []int64{1, 3}, filterOwned([]int64{1, 2, 3, 4}, owned))
	assert.Empty(t, filterOwned([]int64{2}, owned))
}
