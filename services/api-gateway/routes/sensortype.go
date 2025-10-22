package routes

import (
	"github.com/go-chi/chi/v5"
	"github.com/skni-kod/iot-monitor-backend/services/api-gateway/handlers"
	authMiddleware "github.com/skni-kod/iot-monitor-backend/services/api-gateway/middleware"
)

func SetupSensorTypeRoutes(r chi.Router, handler *handlers.SensorTypeHandler) {
	authMw := authMiddleware.NewAuthMiddleware()

	r.Route("/sensor-types", func(r chi.Router) {
		r.Use(authMw.Authenticate)
		r.Get("/", handler.ListSensorTypes)
		r.Post("/", handler.CreateSensorType)
		r.Route("/{id}", func(r chi.Router) {
			r.Get("/", handler.GetSensorType)
			r.Put("/", handler.UpdateSensorType)
			r.Delete("/", handler.DeleteSensorType)
		})
	})
}
