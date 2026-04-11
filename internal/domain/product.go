// Package domain mendefinisikan struktur data dan interface untuk produk.
package domain

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// Product = Struct data (tetap sama)
type Product struct {
	ID          bson.ObjectID `bson:"_id,omitempty" json:"id"`
	URL         string        `bson:"url" json:"url"`
	CreatedAt   time.Time     `bson:"createdAt" json:"createdAt"`
	Location    string        `bson:"location" json:"location"`
	Marketplace string        `bson:"marketplace" json:"marketplace"`
	Name        string        `bson:"name" json:"name"`
	PriceRp     int64         `bson:"price_rp" json:"price_rp"`
	Shop        string        `bson:"shop" json:"shop"`
	UpdatedAt   time.Time     `bson:"updatedAt" json:"updatedAt"`
}

// ProductFilter Struct untuk filter dan pagination
type ProductFilter struct {
	Search      string // Search by name
	Location    string // Filter by location
	Marketplace string // Filter by marketplace
	MinPrice    int64  // Filter min price
	MaxPrice    int64  // Filter max price
	SortBy      string // Sort field (name, price_rp, createdAt)
	SortOrder   string // asc atau desc
	Page        int64  // Halaman (mulai dari 1)
	Limit       int64  // Jumlah item per halaman
}

// ProductResponse Struct untuk response dengan metadata pagination
type ProductResponse struct {
	Data       []Product `json:"data"`
	Total      int64     `json:"total"`
	Page       int64     `json:"page"`
	Limit      int64     `json:"limit"`
	TotalPages int64     `json:"total_pages"`
}

// ProductUseCase dan ProductRepository tetap sama, hanya ditambahkan method baru untuk GetByID
type ProductUseCase interface {
	GetProductsWithFilter(ctx context.Context, filter ProductFilter) (ProductResponse, error)
	GetProductByID(ctx context.Context, id string) (Product, error) // BARU
}

// ProductRepository interface untuk database operations
type ProductRepository interface {
	FetchWithFilter(ctx context.Context, filter ProductFilter) ([]Product, int64, error)
	GetByID(ctx context.Context, id string) (Product, error) // BARU
}
