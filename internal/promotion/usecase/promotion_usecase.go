// Package usecase berisi implementasi logika bisnis untuk entitas Promotion
package usecase

import (
	"context"
	"time"

	"github.com/username/shop-api/internal/domain"
)

type promotionUseCase struct {
	repo domain.PromotionRepository
}

// NewPromotionUseCase membuat instance baru dari promotionUseCase
func NewPromotionUseCase(repo domain.PromotionRepository) domain.PromotionUseCase {
	return &promotionUseCase{repo: repo}
}

func (u *promotionUseCase) Create(ctx context.Context, req domain.PromotionRequest) (domain.Promotion, error) {
	now := time.Now().UTC()
	promo := domain.Promotion{
		Title:       req.Title,
		Description: req.Description,
		ImageURL:    req.ImageURL,
		LinkURL:     req.LinkURL,
		IsActive:    req.IsActive,
		OrderIndex:  req.OrderIndex,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	return u.repo.Create(ctx, promo)
}

func (u *promotionUseCase) GetAll(ctx context.Context, onlyActive bool) ([]domain.Promotion, error) {
	return u.repo.GetAll(ctx, onlyActive)
}

func (u *promotionUseCase) GetByID(ctx context.Context, id string) (domain.Promotion, error) {
	return u.repo.GetByID(ctx, id)
}

func (u *promotionUseCase) Update(ctx context.Context, id string, req domain.PromotionRequest) (domain.Promotion, error) {
	existing, err := u.repo.GetByID(ctx, id)
	if err != nil {
		return domain.Promotion{}, err
	}

	promo := domain.Promotion{
		Title:       req.Title,
		Description: req.Description,
		ImageURL:    req.ImageURL,
		LinkURL:     req.LinkURL,
		IsActive:    req.IsActive,
		OrderIndex:  req.OrderIndex,
		CreatedAt:   existing.CreatedAt, // Pertahankan tanggal dibuat
		UpdatedAt:   time.Now().UTC(),
	}
	return u.repo.Update(ctx, id, promo)
}

func (u *promotionUseCase) Delete(ctx context.Context, id string) error {
	return u.repo.Delete(ctx, id)
}
