package usecase_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/username/shop-api/internal/category/usecase"
	"github.com/username/shop-api/internal/domain"
	"github.com/username/shop-api/internal/mocks"
)

func TestCreate(t *testing.T) {
	type testCase struct {
		name          string
		inputReq      domain.CategoryRequest
		mockSetup     func(m *mocks.CategoryRepository)
		expectedError error
	}

	tests := []testCase{
		{
			name: "Gagal - Database Error",
			inputReq: domain.CategoryRequest{
				Name: "Kategori Baru",
				Slug: "kategori-baru",
			},
			mockSetup: func(m *mocks.CategoryRepository) {
				m.On("Create", mock.Anything, mock.AnythingOfType("domain.Category")).
					Return(domain.Category{}, errors.New("db error")).Once()
			},
			expectedError: errors.New("db error"),
		},
		{
			name: "Sukses Create",
			inputReq: domain.CategoryRequest{
				Name: "Kategori Baru",
				Slug: "kategori-baru",
			},
			mockSetup: func(m *mocks.CategoryRepository) {
				m.On("Create", mock.Anything, mock.AnythingOfType("domain.Category")).
					Return(domain.Category{Name: "Kategori Baru", Slug: "kategori-baru"}, nil).Once()
			},
			expectedError: nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockRepo := new(mocks.CategoryRepository)
			tc.mockSetup(mockRepo)

			catUC := usecase.NewCategoryUseCase(mockRepo)
			_, err := catUC.Create(context.Background(), tc.inputReq)

			assert.Equal(t, tc.expectedError, err)
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestGetAll(t *testing.T) {
	type testCase struct {
		name          string
		mockSetup     func(m *mocks.CategoryRepository)
		expectedError error
	}

	tests := []testCase{
		{
			name: "Gagal Ambil Data",
			mockSetup: func(m *mocks.CategoryRepository) {
				m.On("GetAll", mock.Anything).
					Return(nil, errors.New("db error")).Once()
			},
			expectedError: errors.New("db error"),
		},
		{
			name: "Sukses Ambil Semua Kategori",
			mockSetup: func(m *mocks.CategoryRepository) {
				m.On("GetAll", mock.Anything).
					Return([]domain.Category{{Name: "Kategori 1"}}, nil).Once()
			},
			expectedError: nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockRepo := new(mocks.CategoryRepository)
			tc.mockSetup(mockRepo)

			catUC := usecase.NewCategoryUseCase(mockRepo)
			_, err := catUC.GetAll(context.Background())

			assert.Equal(t, tc.expectedError, err)
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestGetByID(t *testing.T) {
	type testCase struct {
		name          string
		inputID       string
		mockSetup     func(m *mocks.CategoryRepository)
		expectedError error
	}

	tests := []testCase{
		{
			name:    "Gagal - Tidak Ditemukan",
			inputID: "id-ngasal",
			mockSetup: func(m *mocks.CategoryRepository) {
				m.On("GetByID", mock.Anything, "id-ngasal").
					Return(domain.Category{}, errors.New("category not found")).Once()
			},
			expectedError: errors.New("category not found"),
		},
		{
			name:    "Sukses Ambil Data by ID",
			inputID: "valid-id",
			mockSetup: func(m *mocks.CategoryRepository) {
				m.On("GetByID", mock.Anything, "valid-id").
					Return(domain.Category{Name: "Kategori Valid"}, nil).Once()
			},
			expectedError: nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockRepo := new(mocks.CategoryRepository)
			tc.mockSetup(mockRepo)

			catUC := usecase.NewCategoryUseCase(mockRepo)
			_, err := catUC.GetByID(context.Background(), tc.inputID)

			assert.Equal(t, tc.expectedError, err)
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestUpdate(t *testing.T) {
	type testCase struct {
		name          string
		paramID       string
		inputReq      domain.CategoryRequest
		mockSetup     func(m *mocks.CategoryRepository)
		expectedError error
	}

	tests := []testCase{
		{
			name:    "Gagal saat proses Update ke Database",
			paramID: "id-123",
			inputReq: domain.CategoryRequest{
				Name: "Update Kat",
			},
			mockSetup: func(m *mocks.CategoryRepository) {
				// Tidak perlu mock GetByID, langsung mock Update!
				m.On("Update", mock.Anything, "id-123", mock.AnythingOfType("domain.Category")).
					Return(domain.Category{}, errors.New("db update error")).Once()
			},
			expectedError: errors.New("db update error"),
		},
		{
			name:    "Sukses Update Kategori",
			paramID: "id-123",
			inputReq: domain.CategoryRequest{
				Name: "Update Kat",
				Slug: "update-kat",
			},
			mockSetup: func(m *mocks.CategoryRepository) {
				m.On("Update", mock.Anything, "id-123", mock.AnythingOfType("domain.Category")).
					Return(domain.Category{Name: "Update Kat"}, nil).Once()
			},
			expectedError: nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockRepo := new(mocks.CategoryRepository)
			tc.mockSetup(mockRepo)

			catUC := usecase.NewCategoryUseCase(mockRepo)
			_, err := catUC.Update(context.Background(), tc.paramID, tc.inputReq)

			assert.Equal(t, tc.expectedError, err)
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestDelete(t *testing.T) {
	type testCase struct {
		name          string
		paramID       string
		mockSetup     func(m *mocks.CategoryRepository)
		expectedError error
	}

	tests := []testCase{
		{
			name:    "Gagal Delete",
			paramID: "id-error",
			mockSetup: func(m *mocks.CategoryRepository) {
				m.On("Delete", mock.Anything, "id-error").
					Return(errors.New("db timeout")).Once()
			},
			expectedError: errors.New("db timeout"),
		},
		{
			name:    "Sukses Delete",
			paramID: "id-sukses",
			mockSetup: func(m *mocks.CategoryRepository) {
				m.On("Delete", mock.Anything, "id-sukses").
					Return(nil).Once()
			},
			expectedError: nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockRepo := new(mocks.CategoryRepository)
			tc.mockSetup(mockRepo)

			catUC := usecase.NewCategoryUseCase(mockRepo)
			err := catUC.Delete(context.Background(), tc.paramID)

			assert.Equal(t, tc.expectedError, err)
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestSyncCategories(t *testing.T) {
	type testCase struct {
		name          string
		mockSetup     func(m *mocks.CategoryRepository)
		expectedError error
		expectedCount int64
	}

	tests := []testCase{
		{
			name: "Gagal Sinkronisasi",
			mockSetup: func(m *mocks.CategoryRepository) {
				m.On("SyncCategories", mock.Anything).
					Return(int64(0), errors.New("sync failed")).Once()
			},
			expectedError: errors.New("sync failed"),
			expectedCount: 0,
		},
		{
			name: "Sukses Sinkronisasi",
			mockSetup: func(m *mocks.CategoryRepository) {
				m.On("SyncCategories", mock.Anything).
					Return(int64(42), nil).Once() // Anggap 42 kategori tersinkronisasi
			},
			expectedError: nil,
			expectedCount: 42,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockRepo := new(mocks.CategoryRepository)
			tc.mockSetup(mockRepo)

			catUC := usecase.NewCategoryUseCase(mockRepo)
			count, err := catUC.SyncCategories(context.Background())

			assert.Equal(t, tc.expectedError, err)
			assert.Equal(t, tc.expectedCount, count)
			mockRepo.AssertExpectations(t)
		})
	}
}
