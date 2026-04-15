// Package repository berisi implementasi repository untuk entitas Promotion menggunakan MongoDB
package repository

import (
	"context"
	"errors"
	"log"

	"github.com/username/shop-api/internal/domain"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type mongoPromotionRepository struct {
	db         *mongo.Database
	collection string
}

// NewMongoPromotionRepository membuat instance baru dari mongoPromotionRepository
func NewMongoPromotionRepository(db *mongo.Database) domain.PromotionRepository {
	return &mongoPromotionRepository{
		db:         db,
		collection: "promotions",
	}
}

func (m *mongoPromotionRepository) Create(ctx context.Context, promo domain.Promotion) (domain.Promotion, error) {
	collection := m.db.Collection(m.collection)
	promo.ID = bson.NewObjectID()

	_, err := collection.InsertOne(ctx, promo)
	return promo, err
}

func (m *mongoPromotionRepository) GetAll(ctx context.Context, onlyActive bool) ([]domain.Promotion, error) {
	collection := m.db.Collection(m.collection)
	var promos []domain.Promotion

	filter := bson.M{}
	if onlyActive {
		filter["is_active"] = true
	}

	opts := options.Find().SetSort(bson.D{{Key: "order_index", Value: 1}})
	cursor, err := collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := cursor.Close(ctx); err != nil {
			log.Printf("⚠️ Gagal menutup cursor: %v", err)
		}
	}()

	for cursor.Next(ctx) {
		var p domain.Promotion
		if err := cursor.Decode(&p); err == nil {
			promos = append(promos, p)
		}
	}
	return promos, nil
}

func (m *mongoPromotionRepository) GetByID(ctx context.Context, id string) (domain.Promotion, error) {
	objID, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return domain.Promotion{}, errors.New("invalid id")
	}

	var promo domain.Promotion
	err = m.db.Collection(m.collection).FindOne(ctx, bson.M{"_id": objID}).Decode(&promo)
	return promo, err
}

func (m *mongoPromotionRepository) Update(ctx context.Context, id string, promo domain.Promotion) (domain.Promotion, error) {
	objID, _ := bson.ObjectIDFromHex(id)
	filter := bson.M{"_id": objID}
	update := bson.M{"$set": promo}

	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)
	var updated domain.Promotion
	err := m.db.Collection(m.collection).FindOneAndUpdate(ctx, filter, update, opts).Decode(&updated)
	return updated, err
}

func (m *mongoPromotionRepository) Delete(ctx context.Context, id string) error {
	objID, _ := bson.ObjectIDFromHex(id)
	_, err := m.db.Collection(m.collection).DeleteOne(ctx, bson.M{"_id": objID})
	return err
}
