package auth

import (
	"context"
	"gofiber-starterkit/app/models"
	"gofiber-starterkit/app/shared"
	"gofiber-starterkit/pkg/client/dragonfly"
	"gofiber-starterkit/pkg/utils"
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type AuthService interface {
	Login(ctx context.Context, username, password string) (*models.User, error)
	CreateToken(ctx context.Context, userID uuid.UUID) (string, error)
	CreateRefreshToken(ctx context.Context, userID uuid.UUID) (string, error)
	RefreshToken(ctx context.Context, refreshToken string) (string, string, error)
	Logout(ctx context.Context, token string, refreshToken string) error
}

type authService struct {
	db              *bun.DB
	dragonflyClient *dragonfly.DragonflyClient
}

func NewAuthService(db *bun.DB, dragonflyClient *dragonfly.DragonflyClient) AuthService {
	return &authService{
		db:              db,
		dragonflyClient: dragonflyClient,
	}
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

func (s *authService) CreateRefreshToken(ctx context.Context, userID uuid.UUID) (string, error) {
	token := utils.GenerateRandomString(32)
	key := "refresh_token:" + token
	
	err := s.dragonflyClient.Client.Set(ctx, key, userID.String(), 7*24*time.Hour).Err()
	if err != nil {
		return "", shared.ErrInternalServerError("Failed to store refresh token")
	}

	return token, nil
}

func (s *authService) RefreshToken(ctx context.Context, refreshToken string) (string, string, error) {
	key := "refresh_token:" + refreshToken
	
	userIDStr, err := s.dragonflyClient.Client.Get(ctx, key).Result()
	if err != nil {
		return "", "", shared.ErrUnauthorized("Invalid or expired refresh token")
	}

	userID, _ := uuid.Parse(userIDStr)

	// Rotate token: Delete old one
	s.dragonflyClient.Client.Del(ctx, key)

	// Generate new pair
	accessToken, err := s.CreateToken(ctx, userID)
	if err != nil {
		return "", "", err
	}

	newRefreshToken, err := s.CreateRefreshToken(ctx, userID)
	if err != nil {
		return "", "", err
	}

	return accessToken, newRefreshToken, nil
}

func (s *authService) Logout(ctx context.Context, token string, refreshToken string) error {
	if refreshToken != "" {
		s.dragonflyClient.Client.Del(ctx, "refresh_token:"+refreshToken)
	}
	return nil
}
