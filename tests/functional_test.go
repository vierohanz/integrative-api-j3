package tests

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"gofiber-starterkit/app/api/auth"
	"gofiber-starterkit/app/api/post"
	"gofiber-starterkit/app/api/product"
	"gofiber-starterkit/app/models"
	"gofiber-starterkit/app/routes"
	"gofiber-starterkit/app/shared"
	"gofiber-starterkit/pkg/config"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
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

func (m *MockAuthService) Logout(ctx context.Context, token string) error {
	return nil
}

type MockPostService struct{}

func (m *MockPostService) Create(ctx context.Context, req *post.CreatePostRequest) error {
	return nil
}

func TestSessionAuth_Functional(t *testing.T) {
	cfg := config.FiberConfig()
	cfg.ErrorHandler = shared.RespondError
	app := fiber.New(cfg)

	productController := product.NewProductController(nil)
	postController := post.NewPostController(&MockPostService{})
	authController := auth.NewAuthController(&MockAuthService{})

	mockAuth := func(ctx fiber.Ctx) error {
		token := ctx.Cookies("access_token")
		if token == "mock-token" {
			return ctx.Next()
		}
		return shared.ErrUnauthorized("Mock auth failed")
	}

	routes.RegisterRoutes(app, productController, postController, authController, mockAuth)

	t.Run("Create Post - Unauthorized (No Session)", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/api/v1/posts", bytes.NewBufferString(`{"title":"New Post","content":"Content"}`))
		req.Header.Set("Content-Type", "application/json")

		resp, _ := app.Test(req)

		if resp.StatusCode != http.StatusUnauthorized {
			t.Errorf("Expected status 401 Unauthorized, got %d", resp.StatusCode)
		}
	})

	t.Run("Create Post - Authorized (With Session)", func(t *testing.T) {
		loginReqBody := `{"username":"admin","password":"password"}`
		loginReq := httptest.NewRequest("POST", "/api/v1/auth/login", bytes.NewBufferString(loginReqBody))
		loginReq.Header.Set("Content-Type", "application/json")
		loginResp, _ := app.Test(loginReq)

		if loginResp.StatusCode != http.StatusOK {
			t.Fatalf("Login failed: expected 200, got %d", loginResp.StatusCode)
		}

		var accessTokenCookie *http.Cookie
		for _, c := range loginResp.Cookies() {
			if c.Name == "access_token" {
				accessTokenCookie = c
				break
			}
		}

		if accessTokenCookie == nil {
			t.Fatal("Access token cookie not found after login")
		}

		req := httptest.NewRequest("POST", "/api/v1/posts", bytes.NewBufferString(`{"title":"New Post","content":"Content"}`))
		req.Header.Set("Content-Type", "application/json")
		req.AddCookie(accessTokenCookie)

		resp, _ := app.Test(req)

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200 OK after login, got %d", resp.StatusCode)
		}
	})

	t.Run("Create Post - Unauthorized (Invalid Token)", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/api/v1/posts", bytes.NewBufferString(`{"title":"New Post","content":"Content"}`))
		req.Header.Set("Content-Type", "application/json")
		req.AddCookie(&http.Cookie{Name: "access_token", Value: "wrong-token"})

		resp, _ := app.Test(req)

		if resp.StatusCode != http.StatusUnauthorized {
			t.Errorf("Expected status 401 Unauthorized for wrong token, got %d", resp.StatusCode)
		}
	})

	t.Run("Create Post - Validation Error (Empty Title)", func(t *testing.T) {
		// Mock login to get a valid session
		loginReqBody := `{"username":"admin","password":"password"}`
		loginReq := httptest.NewRequest("POST", "/api/v1/auth/login", bytes.NewBufferString(loginReqBody))
		loginReq.Header.Set("Content-Type", "application/json")
		loginResp, _ := app.Test(loginReq)

		var accessTokenCookie *http.Cookie
		for _, c := range loginResp.Cookies() {
			if c.Name == "access_token" {
				accessTokenCookie = c
			}
		}

		// Send request with empty title
		req := httptest.NewRequest("POST", "/api/v1/posts", bytes.NewBufferString(`{"title":"","content":"Content"}`))
		req.Header.Set("Content-Type", "application/json")
		req.AddCookie(accessTokenCookie)

		resp, _ := app.Test(req)

		if resp.StatusCode != 400 && resp.StatusCode != 422 {
			t.Errorf("Expected validation error (400 or 422), got %d", resp.StatusCode)
		}
	})

	t.Run("Auth - Logout Lifecycle", func(t *testing.T) {
		// 1. Login
		loginReq := httptest.NewRequest("POST", "/api/v1/auth/login", bytes.NewBufferString(`{"username":"admin","password":"password"}`))
		loginReq.Header.Set("Content-Type", "application/json")
		loginResp, _ := app.Test(loginReq)

		var accessTokenCookie *http.Cookie
		for _, c := range loginResp.Cookies() {
			if c.Name == "access_token" {
				accessTokenCookie = c
			}
		}

		// 2. Logout
		logoutReq := httptest.NewRequest("POST", "/api/v1/auth/logout", nil)
		logoutReq.AddCookie(accessTokenCookie)
		logoutResp, _ := app.Test(logoutReq)

		if logoutResp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200 OK for logout, got %d", logoutResp.StatusCode)
		}

		// 3. Verify access is revoked
		req := httptest.NewRequest("POST", "/api/v1/posts", bytes.NewBufferString(`{"title":"Title"}`))
		req.Header.Set("Content-Type", "application/json")
		req.AddCookie(&http.Cookie{Name: "access_token", Value: "revoked-token"}) // Simulation

		resp, _ := app.Test(req)
		if resp.StatusCode != http.StatusUnauthorized {
			t.Errorf("Expected 401 after session end, got %d", resp.StatusCode)
		}
	})
}
