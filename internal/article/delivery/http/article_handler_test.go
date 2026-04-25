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
func setupRouter(mockUC *mocks.ArticleUseCase) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	// Kita passing 'r' (router) untuk public dan admin router
	NewArticleHandler(r, r, mockUC)
	return r
}

func TestCreate(t *testing.T) {
	type testCase struct {
		name               string
		inputPayload       interface{}
		mockSetup          func(m *mocks.ArticleUseCase)
		expectedStatusCode int
	}

	tests := []testCase{
		{
			name:               "Gagal format JSON (Invalid)",
			inputPayload:       "invalid-json",
			mockSetup:          func(_ *mocks.ArticleUseCase) {},
			expectedStatusCode: http.StatusUnprocessableEntity, // 422 karena ShouldBindJSON gagal
		},
		{
			name: "Gagal dari Usecase (DB Error)",
			inputPayload: domain.ArticleRequest{
				Title:   "Judul Error",
				Slug:    "judul-error",
				Content: "Isi konten",
				Author:  "Penulis",
			},
			mockSetup: func(m *mocks.ArticleUseCase) {
				m.On("Create", mock.Anything, mock.AnythingOfType("domain.ArticleRequest")).
					Return(domain.Article{}, errors.New("usecase error")).Once()
			},
			expectedStatusCode: http.StatusInternalServerError,
		},
		{
			name: "Sukses Create",
			inputPayload: domain.ArticleRequest{
				Title:   "Judul Sukses",
				Slug:    "judul-sukses",
				Content: "Isi konten artikel yang valid",
				Author:  "Penulis",
			},
			mockSetup: func(m *mocks.ArticleUseCase) {
				m.On("Create", mock.Anything, mock.AnythingOfType("domain.ArticleRequest")).
					Return(domain.Article{Title: "Judul Sukses"}, nil).Once()
			},
			expectedStatusCode: http.StatusCreated,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockUC := new(mocks.ArticleUseCase)
			tc.mockSetup(mockUC)

			r := setupRouter(mockUC)

			var payload []byte
			if strPayload, ok := tc.inputPayload.(string); ok {
				payload = []byte(strPayload)
			} else {
				payload, _ = json.Marshal(tc.inputPayload)
			}

			req := httptest.NewRequest(http.MethodPost, "/articles", bytes.NewBuffer(payload))
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
		queryParam         string
		mockSetup          func(m *mocks.ArticleUseCase)
		expectedStatusCode int
	}

	tests := []testCase{
		{
			name:       "Sukses Get All - Hanya Published",
			queryParam: "?published=true", // Query parameter dikirim
			mockSetup: func(m *mocks.ArticleUseCase) {
				// Pastikan usecase dipanggil dengan param `onlyPublished` bernilai TRUE
				m.On("GetAll", mock.Anything, true).
					Return([]domain.Article{}, nil).Once()
			},
			expectedStatusCode: http.StatusOK,
		},
		{
			name:       "Sukses Get All - Semua Artikel (Termasuk Draft)",
			queryParam: "", // Tanpa query parameter
			mockSetup: func(m *mocks.ArticleUseCase) {
				// Pastikan usecase dipanggil dengan param `onlyPublished` bernilai FALSE
				m.On("GetAll", mock.Anything, false).
					Return([]domain.Article{}, nil).Once()
			},
			expectedStatusCode: http.StatusOK,
		},
		{
			name:       "Gagal Get All dari Usecase",
			queryParam: "",
			mockSetup: func(m *mocks.ArticleUseCase) {
				m.On("GetAll", mock.Anything, false).
					Return(nil, errors.New("db crash")).Once()
			},
			expectedStatusCode: http.StatusInternalServerError,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockUC := new(mocks.ArticleUseCase)
			tc.mockSetup(mockUC)

			r := setupRouter(mockUC)

			req := httptest.NewRequest(http.MethodGet, "/articles"+tc.queryParam, nil)
			rec := httptest.NewRecorder()

			r.ServeHTTP(rec, req)

			assert.Equal(t, tc.expectedStatusCode, rec.Code)
			mockUC.AssertExpectations(t)
		})
	}
}

