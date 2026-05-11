package routes

import (
	"gofiber-starterkit/app/api/auth"
	"gofiber-starterkit/app/api/post"
	"gofiber-starterkit/app/api/product"

	"github.com/gofiber/fiber/v3"
)

func RegisterRoutes(
	app *fiber.App,
	productController *product.ProductController,
	postController *post.PostController,
	authController *auth.AuthController,
	authMiddleware fiber.Handler,
) {
	const ApiVersion = "/api/v1"

	api := app.Group(ApiVersion)

	authGroup := api.Group("/auth")
	authGroup.Post("/login", authController.Login)
	authGroup.Post("/logout", authController.Logout)

	products := api.Group("/products")
	products.Get("", productController.List)
	products.Get("/:id/show", productController.Get)
	products.Post("", authMiddleware, productController.Create)
	products.Put("/:id/update", productController.Update)
	products.Patch("/:id/status", productController.UpdateStatus)
	products.Delete("/:id/delete", productController.Delete)

	posts := api.Group("/posts")
	posts.Post("", authMiddleware, postController.Create)
}
