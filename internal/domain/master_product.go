// Package domain mendefinisikan struktur data dan interface untuk MasterProduct
package domain

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// MasterProduct merepresentasikan katalog utama yang bersih
type MasterProduct struct {
	ID               bson.ObjectID          `bson:"_id,omitempty" json:"id"`
	Name             string                 `bson:"name" json:"name"`
	Slug             string                 `bson:"slug" json:"slug"`
	CategoryID       bson.ObjectID          `bson:"category_id" json:"category_id"`
	Brand            string                 `bson:"brand" json:"brand"`
	Model            string                 `bson:"model" json:"model"`
	Specifications   map[string]interface{} `bson:"specifications" json:"specifications"`
	BaselinePriceMin int64                  `bson:"baseline_price_min" json:"baseline_price_min"`
	BaselinePriceMax int64                  `bson:"baseline_price_max" json:"baseline_price_max"`
	DefaultImage     string                 `bson:"default_image" json:"default_image"`
	CreatedAt        time.Time              `bson:"createdAt" json:"createdAt"`
	UpdatedAt        time.Time              `bson:"updatedAt" json:"updatedAt"`
}

// MasterProductDetail adalah response khusus untuk halaman Detail
type MasterProductDetail struct {
	MasterProduct
	Offers      []Product `json:"offers"`
	MinPrice    int64     `json:"min_price"`
	MaxPrice    int64     `json:"max_price"`
	Savings     int64     `json:"savings"`
	TotalOffers int       `json:"total_offers"`
}

// MasterProductUseCase mendefinisikan business logic untuk MasterProduct
type MasterProductUseCase interface {
	GetDetailByID(ctx context.Context, id string) (MasterProductDetail, error)
}

// MasterProductRepository mendefinisikan operasi database untuk MasterProduct
type MasterProductRepository interface {
	GetByID(ctx context.Context, id string) (MasterProduct, error)
	GetOffersByMasterID(ctx context.Context, masterID bson.ObjectID) ([]Product, error)
}
