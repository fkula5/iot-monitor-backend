package routes

import (
	"github.com/go-chi/chi/v5"

	"github.com/skni-kod/iot-monitor-backend/services/api-gateway/handlers"
	authMiddleware "github.com/skni-kod/iot-monitor-backend/services/api-gateway/middleware"
)

func SetupAuthRoutes(r chi.Router, handler *handlers.AuthHandler) {
	authMw := authMiddleware.NewAuthMiddleware()

	r.Post("/register", handler.Register)
	r.Post("/login", handler.Login)
	r.Post("/forgot-password", handler.ForgotPassword)
	r.Post("/reset-password", handler.ResetPassword)
	r.Group(func(r chi.Router) {
		r.Use(authMw.Authenticate)
		r.Get("/user", handler.GetUser)
		r.Put("/user", handler.UpdateUser)
	})
}
