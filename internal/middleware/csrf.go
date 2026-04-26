// Package middleware untuk menyimpan semua middleware auth
package middleware

import (
	"slices"

	"github.com/gin-gonic/gin"
	"github.com/username/shop-api/internal/response"
)

// CSRFProtection middleware untuk melindungi dari serangan CSRF
// Middleware ini akan mengecek header X-Requested-With atau X-CSRF-Token
// untuk request dengan method yang berpotensi mengubah data (POST, PUT, DELETE, PATCH)
func CSRFProtection() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Daftar method yang aman (tidak perlu CSRF check)
		safeMethods := []string{"GET", "HEAD", "OPTIONS", "TRACE"}

		// Jika method aman, biarkan lewat
		if slices.Contains(safeMethods, c.Request.Method) {
			c.Next()
			return
		}

		// Untuk method yang tidak aman (POST, PUT, DELETE, PATCH),
		// cek keberadaan header X-Requested-With atau X-CSRF-Token
		xRequestedWith := c.GetHeader("X-Requested-With")
		xCsrfToken := c.GetHeader("X-CSRF-Token")

		if xRequestedWith == "" && xCsrfToken == "" {
			response.ErrorForbidden(c, "Permintaan ditolak: Potensi CSRF")
			c.Abort()
			return
		}

		// Header ada, lanjutkan request
		c.Next()
	}
}
