// internal/master_product/repository/mongo_repository.go
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

type mongoMasterProductRepository struct {
	db *mongo.Database
}

func NewMongoMasterProductRepository(db *mongo.Database) domain.MasterProductRepository {
	return &mongoMasterProductRepository{db: db}
}

func (m *mongoMasterProductRepository) GetByID(ctx context.Context, id string) (domain.MasterProduct, error) {
	objID, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return domain.MasterProduct{}, errors.New("invalid master product ID format")
	}

	var master domain.MasterProduct
	err = m.db.Collection("master_products").FindOne(ctx, bson.M{"_id": objID}).Decode(&master)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return domain.MasterProduct{}, errors.New("master product not found")
		}
		return domain.MasterProduct{}, err
	}
	return master, nil
}

func (m *mongoMasterProductRepository) GetOffersByMasterID(ctx context.Context, masterID bson.ObjectID) ([]domain.Product, error) {
	var offers []domain.Product
	
	// Cari product mentah yang terikat dengan master_product_id ini
	// Filter anomali dan pastikan confidence tinggi
	filter := bson.M{
		"master_product_id": masterID,
		"is_anomaly":        false,
		"match_confidence":  bson.M{"$gte": 0.8}, // Minimal keyakinan sistem 80%
	}

	// Langsung urutkan dari harga termurah
	opts := options.Find().SetSort(bson.D{{Key: "price_rp", Value: 1}})
	
	cursor, err := m.db.Collection("products").Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := cursor.Close(ctx); err != nil {
			log.Printf("⚠️ Warning: Gagal menutup cursor: %v", err)
		}
	}()

	for cursor.Next(ctx) {
		var p domain.Product
		if err := cursor.Decode(&p); err == nil {
			offers = append(offers, p)
		}
	}

	return offers, nil
}