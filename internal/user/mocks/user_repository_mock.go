// Package mocks = Mock repository untuk User
package mocks

import (
	"context"

	"github.com/stretchr/testify/mock"
	"github.com/username/shop-api/internal/domain"
)

// MockUserRepository adalah representasi database palsu
type MockUserRepository struct {
	mock.Mock
}

// Create insert user baru
func (m *MockUserRepository) Create(ctx context.Context, user domain.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

// GetByEmail cari user berdasarkan email
func (m *MockUserRepository) GetByEmail(ctx context.Context, email string) (domain.User, error) {
	args := m.Called(ctx, email)
	return args.Get(0).(domain.User), args.Error(1)
}

// GetByID cari user berdasarkan ID
func (m *MockUserRepository) GetByID(ctx context.Context, id string) (domain.User, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(domain.User), args.Error(1)
}

// EmailExists cek apakah email sudah terdaftar
func (m *MockUserRepository) EmailExists(ctx context.Context, email string) (bool, error) {
	args := m.Called(ctx, email)
	return args.Bool(0), args.Error(1)
}
