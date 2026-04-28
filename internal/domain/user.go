package domain

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// UserRole type untuk role user
type UserRole string

const (
	// RoleAdmin memiliki akses penuh
	RoleAdmin UserRole = "admin"
	// RoleUser memiliki akses terbatas
	RoleUser UserRole = "user"
)

// UserStatus type untuk status user
type UserStatus string

const (
	// StatusActive user aktif
	StatusActive UserStatus = "active"
	// StatusInactive user tidak aktif
	StatusInactive UserStatus = "inactive"
)

// IsValid method untuk memvalidasi nilai UserRole
func (r UserRole) IsValid() bool {
	switch r {
	case RoleAdmin, RoleUser:
		return true
	}
	return false
}

// User merepresentasikan user di database
type User struct {
	ID        bson.ObjectID `bson:"_id,omitempty" json:"id"`
	Email     string        `bson:"email" json:"email"`
	Password  string        `bson:"password" json:"-"`
	Name      string        `bson:"name" json:"name"`
	Role      UserRole      `bson:"role" json:"role"`
	Status    UserStatus    `bson:"status" json:"status"`
	CreatedAt time.Time     `bson:"createdAt" json:"createdAt"`
	UpdatedAt time.Time     `bson:"updatedAt" json:"updatedAt"`
}

// UserFilter struct untuk filter dan pagination query user
type UserFilter struct {
	BaseQuery
	Role UserRole
}

// UpdateUserRequest struct untuk request update user
type UpdateUserRequest struct {
	Email  string `json:"email" binding:"required,email"`
	Name   string `json:"name"`
	Role   string `json:"role"`
	Status string `json:"status"`
}

// UserWithPagination struct untuk response user dengan pagination
type UserWithPagination struct {
	Data       []User
	Page       int64
	Limit      int64
	Role       UserRole
	Total      int64
	TotalPages int64
}

// UserRepository interface untuk database operations
type UserRepository interface {
	Create(ctx context.Context, user User) error
	GetByEmail(ctx context.Context, email string) (User, error)
	GetByID(ctx context.Context, id string) (User, error)
	GetAll(ctx context.Context, filter UserFilter) (UserWithPagination, error)
	Update(ctx context.Context, user User) error
	Delete(ctx context.Context, id string) error
	EmailExists(ctx context.Context, email string) (bool, error)
}

// UserUseCase interface untuk business logic
type UserUseCase interface {
	Register(ctx context.Context, user User) error
	Login(ctx context.Context, email, password string) (User, error)
	GetUserByID(ctx context.Context, id string) (User, error)
	GetAllUsers(ctx context.Context, filter UserFilter) (UserWithPagination, error)
	UpdateUser(ctx context.Context, id string, req UpdateUserRequest) (User, error)
	DeleteUser(ctx context.Context, id string) error
}
