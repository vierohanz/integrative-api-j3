package middlewares

import (
	"gofiber-starterkit/app/models"
	"gofiber-starterkit/app/shared"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/uptrace/bun"
)

func AuthRequired(db *bun.DB) fiber.Handler {
	return func(ctx fiber.Ctx) error {
		token := ctx.Cookies("access_token")

		if token == "" {
			return shared.ErrUnauthorized("Authentication required")
		}

		pat := new(models.PersonalAccessToken)
		err := db.NewSelect().
			Model(pat).
			Where("token = ?", token).
			Where("expires_at > ?", time.Now()).
			Scan(ctx.Context())

		if err != nil {
			return shared.ErrUnauthorized("Invalid or expired token")
		}

		ctx.Locals("user_id", pat.UserID)

		return ctx.Next()
	}
}
