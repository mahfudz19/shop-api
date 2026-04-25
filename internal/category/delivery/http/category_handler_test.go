package http

import (
	"bytes"
	"encoding/json"
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

// setupRouter adalah helper untuk menginisialisasi Gin dan CategoryHandler
func setupRouter(mockUC *mocks.CategoryUseCase) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	// Mendaftarkan router public dan admin (menggunakan r yang sama untuk testing)
	NewCategoryHandler(r, r, mockUC)
	return r
}

func TestCreate(t *testing.T) {
	type testCase struct {
		name               string
		inputPayload       interface{}
		mockSetup          func(m *mocks.CategoryUseCase)
		expectedStatusCode int
	}

	tests := []testCase{
		{
			name:               "Gagal format JSON / Validasi Required Kosong",
			inputPayload:       domain.CategoryRequest{Name: "Tanpa Slug"}, // Name diisi, Slug sengaja kosong
			mockSetup:          func(_ *mocks.CategoryUseCase) {},
			expectedStatusCode: http.StatusUnprocessableEntity, // 422 Error Validation
		},
		{
			name: "Gagal karena Usecase Error",
			inputPayload: domain.CategoryRequest{
				Name: "Kategori Baru",
				Slug: "kategori-baru",
			},
			mockSetup: func(m *mocks.CategoryUseCase) {
				m.On("Create", mock.Anything, mock.AnythingOfType("domain.CategoryRequest")).
					Return(domain.Category{}, errors.New("database timeout")).Once()
			},
			expectedStatusCode: http.StatusInternalServerError,
		},
		{
			name: "Sukses Membuat Kategori",
			inputPayload: domain.CategoryRequest{
				Name: "Kategori Baru",
				Slug: "kategori-baru",
			},
			mockSetup: func(m *mocks.CategoryUseCase) {
				m.On("Create", mock.Anything, mock.AnythingOfType("domain.CategoryRequest")).
					Return(domain.Category{Name: "Kategori Baru"}, nil).Once()
			},
			expectedStatusCode: http.StatusCreated,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockUC := new(mocks.CategoryUseCase)
			tc.mockSetup(mockUC)

			r := setupRouter(mockUC)

			var payload []byte
			if strPayload, ok := tc.inputPayload.(string); ok {
				payload = []byte(strPayload)
			} else {
				payload, _ = json.Marshal(tc.inputPayload)
			}

			req := httptest.NewRequest(http.MethodPost, "/categories", bytes.NewBuffer(payload))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			r.ServeHTTP(rec, req)

			assert.Equal(t, tc.expectedStatusCode, rec.Code)
			mockUC.AssertExpectations(t)
		})
	}
}

func TestGetAll(t *testing.T) {
	type testCase struct {
		name               string
		mockSetup          func(m *mocks.CategoryUseCase)
		expectedStatusCode int
	}

	tests := []testCase{
		{
			name: "Gagal Get All",
			mockSetup: func(m *mocks.CategoryUseCase) {
				m.On("GetAll", mock.Anything).
					Return(nil, errors.New("db error")).Once()
			},
			expectedStatusCode: http.StatusInternalServerError,
		},
		{
			name: "Sukses Get All",
			mockSetup: func(m *mocks.CategoryUseCase) {
				m.On("GetAll", mock.Anything).
					Return([]domain.Category{{Name: "Kat 1"}}, nil).Once()
			},
			expectedStatusCode: http.StatusOK,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockUC := new(mocks.CategoryUseCase)
			tc.mockSetup(mockUC)

			r := setupRouter(mockUC)

			req := httptest.NewRequest(http.MethodGet, "/categories", nil)
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
		mockSetup          func(m *mocks.CategoryUseCase)
		expectedStatusCode int
	}

	tests := []testCase{
		{
			name:    "Gagal karena Kategori tidak ditemukan",
			paramID: "invalid-id",
			mockSetup: func(m *mocks.CategoryUseCase) {
				m.On("GetByID", mock.Anything, "invalid-id").
					Return(domain.Category{}, errors.New("category not found")).Once()
			},
			expectedStatusCode: http.StatusNotFound,
		},
		{
			name:    "Sukses Get By ID",
			paramID: "valid-id",
			mockSetup: func(m *mocks.CategoryUseCase) {
				m.On("GetByID", mock.Anything, "valid-id").
					Return(domain.Category{Name: "Kategori"}, nil).Once()
			},
			expectedStatusCode: http.StatusOK, // 200
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockUC := new(mocks.CategoryUseCase)
			tc.mockSetup(mockUC)

			r := setupRouter(mockUC)

			req := httptest.NewRequest(http.MethodGet, "/categories/"+tc.paramID, nil)
			rec := httptest.NewRecorder()

			r.ServeHTTP(rec, req)

			assert.Equal(t, tc.expectedStatusCode, rec.Code)
			mockUC.AssertExpectations(t)
		})
	}
}

