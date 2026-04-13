package services

import (
	"context"
	"time"

	"gofiber-starterkit/app/api/types"
	"gofiber-starterkit/app/models"
	"gofiber-starterkit/app/shared"
	"gofiber-starterkit/pkg/client/dragonfly"
	"gofiber-starterkit/pkg/client/rustfs"
	"gofiber-starterkit/pkg/utils"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type UserService struct {
	db              *bun.DB
	dragonflyClient *dragonfly.DragonflyClient
	rustfsClient    *rustfs.RustfsClient
}

func NewUserService(db *bun.DB, dragonflyClient *dragonfly.DragonflyClient, rustfsClient *rustfs.RustfsClient) *UserService {
	return &UserService{
		db:              db,
		dragonflyClient: dragonflyClient,
		rustfsClient:    rustfsClient,
	}
}

func (s *UserService) Register(ctx context.Context, req *types.RegisterRequest) (*models.User, *types.AuthResponse, error) {
	exists, err := s.db.NewSelect().Model((*models.User)(nil)).Where("email = ?", req.Email).Exists(ctx)
	if err != nil {
		return nil, nil, shared.ErrInternalServerError("Failed to check email")
	}
	if exists {
		return nil, nil, shared.ErrConflict("Email already registered")
	}

	exists, err = s.db.NewSelect().Model((*models.User)(nil)).Where("username = ?", req.Username).Exists(ctx)
	if err != nil {
		return nil, nil, shared.ErrInternalServerError("Failed to check username")
	}
	if exists {
		return nil, nil, shared.ErrConflict("Username already taken")
	}

	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		return nil, nil, shared.ErrInternalServerError("Failed to hash password")
	}

	user := &models.User{
		Email:        req.Email,
		Username:     req.Username,
		PasswordHash: &hashedPassword,
	}

	_, err = s.db.NewInsert().Model(user).Exec(ctx)
	if err != nil {
		return nil, nil, shared.ErrInternalServerError("Failed to create user")
	}

	accessToken, refreshToken, err := utils.GenerateTokenPair(s.dragonflyClient.Client, user.ID.String())
	if err != nil {
		return nil, nil, shared.ErrInternalServerError("Failed to generate tokens")
	}

	return user, &types.AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (s *UserService) Login(ctx context.Context, req *types.LoginRequest) (*models.User, *types.AuthResponse, error) {
	user := new(models.User)
	err := s.db.NewSelect().Model(user).Where("email = ?", req.Email).Scan(ctx)
	if err != nil {
		return nil, nil, shared.ErrUnauthorized("Invalid email or password")
	}

	if user.PasswordHash == nil || !utils.CheckPasswordHash(req.Password, *user.PasswordHash) {
		return nil, nil, shared.ErrUnauthorized("Invalid email or password")
	}

	accessToken, refreshToken, err := utils.GenerateTokenPair(s.dragonflyClient.Client, user.ID.String())
	if err != nil {
		return nil, nil, shared.ErrInternalServerError("Failed to generate tokens")
	}

	return user, &types.AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (s *UserService) RefreshToken(ctx context.Context, refreshTokenStr string) (*types.AuthResponse, error) {
	claims, err := utils.ValidateRefreshToken(s.dragonflyClient.Client, refreshTokenStr)
	if err != nil {
		return nil, shared.ErrUnauthorized("Invalid refresh token")
	}

	accessToken, refreshToken, err := utils.RotateRefreshToken(s.dragonflyClient.Client, claims)
	if err != nil {
		return nil, shared.ErrInternalServerError("Failed to generate tokens")
	}

	return &types.AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (s *UserService) Logout(ctx context.Context, tokenID string, userID string) error {
	return utils.RevokeAccessToken(s.dragonflyClient.Client, tokenID)
}

func (s *UserService) LogoutAll(ctx context.Context, userID string) error {
	return utils.RevokeAllUserTokens(s.dragonflyClient.Client, userID)
}

func (s *UserService) GetByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	user := new(models.User)
	err := s.db.NewSelect().Model(user).Where("id = ?", id).Scan(ctx)
	if err != nil {
		return nil, shared.ErrNotFound("User not found")
	}
	return user, nil
}

func (s *UserService) List(ctx context.Context, page, perPage int) ([]*models.User, int, error) {
	var users []*models.User
	count, err := s.db.NewSelect().
		Model(&users).
		Limit(perPage).
		Offset((page - 1) * perPage).
		Order("created_at DESC").
		ScanAndCount(ctx)
	if err != nil {
		return nil, 0, shared.ErrInternalServerError("Failed to list users")
	}
	return users, count, nil
}

func (s *UserService) Update(ctx context.Context, id uuid.UUID, req *types.UpdateProfileRequest) (*models.User, error) {
	user, err := s.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if req.Username != nil {
		exists, err := s.db.NewSelect().Model((*models.User)(nil)).
			Where("username = ? AND id != ?", *req.Username, id).Exists(ctx)
		if err != nil {
			return nil, shared.ErrInternalServerError("Failed to check username")
		}
		if exists {
			return nil, shared.ErrConflict("Username already taken")
		}
		user.Username = *req.Username
	}
	if req.Avatar != nil {
		user.Avatar = req.Avatar
	}
	if req.Bio != nil {
		user.Bio = req.Bio
	}
	user.UpdatedAt = time.Now()

	_, err = s.db.NewUpdate().Model(user).WherePK().Exec(ctx)
	if err != nil {
		return nil, shared.ErrInternalServerError("Failed to update user")
	}

	return user, nil
}

func (s *UserService) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := s.db.NewDelete().Model((*models.User)(nil)).Where("id = ?", id).Exec(ctx)
	if err != nil {
		return shared.ErrInternalServerError("Failed to delete user")
	}

	utils.RevokeAllUserTokens(s.dragonflyClient.Client, id.String())

	return nil
}

func (s *UserService) GetRustfsClient() *rustfs.RustfsClient {
	return s.rustfsClient
}
