package service

import (
	"context"
	"errors"
	"strings"

	"inventory/internal/model"
	"inventory/internal/repository"
)

type CategoryService struct {
	repo *repository.CategoryRepository
}

func NewCategoryService(repo *repository.CategoryRepository) *CategoryService {
	return &CategoryService{repo: repo}
}

func (s *CategoryService) List(ctx context.Context) ([]model.Category, error) {
	return s.repo.List(ctx)
}

func (s *CategoryService) Create(ctx context.Context, name string) (*model.Category, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, &ValidationError{Fields: map[string]string{"name": "Name is required"}}
	}
	c, err := s.repo.Create(ctx, name)
	if errors.Is(err, repository.ErrDuplicateCategory) {
		return nil, &ValidationError{Fields: map[string]string{"name": "Category already exists"}, Conflict: true}
	}
	return c, err
}
