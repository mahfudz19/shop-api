// internal/master_product/usecase/master_product_usecase.go
package usecase

import (
	"context"

	"github.com/username/shop-api/internal/domain"
)

type masterProductUseCase struct {
	repo domain.MasterProductRepository
}

func NewMasterProductUseCase(repo domain.MasterProductRepository) domain.MasterProductUseCase {
	return &masterProductUseCase{repo: repo}
}

func (u *masterProductUseCase) GetDetailByID(ctx context.Context, id string) (domain.MasterProductDetail, error) {
	// 1. Ambil Data Master
	master, err := u.repo.GetByID(ctx, id)
	if err != nil {
		return domain.MasterProductDetail{}, err
	}

	// 2. Ambil semua penawaran (Offers) yang terkait
	offers, err := u.repo.GetOffersByMasterID(ctx, master.ID)
	if err != nil {
		return domain.MasterProductDetail{}, err
	}

	// 3. Mesin Kalkulasi (Engine Logic)
	var minPrice int64 = 0
	var maxPrice int64 = 0
	var savings int64 = 0

	totalOffers := len(offers)
	if totalOffers > 0 {
		// Karena offers sudah di-sort dari DB berdasarkan harga (Ascending),
		// index 0 adalah yang termurah, index terakhir adalah termahal.
		minPrice = offers[0].PriceRp
		maxPrice = offers[totalOffers-1].PriceRp
		savings = maxPrice - minPrice
	}

	// 4. Susun Response Detail
	detail := domain.MasterProductDetail{
		MasterProduct: master,
		Offers:        offers,
		MinPrice:      minPrice,
		MaxPrice:      maxPrice,
		Savings:       savings,
		TotalOffers:   totalOffers,
	}

	return detail, nil
}