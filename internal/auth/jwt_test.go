package auth

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewJWTService_RequiresSecret(t *testing.T) {
	t.Setenv("JWT_SECRET", "")
	assert.Panics(t, func() { NewJWTService() })

	t.Setenv("JWT_SECRET", "test-secret")
	s := NewJWTService()
	token, err := s.GenerateToken(1, "user", "user@example.com")
	assert.NoError(t, err)
	claims, err := s.ValidateToken(token)
	assert.NoError(t, err)
	assert.Equal(t, 1, claims.UserId)
}
