package usecase_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/username/shop-api/internal/article/usecase"
	"github.com/username/shop-api/internal/domain"
	"github.com/username/shop-api/internal/mocks"
)

func TestCreate(t *testing.T) {
	type testCase struct {
		name          string
		inputReq      domain.ArticleRequest
		mockSetup     func(m *mocks.ArticleRepository)
		expectedError error
	}

	tests := []testCase{
		{
			name: "Gagal - Database Error",
			inputReq: domain.ArticleRequest{
				Title: "Test Article",
			},
			mockSetup: func(m *mocks.ArticleRepository) {
				m.On("Create", mock.Anything, mock.AnythingOfType("domain.Article")).
					Return(domain.Article{}, errors.New("db error")).Once()
			},
			expectedError: errors.New("db error"),
		},
		{
			name: "Sukses Create - Status Draft (IsPublished = false)",
			inputReq: domain.ArticleRequest{
				Title:       "Test Draft",
				IsPublished: false,
			},
			mockSetup: func(m *mocks.ArticleRepository) {
				// Memastikan masuk ke repo tanpa error
				m.On("Create", mock.Anything, mock.AnythingOfType("domain.Article")).
					Return(domain.Article{Title: "Test Draft"}, nil).Once()
			},
			expectedError: nil,
		},
		{
			name: "Sukses Create - Status Published (IsPublished = true)",
			inputReq: domain.ArticleRequest{
				Title:       "Test Published",
				IsPublished: true,
			},
			mockSetup: func(m *mocks.ArticleRepository) {
				m.On("Create", mock.Anything, mock.AnythingOfType("domain.Article")).
					Return(domain.Article{Title: "Test Published"}, nil).Once()
			},
			expectedError: nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockRepo := new(mocks.ArticleRepository)
			tc.mockSetup(mockRepo)

			articleUC := usecase.NewArticleUseCase(mockRepo)
			_, err := articleUC.Create(context.Background(), tc.inputReq)

			assert.Equal(t, tc.expectedError, err)
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestGetAll(t *testing.T) {
	type testCase struct {
		name               string
		inputOnlyPublished bool
		mockSetup          func(m *mocks.ArticleRepository)
		expectedError      error
	}

	tests := []testCase{
		{
			name:               "Gagal Ambil Data",
			inputOnlyPublished: false,
			mockSetup: func(m *mocks.ArticleRepository) {
				m.On("GetAll", mock.Anything, false).
					Return(nil, errors.New("db error")).Once()
			},
			expectedError: errors.New("db error"),
		},
		{
			name:               "Sukses Ambil Semua (Termasuk Draft)",
			inputOnlyPublished: false,
			mockSetup: func(m *mocks.ArticleRepository) {
				m.On("GetAll", mock.Anything, false).
					Return([]domain.Article{{Title: "A"}}, nil).Once()
			},
			expectedError: nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockRepo := new(mocks.ArticleRepository)
			tc.mockSetup(mockRepo)

			articleUC := usecase.NewArticleUseCase(mockRepo)
			_, err := articleUC.GetAll(context.Background(), tc.inputOnlyPublished)

			assert.Equal(t, tc.expectedError, err)
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestGetBySlug(t *testing.T) {
	type testCase struct {
		name          string
		inputSlug     string
		mockSetup     func(m *mocks.ArticleRepository)
		expectedError error
	}

	tests := []testCase{
		{
			name:      "Gagal - Tidak Ditemukan",
			inputSlug: "slug-ngasal",
			mockSetup: func(m *mocks.ArticleRepository) {
				m.On("GetBySlug", mock.Anything, "slug-ngasal").
					Return(domain.Article{}, errors.New("not found")).Once()
			},
			expectedError: errors.New("not found"),
		},
		{
			name:      "Sukses Ambil Data by Slug",
			inputSlug: "judul-artikel",
			mockSetup: func(m *mocks.ArticleRepository) {
				m.On("GetBySlug", mock.Anything, "judul-artikel").
					Return(domain.Article{Title: "Judul Artikel"}, nil).Once()
			},
			expectedError: nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockRepo := new(mocks.ArticleRepository)
			tc.mockSetup(mockRepo)

			articleUC := usecase.NewArticleUseCase(mockRepo)
			_, err := articleUC.GetBySlug(context.Background(), tc.inputSlug)

			assert.Equal(t, tc.expectedError, err)
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestUpdate(t *testing.T) {
	type testCase struct {
		name          string
		paramID       string
		inputReq      domain.ArticleRequest
		mockSetup     func(m *mocks.ArticleRepository)
		expectedError error
	}

	tests := []testCase{
		{
			name:    "Gagal saat proses Update ke Database",
			paramID: "id-123",
			inputReq: domain.ArticleRequest{
				Title: "Update",
			},
			mockSetup: func(m *mocks.ArticleRepository) {
				// Cukup panggil Update, karena kode asli tidak memanggil GetByID
				m.On("Update", mock.Anything, "id-123", mock.AnythingOfType("domain.Article")).
					Return(domain.Article{}, errors.New("db update error")).Once()
			},
			expectedError: errors.New("db update error"),
		},
		{
			name:    "Sukses Update - Status Draft",
			paramID: "id-123",
			inputReq: domain.ArticleRequest{
				Title:       "Update Ke Draft",
				IsPublished: false,
			},
			mockSetup: func(m *mocks.ArticleRepository) {
				m.On("Update", mock.Anything, "id-123", mock.AnythingOfType("domain.Article")).
					Return(domain.Article{Title: "Update Ke Draft"}, nil).Once()
			},
			expectedError: nil,
		},
		{
			name:    "Sukses Update - Status Published",
			paramID: "id-123",
			inputReq: domain.ArticleRequest{
				Title:       "Update Ke Published",
				IsPublished: true,
			},
			mockSetup: func(m *mocks.ArticleRepository) {
				m.On("Update", mock.Anything, "id-123", mock.AnythingOfType("domain.Article")).
					Return(domain.Article{Title: "Update Ke Published"}, nil).Once()
			},
			expectedError: nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockRepo := new(mocks.ArticleRepository)
			tc.mockSetup(mockRepo)

			articleUC := usecase.NewArticleUseCase(mockRepo)
			_, err := articleUC.Update(context.Background(), tc.paramID, tc.inputReq)

			assert.Equal(t, tc.expectedError, err)
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestDelete(t *testing.T) {
	type testCase struct {
		name          string
		paramID       string
		mockSetup     func(m *mocks.ArticleRepository)
		expectedError error
	}

	tests := []testCase{
		{
			name:    "Gagal Delete",
			paramID: "id-error",
			mockSetup: func(m *mocks.ArticleRepository) {
				m.On("Delete", mock.Anything, "id-error").
					Return(errors.New("db timeout")).Once()
			},
			expectedError: errors.New("db timeout"),
		},
		{
			name:    "Sukses Delete",
			paramID: "id-sukses",
			mockSetup: func(m *mocks.ArticleRepository) {
				m.On("Delete", mock.Anything, "id-sukses").
					Return(nil).Once()
			},
			expectedError: nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockRepo := new(mocks.ArticleRepository)
			tc.mockSetup(mockRepo)

			articleUC := usecase.NewArticleUseCase(mockRepo)
			err := articleUC.Delete(context.Background(), tc.paramID)

			assert.Equal(t, tc.expectedError, err)
			mockRepo.AssertExpectations(t)
		})
	}
}
