package usecase_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/username/shop-api/internal/domain"
	"github.com/username/shop-api/internal/mocks"
	"github.com/username/shop-api/internal/promotion/usecase"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func TestCreate(t *testing.T) {
	type testCase struct {
		name          string
		inputReq      domain.PromotionRequest
		mockSetup     func(m *mocks.PromotionRepository)
		expectedError error
	}

	tests := []testCase{
		{
			name: "Sukses Membuat Promosi",
			inputReq: domain.PromotionRequest{
				Title:       "Promo Akhir Tahun",
				Description: "Diskon 50%",
				ImageURL:    "http://image.com/promo.png",
				IsActive:    true,
			},
			mockSetup: func(m *mocks.PromotionRepository) {
				// Gunakan AnythingOfType("domain.Promotion") karena waktu (time.Now) akan selalu dinamis
				m.On("Create", mock.Anything, mock.AnythingOfType("domain.Promotion")).
					Return(domain.Promotion{ID: bson.NewObjectID()}, nil).Once()
			},
			expectedError: nil,
		},
		{
			name: "Gagal karena Repository Error",
			inputReq: domain.PromotionRequest{
				Title: "Promo Gagal",
			},
			mockSetup: func(m *mocks.PromotionRepository) {
				m.On("Create", mock.Anything, mock.AnythingOfType("domain.Promotion")).
					Return(domain.Promotion{}, errors.New("database timeout")).Once()
			},
			expectedError: errors.New("database timeout"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockRepo := new(mocks.PromotionRepository)
			tc.mockSetup(mockRepo)

			promoUC := usecase.NewPromotionUseCase(mockRepo)
			_, err := promoUC.Create(context.Background(), tc.inputReq)

			assert.Equal(t, tc.expectedError, err)
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestGetAll(t *testing.T) {
	type testCase struct {
		name           string
		inputActive    bool
		mockSetup      func(m *mocks.PromotionRepository)
		expectedError  error
		expectedLength int
	}

	tests := []testCase{
		{
			name:        "Sukses mengambil semua promosi",
			inputActive: false,
			mockSetup: func(m *mocks.PromotionRepository) {
				m.On("GetAll", mock.Anything, false).
					Return([]domain.Promotion{{Title: "A"}, {Title: "B"}}, nil).Once()
			},
			expectedError:  nil,
			expectedLength: 2,
		},
		{
			name:        "Gagal mengambil promosi",
			inputActive: true,
			mockSetup: func(m *mocks.PromotionRepository) {
				m.On("GetAll", mock.Anything, true).
					Return(nil, errors.New("db error")).Once()
			},
			expectedError:  errors.New("db error"),
			expectedLength: 0,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockRepo := new(mocks.PromotionRepository)
			tc.mockSetup(mockRepo)

			promoUC := usecase.NewPromotionUseCase(mockRepo)
			res, err := promoUC.GetAll(context.Background(), tc.inputActive)

			assert.Equal(t, tc.expectedError, err)
			assert.Len(t, res, tc.expectedLength)
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestGetByID(t *testing.T) {
	type testCase struct {
		name          string
		paramID       string
		mockSetup     func(m *mocks.PromotionRepository)
		expectedError error
	}

	tests := []testCase{
		{
			name:    "Gagal - Promotion tidak ditemukan",
			paramID: "id-123",
			mockSetup: func(m *mocks.PromotionRepository) {
				m.On("GetByID", mock.Anything, "id-123").
					Return(domain.Promotion{}, errors.New("not found")).Once()
			},
			expectedError: errors.New("not found"),
		},
		{
			name:    "Sukses Get By ID",
			paramID: "id-123",
			mockSetup: func(m *mocks.PromotionRepository) {
				m.On("GetByID", mock.Anything, "id-123").
					Return(domain.Promotion{Title: "Promo 1"}, nil).Once()
			},
			expectedError: nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockRepo := new(mocks.PromotionRepository)
			tc.mockSetup(mockRepo)

			promoUC := usecase.NewPromotionUseCase(mockRepo)
			_, err := promoUC.GetByID(context.Background(), tc.paramID)

			assert.Equal(t, tc.expectedError, err)
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestUpdate(t *testing.T) {
	type testCase struct {
		name          string
		paramID       string
		inputReq      domain.PromotionRequest
		mockSetup     func(m *mocks.PromotionRepository)
		expectedError error
	}

	tests := []testCase{
		{
			name:     "Gagal karena Promotion tidak ada di DB",
			paramID:  "id-123",
			inputReq: domain.PromotionRequest{Title: "Update Promo"},
			mockSetup: func(m *mocks.PromotionRepository) {
				// GetByID langsung me-return error
				m.On("GetByID", mock.Anything, "id-123").
					Return(domain.Promotion{}, errors.New("not found")).Once()
			},
			expectedError: errors.New("not found"),
		},
		{
			name:     "Gagal saat proses Update di Repository",
			paramID:  "id-123",
			inputReq: domain.PromotionRequest{Title: "Update Promo"},
			mockSetup: func(m *mocks.PromotionRepository) {
				// 1. GetByID sukses
				existingPromo := domain.Promotion{ID: bson.NewObjectID(), Title: "Old", CreatedAt: time.Now()}
				m.On("GetByID", mock.Anything, "id-123").Return(existingPromo, nil).Once()

				// 2. Update gagal
				m.On("Update", mock.Anything, "id-123", mock.AnythingOfType("domain.Promotion")).
					Return(domain.Promotion{}, errors.New("db error saat update")).Once()
			},
			expectedError: errors.New("db error saat update"),
		},
		{
			name:     "Sukses Update Promosi",
			paramID:  "id-123",
			inputReq: domain.PromotionRequest{Title: "Update Promo"},
			mockSetup: func(m *mocks.PromotionRepository) {
				// 1. GetByID sukses
				existingPromo := domain.Promotion{ID: bson.NewObjectID(), Title: "Old", CreatedAt: time.Now()}
				m.On("GetByID", mock.Anything, "id-123").Return(existingPromo, nil).Once()

				// 2. Update sukses
				m.On("Update", mock.Anything, "id-123", mock.AnythingOfType("domain.Promotion")).
					Return(domain.Promotion{Title: "Update Promo"}, nil).Once()
			},
			expectedError: nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockRepo := new(mocks.PromotionRepository)
			tc.mockSetup(mockRepo)

			promoUC := usecase.NewPromotionUseCase(mockRepo)
			_, err := promoUC.Update(context.Background(), tc.paramID, tc.inputReq)

			assert.Equal(t, tc.expectedError, err)
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestDelete(t *testing.T) {
	type testCase struct {
		name          string
		paramID       string
		mockSetup     func(m *mocks.PromotionRepository)
		expectedError error
	}

	tests := []testCase{
		{
			name:    "Gagal Delete",
			paramID: "id-error",
			mockSetup: func(m *mocks.PromotionRepository) {
				m.On("Delete", mock.Anything, "id-error").
					Return(errors.New("db timeout")).Once()
			},
			expectedError: errors.New("db timeout"),
		},
		{
			name:    "Sukses Delete",
			paramID: "id-sukses",
			mockSetup: func(m *mocks.PromotionRepository) {
				m.On("Delete", mock.Anything, "id-sukses").
					Return(nil).Once()
			},
			expectedError: nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockRepo := new(mocks.PromotionRepository)
			tc.mockSetup(mockRepo)

			promoUC := usecase.NewPromotionUseCase(mockRepo)
			err := promoUC.Delete(context.Background(), tc.paramID)

			assert.Equal(t, tc.expectedError, err)
			mockRepo.AssertExpectations(t)
		})
	}
}
