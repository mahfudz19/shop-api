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
	"github.com/username/shop-api/internal/mocks"
)

func TestRegister(t *testing.T) {
	// Set Gin ke TestMode agar tidak memenuhi terminal dengan log saat testing
	gin.SetMode(gin.TestMode)

	type testCase struct {
		name               string
		inputPayload       interface{}
		mockSetup          func(m *mocks.UserUseCase)
		expectedStatusCode int
	}

	tests := []testCase{
		{
			name:               "Gagal karena format JSON tidak valid",
			inputPayload:       "payload-acak-bukan-json",
			mockSetup:          func(_ *mocks.UserUseCase) {}, // Tidak memanggil Usecase
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name: "Gagal dari validasi Usecase",
			inputPayload: RegisterRequest{ // Menggunakan struct yang ada di handler Anda
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
			// 1. Inisialisasi Mock Usecase
			mockUC := new(mocks.UserUseCase)
			tc.mockSetup(mockUC)

			// 2. Setup Gin Router untuk Testing
			r := gin.New()
			// Inject mock langsung ke struct handler (bisa dilakukan karena satu package)
			handler := &UserHandler{usecase: mockUC}
			r.POST("/register", handler.Register)

			// 3. Siapkan Request Body
			var payload []byte
			if strPayload, ok := tc.inputPayload.(string); ok {
				payload = []byte(strPayload)
			} else {
				payload, _ = json.Marshal(tc.inputPayload)
			}

			// 4. Buat Request HTTP bohongan
			req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(payload))
			req.Header.Set("Content-Type", "application/json")

			// Buat penangkap respons (Recorder)
			rec := httptest.NewRecorder()

			// 5. Eksekusi Request melalui router Gin
			r.ServeHTTP(rec, req)

			// 6. Pengecekan Hasil
			assert.Equal(t, tc.expectedStatusCode, rec.Code)

			mockUC.AssertExpectations(t)
		})
	}
}
