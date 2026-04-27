// Package middleware untuk menyimpan semua middleware role
package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/username/shop-api/internal/domain"
	"github.com/username/shop-api/internal/response"
)

// RequireRole memblokir akses jika user tidak memiliki role yang diizinkan
func RequireRole(allowedRoles ...domain.UserRole) gin.HandlerFunc {
	return func(c *gin.Context) {
		roleVal, exists := c.Get("role")
		if !exists {
			response.ErrorUnauthorized(c, "Sesi tidak valid, tidak ada data role")
			c.Abort()
			return
		}

		// ✅ Ubah type assertion dari string ke domain.UserRole
		userRole, ok := roleVal.(domain.UserRole)
		if !ok {
			response.ErrorUnauthorized(c, "Role type tidak valid")
			c.Abort()
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
			response.ErrorForbidden(c, "Akses ditolak: Anda tidak memiliki izin (bukan admin)")
			c.Abort()
			return
		}

		c.Next()
	}
}
