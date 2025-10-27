package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/skni-kod/iot-monitor-backend/internal/proto/auth"
)

type AuthHandler struct {
	Client auth.AuthServiceClient
}

func NewAuthHandler(client auth.AuthServiceClient) *AuthHandler {
	return &AuthHandler{Client: client}
}

// @Summary Login authenticates a user and returns a token.
// @Description Authenticates a user with email and password.
// @Tags Auth
// @Accept json
// @Produce json
// @Param credentials body object{email=string,password=string} true "User credentials"
// @Success 200 {object} object{token=string,expires_at=string,user=object} "Authentication successful"
// @Failure 400 {string} string "Bad Request"
// @Failure 401 {string} string "Unauthorized"
// @Failure 500 {string} string "Internal Server Error"
// @Router /auth/login [post]
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

	res, err := h.Client.Login(ctx, &auth.LoginRequest{
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

// @Summary Register creates a new user account.
// @Description Registers a new user with email, username, and password.
// @Tags Auth
// @Accept json
// @Produce json
// @Param user body object{email=string,username=string,password=string,first_name=string,last_name=string} true "User registration data"
// @Success 201 {object} object{token=string,expires_at=string,user=object} "Registration successful"
// @Failure 400 {string} string "Bad Request"
// @Failure 500 {string} string "Internal Server Error"
// @Router /auth/register [post]
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

	res, err := h.Client.Register(ctx, &auth.RegisterRequest{
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
