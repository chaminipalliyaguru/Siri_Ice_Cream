// Package auth creates and checks JWT tokens.
package auth

import (
	"crypto/rand"
	"encoding/hex"
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
	id, err := newTokenID()
	if err != nil {
		return "", err
	}

	now := time.Now()
	claims := Claims{
		UserID: user.ID,
		Role:   user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			// A random ID makes every token unique. Without it, two logins of the
			// same user in the same second would create the same token, and if the
			// first one was logged out, the new one would be rejected too.
			ID:        id,
			Subject:   user.Username,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(TokenLifetime)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

// newTokenID returns 16 random bytes as hex text (the JWT "jti" claim).
func newTokenID() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
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
