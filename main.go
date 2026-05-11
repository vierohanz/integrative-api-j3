package main

import (
	"gofiber-starterkit/app/api/auth"
	"gofiber-starterkit/app/api/product"
	"gofiber-starterkit/app/routes"
	"gofiber-starterkit/app/shared"
	"gofiber-starterkit/pkg/client/db"
	"gofiber-starterkit/pkg/client/dragonfly"
	"gofiber-starterkit/pkg/client/rustfs"
	"gofiber-starterkit/pkg/config"
	"gofiber-starterkit/pkg/middlewares"
	"gofiber-starterkit/pkg/utils"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/compress"
	"github.com/gofiber/fiber/v3/middleware/healthcheck"
	"github.com/gofiber/fiber/v3/middleware/limiter"
	"github.com/rs/zerolog/log"
	"github.com/uptrace/bun"
	"go.uber.org/dig"
)

func main() {
	c := dig.New()

	c.Provide(db.New)
	c.Provide(dragonfly.New)
	c.Provide(rustfs.New)

	c.Provide(product.NewProductService)
	c.Provide(auth.NewAuthService)
	c.Provide(product.NewProductController)
	c.Provide(auth.NewAuthController)

	c.Provide(func(dragonflyClient *dragonfly.DragonflyClient) *fiber.App {
		cfg := config.FiberConfig()
		cfg.ErrorHandler = shared.RespondError

		app := fiber.New(cfg)

		app.Use(compress.New(compress.Config{
			Level: compress.LevelBestSpeed,
		}))

		app.Use(limiter.New(limiter.Config{
			Max:        100,
			Expiration: 1 * time.Minute,
		}))

		middlewares.FiberMiddleware(app)

		app.Get(healthcheck.LivenessEndpoint, healthcheck.New())

		return app
	})

	c.Invoke(func(
		app *fiber.App,
		productController *product.ProductController,
		authController *auth.AuthController,
		dbClient *bun.DB,
		dragonflyClient *dragonfly.DragonflyClient,
	) {
		routes.RegisterRoutes(app, productController, authController, middlewares.AuthRequired)

		defer dbClient.Close()
		defer dragonflyClient.Client.Close()

		if err := utils.StartServerWithGracefulShutdown(app); err != nil {
			log.Error().Err(err).Msg("Server error")
		}
	})
}