func TestUpdate(t *testing.T) {
	type testCase struct {
		name               string
		paramID            string
		inputPayload       interface{}
		mockSetup          func(m *mocks.CategoryUseCase)
		expectedStatusCode int
	}

	validPayload := domain.CategoryRequest{
		Name: "Update Nama",
		Slug: "update-nama",
	}

	tests := []testCase{
		{
			name:               "Gagal karena Validasi JSON (422)",
			paramID:            "valid-id",
			inputPayload:       domain.CategoryRequest{Name: "Tanpa Slug"},
			mockSetup:          func(_ *mocks.CategoryUseCase) {},
			expectedStatusCode: http.StatusUnprocessableEntity,
		},
		{
			name:         "Gagal karena Kategori Tidak Ada (Custom 404)",
			paramID:      "id-ilang",
			inputPayload: validPayload,
			mockSetup: func(m *mocks.CategoryUseCase) {
				// Me-return pesan spesifik yang dicek manual di handler
				m.On("Update", mock.Anything, "id-ilang", mock.AnythingOfType("domain.CategoryRequest")).
					Return(domain.Category{}, errors.New("category not found")).Once()
			},
			expectedStatusCode: http.StatusNotFound, // Diubah jadi 404 oleh handler
		},
		{
			name:         "Gagal karena Internal Error (500)",
			paramID:      "id-error",
			inputPayload: validPayload,
			mockSetup: func(m *mocks.CategoryUseCase) {
				m.On("Update", mock.Anything, "id-error", mock.AnythingOfType("domain.CategoryRequest")).
					Return(domain.Category{}, errors.New("database timeout")).Once()
			},
			expectedStatusCode: http.StatusInternalServerError,
		},
		{
			name:         "Sukses Update",
			paramID:      "valid-id",
			inputPayload: validPayload,
			mockSetup: func(m *mocks.CategoryUseCase) {
				m.On("Update", mock.Anything, "valid-id", mock.AnythingOfType("domain.CategoryRequest")).
					Return(domain.Category{Name: "Update Nama"}, nil).Once()
			},
			expectedStatusCode: http.StatusOK,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockUC := new(mocks.CategoryUseCase)
			tc.mockSetup(mockUC)

			r := setupRouter(mockUC)

			payload, _ := json.Marshal(tc.inputPayload)
			req := httptest.NewRequest(http.MethodPut, "/categories/"+tc.paramID, bytes.NewBuffer(payload))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			r.ServeHTTP(rec, req)

			assert.Equal(t, tc.expectedStatusCode, rec.Code)
			mockUC.AssertExpectations(t)
		})
	}
}

func TestDelete(t *testing.T) {
	type testCase struct {
		name               string
		paramID            string
		mockSetup          func(m *mocks.CategoryUseCase)
		expectedStatusCode int
	}

	tests := []testCase{
		{
			name:    "Gagal Delete karena Tidak Ada (Custom 404)",
			paramID: "id-ilang",
			mockSetup: func(m *mocks.CategoryUseCase) {
				m.On("Delete", mock.Anything, "id-ilang").
					Return(errors.New("category not found")).Once()
			},
			expectedStatusCode: http.StatusNotFound, // Dicek spesifik di handler
		},
		{
			name:    "Gagal Delete karena Error Internal",
			paramID: "id-error",
			mockSetup: func(m *mocks.CategoryUseCase) {
				m.On("Delete", mock.Anything, "id-error").
					Return(errors.New("db error")).Once()
			},
			expectedStatusCode: http.StatusInternalServerError,
		},
		{
			name:    "Sukses Delete",
			paramID: "valid-id",
			mockSetup: func(m *mocks.CategoryUseCase) {
				m.On("Delete", mock.Anything, "valid-id").
					Return(nil).Once()
			},
			expectedStatusCode: http.StatusOK,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockUC := new(mocks.CategoryUseCase)
			tc.mockSetup(mockUC)

			r := setupRouter(mockUC)

			req := httptest.NewRequest(http.MethodDelete, "/categories/"+tc.paramID, nil)
			rec := httptest.NewRecorder()

			r.ServeHTTP(rec, req)

			assert.Equal(t, tc.expectedStatusCode, rec.Code)
			mockUC.AssertExpectations(t)
		})
	}
}

func TestSyncCategories(t *testing.T) {
	type testCase struct {
		name               string
		mockSetup          func(m *mocks.CategoryUseCase)
		expectedStatusCode int
	}

	tests := []testCase{
		{
			name: "Gagal Sinkronisasi",
			mockSetup: func(m *mocks.CategoryUseCase) {
				m.On("SyncCategories", mock.Anything).
					Return(int64(0), errors.New("failed to sync")).Once()
			},
			expectedStatusCode: http.StatusInternalServerError,
		},
		{
			name: "Sukses Sinkronisasi",
			mockSetup: func(m *mocks.CategoryUseCase) {
				m.On("SyncCategories", mock.Anything).
					Return(int64(15), nil).Once() // Asumsi 15 kategori tersinkronisasi
			},
			expectedStatusCode: http.StatusOK,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockUC := new(mocks.CategoryUseCase)
			tc.mockSetup(mockUC)

			r := setupRouter(mockUC)

			req := httptest.NewRequest(http.MethodPost, "/categories/sync", nil)
			rec := httptest.NewRecorder()

			r.ServeHTTP(rec, req)

			assert.Equal(t, tc.expectedStatusCode, rec.Code)
			mockUC.AssertExpectations(t)
		})
	}
}
