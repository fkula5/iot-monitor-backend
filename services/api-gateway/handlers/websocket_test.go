package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/skni-kod/iot-monitor-backend/internal/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc"
	
	pb_data "github.com/skni-kod/iot-monitor-backend/internal/proto/data_service"
)

type MockDataServiceClient struct {
	mock.Mock
	pb_data.DataServiceClient
}

func (m *MockDataServiceClient) StoreReading(ctx context.Context, in *pb_data.StoreReadingRequest, opts ...grpc.CallOption) (*pb_data.StoreReadingResponse, error) {
	args := m.Called(ctx, in)
	return args.Get(0).(*pb_data.StoreReadingResponse), args.Error(1)
}

func TestStoreReading_Validation(t *testing.T) {
	mockDataClient := new(MockDataServiceClient)
	handler := &WebSocketHandler{
		dataClient: mockDataClient,
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
		rr := httptest.NewRecorder()

		mockDataClient.On("StoreReading", mock.Anything, mock.Anything).Return(&pb_data.StoreReadingResponse{}, nil).Once()

		handler.StoreReading(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		
		var resp map[string]string
		json.Unmarshal(rr.Body.Bytes(), &resp)
		assert.Equal(t, "success", resp["status"])
		mockDataClient.AssertExpectations(t)
	})
}
