package auth

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"siri_ice_cream/internal/models"
)

func TestGenerateAndParseToken(t *testing.T) {
	user := models.User{ID: 7, Username: "manager", Role: models.RoleManager}

	token, err := GenerateToken(user, "secret")
	if err != nil {
		t.Fatalf("generate: %v", err)
	}

	claims, err := ParseToken(token, "secret")
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if claims.UserID != 7 || claims.Role != models.RoleManager || claims.Subject != "manager" {
		t.Errorf("unexpected claims: %+v", claims)
	}
}

func TestTokenExpiresAfterLifetime(t *testing.T) {
	token, _ := GenerateToken(models.User{ID: 1, Role: models.RoleOwner}, "secret")
	claims, err := ParseToken(token, "secret")
	if err != nil {
		t.Fatalf("parse: %v", err)
	}

	lifetime := claims.ExpiresAt.Sub(claims.IssuedAt.Time)
	if lifetime != TokenLifetime {
		t.Errorf("token lifetime is %v, want %v", lifetime, TokenLifetime)
	}
}

func TestParseTokenWithWrongSecret(t *testing.T) {
	token, _ := GenerateToken(models.User{ID: 1, Role: models.RoleOwner}, "secret")

	if _, err := ParseToken(token, "other-secret"); err == nil {
		t.Error("expected an error for a token signed with another secret")
	}
}

func TestParseExpiredToken(t *testing.T) {
	// Build a token that expired one hour ago.
	claims := Claims{
		UserID: 1,
		Role:   models.RoleOwner,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-time.Hour)),
		},
	}
	token, _ := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte("secret"))

	if _, err := ParseToken(token, "secret"); err == nil {
		t.Error("expected an error for an expired token")
	}
}

func TestParseGarbageToken(t *testing.T) {
	if _, err := ParseToken("this-is-not-a-jwt", "secret"); err == nil {
		t.Error("expected an error for a malformed token")
	}
}

func TestBlacklist(t *testing.T) {
	b := NewTokenBlacklist()
	if b.Contains("abc") {
		t.Fatal("new blacklist should be empty")
	}

	b.Add("abc", time.Now().Add(time.Hour))
	if !b.Contains("abc") {
		t.Error("token should be on the blacklist after Add")
	}
	if b.Contains("xyz") {
		t.Error("a different token should not be on the blacklist")
	}
}

func TestBlacklistRemovesExpiredTokens(t *testing.T) {
	b := NewTokenBlacklist()
	b.Add("old", time.Now().Add(-time.Minute)) // already expired

	// Adding another token cleans up the expired ones.
	b.Add("new", time.Now().Add(time.Hour))

	if b.Contains("old") {
		t.Error("expired token should have been removed")
	}
	if !b.Contains("new") {
		t.Error("new token should still be on the blacklist")
	}
}
