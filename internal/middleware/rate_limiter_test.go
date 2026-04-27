// Package middleware test untuk Rate Limiter middleware
package middleware

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestRateLimiter(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Reset rate limiter sebelum test
	rateLimiterConfig = nil

	// Set environment variable untuk testing
	_ = os.Setenv("RATE_LIMIT_PER_MINUTE", "5") // 5 request per menit untuk testing
	defer func() {
		_ = os.Unsetenv("RATE_LIMIT_PER_MINUTE")
	}()

	t.Run("Should allow requests within limit", func(t *testing.T) {
		// Reset limiter untuk test ini
		rateLimiterConfig = nil

		for i := 0; i < 5; i++ {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			req, _ := http.NewRequest(http.MethodGet, "/test", nil)
			req.Header.Set("X-Forwarded-For", "192.168.1.100")
			c.Request = req

			RateLimiter()(c)

			assert.Equal(t, http.StatusOK, w.Code, "Request ke-%d seharusnya diizinkan", i+1)
		}
	})

	t.Run("Should reject requests exceeding limit", func(t *testing.T) {
		// Reset limiter untuk test ini
		rateLimiterConfig = nil

		// Lakukan 5 request dulu (should pass)
		for i := 0; i < 5; i++ {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			req, _ := http.NewRequest(http.MethodGet, "/test", nil)
			req.Header.Set("X-Forwarded-For", "192.168.1.101")
			c.Request = req

			RateLimiter()(c)

			assert.Equal(t, http.StatusOK, w.Code)
		}

		// Request ke-6 seharusnya ditolak
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		req, _ := http.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("X-Forwarded-For", "192.168.1.101")
		c.Request = req

		RateLimiter()(c)

		assert.Equal(t, http.StatusTooManyRequests, w.Code, "Request melebihi limit seharusnya ditolak")
	})

	t.Run("Should handle different IPs separately", func(t *testing.T) {
		// Reset limiter untuk test ini
		rateLimiterConfig = nil

		// IP 1 melakukan 5 request (gunakan RemoteAddr yang berbeda)
		for i := 0; i < 5; i++ {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			req, _ := http.NewRequest(http.MethodGet, "/test", nil)
			// Set RemoteAddr langsung karena gin.ClientIP() membaca dari sini
			req.RemoteAddr = "192.168.1.200:12345"
			c.Request = req

			RateLimiter()(c)

			assert.Equal(t, http.StatusOK, w.Code)
		}

		// IP 2 seharusnya masih bisa melakukan request
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		req, _ := http.NewRequest(http.MethodGet, "/test", nil)
		req.RemoteAddr = "192.168.1.201:12345"
		c.Request = req

		RateLimiter()(c)

		assert.Equal(t, http.StatusOK, w.Code, "IP berbeda seharusnya memiliki limit terpisah")
	})

	t.Run("Should use default limit when env is not set", func(t *testing.T) {
		// Reset limiter untuk test ini
		rateLimiterConfig = nil
		_ = os.Unsetenv("RATE_LIMIT_PER_MINUTE")

		// 100 request seharusnya masih diizinkan (default limit)
		for i := 0; i < 50; i++ { // Test dengan 50 request saja untuk efisiensi
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			req, _ := http.NewRequest(http.MethodGet, "/test", nil)
			req.Header.Set("X-Forwarded-For", "192.168.1.250")
			c.Request = req

			RateLimiter()(c)

			assert.Equal(t, http.StatusOK, w.Code)
		}
	})
}
