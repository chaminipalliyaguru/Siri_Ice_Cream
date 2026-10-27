// Package handlers contains the functions that handle each API request.
package handlers

import (
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"

	"siri_ice_cream/internal/models"
)

// CategoryHandler handles the category endpoints.
type CategoryHandler struct {
	db *gorm.DB
}

// NewCategoryHandler creates a CategoryHandler.
func NewCategoryHandler(db *gorm.DB) *CategoryHandler {
	return &CategoryHandler{db: db}
}

// List returns all categories.
func (h *CategoryHandler) List(c *fiber.Ctx) error {
	var categories []models.Category
	if err := h.db.Order("name").Find(&categories).Error; err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "database error")
	}
	return c.JSON(categories)
}
