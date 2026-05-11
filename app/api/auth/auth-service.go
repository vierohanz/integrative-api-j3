package auth

import (
	"context"
	"gofiber-starterkit/app/models"
	"gofiber-starterkit/app/shared"
	"gofiber-starterkit/pkg/utils"

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
	token, err := utils.GenerateJWT(userID)
	if err != nil {
		return "", shared.ErrInternalServerError("Failed to generate JWT")
	}
	return token, nil
}

func (s *authService) Logout(ctx context.Context, token string) error {
	// JWT is stateless, logout is handled by clearing the cookie in the controller
	return nil
}
