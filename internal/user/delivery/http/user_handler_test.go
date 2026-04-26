package http

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/username/shop-api/internal/domain"
	"github.com/username/shop-api/internal/mocks"
	"go.mongodb.org/mongo-driver/v2/bson"
)

// setupRouter adalah fungsi helper untuk menginisialisasi Gin Handler
func setupRouter(mockUC *mocks.UserUseCase) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	// 1. Setup Router Publik (Register, Login, dll)
	public := r.Group("/")

	// 2. Setup Router Protected dengan Dummy Middleware
	// Ini menyimulasikan AuthMiddleware yang berhasil memvalidasi JWT
	// dan menyuntikkan "user_id" ke dalam Context.
	protected := r.Group("/")
	protected.Use(func(c *gin.Context) {
		// Set ID palsu seolah-olah user sedang login
		c.Set("user_id", "user-auth-123")
		c.Next()
	})

	// 3. Setup Router Admin
	admin := r.Group("/")
	admin.Use(func(c *gin.Context) {
		c.Set("user_id", "admin-auth-123")
		c.Set("role", domain.RoleAdmin)
		c.Next()
	})

	NewUserHandler(public, protected, admin, mockUC)
	return r
}

func TestRegister(t *testing.T) {
	type testCase struct {
		name               string
		inputPayload       interface{}
		mockSetup          func(m *mocks.UserUseCase)
		expectedStatusCode int
	}

	tests := []testCase{
		{
			name:               "Gagal format JSON (Email tidak valid)",
			inputPayload:       RegisterRequest{Email: "bukan-email", Password: "password123"},
			mockSetup:          func(_ *mocks.UserUseCase) {},
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name: "Gagal dari validasi Usecase",
			inputPayload: RegisterRequest{
				Email:    "ada@test.com",
				Password: "password123",
				Name:     "User Ada",
			},
			mockSetup: func(m *mocks.UserUseCase) {
				m.On("Register", mock.Anything, mock.AnythingOfType("domain.User")).
					Return(errors.New("email already registered")).Once()
			},
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name: "Sukses Register",
			inputPayload: RegisterRequest{
				Email:    "baru@test.com",
				Password: "password123",
				Name:     "User Baru",
			},
			mockSetup: func(m *mocks.UserUseCase) {
				m.On("Register", mock.Anything, mock.AnythingOfType("domain.User")).
					Return(nil).Once()
			},
			expectedStatusCode: http.StatusCreated,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockUC := new(mocks.UserUseCase)
			tc.mockSetup(mockUC)

			r := setupRouter(mockUC)

			payload, _ := json.Marshal(tc.inputPayload)
			req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewBuffer(payload))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			r.ServeHTTP(rec, req)

			assert.Equal(t, tc.expectedStatusCode, rec.Code)
			mockUC.AssertExpectations(t)
		})
	}
}

func TestLogin(t *testing.T) {
	// Set Env Var palsu agar util.GenerateToken tidak error/panic saat di-test
	require.NoError(t, os.Setenv("JWT_SECRET", "test-secret-key"))

	type testCase struct {
		name               string
		inputPayload       interface{}
		mockSetup          func(m *mocks.UserUseCase)
		expectedStatusCode int
	}

	tests := []testCase{
		{
			name:               "Gagal Validasi JSON",
			inputPayload:       LoginRequest{Email: "bukan-email"},
			mockSetup:          func(_ *mocks.UserUseCase) {},
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name: "Gagal Login (Password Salah/User Tidak Ada)",
			inputPayload: LoginRequest{
				Email:    "salah@test.com",
				Password: "wrongpassword",
			},
			mockSetup: func(m *mocks.UserUseCase) {
				m.On("Login", mock.Anything, "salah@test.com", "wrongpassword").
					Return(domain.User{}, errors.New("invalid password")).Once()
			},
			expectedStatusCode: http.StatusUnauthorized,
		},
		{
			name: "Sukses Login",
			inputPayload: LoginRequest{
				Email:    "sukses@test.com",
				Password: "password123",
			},
			mockSetup: func(m *mocks.UserUseCase) {
				// Return user yang valid dengan ObjectID agar GenerateToken berhasil
				m.On("Login", mock.Anything, "sukses@test.com", "password123").
					Return(domain.User{
						ID:    bson.NewObjectID(),
						Email: "sukses@test.com",
						Role:  domain.RoleUser,
					}, nil).Once()
			},
			expectedStatusCode: http.StatusOK,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockUC := new(mocks.UserUseCase)
			tc.mockSetup(mockUC)

			r := setupRouter(mockUC)

			payload, _ := json.Marshal(tc.inputPayload)
			req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewBuffer(payload))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			r.ServeHTTP(rec, req)

			assert.Equal(t, tc.expectedStatusCode, rec.Code)

			// Jika sukses, pastikan Cookie auth_token ikut terkirim
			if tc.expectedStatusCode == http.StatusOK {
				cookies := rec.Result().Cookies()
				assert.NotEmpty(t, cookies)
				assert.Equal(t, "auth_token", cookies[0].Name)
			}
			mockUC.AssertExpectations(t)
		})
	}
}

