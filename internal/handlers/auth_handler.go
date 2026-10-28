package handlers

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"siri_ice_cream/internal/auth"
	"siri_ice_cream/internal/middleware"
	"siri_ice_cream/internal/models"
)

// AuthHandler handles login (and logout in the next step).
type AuthHandler struct {
	db        *gorm.DB
	jwtSecret string
	blacklist *auth.TokenBlacklist
}

// NewAuthHandler creates an AuthHandler.
func NewAuthHandler(db *gorm.DB, jwtSecret string, blacklist *auth.TokenBlacklist) *AuthHandler {
	return &AuthHandler{db: db, jwtSecret: jwtSecret, blacklist: blacklist}
}

// LoginRequest is the JSON body for POST /api/auth/login.
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// Login checks the username and password and returns a JWT token.
func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var req LoginRequest
	if err := parseBody(c, &req); err != nil {
		return err
	}
	req.Username = strings.TrimSpace(req.Username)
	if req.Username == "" || req.Password == "" {
		return badRequest("username and password are required")
	}

	var user models.User
	err := h.db.Where("username = ?", req.Username).First(&user).Error
	// Use the same message for "no such user" and "wrong password",
	// so attackers cannot guess which usernames exist.
	if err != nil || bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)) != nil {
		return fiber.NewError(fiber.StatusUnauthorized, "invalid username or password")
	}

	token, err := auth.GenerateToken(user, h.jwtSecret)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "could not create token")
	}

	return c.JSON(fiber.Map{
		"token": token,
		"user":  user,
	})
}

// Logout puts the current token on the blacklist so it cannot be used again.
func (h *AuthHandler) Logout(c *fiber.Ctx) error {
	token := c.Locals(middleware.LocalToken).(string)
	claims := c.Locals(middleware.LocalClaims).(*auth.Claims)

	h.blacklist.Add(token, claims.ExpiresAt.Time)
	return c.JSON(fiber.Map{"message": "logged out successfully"})
}

// Me returns the currently logged-in user.
func (h *AuthHandler) Me(c *fiber.Ctx) error {
	userID := c.Locals(middleware.LocalUserID).(uint)

	var user models.User
	if err := h.db.First(&user, userID).Error; err != nil {
		return notFoundOr(err, "user")
	}
	return c.JSON(user)
}
