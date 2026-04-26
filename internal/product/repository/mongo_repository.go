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

	// Filter by rating (products with rating >= requested rating)
	if filter.Rating > 0 {
		bsonFilter["rating"] = bson.M{"$gte": filter.Rating}
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

// GetDeals: Mengambil produk dengan diskon tertinggi
func (m *mongoProductRepository) GetDeals(ctx context.Context, limit int64) ([]domain.Product, error) {
	var products []domain.Product
	collection := m.db.Collection(m.collection)

	// Filter: Hanya produk yang punya diskon > 0
	filter := bson.M{"discount_percent": bson.M{"$gt": 0}}

	// Sort berdasarkan discount_percent dari yang paling besar (descending: -1)
	opts := options.Find().
		SetSort(bson.D{{Key: "discount_percent", Value: -1}}).
		SetLimit(limit)

	cursor, err := collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := cursor.Close(ctx); err != nil {
			log.Printf("⚠️ Warning: Gagal menutup cursor di GetDeals: %v", err)
		}
	}()

	for cursor.Next(ctx) {
		var p domain.Product
		if err := cursor.Decode(&p); err != nil {
			return nil, err
		}
		products = append(products, p)
	}

	return products, nil
}

// GetStats: Mengambil agregasi statistik total produk & toko unik
func (m *mongoProductRepository) GetStats(ctx context.Context) (domain.ProductStats, error) {
	collection := m.db.Collection(m.collection)

	// 1. Hitung total semua dokumen produk
	totalProducts, err := collection.CountDocuments(ctx, bson.M{})
	if err != nil {
		return domain.ProductStats{}, err
	}

	// 2. Hitung jumlah toko unik menggunakan Aggregation (Lebih aman dari Distinct untuk Data Besar & Kompatibel v2)
	pipeline := mongo.Pipeline{
		bson.D{{Key: "$group", Value: bson.D{{Key: "_id", Value: "$shop"}}}},
		bson.D{{Key: "$count", Value: "total_shops"}},
	}

	cursor, err := collection.Aggregate(ctx, pipeline)
	if err != nil {
		return domain.ProductStats{}, err
	}
	defer func() {
		if err := cursor.Close(ctx); err != nil {
			log.Printf("⚠️ Warning: Gagal menutup cursor di GetDeals: %v", err)
		}
	}()

	var totalShops int
	if cursor.Next(ctx) {
		// Struct sementara untuk menangkap hasil $count
		var result struct {
			TotalShops int `bson:"total_shops"`
		}
		if err := cursor.Decode(&result); err == nil {
			totalShops = result.TotalShops
		}
	}

	return domain.ProductStats{
		TotalProducts: totalProducts,
		TotalShops:    totalShops,
	}, nil
}
