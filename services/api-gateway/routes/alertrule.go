package routes

import (
	"github.com/go-chi/chi/v5"

	"github.com/skni-kod/iot-monitor-backend/services/api-gateway/handlers"
	authMiddleware "github.com/skni-kod/iot-monitor-backend/services/api-gateway/middleware"
)

func SetupAlertRuleRoutes(r chi.Router, handler *handlers.AlertRuleHandler) {
	authMw := authMiddleware.NewAuthMiddleware()
	r.Route("/alert-rules", func(r chi.Router) {
		r.Use(authMw.Authenticate)
		r.Get("/", handler.ListAlertRules)
		r.Post("/", handler.CreateAlertRule)
		r.Put("/{id}", handler.UpdateAlertRule)
		r.Delete("/{id}", handler.DeleteAlertRule)
		r.Get("/{id}", handler.GetAlertRule)
	})
}
