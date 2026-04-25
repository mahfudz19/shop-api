package usecase_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/username/shop-api/internal/domain"
	"github.com/username/shop-api/internal/master_product/usecase"
	"github.com/username/shop-api/internal/mocks"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func TestGetDetailByID(t *testing.T) {
	type testCase struct {
		name          string
		paramID       string
		mockSetup     func(m *mocks.MasterProductRepository)
		expectedError error
		expectedData  domain.MasterProductDetail
	}

	// ID dummy untuk pengujian
	fakeObjectID := bson.NewObjectID()
	fakeMasterProduct := domain.MasterProduct{
		ID:   fakeObjectID,
		Name: "Samsung Galaxy S24 Ultra",
	}

	tests := []testCase{
		{
			name:    "Gagal - Master Product tidak ditemukan di database pertama",
			paramID: "id-error",
			mockSetup: func(m *mocks.MasterProductRepository) {
				m.On("GetByID", mock.Anything, "id-error").
					Return(domain.MasterProduct{}, errors.New("master product not found")).Once()
				// Tidak memanggil GetOffersByMasterID
			},
			expectedError: errors.New("master product not found"),
			expectedData:  domain.MasterProductDetail{},
		},
		{
			name:    "Gagal - Error saat mengambil Offers (Penawaran)",
			paramID: "id-valid",
			mockSetup: func(m *mocks.MasterProductRepository) {
				// 1. GetByID sukses
				m.On("GetByID", mock.Anything, "id-valid").
					Return(fakeMasterProduct, nil).Once()

				// 2. GetOffers gagal
				m.On("GetOffersByMasterID", mock.Anything, fakeObjectID).
					Return(nil, errors.New("db error on offers")).Once()
			},
			expectedError: errors.New("db error on offers"),
			expectedData:  domain.MasterProductDetail{},
		},
		{
			name:    "Sukses - Produk ditemukan tapi TIDAK ADA penawaran (Offers = 0)",
			paramID: "id-no-offers",
			mockSetup: func(m *mocks.MasterProductRepository) {
				m.On("GetByID", mock.Anything, "id-no-offers").
					Return(fakeMasterProduct, nil).Once()

				// Me-return array kosong
				m.On("GetOffersByMasterID", mock.Anything, fakeObjectID).
					Return([]domain.Product{}, nil).Once()
			},
			expectedError: nil,
			expectedData: domain.MasterProductDetail{
				MasterProduct: fakeMasterProduct,
				Offers:        []domain.Product{},
				MinPrice:      0, // Harus 0
				MaxPrice:      0, // Harus 0
				Savings:       0, // Harus 0
				TotalOffers:   0,
			},
		},
		{
			name:    "Sukses - Kalkulasi Harga Termurah, Termahal, dan Savings Akurat",
			paramID: "id-with-offers",
			mockSetup: func(m *mocks.MasterProductRepository) {
				m.On("GetByID", mock.Anything, "id-with-offers").
					Return(fakeMasterProduct, nil).Once()

				// Membuat simulasi 3 penawaran harga yang sudah diurutkan (Ascending)
				mockOffers := []domain.Product{
					{PriceRp: 18000000}, // Index 0: Termurah
					{PriceRp: 19500000}, // Index 1: Tengah
					{PriceRp: 21000000}, // Index 2: Termahal
				}

				m.On("GetOffersByMasterID", mock.Anything, fakeObjectID).
					Return(mockOffers, nil).Once()
			},
			expectedError: nil,
			expectedData: domain.MasterProductDetail{
				MasterProduct: fakeMasterProduct,
				// Offers di-skip untuk asersi detail, tapi kita cek nilai kalkulasinya:
				MinPrice:    18000000,
				MaxPrice:    21000000,
				Savings:     3000000, // (21 juta - 18 juta)
				TotalOffers: 3,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockRepo := new(mocks.MasterProductRepository)
			tc.mockSetup(mockRepo)

			mpUC := usecase.NewMasterProductUseCase(mockRepo)
			detail, err := mpUC.GetDetailByID(context.Background(), tc.paramID)

			assert.Equal(t, tc.expectedError, err)

			// Asersi hasil data (hanya jika sukses)
			if err == nil {
				assert.Equal(t, tc.expectedData.MinPrice, detail.MinPrice)
				assert.Equal(t, tc.expectedData.MaxPrice, detail.MaxPrice)
				assert.Equal(t, tc.expectedData.Savings, detail.Savings)
				assert.Equal(t, tc.expectedData.TotalOffers, detail.TotalOffers)
				assert.Equal(t, tc.expectedData.MasterProduct.Name, detail.MasterProduct.Name)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}
