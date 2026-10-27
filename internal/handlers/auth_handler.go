package handlers

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"siri_ice_cream/internal/auth"
	"siri_ice_cream/internal/models"
)

// AuthHandler handles login (and logout in the next step).
type AuthHandler struct {
	db        *gorm.DB
	jwtSecret string
}

// NewAuthHandler creates an AuthHandler.
func NewAuthHandler(db *gorm.DB, jwtSecret string) *AuthHandler {
	return &AuthHandler{db: db, jwtSecret: jwtSecret}
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
