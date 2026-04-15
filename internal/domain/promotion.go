// Package domain berisi definisi struct dan interface untuk entitas Promotion
package domain

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// Promotion merepresentasikan data banner untuk Hero Section
type Promotion struct {
	ID          bson.ObjectID `bson:"_id,omitempty" json:"id"`
	Title       string        `bson:"title" json:"title"`
	Description string        `bson:"description" json:"description"`
	ImageURL    string        `bson:"image_url" json:"image_url"` // Path di S3
	LinkURL     string        `bson:"link_url" json:"link_url"`   // Link tujuan saat diklik
	IsActive    bool          `bson:"is_active" json:"is_active"` // Status aktif
	OrderIndex  int           `bson:"order_index" json:"order_index"`
	CreatedAt   time.Time     `bson:"createdAt" json:"createdAt"`
	UpdatedAt   time.Time     `bson:"updatedAt" json:"updatedAt"`
}

// PromotionRequest untuk validasi input JSON
type PromotionRequest struct {
	Title       string `json:"title" binding:"required"`
	Description string `json:"description"`
	ImageURL    string `json:"image_url" binding:"required"`
	LinkURL     string `json:"link_url"`
	IsActive    bool   `json:"is_active"`
	OrderIndex  int    `json:"order_index"`
}

// PromotionUseCase mendefinisikan operasi bisnis untuk entitas Promotion
type PromotionUseCase interface {
	Create(ctx context.Context, req PromotionRequest) (Promotion, error)
	GetAll(ctx context.Context, onlyActive bool) ([]Promotion, error)
	GetByID(ctx context.Context, id string) (Promotion, error)
	Update(ctx context.Context, id string, req PromotionRequest) (Promotion, error)
	Delete(ctx context.Context, id string) error
}

// PromotionRepository mendefinisikan operasi database untuk entitas Promotion
type PromotionRepository interface {
	Create(ctx context.Context, promo Promotion) (Promotion, error)
	GetAll(ctx context.Context, onlyActive bool) ([]Promotion, error)
	GetByID(ctx context.Context, id string) (Promotion, error)
	Update(ctx context.Context, id string, promo Promotion) (Promotion, error)
	Delete(ctx context.Context, id string) error
}
