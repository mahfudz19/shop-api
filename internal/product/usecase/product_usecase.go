// Package usecase = Business logic untuk Product
package usecase

import (
	"context"
	"errors"
	"math"

	"github.com/username/shop-api/internal/domain"
)

type productUseCase struct {
	repo domain.ProductRepository
}

// NewProductUseCase = Inisialisasi Product UseCase
func NewProductUseCase(repo domain.ProductRepository) domain.ProductUseCase {
	return &productUseCase{repo: repo}
}

// BARU: GetProductsWithFilter
func (u *productUseCase) GetProductsWithFilter(
	ctx context.Context,
	filter domain.ProductFilter,
) (domain.ProductResponse, error) {

	// Validasi input
	if filter.Limit <= 0 {
		filter.Limit = 10
	}
	if filter.Limit > 100 {
		filter.Limit = 100 // max 100 item per halaman
	}
	if filter.Page <= 0 {
		filter.Page = 1
	}

	// Panggil repository
	products, total, err := u.repo.FetchWithFilter(ctx, filter)
	if err != nil {
		return domain.ProductResponse{}, err
	}

	// Hitung total pages
	totalPages := int64(math.Ceil(float64(total) / float64(filter.Limit)))

	// Return response dengan metadata
	return domain.ProductResponse{
		Data:       products,
		Total:      total,
		Page:       filter.Page,
		Limit:      filter.Limit,
		TotalPages: totalPages,
	}, nil
}

// GetProductByID = Business logic untuk get single product
func (u *productUseCase) GetProductByID(ctx context.Context, id string) (domain.Product, error) {
	if id == "" {
		return domain.Product{}, errors.New("product ID is required")
	}
	return u.repo.GetByID(ctx, id)
}
