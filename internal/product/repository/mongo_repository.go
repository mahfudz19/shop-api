package repository

import (
	"context"
	// PENTING: Ganti "github.com/username/shop-api" dengan nama module yang ada di file go.mod kamu
	"github.com/username/shop-api/internal/domain"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// mongoProductRepository adalah implementasi konkrit dari antarmuka domain.ProductRepository
type mongoProductRepository struct {
	db         *mongo.Database
	collection string
}

// NewMongoProductRepository adalah constructor (pembuat) instance repository
func NewMongoProductRepository(db *mongo.Database) domain.ProductRepository {
	return &mongoProductRepository{
		db:         db,
		collection: "products", // Sesuaikan dengan nama collection di MongoDB Atlas kamu
	}
}

// FetchAll mengambil semua data produk dari MongoDB
func (m *mongoProductRepository) FetchAll(ctx context.Context) ([]domain.Product, error) {
	var products []domain.Product

	// Query tanpa filter (bson.M{}) artinya ambil semua data
	cursor, err := m.db.Collection(m.collection).Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	// Pastikan cursor ditutup setelah selesai
	defer cursor.Close(ctx)

	// Looping untuk memasukkan data dari MongoDB ke dalam slice struct Product
	for cursor.Next(ctx) {
		var p domain.Product
		if err := cursor.Decode(&p); err != nil {
			return nil, err
		}
		products = append(products, p)
	}

	return products, nil
}

// GetByID mengambil satu data produk berdasarkan ID
func (m *mongoProductRepository) GetByID(ctx context.Context, id string) (domain.Product, error) {
	var product domain.Product

	// Convert string ID dari URL parameter menjadi format ObjectID MongoDB
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return product, err // Gagal karena format ID tidak valid
	}

	// Query FindOne berdasarkan _id
	err = m.db.Collection(m.collection).FindOne(ctx, bson.M{"_id": objID}).Decode(&product)
	if err != nil {
		return product, err // Gagal karena data tidak ditemukan atau error DB
	}

	return product, nil
}