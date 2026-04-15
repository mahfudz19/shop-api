// Package repository menyediakan implementasi penyimpanan data untuk artikel menggunakan MongoDB
package repository

import (
	"context"
	"log"

	"github.com/username/shop-api/internal/domain"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type mongoArticleRepository struct {
	db         *mongo.Database
	collection string
}

// NewMongoArticleRepository membuat instance baru dari mongoArticleRepository
func NewMongoArticleRepository(db *mongo.Database) domain.ArticleRepository {
	return &mongoArticleRepository{
		db:         db,
		collection: "articles",
	}
}

func (m *mongoArticleRepository) Create(ctx context.Context, art domain.Article) (domain.Article, error) {
	art.ID = bson.NewObjectID()
	_, err := m.db.Collection(m.collection).InsertOne(ctx, art)
	return art, err
}

func (m *mongoArticleRepository) GetAll(ctx context.Context, onlyPublished bool) ([]domain.Article, error) {
	var articles []domain.Article
	filter := bson.M{}
	if onlyPublished {
		filter["is_published"] = true
	}

	opts := options.Find().SetSort(bson.D{{Key: "createdAt", Value: -1}})
	cursor, err := m.db.Collection(m.collection).Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := cursor.Close(ctx); err != nil {
			log.Printf("⚠️ Warning: Gagal menutup cursor: %v", err)
		}
	}()

	for cursor.Next(ctx) {
		var a domain.Article
		if err := cursor.Decode(&a); err == nil {
			articles = append(articles, a)
		}
	}
	return articles, nil
}

func (m *mongoArticleRepository) GetBySlug(ctx context.Context, slug string) (domain.Article, error) {
	var art domain.Article
	err := m.db.Collection(m.collection).FindOne(ctx, bson.M{"slug": slug}).Decode(&art)
	return art, err
}

func (m *mongoArticleRepository) Update(ctx context.Context, id string, art domain.Article) (domain.Article, error) {
	objID, _ := bson.ObjectIDFromHex(id)
	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)
	var updated domain.Article
	err := m.db.Collection(m.collection).FindOneAndUpdate(ctx, bson.M{"_id": objID}, bson.M{"$set": art}, opts).Decode(&updated)
	return updated, err
}

func (m *mongoArticleRepository) Delete(ctx context.Context, id string) error {
	objID, _ := bson.ObjectIDFromHex(id)
	_, err := m.db.Collection(m.collection).DeleteOne(ctx, bson.M{"_id": objID})
	return err
}
