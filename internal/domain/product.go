package domain

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Product merepresentasikan skema data di MongoDB
type Product struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	URL         string             `bson:"url" json:"url"`
	CreatedAt   time.Time          `bson:"createdAt" json:"created_at"`
	Location    string             `bson:"location" json:"location"`
	Marketplace string             `bson:"marketplace" json:"marketplace"`
	Name        string             `bson:"name" json:"name"`
	PriceRp     int64              `bson:"price_rp" json:"price_rp"`
	Shop        string             `bson:"shop" json:"shop"`
	UpdatedAt   time.Time          `bson:"updatedAt" json:"updated_at"`
}

// ProductRepository adalah antarmuka untuk operasi Database
type ProductRepository interface {
	FetchAll(ctx context.Context) ([]Product, error)
	GetByID(ctx context.Context, id string) (Product, error)
}

// ProductUseCase adalah antarmuka untuk Logika Bisnis
type ProductUseCase interface {
	GetAllProducts(ctx context.Context) ([]Product, error)
	GetProductDetail(ctx context.Context, id string) (Product, error)
}