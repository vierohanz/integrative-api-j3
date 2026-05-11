package post

import (
	"gofiber-starterkit/app/shared"
	"gofiber-starterkit/pkg/middlewares"

	"github.com/gofiber/fiber/v3"
)

type PostController struct {
	service PostService
}

func NewPostController(service PostService) *PostController {
	return &PostController{service: service}
}

func (c *PostController) Create(ctx fiber.Ctx) error {
	var req CreatePostRequest
	if err := ctx.Bind().Body(&req); err != nil {
		return err
	}

	if err := middlewares.ValidateStruct(&req); err != nil {
		return err
	}

	if err := c.service.Create(ctx.Context(), &req); err != nil {
		return err
	}

	return shared.RespondSuccess(ctx, "Post created successfully", nil)
}
