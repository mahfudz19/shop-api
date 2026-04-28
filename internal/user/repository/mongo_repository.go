// Package repository = Implementasi Repository untuk User dengan MongoDB
package repository

import (
	"context"

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
	opts := options.Count().SetCollation(&options.Collation{Locale: "en", Strength: 2})

	count, err := m.db.Collection(m.collection).CountDocuments(ctx, bson.M{"email": email}, opts)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// GetAll ambil semua user dengan filter dan pagination
func (m *mongoUserRepository) GetAll(ctx context.Context, filter domain.UserFilter) (domain.UserWithPagination, error) {
	result := domain.UserWithPagination{
		Data:  []domain.User{},
		Page:  filter.Page,
		Limit: filter.Limit,
		Role:  filter.Role,
	}

	// Build filter query
	query := bson.M{}
	if filter.Search != "" {
		query["$or"] = []bson.M{
			{"name": bson.M{"$regex": filter.Search, "$options": "i"}},
			{"email": bson.M{"$regex": filter.Search, "$options": "i"}},
		}
	}

	if filter.Role != "" {
		query["role"] = filter.Role
	}

	// Count total documents
	total, err := m.db.Collection(m.collection).CountDocuments(ctx, query)
	if err != nil {
		return result, err
	}
	result.Total = total

	// Calculate total pages
	if filter.Limit > 0 {
		result.TotalPages = (total + filter.Limit - 1) / filter.Limit
	}

	// Set default page and limit
	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.Limit <= 0 {
		filter.Limit = 10
	}

	// Set options for pagination and sorting
	opts := options.Find().
		SetSkip((filter.Page - 1) * filter.Limit).
		SetLimit(filter.Limit)

	// Apply sorting
	if filter.SortBy != "" {
		sortOrder := 1 // ascending
		if filter.SortOrder == "desc" {
			sortOrder = -1
		}
		opts.SetSort(bson.M{filter.SortBy: sortOrder})
	}

	cursor, err := m.db.Collection(m.collection).Find(ctx, query, opts)
	if err != nil {
		return result, err
	}
	defer cursor.Close(ctx)

	if err := cursor.All(ctx, &result.Data); err != nil {
		return result, err
	}

	return result, nil
}

// Update update user
func (m *mongoUserRepository) Update(ctx context.Context, user domain.User) error {
	objID, err := bson.ObjectIDFromHex(user.ID.Hex())
	if err != nil {
		return err
	}

	_, err = m.db.Collection(m.collection).UpdateOne(
		ctx,
		bson.M{"_id": objID},
		bson.M{
			"$set": bson.M{
				"email":     user.Email,
				"name":      user.Name,
				"role":      user.Role,
				"status":    user.Status,
				"updatedAt": user.UpdatedAt,
			},
		},
	)
	return err
}

// Delete hapus user
func (m *mongoUserRepository) Delete(ctx context.Context, id string) error {
	objID, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	_, err = m.db.Collection(m.collection).DeleteOne(ctx, bson.M{"_id": objID})
	return err
}