func TestGetBySlug(t *testing.T) {
	type testCase struct {
		name               string
		paramSlug          string
		mockSetup          func(m *mocks.ArticleUseCase)
		expectedStatusCode int
	}

	tests := []testCase{
		{
			name:      "Gagal - Article Tidak Ditemukan",
			paramSlug: "slug-ngasal",
			mockSetup: func(m *mocks.ArticleUseCase) {
				m.On("GetBySlug", mock.Anything, "slug-ngasal").
					Return(domain.Article{}, errors.New("not found")).Once()
			},
			expectedStatusCode: http.StatusNotFound, // 404
		},
		{
			name:      "Sukses Get By Slug",
			paramSlug: "5-tenda-camping-terbaik",
			mockSetup: func(m *mocks.ArticleUseCase) {
				m.On("GetBySlug", mock.Anything, "5-tenda-camping-terbaik").
					Return(domain.Article{Title: "5 Tenda Camping"}, nil).Once()
			},
			expectedStatusCode: http.StatusOK, // 200
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockUC := new(mocks.ArticleUseCase)
			tc.mockSetup(mockUC)

			r := setupRouter(mockUC)

			// Tembak URL khusus Slug
			req := httptest.NewRequest(http.MethodGet, "/articles/slug/"+tc.paramSlug, nil)
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
		mockSetup          func(m *mocks.ArticleUseCase)
		expectedStatusCode int
	}

	tests := []testCase{
		{
			name:               "Gagal format JSON",
			paramID:            "valid-id",
			inputPayload:       "invalid-json",
			mockSetup:          func(_ *mocks.ArticleUseCase) {},
			expectedStatusCode: http.StatusUnprocessableEntity, // 422
		},
		{
			name:    "Gagal karena Artikel tidak ada di DB",
			paramID: "non-existent-id",
			inputPayload: domain.ArticleRequest{
				Title:   "Update Judul",
				Slug:    "update-judul",
				Content: "Update Konten",
				Author:  "Penulis",
			},
			mockSetup: func(m *mocks.ArticleUseCase) {
				m.On("Update", mock.Anything, "non-existent-id", mock.AnythingOfType("domain.ArticleRequest")).
					Return(domain.Article{}, errors.New("not found")).Once()
			},
			expectedStatusCode: http.StatusInternalServerError,
		},
		{
			name:    "Sukses Update",
			paramID: "valid-id",
			inputPayload: domain.ArticleRequest{
				Title:   "Update Judul",
				Slug:    "update-judul",
				Content: "Update Konten",
				Author:  "Penulis",
			},
			mockSetup: func(m *mocks.ArticleUseCase) {
				m.On("Update", mock.Anything, "valid-id", mock.AnythingOfType("domain.ArticleRequest")).
					Return(domain.Article{Title: "Update Judul"}, nil).Once()
			},
			expectedStatusCode: http.StatusOK,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockUC := new(mocks.ArticleUseCase)
			tc.mockSetup(mockUC)

			r := setupRouter(mockUC)

			var payload []byte
			if strPayload, ok := tc.inputPayload.(string); ok {
				payload = []byte(strPayload)
			} else {
				payload, _ = json.Marshal(tc.inputPayload)
			}

			req := httptest.NewRequest(http.MethodPut, "/articles/"+tc.paramID, bytes.NewBuffer(payload))
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
		mockSetup          func(m *mocks.ArticleUseCase)
		expectedStatusCode int
	}

	tests := []testCase{
		{
			name:    "Gagal Delete karena DB Error/Tidak Ada",
			paramID: "error-id",
			mockSetup: func(m *mocks.ArticleUseCase) {
				m.On("Delete", mock.Anything, "error-id").
					Return(errors.New("failed to delete")).Once()
			},
			expectedStatusCode: http.StatusInternalServerError,
		},
		{
			name:    "Sukses Delete",
			paramID: "valid-id",
			mockSetup: func(m *mocks.ArticleUseCase) {
				m.On("Delete", mock.Anything, "valid-id").
					Return(nil).Once()
			},
			expectedStatusCode: http.StatusOK,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockUC := new(mocks.ArticleUseCase)
			tc.mockSetup(mockUC)

			r := setupRouter(mockUC)

			req := httptest.NewRequest(http.MethodDelete, "/articles/"+tc.paramID, nil)
			rec := httptest.NewRecorder()

			r.ServeHTTP(rec, req)

			assert.Equal(t, tc.expectedStatusCode, rec.Code)
			mockUC.AssertExpectations(t)
		})
	}
}
