package routes

import (
	"github.com/go-chi/chi/v5"

	"github.com/skni-kod/iot-monitor-backend/services/api-gateway/handlers"
	authMiddleware "github.com/skni-kod/iot-monitor-backend/services/api-gateway/middleware"
)

func SetupAlertRoutes(r chi.Router, handler *handlers.AlertHandler) {
	authMw := authMiddleware.NewAuthMiddleware()

	r.Route("/alerts", func(r chi.Router) {
		r.Use(authMw.Authenticate)
		r.Get("/", handler.ListAlerts)
		r.Post("/{id}/read", handler.MarkAlertAsRead)
	})
}
