package routes

import (
	"github.com/go-chi/chi/v5"

	"github.com/skni-kod/iot-monitor-backend/services/api-gateway/handlers"
	authMiddleware "github.com/skni-kod/iot-monitor-backend/services/api-gateway/middleware"
)

func SetupSensorGroupRoutes(r chi.Router, handler *handlers.SensorGroupHandler) {
	authMw := authMiddleware.NewAuthMiddleware()

	r.Route("/sensor-groups", func(r chi.Router) {
		r.Use(authMw.Authenticate)
		r.Get("/", handler.ListGroups)
		r.Post("/", handler.CreateGroup)
		r.Route("/{id}", func(r chi.Router) {
			r.Get("/", handler.GetGroup)
			r.Put("/", handler.UpdateGroup)
			r.Delete("/", handler.DeleteGroup)
			r.Post("/sensors", handler.AddSensorsToGroup)
			r.Delete("/sensors", handler.RemoveSensorsFromGroup)
		})
	})
}
