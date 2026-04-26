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

// setupRouter adalah helper untuk menginisialisasi Gin dan ProductHandler
func setupRouter(mockUC *mocks.ProductUseCase) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	public := r.Group("/")
	admin := r.Group("/") // Untuk testing fungsi admin yang masih statis

	NewProductHandler(public, admin, mockUC)
	return r
}

func TestFetchAll(t *testing.T) {
	type testCase struct {
		name               string
		queryString        string
		mockSetup          func(m *mocks.ProductUseCase)
		expectedStatusCode int
	}

	tests := []testCase{
		{
			name:        "Gagal karena Internal Error Usecase",
			queryString: "?search=laptop",
			mockSetup: func(m *mocks.ProductUseCase) {
				m.On("GetProductsWithFilter", mock.Anything, mock.AnythingOfType("domain.ProductFilter")).
					Return(domain.ProductResponse{}, errors.New("db timeout")).Once()
			},
			expectedStatusCode: http.StatusInternalServerError,
		},
		{
			name:        "Sukses Ambil Data dengan Filter Lengkap",
			queryString: "?search=hp&min_price=1000000&max_price=5000000&page=2&limit=20&rating=4.5",
			mockSetup: func(m *mocks.ProductUseCase) {
				// Memastikan parsing strconv berjalan dengan membuat mock return data sukses
				mockResp := domain.ProductResponse{
					Data:       []domain.Product{{Name: "HP Gaming"}},
					Total:      100,
					Page:       2,
					Limit:      20,
					TotalPages: 5,
				}
				m.On("GetProductsWithFilter", mock.Anything, mock.AnythingOfType("domain.ProductFilter")).
					Return(mockResp, nil).Once()
			},
			expectedStatusCode: http.StatusOK,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockUC := new(mocks.ProductUseCase)
			tc.mockSetup(mockUC)

			r := setupRouter(mockUC)

			req := httptest.NewRequest(http.MethodGet, "/products"+tc.queryString, nil)
			rec := httptest.NewRecorder()

			r.ServeHTTP(rec, req)

			assert.Equal(t, tc.expectedStatusCode, rec.Code)
			mockUC.AssertExpectations(t)
		})
	}
}

func TestGetByID(t *testing.T) {
	type testCase struct {
		name               string
		paramID            string
		mockSetup          func(m *mocks.ProductUseCase)
		expectedStatusCode int
	}

	tests := []testCase{
		{
			name:    "Gagal - Product Tidak Ditemukan (Pasti 404)",
			paramID: "invalid-id",
			mockSetup: func(m *mocks.ProductUseCase) {
				m.On("GetProductByID", mock.Anything, "invalid-id").
					Return(domain.Product{}, errors.New("not found")).Once()
			},
			expectedStatusCode: http.StatusNotFound, // Ekspektasi 404 karena hardcoded di handler
		},
		{
			name:    "Sukses Get Detail Product",
			paramID: "valid-id",
			mockSetup: func(m *mocks.ProductUseCase) {
				m.On("GetProductByID", mock.Anything, "valid-id").
					Return(domain.Product{Name: "Tenda"}, nil).Once()
			},
			expectedStatusCode: http.StatusOK,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockUC := new(mocks.ProductUseCase)
			tc.mockSetup(mockUC)

			r := setupRouter(mockUC)

			req := httptest.NewRequest(http.MethodGet, "/product/"+tc.paramID, nil)
			rec := httptest.NewRecorder()

			r.ServeHTTP(rec, req)

			assert.Equal(t, tc.expectedStatusCode, rec.Code)
			mockUC.AssertExpectations(t)
		})
	}
}

func TestGetDeals(t *testing.T) {
	type testCase struct {
		name               string
		queryString        string
		mockSetup          func(m *mocks.ProductUseCase)
		expectedStatusCode int
	}

	tests := []testCase{
		{
			name:        "Gagal saat ambil Deals",
			queryString: "?limit=5",
			mockSetup: func(m *mocks.ProductUseCase) {
				m.On("GetDeals", mock.Anything, int64(5)). // Limit di-parse menjadi 5
										Return(nil, errors.New("db error")).Once()
			},
			expectedStatusCode: http.StatusInternalServerError,
		},
		{
			name:        "Sukses ambil Deals dengan limit default (10)",
			queryString: "", // Tanpa query limit
			mockSetup: func(m *mocks.ProductUseCase) {
				m.On("GetDeals", mock.Anything, int64(10)). // Limit default 10
										Return([]domain.Product{{Name: "Flash Sale"}}, nil).Once()
			},
			expectedStatusCode: http.StatusOK,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockUC := new(mocks.ProductUseCase)
			tc.mockSetup(mockUC)

			r := setupRouter(mockUC)

			req := httptest.NewRequest(http.MethodGet, "/products/deals"+tc.queryString, nil)
			rec := httptest.NewRecorder()

			r.ServeHTTP(rec, req)

			assert.Equal(t, tc.expectedStatusCode, rec.Code)
			mockUC.AssertExpectations(t)
		})
	}
}

func TestGetStats(t *testing.T) {
	type testCase struct {
		name               string
		mockSetup          func(m *mocks.ProductUseCase)
		expectedStatusCode int
	}

	tests := []testCase{
		{
			name: "Gagal ambil Stats",
			mockSetup: func(m *mocks.ProductUseCase) {
				m.On("GetStats", mock.Anything).
					Return(domain.ProductStats{}, errors.New("db error")).Once()
			},
			expectedStatusCode: http.StatusInternalServerError,
		},
		{
			name: "Sukses ambil Stats",
			mockSetup: func(m *mocks.ProductUseCase) {
				m.On("GetStats", mock.Anything).
					Return(domain.ProductStats{TotalProducts: 100, TotalShops: 10}, nil).Once()
			},
			expectedStatusCode: http.StatusOK,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockUC := new(mocks.ProductUseCase)
			tc.mockSetup(mockUC)

			r := setupRouter(mockUC)

			req := httptest.NewRequest(http.MethodGet, "/products/stats", nil)
			rec := httptest.NewRecorder()

			r.ServeHTTP(rec, req)

			assert.Equal(t, tc.expectedStatusCode, rec.Code)
			mockUC.AssertExpectations(t)
		})
	}
}

func TestGetStatsAdmin(t *testing.T) {
	// Fungsi ini tidak memanggil Usecase, jadi Mock tidak dibutuhkan
	mockUC := new(mocks.ProductUseCase)
	r := setupRouter(mockUC)

	req := httptest.NewRequest(http.MethodGet, "/products-admin/stats", nil)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	// Pastikan status kembaliannya adalah 200 OK
	assert.Equal(t, http.StatusOK, rec.Code)
}
