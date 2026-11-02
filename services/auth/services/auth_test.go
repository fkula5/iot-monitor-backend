package services

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/skni-kod/iot-monitor-backend/services/auth/ent"
	"github.com/skni-kod/iot-monitor-backend/services/auth/storage"
)

type fakeUserStorage struct {
	storage.IUserStorage
	getErr error
}

func (f *fakeUserStorage) GetByEmail(_ context.Context, _ string) (*ent.User, error) {
	return nil, f.getErr
}

func TestForgotPassword_UnknownEmail(t *testing.T) {
	t.Setenv("JWT_SECRET", "test-secret")

	// The mailer is nil: reaching it would panic, so this also checks that no
	// email is sent for an unknown address.
	s := NewAuthService(&fakeUserStorage{getErr: &ent.NotFoundError{}}, nil)
	assert.NoError(t, s.ForgotPassword(context.Background(), "nobody@example.com"))

	s = NewAuthService(&fakeUserStorage{getErr: errors.New("db down")}, nil)
	assert.Error(t, s.ForgotPassword(context.Background(), "nobody@example.com"))
}
