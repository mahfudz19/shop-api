package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/username/shop-api/internal/domain"
	"github.com/username/shop-api/internal/middleware"
)

func TestRequireRole(t *testing.T) {
	gin.SetMode(gin.TestMode)

	type testCase struct {
		name               string
		injectContext      func(c *gin.Context)
		allowedRoles       []domain.UserRole
		expectedStatusCode int
	}

	tests := []testCase{
		{
			name: "Gagal - Context Kosong (Sesi Tidak Valid)",
			injectContext: func(_ *gin.Context) {
				// Tidak menyuntikkan apapun
			},
			allowedRoles:       []domain.UserRole{domain.RoleAdmin},
			expectedStatusCode: http.StatusUnauthorized,
		},
		{
			name: "Gagal - Tipe Data Role Salah (Bukan UserRole)",
			injectContext: func(c *gin.Context) {
				c.Set("role", "admin") // String biasa, bukan domain.UserRole
			},
			allowedRoles:       []domain.UserRole{domain.RoleAdmin},
			expectedStatusCode: http.StatusUnauthorized,
		},
		{
			name: "Gagal - Akses Ditolak (User mencoba akses Admin area)",
			injectContext: func(c *gin.Context) {
				c.Set("role", domain.RoleUser)
			},
			allowedRoles:       []domain.UserRole{domain.RoleAdmin},
			expectedStatusCode: http.StatusForbidden,
		},
		{
			name: "Sukses - Role Cocok (Admin akses Admin area)",
			injectContext: func(c *gin.Context) {
				c.Set("role", domain.RoleAdmin)
			},
			allowedRoles:       []domain.UserRole{domain.RoleAdmin},
			expectedStatusCode: http.StatusOK,
		},
		{
			name: "Sukses - Role Cocok (User akses area umum/admin)",
			injectContext: func(c *gin.Context) {
				c.Set("role", domain.RoleUser)
			},
			allowedRoles:       []domain.UserRole{domain.RoleAdmin, domain.RoleUser},
			expectedStatusCode: http.StatusOK,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			r := gin.New()

			// 1. Injector Middleware: Mensimulasikan AuthMiddleware yang berhasil
			r.Use(func(c *gin.Context) {
				tc.injectContext(c)
				c.Next()
			})

			// 2. Middleware yang sedang dites
			r.Use(middleware.RequireRole(tc.allowedRoles...))

			r.GET("/protected", func(c *gin.Context) {
				c.Status(http.StatusOK)
			})

			req := httptest.NewRequest(http.MethodGet, "/protected", nil)
			rec := httptest.NewRecorder()

			r.ServeHTTP(rec, req)

			assert.Equal(t, tc.expectedStatusCode, rec.Code)
		})
	}
}
