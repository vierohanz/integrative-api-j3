package post

import (
	"context"

	"github.com/uptrace/bun"
)

type PostService interface {
	Create(ctx context.Context, req *CreatePostRequest) error
}

type postService struct {
	db *bun.DB
}

func NewPostService(db *bun.DB) PostService {
	return &postService{db: db}
}

func (s *postService) Create(ctx context.Context, req *CreatePostRequest) error {
	// For this task, we just mock the creation
	return nil
}
