// Package auth creates and checks JWT tokens.
package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"siri_ice_cream/internal/models"
)

// TokenLifetime is how long a login token stays valid.
const TokenLifetime = 24 * time.Hour

// Claims is the information stored inside a JWT token.
type Claims struct {
	UserID uint        `json:"user_id"`
	Role   models.Role `json:"role"`
	jwt.RegisteredClaims
}

// GenerateToken creates a signed JWT token for the given user.
func GenerateToken(user models.User, secret string) (string, error) {
	now := time.Now()
	claims := Claims{
		UserID: user.ID,
		Role:   user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   user.Username,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(TokenLifetime)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

// ParseToken checks the token's signature and expiry and returns its claims.
func ParseToken(tokenString, secret string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
		// Only accept the HMAC algorithm we used to sign the token.
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(secret), nil
	})
	if err != nil || !token.Valid {
		return nil, errors.New("invalid or expired token")
	}
	return claims, nil
}
