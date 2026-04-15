// Package repository = Implementasi Repository untuk Category menggunakan MongoDB
package repository

import (
	"context"
	"errors"
	"log"
	"strings"

	"github.com/username/shop-api/internal/domain"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type mongoCategoryRepository struct {
	db         *mongo.Database
	collection string
}

// NewMongoCategoryRepository membuat instance baru dari mongoCategoryRepository
func NewMongoCategoryRepository(db *mongo.Database) domain.CategoryRepository {
	return &mongoCategoryRepository{
		db:         db,
		collection: "categories",
	}
}

func (m *mongoCategoryRepository) Create(ctx context.Context, cat domain.Category) (domain.Category, error) {
	collection := m.db.Collection(m.collection)

	cat.ID = bson.NewObjectID()
	_, err := collection.InsertOne(ctx, cat)
	if err != nil {
		return domain.Category{}, err
	}

	return cat, nil
}

func (m *mongoCategoryRepository) GetAll(ctx context.Context) ([]domain.Category, error) {
	collection := m.db.Collection(m.collection)
	var categories []domain.Category

	// Urutkan berdasarkan order_index
	opts := options.Find().SetSort(bson.D{{Key: "order_index", Value: 1}})
	cursor, err := collection.Find(ctx, bson.M{}, opts)
	if err != nil {
		return nil, err
	}

	defer func() {
		if err := cursor.Close(ctx); err != nil {
			log.Printf("⚠️ Warning: Gagal menutup cursor: %v", err)
		}
	}()

	for cursor.Next(ctx) {
		var cat domain.Category
		if err := cursor.Decode(&cat); err != nil {
			return nil, err
		}
		categories = append(categories, cat)
	}

	return categories, nil
}

func (m *mongoCategoryRepository) GetByID(ctx context.Context, id string) (domain.Category, error) {
	collection := m.db.Collection(m.collection)

	objID, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return domain.Category{}, errors.New("invalid category ID format")
	}

	var cat domain.Category
	err = collection.FindOne(ctx, bson.M{"_id": objID}).Decode(&cat)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return domain.Category{}, errors.New("category not found")
		}
		return domain.Category{}, err
	}

	return cat, nil
}

func (m *mongoCategoryRepository) Update(ctx context.Context, id string, cat domain.Category) (domain.Category, error) {
	collection := m.db.Collection(m.collection)

	objID, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return domain.Category{}, errors.New("invalid category ID format")
	}

	filter := bson.M{"_id": objID}
	update := bson.M{"$set": cat}

	// Option untuk mengembalikan dokumen setelah diupdate
	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)

	var updatedCat domain.Category
	err = collection.FindOneAndUpdate(ctx, filter, update, opts).Decode(&updatedCat)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return domain.Category{}, errors.New("category not found")
		}
		return domain.Category{}, err
	}

	return updatedCat, nil
}

func (m *mongoCategoryRepository) Delete(ctx context.Context, id string) error {
	collection := m.db.Collection(m.collection)

	objID, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return errors.New("invalid category ID format")
	}

	result, err := collection.DeleteOne(ctx, bson.M{"_id": objID})
	if err != nil {
		return err
	}
	if result.DeletedCount == 0 {
		return errors.New("category not found")
	}

	return nil
}

// SyncCategories mensinkronisasi kategori unik dari products ke categories
func (m *mongoCategoryRepository) SyncCategories(ctx context.Context) (int64, error) {
	productsColl := m.db.Collection("products")
	categoriesColl := m.db.Collection(m.collection)

	// 1. Ambil semua unique category dari koleksi "products"
	var distinctCategories []string

	// Di MongoDB v2, Distinct() dirangkai langsung dengan Decode()
	// sama seperti cara kerja FindOne()
	err := productsColl.Distinct(ctx, "category", bson.M{}).Decode(&distinctCategories)
	if err != nil {
		return 0, err
	}

	// 2. Ambil semua kategori yang saat ini sudah ada di Master Data ("categories")
	cursor, err := categoriesColl.Find(ctx, bson.M{})
	if err != nil {
		return 0, err
	}
	defer func() {
		if err := cursor.Close(ctx); err != nil {
			log.Printf("⚠️ Warning: Gagal menutup cursor di SyncCategories: %v", err)
		}
	}()

	existingCategories := make(map[string]bool)
	for cursor.Next(ctx) {
		var cat domain.Category
		if err := cursor.Decode(&cat); err == nil {
			existingCategories[cat.Name] = true
		}
	}

	// 3. Bandingkan dan siapkan kategori yang benar-benar baru
	var newCategories []interface{}
	for _, catName := range distinctCategories {
		// Validasi string kosong
		if catName == "" {
			continue
		}

		if !existingCategories[catName] {
			slug := strings.ToLower(strings.ReplaceAll(catName, " ", "-"))

			newCat := domain.Category{
				ID:         bson.NewObjectID(),
				Name:       catName,
				Slug:       slug,
				IconURL:    "",
				IsPopular:  false,
				OrderIndex: 0,
			}
			newCategories = append(newCategories, newCat)
		}
	}

	// 4. Eksekusi InsertMany jika ada kategori baru
	if len(newCategories) == 0 {
		return 0, nil
	}

	res, err := categoriesColl.InsertMany(ctx, newCategories)
	if err != nil {
		return 0, err
	}

	return int64(len(res.InsertedIDs)), nil
}
