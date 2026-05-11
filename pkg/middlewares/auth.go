package middlewares

import (
	"gofiber-starterkit/app/shared"
	"gofiber-starterkit/pkg/utils"

	"github.com/gofiber/fiber/v3"
)

func AuthRequired(ctx fiber.Ctx) error {
	token := ctx.Cookies("access_token")

	if token == "" {
		return shared.ErrUnauthorized("Authentication required")
	}

	claims, err := utils.ValidateJWT(token)
	if err != nil {
		return shared.ErrUnauthorized("Invalid or expired token")
	}

	ctx.Locals("user_id", claims.UserID)

	return ctx.Next()
}
