package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/skni-kod/iot-monitor-backend/internal/proto/auth"
	authMiddleware "github.com/skni-kod/iot-monitor-backend/services/api-gateway/middleware"
)

type AuthHandler struct {
	Client auth.AuthServiceClient
}

func NewAuthHandler(client auth.AuthServiceClient) *AuthHandler {
	return &AuthHandler{Client: client}
}

type UserResponse struct {
	ID        int64  `json:"id"`
	Email     string `json:"email"`
	Username  string `json:"username"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

// @Summary Login authenticates a user and returns a token.
// @Description Authenticates a user with email and password.
// @Tags Auth
// @Accept json
// @Produce json
// @Param credentials body object{email=string,password=string} true "User credentials"
// @Success 200 {object} Response{data=object{token=string,expires_at=string,user=UserResponse}} "Authentication successful"
// @Failure 400 {object} Response{error=string} "Bad Request"
// @Failure 401 {object} Response{error=string} "Unauthorized"
// @Failure 500 {object} Response{error=string} "Internal Server Error"
// @Router /auth/login [post]
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		Error(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Email == "" {
		Error(w, http.StatusBadRequest, "Email is required")
		return
	}

	if req.Password == "" {
		Error(w, http.StatusBadRequest, "Password is required")
		return
	}

	res, err := h.Client.Login(ctx, &auth.LoginRequest{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		Error(w, http.StatusUnauthorized, "Invalid email or password")
		return
	}

	if res.User == nil {
		Error(w, http.StatusInternalServerError, "User data is missing in response")
		return
	}

	response := map[string]interface{}{
		"token":      res.Token,
		"expires_at": res.ExpiresAt.AsTime(),
		"user":       mapUserToResponse(res.User),
	}

	JSON(w, http.StatusOK, response)
}

// @Summary Register creates a new user account.
// @Description Registers a new user with email, username, and password.
// @Tags Auth
// @Accept json
// @Produce json
// @Param user body object{email=string,username=string,password=string,first_name=string,last_name=string} true "User registration data"
// @Success 201 {object} Response{data=object{token=string,expires_at=string,user=UserResponse}} "Registration successful"
// @Failure 400 {object} Response{error=string} "Bad Request"
// @Failure 500 {object} Response{error=string} "Internal Server Error"
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
		Error(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Email == "" {
		Error(w, http.StatusBadRequest, "Email is required")
		return
	}

	if req.Username == "" {
		Error(w, http.StatusBadRequest, "Username is required")
		return
	}

	if req.Password == "" {
		Error(w, http.StatusBadRequest, "Password is required")
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
		Error(w, http.StatusBadRequest, err.Error())
		return
	}

	if res.User == nil {
		Error(w, http.StatusInternalServerError, "User data is missing in response")
		return
	}

	response := map[string]interface{}{
		"token":      res.Token,
		"expires_at": res.ExpiresAt.AsTime(),
		"user":       mapUserToResponse(res.User),
	}

	Created(w, response)
}

// @Summary Get User
// @Description Get current user profile
// @Tags Auth
// @Security ApiKeyAuth
// @Produce json
// @Success 200 {object} Response{data=UserResponse} "User Profile"
// @Failure 401 {object} Response{error=string} "Unauthorized"
// @Failure 404 {object} Response{error=string} "Not Found"
// @Failure 500 {object} Response{error=string} "Internal Server Error"
// @Router /auth/user [get]
func (h *AuthHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	claims, ok := authMiddleware.GetUserFromContext(r.Context())
	if !ok {
		Error(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	resp, err := h.Client.GetUser(ctx, &auth.GetUserRequest{
		Id: int64(claims.UserId),
	})
	if err != nil {
		Error(w, http.StatusInternalServerError, "Failed to fetch user profile")
		return
	}

	if resp.User == nil {
		Error(w, http.StatusNotFound, "User not found")
		return
	}

	JSON(w, http.StatusOK, mapUserToResponse(resp.User))
}

// @Summary Update Profile
// @Description Update current user profile
// @Tags Auth
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param user body object{first_name=string,last_name=string} true "Update data"
// @Success 200 {object} Response{data=UserResponse} "Updated User"
// @Failure 401 {object} Response{error=string} "Unauthorized"
// @Failure 500 {object} Response{error=string} "Internal Server Error"
// @Router /auth/user [put]
func (h *AuthHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	claims, ok := authMiddleware.GetUserFromContext(r.Context())
	if !ok {
		Error(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req struct {
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		Error(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	resp, err := h.Client.UpdateUser(ctx, &auth.UpdateUserRequest{
		Id:        int64(claims.UserId),
		FirstName: req.FirstName,
		LastName:  req.LastName,
	})
	if err != nil {
		Error(w, http.StatusInternalServerError, "Failed to update profile")
		return
	}

	if resp.User == nil {
		Error(w, http.StatusInternalServerError, "Failed to retrieve updated user data")
		return
	}

	JSON(w, http.StatusOK, mapUserToResponse(resp.User))
}

// @Summary Forgot Password
// @Description Request a password reset email
// @Tags Auth
// @Accept json
// @Produce json
// @Param body body object{email=string} true "Email address"
// @Success 200 {object} object{success=bool,message=string} "Success response"
// @Router /auth/forgot-password [post]
func (h *AuthHandler) ForgotPassword(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email string `json:"email"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	res, err := h.Client.ForgotPassword(ctx, &auth.ForgotPasswordRequest{Email: req.Email})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}

// @Summary Reset Password
// @Description Reset password using a token
// @Tags Auth
// @Accept json
// @Produce json
// @Param body body object{token=string,new_password=string} true "Reset data"
// @Success 200 {object} object{success=bool,message=string} "Success response"
// @Router /auth/reset-password [post]
func (h *AuthHandler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Token       string `json:"token"`
		NewPassword string `json:"new_password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	res, err := h.Client.ResetPassword(ctx, &auth.ResetPasswordRequest{
		Token:       req.Token,
		NewPassword: req.NewPassword,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}

func mapUserToResponse(u *auth.User) UserResponse {
	return UserResponse{
		ID:        u.Id,
		Email:     u.Email,
		Username:  u.Username,
		FirstName: u.FirstName,
		LastName:  u.LastName,
	}
}
