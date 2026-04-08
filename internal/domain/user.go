package domain

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// User merepresentasikan user di database
type User struct {
	ID        bson.ObjectID `bson:"_id,omitempty" json:"id"`
	Email     string        `bson:"email" json:"email"`
	Password  string        `bson:"password" json:"-"` // json:"-" = tidak dimasukkan ke JSON response
	Name      string        `bson:"name" json:"name"`
	Role      string        `bson:"role" json:"role"` // admin, user
	CreatedAt time.Time     `bson:"createdAt" json:"created_at"`
	UpdatedAt time.Time     `bson:"updatedAt" json:"updated_at"`
}

// UserRepository interface untuk database operations
type UserRepository interface {
	Create(ctx context.Context, user User) error
	GetByEmail(ctx context.Context, email string) (User, error)
	GetByID(ctx context.Context, id string) (User, error)
	EmailExists(ctx context.Context, email string) (bool, error)
}

// UserUseCase interface untuk business logic
type UserUseCase interface {
	Register(ctx context.Context, user User) error
	Login(ctx context.Context, email, password string) (User, error)
	GetUserByID(ctx context.Context, id string) (User, error)
}
