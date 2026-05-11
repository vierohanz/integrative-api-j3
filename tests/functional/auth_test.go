package functional

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"gofiber-starterkit/app/api/auth"
	"gofiber-starterkit/app/api/product"
	"gofiber-starterkit/app/models"
	"gofiber-starterkit/app/routes"
	"gofiber-starterkit/app/shared"
	"gofiber-starterkit/pkg/config"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type MockAuthService struct{}

func (m *MockAuthService) Login(ctx context.Context, username, password string) (*models.User, error) {
	return &models.User{
		ID:       uuid.New(),
		Username: username,
		Password: password,
	}, nil
}

func (m *MockAuthService) CreateToken(ctx context.Context, userID uuid.UUID) (string, error) {
	return "mock-token", nil
}

func (m *MockAuthService) CreateRefreshToken(ctx context.Context, userID uuid.UUID) (string, error) {
	return "mock-refresh-token", nil
}

func (m *MockAuthService) RefreshToken(ctx context.Context, refreshToken string) (string, string, error) {
	return "new-mock-token", "new-mock-refresh-token", nil
}

func (m *MockAuthService) Logout(ctx context.Context, token string, refreshToken string) error {
	return nil
}

func TestSessionAuth_Functional(t *testing.T) {
	cfg := config.FiberConfig()
	cfg.ErrorHandler = shared.RespondError
	app := fiber.New(cfg)

	t.Log("==================================================")
	t.Log(" TEST SUITE: Auth Functional (API Lifecycle)")
	t.Log("==================================================")

	productController := product.NewProductController(nil)
	authController := auth.NewAuthController(&MockAuthService{})

	mockAuth := func(ctx fiber.Ctx) error {
		token := ctx.Cookies("access_token")
		if token == "mock-token" {
			return ctx.Next()
		}
		return shared.ErrUnauthorized("Mock auth failed")
	}

	routes.RegisterRoutes(app, productController, authController, mockAuth)

	t.Run("Auth - Login Lifecycle", func(t *testing.T) {
		t.Log("[STEP 1] Testing Login API...")
		loginReqBody := `{"username":"admin","password":"password"}`
		loginReq := httptest.NewRequest("POST", "/api/v1/auth/login", bytes.NewBufferString(loginReqBody))
		loginReq.Header.Set("Content-Type", "application/json")
		loginResp, err := app.Test(loginReq)

		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, loginResp.StatusCode, "Login should return 200 OK")
		t.Log("  >> Result: LOGIN SUCCESS")
	})

	t.Run("Auth - Logout Lifecycle", func(t *testing.T) {
		t.Log("[STEP 2] Testing Logout API...")
		// 1. Login
		loginReq := httptest.NewRequest("POST", "/api/v1/auth/login", bytes.NewBufferString(`{"username":"admin","password":"password"}`))
		loginReq.Header.Set("Content-Type", "application/json")
		loginResp, err := app.Test(loginReq)
		require.NoError(t, err)

		var accessTokenCookie *http.Cookie
		var refreshTokenCookie *http.Cookie
		for _, c := range loginResp.Cookies() {
			if c.Name == "access_token" {
				accessTokenCookie = c
			}
			if c.Name == "refresh_token" {
				refreshTokenCookie = c
			}
		}
		require.NotNil(t, accessTokenCookie, "Access token should be present")

		// 2. Logout
		logoutReq := httptest.NewRequest("POST", "/api/v1/auth/logout", nil)
		logoutReq.AddCookie(accessTokenCookie)
		logoutReq.AddCookie(refreshTokenCookie)
		logoutResp, err := app.Test(logoutReq)
		require.NoError(t, err)

		assert.Equal(t, http.StatusOK, logoutResp.StatusCode, "Logout should return 200 OK")
		t.Log("  >> Result: LOGOUT SUCCESS")
	})

	t.Run("Auth - Refresh Token Rotation", func(t *testing.T) {
		t.Log("[STEP 3] Testing Dynamic Refresh Token Rotation...")
		// 1. Login to get initial refresh token
		loginReq := httptest.NewRequest("POST", "/api/v1/auth/login", bytes.NewBufferString(`{"username":"admin","password":"password"}`))
		loginReq.Header.Set("Content-Type", "application/json")
		loginResp, err := app.Test(loginReq)
		require.NoError(t, err)

		var oldRefreshToken string
		for _, c := range loginResp.Cookies() {
			if c.Name == "refresh_token" {
				oldRefreshToken = c.Value
			}
		}
		require.NotEmpty(t, oldRefreshToken, "Refresh token should be present after login")

		// 2. Call Refresh endpoint
		refreshReq := httptest.NewRequest("POST", "/api/v1/auth/refresh", nil)
		refreshReq.AddCookie(&http.Cookie{Name: "refresh_token", Value: oldRefreshToken})
		refreshResp, err := app.Test(refreshReq)
		require.NoError(t, err)

		assert.Equal(t, http.StatusOK, refreshResp.StatusCode, "Refresh should return 200 OK")

		// 3. Verify we got NEW tokens (rotation)
		var hasNewAccess, hasNewRefresh bool
		for _, c := range refreshResp.Cookies() {
			if c.Name == "access_token" && c.Value == "new-mock-token" {
				hasNewAccess = true
			}
			if c.Name == "refresh_token" && c.Value == "new-mock-refresh-token" {
				hasNewRefresh = true
			}
		}

		assert.True(t, hasNewAccess, "Did not receive new access token")
		assert.True(t, hasNewRefresh, "Did not receive new rotated refresh token")
		t.Log("  >> Result: ROTATION SUCCESS")
	})
	
	t.Log("--------------------------------------------------")
}
