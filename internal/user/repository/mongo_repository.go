// Package repository = Implementasi Repository untuk User dengan MongoDB
package repository

import (
	"context"
	"strings"

	"github.com/username/shop-api/internal/domain"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// mongoUserRepository implementasi UserRepository
type mongoUserRepository struct {
	db         *mongo.Database
	collection string
}

// NewMongoUserRepository constructor
func NewMongoUserRepository(db *mongo.Database) domain.UserRepository {
	return &mongoUserRepository{
		db:         db,
		collection: "users",
	}
}

// Create insert user baru
func (m *mongoUserRepository) Create(ctx context.Context, user domain.User) error {
	_, err := m.db.Collection(m.collection).InsertOne(ctx, user)
	return err
}

// GetByEmail cari user berdasarkan email (untuk login)
func (m *mongoUserRepository) GetByEmail(ctx context.Context, email string) (domain.User, error) {
	var user domain.User

	// Case-insensitive search dengan collation
	opts := options.FindOne().SetCollation(&options.Collation{Locale: "en", Strength: 2})

	err := m.db.Collection(m.collection).FindOne(ctx, bson.M{"email": email}, opts).Decode(&user)
	return user, err
}

// GetByID cari user berdasarkan ID
func (m *mongoUserRepository) GetByID(ctx context.Context, id string) (domain.User, error) {
	var user domain.User

	objID, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return user, err
	}

	err = m.db.Collection(m.collection).FindOne(ctx, bson.M{"_id": objID}).Decode(&user)
	return user, err
}

// EmailExists cek apakah email sudah terdaftar
func (m *mongoUserRepository) EmailExists(ctx context.Context, email string) (bool, error) {
	// Normalize email (lowercase)
	email = strings.ToLower(email)

	count, err := m.db.Collection(m.collection).CountDocuments(ctx, bson.M{"email": email})
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
