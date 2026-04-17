// Package domain mendefinisikan struktur data dan interface untuk artikel
package domain

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// Article merepresentasikan data konten blog/magazine
type Article struct {
	ID          bson.ObjectID `bson:"_id,omitempty" json:"id"`
	Title       string        `bson:"title" json:"title"`
	Slug        string        `bson:"slug" json:"slug"`
	Content     string        `bson:"content" json:"content"`
	Author      string        `bson:"author" json:"author"`
	Thumbnail   string        `bson:"thumbnail" json:"thumbnail"`
	Tags        []string      `bson:"tags" json:"tags"`
	IsPublished bool          `bson:"is_published" json:"is_published"`
	PublishedAt time.Time     `bson:"publishedAt" json:"publishedAt"`
	CreatedAt   time.Time     `bson:"createdAt" json:"createdAt"`
	UpdatedAt   time.Time     `bson:"updatedAt" json:"updatedAt"`
}

// ArticleRequest untuk validasi input
type ArticleRequest struct {
	Title       string   `json:"title" binding:"required"`
	Slug        string   `json:"slug" binding:"required"`
	Content     string   `json:"content" binding:"required"`
	Author      string   `json:"author" binding:"required"`
	Thumbnail   string   `json:"thumbnail"`
	Tags        []string `json:"tags"`
	IsPublished bool     `json:"is_published"`
}

// ArticleUseCase mendefinisikan operasi bisnis untuk artikel
type ArticleUseCase interface {
	Create(ctx context.Context, req ArticleRequest) (Article, error)
	GetAll(ctx context.Context, onlyPublished bool) ([]Article, error)
	GetBySlug(ctx context.Context, slug string) (Article, error) // Penting untuk SEO
	Update(ctx context.Context, id string, req ArticleRequest) (Article, error)
	Delete(ctx context.Context, id string) error
}

// ArticleRepository mendefinisikan operasi penyimpanan data untuk artikel
type ArticleRepository interface {
	Create(ctx context.Context, art Article) (Article, error)
	GetAll(ctx context.Context, onlyPublished bool) ([]Article, error)
	GetBySlug(ctx context.Context, slug string) (Article, error)
	Update(ctx context.Context, id string, art Article) (Article, error)
	Delete(ctx context.Context, id string) error
}
