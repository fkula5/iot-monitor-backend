package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/skni-kod/iot-monitor-backend/internal/auth"
	"github.com/skni-kod/iot-monitor-backend/pkg/logger"
)

func TestAuthenticate_QueryToken(t *testing.T) {
	t.Setenv("JWT_SECRET", "test-secret")
	assert.NoError(t, logger.Init(logger.Config{Level: "error", Environment: "test", OutputPaths: []string{"stdout"}}))

	token, err := auth.NewJWTService().GenerateToken(1, "user", "user@example.com")
	assert.NoError(t, err)

	handler := NewAuthMiddleware().Authenticate(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	t.Run("accepted on WebSocket upgrade", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/data/ws/readings?token="+token, nil)
		req.Header.Set("Connection", "Upgrade")
		req.Header.Set("Upgrade", "websocket")
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
	})

	t.Run("rejected on regular request", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/data/readings/latest?token="+token, nil)
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusUnauthorized, rr.Code)
	})
}
