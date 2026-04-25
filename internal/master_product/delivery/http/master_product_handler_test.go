package http

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/username/shop-api/internal/domain"
	"github.com/username/shop-api/internal/mocks"
)

// setupRouter adalah helper untuk menginisialisasi Gin dan Handler dengan Dummy Middleware
func setupRouter(mockUC *mocks.MasterProductUseCase) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	public := r.Group("/")

	// Setup Router Protected dengan Dummy Middleware untuk simulasi Auth
	protected := r.Group("/")
	protected.Use(func(c *gin.Context) {
		c.Set("user_id", "user-123")
		c.Set("role", "user")
		c.Next()
	})

	NewMasterProductHandler(public, protected, mockUC)
	return r
}

func TestGetDetailByID(t *testing.T) {
	type testCase struct {
		name               string
		paramID            string
		mockSetup          func(m *mocks.MasterProductUseCase)
		expectedStatusCode int
	}

	tests := []testCase{
		{
			name:    "Gagal - Master Product Tidak Ditemukan (Atau DB Error)",
			paramID: "invalid-id",
			mockSetup: func(m *mocks.MasterProductUseCase) {
				// Handler selalu mengembalikan 404 jika ada error apapun dari usecase
				m.On("GetDetailByID", mock.Anything, "invalid-id").
					Return(domain.MasterProductDetail{}, errors.New("db error or not found")).Once()
			},
			expectedStatusCode: http.StatusNotFound, // Ekspektasi pasti 404 sesuai kode handler
		},
		{
			name:    "Sukses Get Detail By ID",
			paramID: "valid-id",
			mockSetup: func(m *mocks.MasterProductUseCase) {
				// Harus me-return struct MasterProductDetail
				mockDetail := domain.MasterProductDetail{
					MasterProduct: domain.MasterProduct{Name: "Laptop Gaming"},
					MinPrice:      10000000,
					MaxPrice:      15000000,
					TotalOffers:   3,
				}
				m.On("GetDetailByID", mock.Anything, "valid-id").
					Return(mockDetail, nil).Once()
			},
			expectedStatusCode: http.StatusOK, // Ekspektasi 200
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockUC := new(mocks.MasterProductUseCase)
			tc.mockSetup(mockUC)

			r := setupRouter(mockUC)

			req := httptest.NewRequest(http.MethodGet, "/master-product/"+tc.paramID, nil)
			rec := httptest.NewRecorder()

			r.ServeHTTP(rec, req)

			assert.Equal(t, tc.expectedStatusCode, rec.Code)
			mockUC.AssertExpectations(t)
		})
	}
}

func TestTestAuth(t *testing.T) {
	type testCase struct {
		name               string
		paramID            string
		expectedStatusCode int
	}

	tests := []testCase{
		{
			name:               "Sukses Tembus Protected Route",
			paramID:            "test-id-123",
			expectedStatusCode: http.StatusOK,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Kita tidak perlu setup mockUseCase karena TestAuth tidak memanggil usecase sama sekali
			mockUC := new(mocks.MasterProductUseCase)
			r := setupRouter(mockUC)

			// Tembak URL protected route
			req := httptest.NewRequest(http.MethodGet, "/master-product/"+tc.paramID+"/test", nil)
			rec := httptest.NewRecorder()

			r.ServeHTTP(rec, req)

			assert.Equal(t, tc.expectedStatusCode, rec.Code)
		})
	}
}