func TestLogout(t *testing.T) {
	mockUC := new(mocks.UserUseCase)
	r := setupRouter(mockUC)

	req := httptest.NewRequest(http.MethodPost, "/auth/logout", nil)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	// Pastikan HTTP 200 OK
	assert.Equal(t, http.StatusOK, rec.Code)

	// Pastikan cookie dihapus (MaxAge negatif atau nilainya kosong)
	cookies := rec.Result().Cookies()
	assert.NotEmpty(t, cookies)
	assert.Equal(t, "auth_token", cookies[0].Name)
	assert.Equal(t, "", cookies[0].Value) // Nilai cookie harus di-reset menjadi kosong
}

func TestGetByID(t *testing.T) {
	type testCase struct {
		name               string
		paramID            string
		mockSetup          func(m *mocks.UserUseCase)
		expectedStatusCode int
	}

	tests := []testCase{
		{
			name:    "Gagal - User Tidak Ditemukan",
			paramID: "id-ngasal",
			mockSetup: func(m *mocks.UserUseCase) {
				m.On("GetUserByID", mock.Anything, "id-ngasal").
					Return(domain.User{}, errors.New("not found")).Once()
			},
			expectedStatusCode: http.StatusNotFound,
		},
		{
			name:    "Sukses Ditemukan",
			paramID: "id-valid",
			mockSetup: func(m *mocks.UserUseCase) {
				m.On("GetUserByID", mock.Anything, "id-valid").
					Return(domain.User{Name: "Pencari"}, nil).Once()
			},
			expectedStatusCode: http.StatusOK,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockUC := new(mocks.UserUseCase)
			tc.mockSetup(mockUC)

			r := setupRouter(mockUC)

			req := httptest.NewRequest(http.MethodGet, "/users/"+tc.paramID, nil)
			rec := httptest.NewRecorder()

			r.ServeHTTP(rec, req)

			assert.Equal(t, tc.expectedStatusCode, rec.Code)
			mockUC.AssertExpectations(t)
		})
	}
}

func TestGetMyProfile(t *testing.T) {
	type testCase struct {
		name               string
		mockSetup          func(m *mocks.UserUseCase)
		expectedStatusCode int
	}

	tests := []testCase{
		{
			name: "Gagal (User di DB terhapus meski token valid)",
			mockSetup: func(m *mocks.UserUseCase) {
				// Perhatikan bahwa "user-auth-123" berasal dari Dummy Middleware di fungsi setupRouter
				m.On("GetUserByID", mock.Anything, "user-auth-123").
					Return(domain.User{}, errors.New("user not found")).Once()
			},
			expectedStatusCode: http.StatusNotFound,
		},
		{
			name: "Sukses Ambil Profil Sendiri",
			mockSetup: func(m *mocks.UserUseCase) {
				m.On("GetUserByID", mock.Anything, "user-auth-123").
					Return(domain.User{Name: "Saya Sendiri"}, nil).Once()
			},
			expectedStatusCode: http.StatusOK,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockUC := new(mocks.UserUseCase)
			tc.mockSetup(mockUC)

			r := setupRouter(mockUC)

			// Tembak endpoint Rute Terlindungi
			req := httptest.NewRequest(http.MethodGet, "/auth/my", nil)
			rec := httptest.NewRecorder()

			r.ServeHTTP(rec, req)

			assert.Equal(t, tc.expectedStatusCode, rec.Code)
			mockUC.AssertExpectations(t)
		})
	}
}
