// Package usecase mengimplementasikan logika bisnis untuk artikel
package usecase

import (
	"context"
	"time"

	"github.com/username/shop-api/internal/domain"
)

type articleUseCase struct {
	repo domain.ArticleRepository
}

// NewArticleUseCase membuat instance baru dari articleUseCase
func NewArticleUseCase(repo domain.ArticleRepository) domain.ArticleUseCase {
	return &articleUseCase{repo: repo}
}

func (u *articleUseCase) Create(ctx context.Context, req domain.ArticleRequest) (domain.Article, error) {
	now := time.Now().UTC()
	art := domain.Article{
		Title:       req.Title,
		Slug:        req.Slug,
		Content:     req.Content,
		Author:      req.Author,
		Thumbnail:   req.Thumbnail,
		Tags:        req.Tags,
		IsPublished: req.IsPublished,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if req.IsPublished {
		art.PublishedAt = now
	}
	return u.repo.Create(ctx, art)
}

func (u *articleUseCase) GetAll(ctx context.Context, onlyPublished bool) ([]domain.Article, error) {
	return u.repo.GetAll(ctx, onlyPublished)
}

func (u *articleUseCase) GetBySlug(ctx context.Context, slug string) (domain.Article, error) {
	return u.repo.GetBySlug(ctx, slug)
}

func (u *articleUseCase) Update(ctx context.Context, id string, req domain.ArticleRequest) (domain.Article, error) {
	// Logika sederhana: ambil data lama untuk pertahankan CreatedAt
	// Idealnya pakai ID untuk update
	now := time.Now().UTC()
	art := domain.Article{
		Title:       req.Title,
		Slug:        req.Slug,
		Content:     req.Content,
		Author:      req.Author,
		Thumbnail:   req.Thumbnail,
		Tags:        req.Tags,
		IsPublished: req.IsPublished,
		UpdatedAt:   now,
	}
	return u.repo.Update(ctx, id, art)
}

func (u *articleUseCase) Delete(ctx context.Context, id string) error {
	return u.repo.Delete(ctx, id)
}
