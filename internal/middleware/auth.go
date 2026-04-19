// Package middleware untuk menyimpan semua middleware auth
package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/username/shop-api/internal/domain"
	"github.com/username/shop-api/internal/response"
	"github.com/username/shop-api/internal/util"
)

// AuthMiddleware memeriksa keberadaan dan kevalidan HttpOnly Cookie
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. Ambil token dari cookie bernama "auth_token"
		tokenString, err := c.Cookie("auth_token")
		if err != nil {
			response.ErrorUnauthorized(c, "Cookie auth_token tidak ditemukan. Silakan login.")
			c.Abort()
			return
		}

		// 2. Validasi token tersebut
		claims, err := util.ValidateToken(tokenString)
		if err != nil {
			response.ErrorUnauthorized(c, "Token tidak valid atau sudah kadaluarsa.")
			c.Abort()
			return
		}

		// 3. Jika valid, simpan User ID ke context agar bisa dibaca oleh Handler berikutnya
		c.Set("user_id", claims.UserID)
		c.Set("user_email", claims.Email)
		c.Set("role", domain.UserRole(claims.Role))

		// 4. Lanjut ke proses berikutnya
		c.Next()
	}
}
