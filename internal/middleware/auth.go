// Package middleware contains functions that run before the route handlers,
// such as checking the login token.
package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"

	"siri_ice_cream/internal/auth"
	"siri_ice_cream/internal/models"
)

const (
	LocalUserID = "userID"
	LocalRole   = "role"
	LocalToken  = "token"
	LocalClaims = "claims"
)

func RequireAuth(secret string, blacklist *auth.TokenBlacklist) fiber.Handler {
	return func(c *fiber.Ctx) error {
		header := c.Get(fiber.HeaderAuthorization)
		token, found := strings.CutPrefix(header, "Bearer ")
		if !found || token == "" {
			return fiber.NewError(fiber.StatusUnauthorized, "missing or malformed token")
		}

		if blacklist.Contains(token) {
			return fiber.NewError(fiber.StatusUnauthorized, "token has been logged out")
		}

		claims, err := auth.ParseToken(token, secret)
		if err != nil {
			return fiber.NewError(fiber.StatusUnauthorized, err.Error())
		}

		c.Locals(LocalUserID, claims.UserID)
		c.Locals(LocalRole, claims.Role)
		c.Locals(LocalToken, token)
		c.Locals(LocalClaims, claims)
		return c.Next()
	}
}

func RequireRoles(allowed ...models.Role) fiber.Handler {
	return func(c *fiber.Ctx) error {
		role, _ := c.Locals(LocalRole).(models.Role)
		for _, r := range allowed {
			if role == r {
				return c.Next()
			}
		}
		return fiber.NewError(fiber.StatusForbidden, "you do not have permission to do this")
	}
}
