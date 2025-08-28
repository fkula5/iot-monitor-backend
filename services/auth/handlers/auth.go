package handlers

import "github.com/skni-kod/iot-monitor-backend/services/auth/services"

type AuthHandler struct {
	authService services.IAuthService
}

func NewAuthHandler(authService services.IAuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}
