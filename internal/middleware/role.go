// Package middleware untuk menyimpan semua middleware role
package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/username/shop-api/internal/domain"
)

// RequireRole memblokir akses jika user tidak memiliki role yang diizinkan
func RequireRole(allowedRoles ...domain.UserRole) gin.HandlerFunc {
	return func(c *gin.Context) {
		roleVal, exists := c.Get("role")
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "Sesi tidak valid, tidak ada data role",
			})
			return
		}

		// ✅ Ubah type assertion dari string ke domain.UserRole
		userRole, ok := roleVal.(domain.UserRole)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "Role type tidak valid",
			})
			return
		}

		isAllowed := false

		for _, r := range allowedRoles {
			if userRole == r {
				isAllowed = true
				break
			}
		}

		if !isAllowed {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"success": false,
				"message": "Akses ditolak: Anda tidak memiliki izin (bukan admin)",
			})
			return
		}

		c.Next()
	}
}
