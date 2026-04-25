package util_test

import (
	"os"
	"testing"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/username/shop-api/internal/domain"
	"github.com/username/shop-api/internal/util"
)

func TestJWT(t *testing.T) {
	// 1. STATE ISOLATION: Paksa set environment variable khusus untuk testing
	require.NoError(t, os.Setenv("JWT_SECRET", "test-secret-key"))

	userID := "user-id-123"
	email := "hacker@baik.com"
	role := domain.RoleAdmin

	var generatedToken string

	t.Run("1. Sukses Generate Token", func(t *testing.T) {
		token, err := util.GenerateToken(userID, email, role)

		assert.NoError(t, err)
		assert.NotEmpty(t, token, "Token tidak boleh kosong")

		generatedToken = token
	})

	t.Run("2. Sukses Validasi Token yang Sah", func(t *testing.T) {
		claims, err := util.ValidateToken(generatedToken)

		assert.NoError(t, err)
		assert.NotNil(t, claims)

		assert.Equal(t, userID, claims.UserID)
		assert.Equal(t, email, claims.Email)
		assert.Equal(t, role, claims.Role)
	})

	t.Run("3. Gagal Validasi Token Palsu/Rusak", func(t *testing.T) {
		fakeToken := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.palsu.sekali"

		claims, err := util.ValidateToken(fakeToken)

		assert.Error(t, err)
		assert.Nil(t, claims, "Claims harus nil jika token tidak valid")
	})

	t.Run("4. Gagal Validasi Token dengan Signature Berbeda", func(t *testing.T) {
		// Skenario Hacker: Membuat token dengan struktur valid, tapi ditandatangani dengan kunci rahasia milik hacker
		rogueTokenObj := jwt.NewWithClaims(jwt.SigningMethodHS256, &util.JWTClaim{
			UserID: "hacker-123",
			Email:  "hacker@jahat.com",
			Role:   domain.RoleUser,
		})
		rogueSignedToken, _ := rogueTokenObj.SignedString([]byte("kunci-rahasia-hacker"))

		// Server mencoba memvalidasi token yang disusupkan hacker
		claims, err := util.ValidateToken(rogueSignedToken)

		// HARUS ERROR: Karena "test-secret-key" milik server tidak cocok dengan "kunci-rahasia-hacker"
		assert.Error(t, err)
		assert.Nil(t, claims)
	})
}
