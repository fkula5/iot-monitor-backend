package services

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/skni-kod/iot-monitor-backend/internal/auth"
	"github.com/skni-kod/iot-monitor-backend/services/auth/ent"
	"github.com/skni-kod/iot-monitor-backend/services/auth/storage"
)

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}

type UpdateRequest struct {
	FirstName string `json:"first_name,omitempty" validate:"max=100"`
	LastName  string `json:"last_name,omitempty" validate:"max=100"`
}

type RegisterRequest struct {
	Email     string `json:"email" validate:"required,email"`
	Username  string `json:"username" validate:"required,min=3,max=50"`
	Password  string `json:"password" validate:"required,min=8"`
	FirstName string `json:"first_name,omitempty" validate:"max=100"`
	LastName  string `json:"last_name,omitempty" validate:"max=100"`
}

type AuthResponse struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
	User      UserInfo  `json:"user"`
}

type UserInfo struct {
	ID        int    `json:"id"`
	Email     string `json:"email"`
	Username  string `json:"username"`
	FirstName string `json:"first_name,omitempty"`
	LastName  string `json:"last_name,omitempty"`
}

type IAuthService interface {
	Login(ctx context.Context, req *LoginRequest) (*AuthResponse, error)
	Register(ctx context.Context, req *RegisterRequest) (*AuthResponse, error)
	ValidateToken(ctx context.Context, token string) (*ent.User, error)
	GetUserByID(ctx context.Context, userID int) (*ent.User, error)
	Update(ctx context.Context, userID int, req *UpdateRequest) (*ent.User, error)
	ForgotPassword(ctx context.Context, email string) error
	ResetPassword(ctx context.Context, token string, newPassword string) error
}

type AuthService struct {
	userStorage     storage.IUserStorage
	jwtService      *auth.JWTService
	passwordService *auth.PasswordService
	mailer          *Mailer
}

func NewAuthService(userStorage storage.IUserStorage, mailer *Mailer) IAuthService {
	return &AuthService{
		userStorage:     userStorage,
		jwtService:      auth.NewJWTService(),
		passwordService: auth.NewPasswordService(),
		mailer:          mailer,
	}
}

// GetUserByID implements IAuthService.
func (s *AuthService) GetUserByID(ctx context.Context, userID int) (*ent.User, error) {
	user, err := s.userStorage.Get(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}
	return user, nil
}

// Login implements IAuthService.
func (s *AuthService) Login(ctx context.Context, req *LoginRequest) (*AuthResponse, error) {
	user, err := s.userStorage.GetByEmail(ctx, req.Email)
	if err != nil {
		return nil, fmt.Errorf("invalid email or password")
	}

	if err := s.passwordService.ValidatePassword(req.Password, user.PasswordHash); err != nil {
		return nil, fmt.Errorf("invalid email or password")
	}

	if !user.Active {
		return nil, fmt.Errorf("account is inactive")
	}

	token, err := s.jwtService.GenerateToken(user.ID, user.Username, user.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	expirationHours := 24
	if envHours := os.Getenv("JWT_EXPIRATION_HOURS"); envHours != "" {
		if hours, err := strconv.Atoi(envHours); err == nil {
			expirationHours = hours
		}
	}

	return &AuthResponse{
		Token:     token,
		ExpiresAt: time.Now().Add(time.Duration(expirationHours) * time.Hour),
		User: UserInfo{
			ID:        user.ID,
			Email:     user.Email,
			Username:  user.Username,
			FirstName: user.FirstName,
			LastName:  user.LastName,
		},
	}, nil
}

// ValidateToken implements IAuthService.
func (s *AuthService) ValidateToken(ctx context.Context, token string) (*ent.User, error) {
	claims, err := s.jwtService.ValidateToken(token)
	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	user, err := s.userStorage.Get(ctx, claims.UserId)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	if !user.Active {
		return nil, fmt.Errorf("account is inactive")
	}

	return user, nil
}

func (s *AuthService) Register(ctx context.Context, req *RegisterRequest) (*AuthResponse, error) {
	if exists, _ := s.userStorage.ExistsByEmail(ctx, req.Email); exists {
		return nil, fmt.Errorf("user with this email already exists")
	}

	if exists, _ := s.userStorage.ExistsByUsername(ctx, req.Username); exists {
		return nil, fmt.Errorf("user with this username already exists")
	}

	hashedPassword, err := s.passwordService.HashPassword(req.Password)
	if err != nil {
		return nil, err
	}

	user := &ent.User{
		Email:        req.Email,
		Username:     req.Username,
		PasswordHash: hashedPassword,
		FirstName:    req.FirstName,
		LastName:     req.LastName,
		Active:       true,
	}

	createdUser, err := s.userStorage.Create(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	token, err := s.jwtService.GenerateToken(createdUser.ID, createdUser.Username, createdUser.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	expirationHours := 24
	if envHours := os.Getenv("JWT_EXPIRATION_HOURS"); envHours != "" {
		if hours, err := strconv.Atoi(envHours); err == nil {
			expirationHours = hours
		}
	}

	return &AuthResponse{
		Token:     token,
		ExpiresAt: time.Now().Add(time.Duration(expirationHours) * time.Hour),
		User: UserInfo{
			ID:        createdUser.ID,
			Email:     createdUser.Email,
			Username:  createdUser.Username,
			FirstName: createdUser.FirstName,
			LastName:  createdUser.LastName,
		},
	}, nil
}

func (s *AuthService) Update(ctx context.Context, userID int, req *UpdateRequest) (*ent.User, error) {
	existingUser, err := s.userStorage.Get(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	existingUser.FirstName = req.FirstName
	existingUser.LastName = req.LastName

	updatedUser, err := s.userStorage.Update(ctx, existingUser)
	if err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	return updatedUser, nil
}

func (s *AuthService) ForgotPassword(ctx context.Context, email string) error {
	u, err := s.userStorage.GetByEmail(ctx, email)
	if err != nil {
		return err // Or return nil to avoid email enumeration
	}

	token := uuid.New().String()
	expiresAt := time.Now().Add(time.Hour)

	if err := s.userStorage.SetResetToken(ctx, email, token, expiresAt); err != nil {
		return err
	}

	return s.mailer.SendResetPasswordEmail(u.Email, token)
}

func (s *AuthService) ResetPassword(ctx context.Context, token string, newPassword string) error {
	u, err := s.userStorage.GetByResetToken(ctx, token)
	if err != nil {
		return fmt.Errorf("invalid or expired token")
	}

	if u.ResetTokenExpires.Before(time.Now()) {
		return fmt.Errorf("token expired")
	}

	hash, err := s.passwordService.HashPassword(newPassword)
	if err != nil {
		return err
	}

	return s.userStorage.UpdatePassword(ctx, u.ID, hash)
}
