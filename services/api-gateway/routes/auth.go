package routes

import (
	"github.com/go-chi/chi/v5"
	"github.com/skni-kod/iot-monitor-backend/services/api-gateway/handlers"
)

func SetupAuthRoutes(r chi.Router, handler *handlers.AuthHandler) {
	r.Post("/register", handler.Register)
	r.Post("/login", handler.Login)
}
