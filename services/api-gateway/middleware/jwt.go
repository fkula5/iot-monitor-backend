package middleware

import (
	"context"
	"net/http"
	"strings"
	"time"

	"go.uber.org/zap"

	"github.com/skni-kod/iot-monitor-backend/internal/auth"
	"github.com/skni-kod/iot-monitor-backend/pkg/logger"
)

type contextKey string

const UserContextKey contextKey = "user"

type AuthMiddleware struct {
	jwtService *auth.JWTService
}

func NewAuthMiddleware() *AuthMiddleware {
	return &AuthMiddleware{
		jwtService: auth.NewJWTService(),
	}
}

func (m *AuthMiddleware) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var token string

		authHeader := r.Header.Get("Authorization")
		if authHeader != "" {
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) == 2 && parts[0] == "Bearer" {
				token = parts[1]
			}
		}

		if token == "" {
			cookie, err := r.Cookie("token")
			if err == nil {
				token = cookie.Value
			}
		}

		if token == "" {
			logger.Warn("Missing authentication token", zap.String("path", r.URL.Path))
			http.Error(w, "Authorization header or token cookie required", http.StatusUnauthorized)
			return
		}

		claims, err := m.jwtService.ValidateToken(token)
		if err != nil {
			logger.Warn("Invalid or expired token", zap.Error(err), zap.String("path", r.URL.Path))
			http.Error(w, "Invalid or expired token: "+err.Error(), http.StatusUnauthorized)
			return
		}

		if claims.ExpiresAt != nil && claims.ExpiresAt.Time.Before(time.Now()) {
			logger.Warn("Token expired", zap.Int("userId", claims.UserId), zap.String("path", r.URL.Path))
			http.Error(w, "Token expired", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), UserContextKey, claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func GetUserFromContext(ctx context.Context) (*auth.Claims, bool) {
	claims, ok := ctx.Value(UserContextKey).(*auth.Claims)
	return claims, ok
}
