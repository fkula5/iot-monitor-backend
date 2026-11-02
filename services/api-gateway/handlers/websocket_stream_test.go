package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"

	pb_data "github.com/skni-kod/iot-monitor-backend/internal/proto/data_service"
	pb_sensor "github.com/skni-kod/iot-monitor-backend/internal/proto/sensor_service"
)

// fakeReadingStream emits a reading every millisecond until its context ends.
type fakeReadingStream struct {
	grpc.ServerStreamingClient[pb_data.ReadingUpdate]
	ctx context.Context
}

func (s *fakeReadingStream) Recv() (*pb_data.ReadingUpdate, error) {
	select {
	case <-s.ctx.Done():
		return nil, s.ctx.Err()
	case <-time.After(time.Millisecond):
		return &pb_data.ReadingUpdate{SensorId: 1, Value: 1, Timestamp: timestamppb.Now()}, nil
	}
}

type streamingDataClient struct {
	pb_data.DataServiceClient
	streamCtx chan context.Context
}

func (c *streamingDataClient) StreamReadings(ctx context.Context, _ *pb_data.StreamReadingsRequest, _ ...grpc.CallOption) (grpc.ServerStreamingClient[pb_data.ReadingUpdate], error) {
	c.streamCtx <- ctx
	return &fakeReadingStream{ctx: ctx}, nil
}

func TestHandleReadings_ConcurrentWritesAndCleanup(t *testing.T) {
	sensorClient := new(MockSensorServiceClient)
	sensorClient.On("ListSensors", mock.Anything, mock.Anything).Return(&pb_sensor.ListSensorsResponse{
		Sensors: []*pb_sensor.Sensor{{Id: 1, Active: true}},
	}, nil)
	dataClient := &streamingDataClient{streamCtx: make(chan context.Context, 1)}
	alerts := make(chan amqp.Delivery)
	defer close(alerts)

	h := NewWebSocketHandler(dataClient, sensorClient, alerts)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h.HandleReadings(w, withUser(r, 7))
	}))
	defer srv.Close()

	conn, _, err := websocket.DefaultDialer.Dial("ws"+strings.TrimPrefix(srv.URL, "http"), nil)
	require.NoError(t, err)

	var streamCtx context.Context
	select {
	case streamCtx = <-dataClient.streamCtx:
	case <-time.After(time.Second):
		t.Fatal("stream was not opened")
	}

	// Alerts are written from the broadcast goroutine while readings are
	// streamed from another; without a write lock gorilla/websocket panics.
	go func() {
		for i := 0; i < 50; i++ {
			alerts <- amqp.Delivery{Body: []byte(`{"user_id":7,"message":"high"}`)}
		}
	}()
	for i := 0; i < 100; i++ {
		_, _, err := conn.ReadMessage()
		require.NoError(t, err)
	}

	conn.Close()
	select {
	case <-streamCtx.Done():
	case <-time.After(time.Second):
		t.Fatal("stream context was not cancelled after the connection closed")
	}
	assert.Eventually(t, func() bool {
		h.clientsMu.RLock()
		defer h.clientsMu.RUnlock()
		return len(h.clients) == 0
	}, time.Second, 10*time.Millisecond)
}
