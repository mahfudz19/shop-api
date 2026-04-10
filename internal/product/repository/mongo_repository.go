// Package repository = Implementasi Repository untuk Product menggunakan MongoDB
package repository

import (
	"context"
	"log"

	"github.com/username/shop-api/internal/domain"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type mongoProductRepository struct {
	db         *mongo.Database
	collection string
}

// NewMongoProductRepository = Inisialisasi MongoDB Repository untuk Product
func NewMongoProductRepository(db *mongo.Database) domain.ProductRepository {
	return &mongoProductRepository{
		db:         db,
		collection: "products",
	}
}

// BARU: FetchWithFilter dengan pagination, search, sort
func (m *mongoProductRepository) FetchWithFilter(
	ctx context.Context,
	filter domain.ProductFilter,
) ([]domain.Product, int64, error) {

	var products []domain.Product
	collection := m.db.Collection(m.collection)

	// 1. BUILD FILTER (Search + Filter)
	bsonFilter := bson.M{}

	// Search by name (case-insensitive)
	if filter.Search != "" {
		bsonFilter["name"] = bson.M{
			"$regex":   filter.Search,
			"$options": "i", // i = case insensitive
		}
	}

	// Filter by location
	if filter.Location != "" {
		bsonFilter["location"] = filter.Location
	}

	// Filter by marketplace
	if filter.Marketplace != "" {
		bsonFilter["marketplace"] = filter.Marketplace
	}

	// Filter by price range
	if filter.MinPrice > 0 || filter.MaxPrice > 0 {
		priceFilter := bson.M{}
		if filter.MinPrice > 0 {
			priceFilter["$gte"] = filter.MinPrice
		}
		if filter.MaxPrice > 0 {
			priceFilter["$lte"] = filter.MaxPrice
		}
		bsonFilter["price_rp"] = priceFilter
	}

	// 2. HITUNG TOTAL (untuk metadata pagination)
	total, err := collection.CountDocuments(ctx, bsonFilter)
	if err != nil {
		return nil, 0, err
	}

	// 3. SETUP OPTIONS (Sort + Pagination)
	opts := options.Find()

	// Sorting
	sortField := filter.SortBy
	if sortField == "" {
		sortField = "createdAt" // default sort
	}

	sortOrder := 1 // 1 = ascending, -1 = descending
	if filter.SortOrder == "desc" {
		sortOrder = -1
	}
	opts.SetSort(bson.D{{Key: sortField, Value: sortOrder}})

	// Pagination
	if filter.Limit <= 0 {
		filter.Limit = 10 // default 10 item per halaman
	}
	if filter.Page <= 0 {
		filter.Page = 1 // default halaman 1
	}

	skip := (filter.Page - 1) * filter.Limit
	opts.SetLimit(filter.Limit)
	opts.SetSkip(skip)

	// 4. EXECUTE QUERY
	cursor, err := collection.Find(ctx, bsonFilter, opts)
	if err != nil {
		return nil, 0, err
	}
	defer func() {
		if err := cursor.Close(ctx); err != nil {
			log.Printf("⚠️ Warning: Gagal menutup cursor: %v", err)
		}
	}()

	// Decode hasil
	for cursor.Next(ctx) {
		var p domain.Product
		if err := cursor.Decode(&p); err != nil {
			return nil, 0, err
		}
		products = append(products, p)
	}

	return products, total, nil
}

func (m *mongoProductRepository) GetByID(ctx context.Context, id string) (domain.Product, error) {
	var product domain.Product

	// Convert string ID ke ObjectID
	objID, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return product, err // Invalid ID format
	}

	// Query
	err = m.db.Collection(m.collection).FindOne(ctx, bson.M{"_id": objID}).Decode(&product)
	if err != nil {
		return product, err // Not found atau error DB
	}

	return product, nil
}
