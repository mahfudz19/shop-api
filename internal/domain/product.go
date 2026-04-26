// Package domain mendefinisikan struktur data dan interface untuk produk.
package domain

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// Product = Struct data (tetap sama)
type Product struct {
	ID                   bson.ObjectID `bson:"_id,omitempty" json:"id"`
	MasterProductID      bson.ObjectID `bson:"master_product_id,omitempty" json:"master_product_id,omitempty"`
	URL                  string        `bson:"url" json:"url"`
	Category             []string      `bson:"category" json:"category"`
	CleanURL             string        `bson:"clean_url" json:"clean_url"`
	CreatedAt            time.Time     `bson:"createdAt" json:"createdAt"`
	DiscountPercent      int           `bson:"discount_percent" json:"discount_percent"`
	ImageURL             string        `bson:"image_url" json:"image_url"`
	IsAnomaly            bool          `bson:"is_anomaly" json:"is_anomaly"`
	Location             string        `bson:"location" json:"location"`
	Marketplace          string        `bson:"marketplace" json:"marketplace"`
	MarketplaceProductID string        `bson:"marketplace_product_id" json:"marketplace_product_id"`
	MatchConfidence      float64       `bson:"match_confidence" json:"match_confidence"`
	Name                 string        `bson:"name" json:"name"`
	PriceOriginal        int64         `bson:"price_original" json:"price_original"`
	PriceRp              int64         `bson:"price_rp" json:"price_rp"`
	Rating               float64       `bson:"rating" json:"rating"`
	SearchKeyword        string        `bson:"search_keyword" json:"search_keyword"`
	Shop                 string        `bson:"shop" json:"shop"`
	SoldCount            int           `bson:"sold_count" json:"sold_count"`
	UpdatedAt            time.Time     `bson:"updatedAt" json:"updatedAt"`
}

// ProductStats = Struct baru untuk response Trust Section
type ProductStats struct {
	TotalProducts int64 `json:"total_products"`
	TotalShops    int   `json:"total_shops"`
}

// ProductFilter Struct untuk filter dan pagination
type ProductFilter struct {
	Search      string
	Location    string
	Marketplace string
	MinPrice    int64
	MaxPrice    int64
	SortBy      string
	SortOrder   string
	Rating      float64
	Page        int64
	Limit       int64
}

// ProductResponse Struct untuk response dengan metadata pagination
type ProductResponse struct {
	Data       []Product `json:"data"`
	Total      int64     `json:"total"`
	Page       int64     `json:"page"`
	Limit      int64     `json:"limit"`
	TotalPages int64     `json:"total_pages"`
}

// MasterInfo = Data ringkas MasterProduct
type MasterInfo struct {
	ID               string                 `json:"id"`
	Name             string                 `json:"name"`
	Slug             string                 `json:"slug"`
	Brand            string                 `json:"brand"`
	Model            string                 `json:"model"`
	Specifications   map[string]interface{} `json:"specifications,omitempty"`
	BaselinePriceMin int64                  `json:"baseline_price_min"`
	BaselinePriceMax int64                  `json:"baseline_price_max"`
	DefaultImage     string                 `json:"default_image"`
}

// ProductUseCase dan ProductRepository tetap sama, hanya ditambahkan method baru untuk GetByID
type ProductUseCase interface {
	GetProductsWithFilter(ctx context.Context, filter ProductFilter) (ProductResponse, error)
	GetProductByID(ctx context.Context, id string) (Product, error)
	GetDeals(ctx context.Context, limit int64) ([]Product, error)
	GetStats(ctx context.Context) (ProductStats, error)
}

// ProductRepository interface untuk database operations
type ProductRepository interface {
	FetchWithFilter(ctx context.Context, filter ProductFilter) ([]Product, int64, error)
	GetByID(ctx context.Context, id string) (Product, error)
	GetDeals(ctx context.Context, limit int64) ([]Product, error)
	GetStats(ctx context.Context) (ProductStats, error)
}
