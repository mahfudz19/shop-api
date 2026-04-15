// Package usecase mengimplementasikan logika bisnis untuk kategori
package usecase

import (
	"context"

	"github.com/username/shop-api/internal/domain"
)

type categoryUseCase struct {
	repo domain.CategoryRepository
}

// NewCategoryUseCase membuat instance baru dari categoryUseCase
func NewCategoryUseCase(repo domain.CategoryRepository) domain.CategoryUseCase {
	return &categoryUseCase{repo: repo}
}

func (u *categoryUseCase) Create(ctx context.Context, req domain.CategoryRequest) (domain.Category, error) {
	cat := domain.Category{
		Name:       req.Name,
		Slug:       req.Slug,
		IconURL:    req.IconURL,
		IsPopular:  req.IsPopular,
		OrderIndex: req.OrderIndex,
	}

	return u.repo.Create(ctx, cat)
}

func (u *categoryUseCase) GetAll(ctx context.Context) ([]domain.Category, error) {
	return u.repo.GetAll(ctx)
}

func (u *categoryUseCase) GetByID(ctx context.Context, id string) (domain.Category, error) {
	return u.repo.GetByID(ctx, id)
}

func (u *categoryUseCase) Update(ctx context.Context, id string, req domain.CategoryRequest) (domain.Category, error) {
	cat := domain.Category{
		Name:       req.Name,
		Slug:       req.Slug,
		IconURL:    req.IconURL,
		IsPopular:  req.IsPopular,
		OrderIndex: req.OrderIndex,
	}

	return u.repo.Update(ctx, id, cat)
}

func (u *categoryUseCase) Delete(ctx context.Context, id string) error {
	return u.repo.Delete(ctx, id)
}

// BARU: Implementasi SyncCategories
func (u *categoryUseCase) SyncCategories(ctx context.Context) (int64, error) {
	return u.repo.SyncCategories(ctx)
}
