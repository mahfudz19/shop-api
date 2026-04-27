// Package middleware test untuk Security Headers
package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestSecurityHeaders(t *testing.T) {
	// 1. Set Gin ke mode testing agar log tidak berisik
	gin.SetMode(gin.TestMode)

	// 2. Buat instance router Gin kosong dan pasang middleware
	r := gin.New()
	r.Use(SecurityHeaders())

	// 3. Buat rute dummy/palsu untuk mengetes middleware
	r.GET("/test-security", func(c *gin.Context) {
		c.String(http.StatusOK, "aman")
	})

	// 4. Siapkan request tiruan dan alat penangkap respons (Recorder)
	req, _ := http.NewRequest(http.MethodGet, "/test-security", nil)
	w := httptest.NewRecorder()

	// 5. Tembakkan request tiruan ke router
	r.ServeHTTP(w, req)

	// 6. Lakukan Pengecekan (Assertions)
	assert.Equal(t, http.StatusOK, w.Code, "Status code harus 200 OK")

	// Pastikan semua 4 tameng keamanan OWASP menempel di header
	assert.Equal(t, "DENY", w.Header().Get("X-Frame-Options"), "X-Frame-Options harus DENY")
	assert.Equal(t, "nosniff", w.Header().Get("X-Content-Type-Options"), "X-Content-Type-Options harus nosniff")
	assert.Equal(t, "1; mode=block", w.Header().Get("X-XSS-Protection"), "X-XSS-Protection harus aktif")
	assert.Equal(t, "max-age=31536000; includeSubDomains", w.Header().Get("Strict-Transport-Security"), "HSTS harus aktif selama 1 tahun")
}
