// Package middleware untuk menyimpan semua middleware auth
package middleware

import (
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/username/shop-api/internal/response"
	"golang.org/x/time/rate"
)

// RateLimiterConfig konfigurasi untuk rate limiter per IP
type RateLimiterConfig struct {
	limiters map[string]*rate.Limiter
	mu       *sync.RWMutex
	rate     rate.Limit
	burst    int
}

// Global rate limiter instance
var rateLimiterConfig *RateLimiterConfig

// initRateLimiter inisialisasi konfigurasi rate limiter
func initRateLimiter() {
	// Baca limit dari environment variable, default 100 request per menit
	limitStr := os.Getenv("RATE_LIMIT_PER_MINUTE")
	limitPerMinute, err := strconv.Atoi(limitStr)
	if err != nil || limitPerMinute <= 0 {
		limitPerMinute = 100 // Default fallback
	}

	rateLimiterConfig = &RateLimiterConfig{
		limiters: make(map[string]*rate.Limiter),
		mu:       &sync.RWMutex{},
		rate:     rate.Limit(float64(limitPerMinute) / 60.0),
		burst:    limitPerMinute,
	}

	// Cleanup limiter yang tidak aktif setiap 10 menit
	go cleanupUnusedLimiters()
}

// getLimiter mendapatkan atau membuat rate limiter baru untuk IP tertentu
func getLimiter(ip string) *rate.Limiter {
	// Coba baca dengan read lock dulu
	rateLimiterConfig.mu.RLock()
	limiter, exists := rateLimiterConfig.limiters[ip]
	rateLimiterConfig.mu.RUnlock()

	if exists {
		return limiter
	}

	// Jika tidak ada, buat limiter baru dengan write lock
	rateLimiterConfig.mu.Lock()
	defer rateLimiterConfig.mu.Unlock()

	// Double-check setelah dapat write lock
	if limiter, exists = rateLimiterConfig.limiters[ip]; exists {
		return limiter
	}

	limiter = rate.NewLimiter(rateLimiterConfig.rate, rateLimiterConfig.burst)
	rateLimiterConfig.limiters[ip] = limiter

	return limiter
}

// cleanupUnusedLimiters membersihkan limiter yang sudah lama tidak digunakan
func cleanupUnusedLimiters() {
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		rateLimiterConfig.mu.Lock()
		for ip, limiter := range rateLimiterConfig.limiters {
			// Jika limiter sudah tidak memiliki token yang tersedia (sudah lama tidak dipakai)
			// dan sudah lebih dari 10 menit, hapus
			if limiter.Allow() {
				// Reset limiter jika masih aktif
				limiter.Allow()
			} else {
				delete(rateLimiterConfig.limiters, ip)
			}
		}
		rateLimiterConfig.mu.Unlock()
	}
}

// RateLimiter middleware untuk membatasi request per IP address
func RateLimiter() gin.HandlerFunc {
	// Inisialisasi jika belum
	if rateLimiterConfig == nil {
		initRateLimiter()
	}

	return func(c *gin.Context) {
		// Dapatkan IP address client
		clientIP := c.ClientIP()

		// Dapatkan limiter untuk IP ini
		limiter := getLimiter(clientIP)

		// Cek apakah request diizinkan
		if !limiter.Allow() {
			response.ErrorTooManyRequests(c, "Terlalu banyak permintaan, silakan coba lagi nanti")
			c.Abort()
			return
		}

		// Lanjutkan request
		c.Next()
	}
}
