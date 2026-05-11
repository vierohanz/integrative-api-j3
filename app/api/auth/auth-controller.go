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

	accessToken, err := c.service.CreateToken(ctx.Context(), user.ID)
	if err != nil {
		return err
	}

	refreshToken, err := c.service.CreateRefreshToken(ctx.Context(), user.ID)
	if err != nil {
		return err
	}

	// Set Cookies
	ctx.Cookie(&fiber.Cookie{
		Name:     "access_token",
		Value:    accessToken,
		Expires:  time.Now().Add(24 * time.Hour),
		HTTPOnly: true,
		SameSite: "Lax",
	})

	ctx.Cookie(&fiber.Cookie{
		Name:     "refresh_token",
		Value:    refreshToken,
		Expires:  time.Now().Add(7 * 24 * time.Hour),
		HTTPOnly: true,
		SameSite: "Lax",
	})
	
	return shared.RespondSuccess(ctx, "Logged in successfully", TransformUser(user))
}

func (c *AuthController) Refresh(ctx fiber.Ctx) error {
	refreshToken := ctx.Cookies("refresh_token")
	if refreshToken == "" {
		return shared.ErrUnauthorized("Refresh token missing")
	}

	newAccessToken, newRefreshToken, err := c.service.RefreshToken(ctx.Context(), refreshToken)
	if err != nil {
		return err
	}

	// Update Cookies
	ctx.Cookie(&fiber.Cookie{
		Name:     "access_token",
		Value:    newAccessToken,
		Expires:  time.Now().Add(24 * time.Hour),
		HTTPOnly: true,
		SameSite: "Lax",
	})

	ctx.Cookie(&fiber.Cookie{
		Name:     "refresh_token",
		Value:    newRefreshToken,
		Expires:  time.Now().Add(7 * 24 * time.Hour),
		HTTPOnly: true,
		SameSite: "Lax",
	})

	return shared.RespondSuccess(ctx, "Token refreshed successfully", nil)
}

func (c *AuthController) Logout(ctx fiber.Ctx) error {
	accessToken := ctx.Cookies("access_token")
	refreshToken := ctx.Cookies("refresh_token")
	
	_ = c.service.Logout(ctx.Context(), accessToken, refreshToken)

	// Clear Cookies
	ctx.Cookie(&fiber.Cookie{
		Name:     "access_token",
		Value:    "",
		Expires:  time.Now().Add(-1 * time.Hour),
		HTTPOnly: true,
	})

	ctx.Cookie(&fiber.Cookie{
		Name:     "refresh_token",
		Value:    "",
		Expires:  time.Now().Add(-1 * time.Hour),
		HTTPOnly: true,
	})
	
	return shared.RespondSuccess(ctx, "Logged out successfully", nil)
}
