// Package util = Helper untuk fungsi-fungsi umum seperti JWT handling
package util

import (
	"errors"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var jwtSecret = []byte(os.Getenv("JWT_SECRET"))

// JWTClaim struct untuk payload JWT
type JWTClaim struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
}

// GenerateToken membuat JWT baru yang valid selama 24 jam
func GenerateToken(userID string, email string) (string, error) {
	// Jika env belum terbaca, coba ambil lagi
	if len(jwtSecret) == 0 {
		jwtSecret = []byte(os.Getenv("JWT_SECRET"))
	}

	expirationTime := time.Now().Add(24 * time.Hour)
	claims := &JWTClaim{
		UserID: userID,
		Email:  email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

// ValidateToken memeriksa keaslian token
func ValidateToken(signedToken string) (*JWTClaim, error) {
	if len(jwtSecret) == 0 {
		jwtSecret = []byte(os.Getenv("JWT_SECRET"))
	}

	token, err := jwt.ParseWithClaims(
		signedToken,
		&JWTClaim{},
		func(_ *jwt.Token) (interface{}, error) {
			return jwtSecret, nil
		},
	)

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*JWTClaim)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}

	return claims, nil
}
