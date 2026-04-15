// Package domain mendefinisikan struktur data dan interface untuk business logic serta repository
package domain

import (
	"context"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// Category merepresentasikan master data kategori di database
type Category struct {
	ID         bson.ObjectID `bson:"_id,omitempty" json:"id"`
	Name       string        `bson:"name" json:"name"`
	Slug       string        `bson:"slug" json:"slug"`
	IconURL    string        `bson:"icon_url" json:"icon_url"`
	IsPopular  bool          `bson:"is_popular" json:"is_popular"`
	OrderIndex int           `bson:"order_index" json:"order_index"`
}

// CategoryRequest adalah struct untuk menangkap input dari body request (JSON)
type CategoryRequest struct {
	Name       string `json:"name" binding:"required"`
	Slug       string `json:"slug" binding:"required"`
	IconURL    string `json:"icon_url"`
	IsPopular  bool   `json:"is_popular"`
	OrderIndex int    `json:"order_index"`
}

// CategoryUseCase mendefinisikan business logic
type CategoryUseCase interface {
	Create(ctx context.Context, req CategoryRequest) (Category, error)
	GetAll(ctx context.Context) ([]Category, error)
	GetByID(ctx context.Context, id string) (Category, error)
	Update(ctx context.Context, id string, req CategoryRequest) (Category, error)
	Delete(ctx context.Context, id string) error
	SyncCategories(ctx context.Context) (int64, error)
}

// CategoryRepository mendefinisikan operasi database
type CategoryRepository interface {
	Create(ctx context.Context, cat Category) (Category, error)
	GetAll(ctx context.Context) ([]Category, error)
	GetByID(ctx context.Context, id string) (Category, error)
	Update(ctx context.Context, id string, cat Category) (Category, error)
	Delete(ctx context.Context, id string) error
	SyncCategories(ctx context.Context) (int64, error)
}
