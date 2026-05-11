package auth

import (
	"gofiber-starterkit/app/shared"
	"gofiber-starterkit/pkg/middlewares"
	"time"

	"github.com/gofiber/fiber/v3"
)

type AuthController struct {
	service AuthService
}

func NewAuthController(service AuthService) *AuthController {
	return &AuthController{service: service}
}

func (c *AuthController) Login(ctx fiber.Ctx) error {
	var req LoginRequest
	if err := ctx.Bind().Body(&req); err != nil {
		return err
	}

	if err := middlewares.ValidateStruct(&req); err != nil {
		return err
	}

	user, err := c.service.Login(ctx.Context(), req.Username, req.Password)
	if err != nil {
		return err
	}

	token, err := c.service.CreateToken(ctx.Context(), user.ID)
	if err != nil {
		return err
	}

	// Set Cookie
	ctx.Cookie(&fiber.Cookie{
		Name:     "access_token",
		Value:    token,
		Expires:  time.Now().Add(24 * time.Hour),
		HTTPOnly: true,
		SameSite: "Lax",
	})
	
	return shared.RespondSuccess(ctx, "Logged in successfully", LoginResponse{
		User_ID:  user.ID.String(),
		Username: user.Username,
	})
}

func (c *AuthController) Logout(ctx fiber.Ctx) error {
	token := ctx.Cookies("access_token")
	if token != "" {
		_ = c.service.Logout(ctx.Context(), token)
	}

	// Clear Cookie
	ctx.Cookie(&fiber.Cookie{
		Name:     "access_token",
		Value:    "",
		Expires:  time.Now().Add(-1 * time.Hour),
		HTTPOnly: true,
	})
	
	return shared.RespondSuccess(ctx, "Logged out successfully", nil)
}
