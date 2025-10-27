package routes

import (
	"github.com/go-chi/chi/v5"
	"github.com/skni-kod/iot-monitor-backend/services/api-gateway/handlers"
	authMiddleware "github.com/skni-kod/iot-monitor-backend/services/api-gateway/middleware"
)

func SetupSensorRoutes(r chi.Router, handler *handlers.SensorHandler) {
	authMw := authMiddleware.NewAuthMiddleware()

	r.Route("/sensors", func(r chi.Router) {
		r.Use(authMw.Authenticate)
		r.Get("/", handler.ListSensors)
		r.Post("/", handler.CreateSensor)
		r.Route("/{id}", func(r chi.Router) {
			r.Get("/", handler.GetSensor)
			r.Put("/", handler.UpdateSensor)
			r.Delete("/", handler.DeleteSensor)
			r.Put("/active", handler.SetSensorActive)
		})
	})
}
