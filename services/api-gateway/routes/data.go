package routes

import (
	"github.com/go-chi/chi/v5"
	"github.com/skni-kod/iot-monitor-backend/services/api-gateway/handlers"
)

func SetupDataRoutes(r chi.Router, handler *handlers.WebSocketHandler) {
	r.Route("/data", func(r chi.Router) {
		r.Get("/ws/readings", handler.HandleReadings)
		// r.Get("/readings/latest", handler.GetLatestReadings)
		r.Get("/sensors/{sensor_id}/readings", handler.GetHistoricalReadings)
		r.Get("/ws/test", handler.WsHandler)
	})
}
