package middleware

import "github.com/gin-gonic/gin"

// SecurityHeaders menambahkan header ke HTTP response
func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Mencegah UI Anda dimuat dalam <iframe> (Cegah Clickjacking)
		c.Header("X-Frame-Options", "DENY")
		// Memaksa browser mengikuti tipe konten yang ditentukan (Cegah MIME Sniffing)
		c.Header("X-Content-Type-Options", "nosniff")
		// Perlindungan dasar terhadap serangan Cross-Site Scripting
		c.Header("X-XSS-Protection", "1; mode=block")
		// Memastikan koneksi hanya melalui HTTPS (HSTS)
		c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")

		c.Next()
	}
}
