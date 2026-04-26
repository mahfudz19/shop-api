package usecase_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/username/shop-api/internal/domain"
	"github.com/username/shop-api/internal/mocks"
	"github.com/username/shop-api/internal/product/usecase"
)

func TestGetProductsWithFilter(t *testing.T) {
	type testCase struct {
		name          string
		inputFilter   domain.ProductFilter
		mockSetup     func(m *mocks.ProductRepository)
		expectedError error
		expectedResp  domain.ProductResponse
	}

	tests := []testCase{
		{
			name: "Gagal karena Database Error",
			inputFilter: domain.ProductFilter{
				Page:  1,
				Limit: 10,
			},
			mockSetup: func(m *mocks.ProductRepository) {
				m.On("FetchWithFilter", mock.Anything, mock.AnythingOfType("domain.ProductFilter")).
					Return(nil, int64(0), errors.New("db error")).Once()
			},
			expectedError: errors.New("db error"),
			expectedResp:  domain.ProductResponse{},
		},
		{
			name: "Sukses & Sanitasi - Page 0/Limit 0 diubah jadi Page 1/Limit 10",
			inputFilter: domain.ProductFilter{
				Page:  -5, // User iseng mengirim page negatif
				Limit: 0,  // Limit kosong
			},
			mockSetup: func(m *mocks.ProductRepository) {
				m.On("FetchWithFilter", mock.Anything, mock.MatchedBy(func(f domain.ProductFilter) bool {
					// Pastikan repo dipanggil dengan nilai yang SUDAH dikoreksi
					return f.Page == 1 && f.Limit == 10
				})).Return([]domain.Product{{Name: "Produk 1"}}, int64(25), nil).Once() // Total 25 item
			},
			expectedError: nil,
			expectedResp: domain.ProductResponse{
				Total:      25,
				Page:       1,
				Limit:      10,
				TotalPages: 3, // 25 / 10 = 2.5, dibulatkan ke atas (math.Ceil) jadi 3
			},
		},
		{
			name: "Sukses & Sanitasi - Limit Ekstrem di-cap ke 100",
			inputFilter: domain.ProductFilter{
				Page:  2,
				Limit: 5000, // User mencoba mengambil 5000 data
			},
			mockSetup: func(m *mocks.ProductRepository) {
				m.On("FetchWithFilter", mock.Anything, mock.MatchedBy(func(f domain.ProductFilter) bool {
					return f.Page == 2 && f.Limit == 100 // Dipaksa maksimal 100
				})).Return([]domain.Product{{Name: "Produk Massal"}}, int64(250), nil).Once() // Total 250
			},
			expectedError: nil,
			expectedResp: domain.ProductResponse{
				Total:      250,
				Page:       2,
				Limit:      100,
				TotalPages: 3, // 250 / 100 = 2.5 -> Ceil jadi 3
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockRepo := new(mocks.ProductRepository)
			tc.mockSetup(mockRepo)

			prodUC := usecase.NewProductUseCase(mockRepo)
			resp, err := prodUC.GetProductsWithFilter(context.Background(), tc.inputFilter)

			assert.Equal(t, tc.expectedError, err)
			if err == nil {
				assert.Equal(t, tc.expectedResp.Total, resp.Total)
				assert.Equal(t, tc.expectedResp.Page, resp.Page)
				assert.Equal(t, tc.expectedResp.Limit, resp.Limit)
				assert.Equal(t, tc.expectedResp.TotalPages, resp.TotalPages)
			}
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestGetProductByID(t *testing.T) {
	type testCase struct {
		name          string
		inputID       string
		mockSetup     func(m *mocks.ProductRepository)
		expectedError error
	}

	tests := []testCase{
		{
			name:    "Gagal - ID Kosong",
			inputID: "",
			mockSetup: func(_ *mocks.ProductRepository) {
				// Tidak memanggil database sama sekali
			},
			expectedError: errors.New("product ID is required"),
		},
		{
			name:    "Gagal - Tidak Ditemukan",
			inputID: "id-ngasal",
			mockSetup: func(m *mocks.ProductRepository) {
				m.On("GetByID", mock.Anything, "id-ngasal").
					Return(domain.Product{}, errors.New("not found")).Once()
			},
			expectedError: errors.New("not found"),
		},
		{
			name:    "Sukses",
			inputID: "id-valid",
			mockSetup: func(m *mocks.ProductRepository) {
				m.On("GetByID", mock.Anything, "id-valid").
					Return(domain.Product{Name: "Valid"}, nil).Once()
			},
			expectedError: nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockRepo := new(mocks.ProductRepository)
			tc.mockSetup(mockRepo)

			prodUC := usecase.NewProductUseCase(mockRepo)
			_, err := prodUC.GetProductByID(context.Background(), tc.inputID)

			assert.Equal(t, tc.expectedError, err)
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestGetDeals(t *testing.T) {
	type testCase struct {
		name          string
		inputLimit    int64
		mockSetup     func(m *mocks.ProductRepository)
		expectedError error
	}

	tests := []testCase{
		{
			name:       "Sanitasi - Limit Negatif dikembalikan ke 10",
			inputLimit: -5,
			mockSetup: func(m *mocks.ProductRepository) {
				m.On("GetDeals", mock.Anything, int64(10)).
					Return([]domain.Product{}, nil).Once()
			},
			expectedError: nil,
		},
		{
			name:       "Sanitasi - Limit di atas 50 dipaksa maksimal 50",
			inputLimit: 99,
			mockSetup: func(m *mocks.ProductRepository) {
				m.On("GetDeals", mock.Anything, int64(50)).
					Return([]domain.Product{}, nil).Once()
			},
			expectedError: nil,
		},
		{
			name:       "Normal - Limit wajar 20",
			inputLimit: 20,
			mockSetup: func(m *mocks.ProductRepository) {
				m.On("GetDeals", mock.Anything, int64(20)).
					Return([]domain.Product{{Name: "Deal"}}, nil).Once()
			},
			expectedError: nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockRepo := new(mocks.ProductRepository)
			tc.mockSetup(mockRepo)

			prodUC := usecase.NewProductUseCase(mockRepo)
			_, err := prodUC.GetDeals(context.Background(), tc.inputLimit)

			assert.Equal(t, tc.expectedError, err)
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestGetStats(t *testing.T) {
	mockRepo := new(mocks.ProductRepository)
	mockRepo.On("GetStats", mock.Anything).Return(domain.ProductStats{TotalProducts: 100}, nil).Once()

	prodUC := usecase.NewProductUseCase(mockRepo)
	stats, err := prodUC.GetStats(context.Background())

	assert.NoError(t, err)
	assert.Equal(t, int64(100), stats.TotalProducts)
	mockRepo.AssertExpectations(t)
}
