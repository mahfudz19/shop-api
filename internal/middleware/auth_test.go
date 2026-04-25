package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/username/shop-api/internal/domain"
	"github.com/username/shop-api/internal/middleware"
	"github.com/username/shop-api/internal/util"
)

func TestAuthMiddleware(t *testing.T) {
	// Setup environment untuk JWT
	require.NoError(t, os.Setenv("JWT_SECRET", "test-secret-key"))
	gin.SetMode(gin.TestMode)

	// Buat token sah untuk testing
	validToken, _ := util.GenerateToken("user-123", "test@test.com", domain.RoleUser)

	type testCase struct {
		name               string
		setupCookie        func(req *http.Request)
		expectedStatusCode int
		checkContext       bool
	}

	tests := []testCase{
		{
			name: "Gagal - Tanpa Cookie",
			setupCookie: func(_ *http.Request) {
				// Tidak melakukan apa-apa
			},
			expectedStatusCode: http.StatusUnauthorized,
		},
		{
			name: "Gagal - Token Palsu/Rusak",
			setupCookie: func(req *http.Request) {
				req.AddCookie(&http.Cookie{Name: "auth_token", Value: "token-ngasal"})
			},
			expectedStatusCode: http.StatusUnauthorized,
		},
		{
			name: "Sukses - Token Sah",
			setupCookie: func(req *http.Request) {
				req.AddCookie(&http.Cookie{Name: "auth_token", Value: validToken})
			},
			expectedStatusCode: http.StatusOK,
			checkContext:       true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			r := gin.New()

			// Pasang middleware yang sedang dites
			r.Use(middleware.AuthMiddleware())

			// Dummy Handler untuk verifikasi
			r.GET("/test", func(c *gin.Context) {
				if tc.checkContext {
					// Pastikan middleware menyuntikkan data ke context
					uid, _ := c.Get("user_id")
					role, _ := c.Get("role")
					assert.Equal(t, "user-123", uid)
					assert.Equal(t, domain.RoleUser, role)
				}
				c.Status(http.StatusOK)
			})

			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			tc.setupCookie(req)
			rec := httptest.NewRecorder()

			r.ServeHTTP(rec, req)

			assert.Equal(t, tc.expectedStatusCode, rec.Code)
		})
	}
}
