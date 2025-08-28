package routes

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/skni-kod/iot-monitor-backend/services/sensor-service/services"
)

type AuthHandler struct {
	authService services.IAuthService
}

func SetupAuthRoutes(r chi.Router, authService services.IAuthService) {
	handler := &AuthHandler{authService: authService}

	r.Route("/auth", func(r chi.Router) {
		r.Post("/register", handler.Register)
	})
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req services.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	res, err := h.authService.Register(r.Context(), &req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}
