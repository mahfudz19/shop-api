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

// setupRouter adalah fungsi bantuan (helper) untuk menginisialisasi Gin dan Handler secara konsisten
func setupRouter(mockUC *mocks.PromotionUseCase) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	NewPromotionHandler(r, r, mockUC)
	return r
}

func TestCreate(t *testing.T) {
	type testCase struct {
		name               string
		inputPayload       interface{}
		mockSetup          func(m *mocks.PromotionUseCase)
		expectedStatusCode int
	}

	tests := []testCase{
		{
			name:         "Gagal format JSON",
			inputPayload: "invalid-json",
			mockSetup:    func(_ *mocks.PromotionUseCase) {},
			// Ubah ekspektasi menjadi 422 sesuai standar respons validasi Anda
			expectedStatusCode: http.StatusUnprocessableEntity,
		},
		{
			name: "Gagal dari Usecase",
			inputPayload: domain.PromotionRequest{
				Title:    "Promo Error",
				ImageURL: "http://image.com/error.png",
			},
			mockSetup: func(m *mocks.PromotionUseCase) {
				m.On("Create", mock.Anything, mock.AnythingOfType("domain.PromotionRequest")).
					Return(domain.Promotion{}, errors.New("usecase error")).Once()
			},
			expectedStatusCode: http.StatusInternalServerError,
		},
		{
			name: "Sukses Create",
			inputPayload: domain.PromotionRequest{
				Title:    "Promo Berhasil",
				ImageURL: "http://image.com/sukses.png",
			},
			mockSetup: func(m *mocks.PromotionUseCase) {
				m.On("Create", mock.Anything, mock.AnythingOfType("domain.PromotionRequest")).
					Return(domain.Promotion{Title: "Promo Berhasil"}, nil).Once()
			},
			expectedStatusCode: http.StatusCreated,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockUC := new(mocks.PromotionUseCase)
			tc.mockSetup(mockUC)

			// Setup otomatis menggunakan fungsi helper
			r := setupRouter(mockUC)

			var payload []byte
			if strPayload, ok := tc.inputPayload.(string); ok {
				payload = []byte(strPayload)
			} else {
				payload, _ = json.Marshal(tc.inputPayload)
			}

			req := httptest.NewRequest(http.MethodPost, "/promotions", bytes.NewBuffer(payload))
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
		queryParam         string // Misalnya: "?active=true"
		mockSetup          func(m *mocks.PromotionUseCase)
		expectedStatusCode int
	}

	tests := []testCase{
		{
			name:       "Sukses Get All Aktif",
			queryParam: "?active=true",
			mockSetup: func(m *mocks.PromotionUseCase) {
				m.On("GetAll", mock.Anything, true).
					Return([]domain.Promotion{}, nil).Once()
			},
			expectedStatusCode: http.StatusOK,
		},
		{
			name:       "Gagal Get All dari Usecase",
			queryParam: "",
			mockSetup: func(m *mocks.PromotionUseCase) {
				m.On("GetAll", mock.Anything, false).
					Return(nil, errors.New("db crash")).Once()
			},
			expectedStatusCode: http.StatusInternalServerError,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockUC := new(mocks.PromotionUseCase)
			tc.mockSetup(mockUC)

			r := setupRouter(mockUC)

			req := httptest.NewRequest(http.MethodGet, "/promotions"+tc.queryParam, nil)
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
		mockSetup          func(m *mocks.PromotionUseCase)
		expectedStatusCode int
	}

	tests := []testCase{
		{
			name:    "Gagal - Promotion Tidak Ditemukan",
			paramID: "non-existent-id",
			mockSetup: func(m *mocks.PromotionUseCase) {
				m.On("GetByID", mock.Anything, "non-existent-id").
					Return(domain.Promotion{}, errors.New("not found")).Once()
			},
			expectedStatusCode: http.StatusNotFound,
		},
		{
			name:    "Sukses Get By ID",
			paramID: "valid-id",
			mockSetup: func(m *mocks.PromotionUseCase) {
				m.On("GetByID", mock.Anything, "valid-id").
					Return(domain.Promotion{Title: "Promo Valid"}, nil).Once()
			},
			expectedStatusCode: http.StatusOK,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockUC := new(mocks.PromotionUseCase)
			tc.mockSetup(mockUC)

			r := setupRouter(mockUC)

			req := httptest.NewRequest(http.MethodGet, "/promotions/"+tc.paramID, nil)
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
		mockSetup          func(m *mocks.PromotionUseCase)
		expectedStatusCode int
	}

	tests := []testCase{
		{
			name:               "Gagal format JSON",
			paramID:            "valid-id",
			inputPayload:       "invalid-json",
			mockSetup:          func(_ *mocks.PromotionUseCase) {},
			expectedStatusCode: http.StatusUnprocessableEntity,
		},
		{
			name:    "Gagal karena Promotion tidak ada di DB",
			paramID: "non-existent-id",
			inputPayload: domain.PromotionRequest{
				Title:    "Update Title",
				ImageURL: "http://image.com/update.png",
			},
			mockSetup: func(m *mocks.PromotionUseCase) {
				m.On("Update", mock.Anything, "non-existent-id", mock.AnythingOfType("domain.PromotionRequest")).
					Return(domain.Promotion{}, errors.New("not found")).Once()
			},
			expectedStatusCode: http.StatusInternalServerError,
		},
		{
			name:    "Sukses Update",
			paramID: "valid-id",
			inputPayload: domain.PromotionRequest{
				Title:    "Updated Title",
				ImageURL: "http://image.com/update.png",
			},
			mockSetup: func(m *mocks.PromotionUseCase) {
				m.On("Update", mock.Anything, "valid-id", mock.AnythingOfType("domain.PromotionRequest")).
					Return(domain.Promotion{Title: "Updated Title"}, nil).Once()
			},
			expectedStatusCode: http.StatusOK,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockUC := new(mocks.PromotionUseCase)
			tc.mockSetup(mockUC)

			r := setupRouter(mockUC)

			var payload []byte
			if strPayload, ok := tc.inputPayload.(string); ok {
				payload = []byte(strPayload)
			} else {
				payload, _ = json.Marshal(tc.inputPayload)
			}

			req := httptest.NewRequest(http.MethodPut, "/promotions/"+tc.paramID, bytes.NewBuffer(payload))
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
		mockSetup          func(m *mocks.PromotionUseCase)
		expectedStatusCode int
	}

	tests := []testCase{
		{
			name:    "Gagal Delete karena DB Error/Tidak Ada",
			paramID: "error-id",
			mockSetup: func(m *mocks.PromotionUseCase) {
				m.On("Delete", mock.Anything, "error-id").
					Return(errors.New("failed to delete")).Once()
			},
			expectedStatusCode: http.StatusInternalServerError,
		},
		{
			name:    "Sukses Delete",
			paramID: "valid-id",
			mockSetup: func(m *mocks.PromotionUseCase) {
				m.On("Delete", mock.Anything, "valid-id").
					Return(nil).Once()
			},
			expectedStatusCode: http.StatusOK,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockUC := new(mocks.PromotionUseCase)
			tc.mockSetup(mockUC)

			r := setupRouter(mockUC)

			req := httptest.NewRequest(http.MethodDelete, "/promotions/"+tc.paramID, nil)
			rec := httptest.NewRecorder()

			r.ServeHTTP(rec, req)

			assert.Equal(t, tc.expectedStatusCode, rec.Code)
			mockUC.AssertExpectations(t)
		})
	}
}
