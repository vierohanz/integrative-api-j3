package auth

import (
	"context"
	"gofiber-starterkit/app/models"
	"gofiber-starterkit/app/shared"
	"gofiber-starterkit/pkg/utils"
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type AuthService interface {
	Login(ctx context.Context, username, password string) (*models.User, error)
	CreateToken(ctx context.Context, userID uuid.UUID) (string, error)
	Logout(ctx context.Context, token string) error
}

type authService struct {
	db *bun.DB
}

func NewAuthService(db *bun.DB) AuthService {
	return &authService{db: db}
}

func (s *authService) Login(ctx context.Context, username, password string) (*models.User, error) {
	user := new(models.User)
	err := s.db.NewSelect().Model(user).Where("username = ?", username).Scan(ctx)
	if err != nil {
		return nil, shared.ErrUnauthorized("Invalid username or password")
	}

	if user.Password != password {
		return nil, shared.ErrUnauthorized("Invalid username or password")
	}

	return user, nil
}

func (s *authService) CreateToken(ctx context.Context, userID uuid.UUID) (string, error) {
	token := utils.GenerateRandomString(32)
	pat := &models.PersonalAccessToken{
		UserID:    userID,
		Token:     token,
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}

	_, err := s.db.NewInsert().Model(pat).Exec(ctx)
	if err != nil {
		return "", shared.ErrInternalServerError("Failed to create access token")
	}

	return token, nil
}

func (s *authService) Logout(ctx context.Context, token string) error {
	_, err := s.db.NewDelete().
		Model((*models.PersonalAccessToken)(nil)).
		Where("token = ?", token).
		Exec(ctx)
	return err
}
