package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/skni-kod/iot-monitor-backend/services/auth/ent"
	"github.com/skni-kod/iot-monitor-backend/services/auth/ent/user"
)

type IUserStorage interface {
	Create(ctx context.Context, user *ent.User) (*ent.User, error)
	Get(ctx context.Context, id int) (*ent.User, error)
	ExistsByEmail(ctx context.Context, email string) (bool, error)
	ExistsByUsername(ctx context.Context, username string) (bool, error)
	Update(ctx context.Context, userData *ent.User) (*ent.User, error)
	List(ctx context.Context) ([]*ent.User, error)
	SetRefreshToken(ctx context.Context, userID int, token string, expires time.Time) error
	ClearRefreshToken(ctx context.Context, userID int) error
	GetByEmail(ctx context.Context, email string) (*ent.User, error)
}

type UserStorage struct {
	client *ent.Client
}

func NewUserStorage(client *ent.Client) IUserStorage {
	return &UserStorage{client: client}
}

func (s *UserStorage) Create(ctx context.Context, userData *ent.User) (*ent.User, error) {
	user, err := s.client.User.Create().
		SetEmail(userData.Email).
		SetUsername(userData.Username).
		SetPasswordHash(userData.PasswordHash).
		SetNillableFirstName(&userData.FirstName).
		SetNillableLastName(&userData.LastName).
		SetActive(userData.Active).
		Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return user, nil
}

func (s *UserStorage) Get(ctx context.Context, id int) (*ent.User, error) {
	return s.client.User.Query().Where(user.ID(id)).Only(ctx)
}

func (s *UserStorage) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	return s.client.User.Query().Where(user.Email(email)).Exist(ctx)
}

func (s *UserStorage) Update(ctx context.Context, userData *ent.User) (*ent.User, error) {
	return s.client.User.UpdateOneID(userData.ID).
		SetNillableFirstName(&userData.FirstName).
		SetNillableLastName(&userData.LastName).
		SetActive(userData.Active).
		Save(ctx)
}

func (s *UserStorage) List(ctx context.Context) ([]*ent.User, error) {
	return s.client.User.Query().All(ctx)
}

func (s *UserStorage) SetRefreshToken(ctx context.Context, userID int, token string, expires time.Time) error {
	_, err := s.client.User.UpdateOneID(userID).
		SetRefreshToken(token).
		SetRefreshTokenExpires(expires).
		Save(ctx)
	return err
}

func (s *UserStorage) ClearRefreshToken(ctx context.Context, userID int) error {
	_, err := s.client.User.UpdateOneID(userID).
		ClearRefreshToken().
		ClearRefreshTokenExpires().
		Save(ctx)
	return err
}

func (s *UserStorage) ExistsByUsername(ctx context.Context, username string) (bool, error) {
	return s.client.User.Query().Where(user.Username(username)).Exist(ctx)
}

func (s *UserStorage) GetByEmail(ctx context.Context, email string) (*ent.User, error) {
	return s.client.User.Query().Where(user.Email(email)).Only(ctx)
}
