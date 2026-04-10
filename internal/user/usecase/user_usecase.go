// Package usecase = Business logic untuk User
package usecase

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/username/shop-api/internal/domain"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"golang.org/x/crypto/bcrypt"
)

// userUseCase implementasi UserUseCase
type userUseCase struct {
	repo domain.UserRepository
}

// NewUserUseCase constructor
func NewUserUseCase(repo domain.UserRepository) domain.UserUseCase {
	return &userUseCase{repo: repo}
}

// Register mendaftarkan user baru
func (u *userUseCase) Register(ctx context.Context, user domain.User) error {
	// 1. Validasi email
	user.Email = strings.ToLower(strings.TrimSpace(user.Email))
	if user.Email == "" {
		return errors.New("email is required")
	}

	// 2. Cek email sudah terdaftar?
	exists, err := u.repo.EmailExists(ctx, user.Email)
	if err != nil {
		return err
	}
	if exists {
		return errors.New("email already registered")
	}

	// 3. Validasi password
	if len(user.Password) < 6 {
		return errors.New("password must be at least 6 characters")
	}

	// 4. Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	user.Password = string(hashedPassword)

	// 5. Set default values
	if user.Role == "" {
		user.Role = "user" // default: bukan admin
	}
	if user.Name == "" {
		user.Name = "User" // default name
	}

	now := time.Now()
	user.CreatedAt = now
	user.UpdatedAt = now

	// 6. Insert ke database
	return u.repo.Create(ctx, user)
}

// Login autentikasi user
func (u *userUseCase) Login(ctx context.Context, email, password string) (domain.User, error) {
	// Normalize email
	email = strings.ToLower(strings.TrimSpace(email))

	// Cari user by email
	user, err := u.repo.GetByEmail(ctx, email)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return domain.User{}, errors.New("invalid email or password")
		}
		return domain.User{}, err
	}

	// Verify password
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return domain.User{}, errors.New("invalid email or password")
	}

	// Jangan return password
	user.Password = ""
	return user, nil
}

// GetUserByID ambil user by ID
func (u *userUseCase) GetUserByID(ctx context.Context, id string) (domain.User, error) {
	if id == "" {
		return domain.User{}, errors.New("user ID is required")
	}

	user, err := u.repo.GetByID(ctx, id)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return domain.User{}, errors.New("user not found")
		}
		return domain.User{}, err
	}

	// Jangan return password
	user.Password = ""
	return user, nil
}
