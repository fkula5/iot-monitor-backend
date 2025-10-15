package routes

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/skni-kod/iot-monitor-backend/internal/proto/auth"
)

type AuthHandler struct {
	client auth.AuthServiceClient
}

func SetupAuthRoutes(r chi.Router, client auth.AuthServiceClient) {
	handler := &AuthHandler{client: client}
	r.Post("/register", handler.Register)
	r.Post("/login", handler.Login)
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Email == "" {
		http.Error(w, "Email is required", http.StatusBadRequest)
		return
	}

	if req.Password == "" {
		http.Error(w, "Password is required", http.StatusBadRequest)
		return
	}

	res, err := h.client.Login(ctx, &auth.LoginRequest{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		http.Error(w, "Invalid email or password", http.StatusUnauthorized)
		return
	}

	if res.User == nil {
		http.Error(w, "User data is missing in response", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"token":      res.Token,
		"expires_at": res.ExpiresAt.AsTime(),
		"user":       res.User,
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		http.Error(w, "Failed to encode response: "+err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email     string `json:"email"`
		Username  string `json:"username"`
		Password  string `json:"password"`
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Email == "" {
		http.Error(w, "Email is required", http.StatusBadRequest)
		return
	}

	if req.Username == "" {
		http.Error(w, "Username is required", http.StatusBadRequest)
		return
	}

	if req.Password == "" {
		http.Error(w, "Password is required", http.StatusBadRequest)
		return
	}

	res, err := h.client.Register(ctx, &auth.RegisterRequest{
		Email:     req.Email,
		Username:  req.Username,
		Password:  req.Password,
		FirstName: req.FirstName,
		LastName:  req.LastName,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if res.User == nil {
		http.Error(w, "User data is missing in response", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"token":      res.Token,
		"expires_at": res.ExpiresAt.AsTime(),
		"user":       res.User,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		http.Error(w, "Failed to encode response: "+err.Error(), http.StatusInternalServerError)
		return
	}
}
