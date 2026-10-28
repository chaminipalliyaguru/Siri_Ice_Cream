package handlers

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"

	"siri_ice_cream/internal/models"
)

// ItemHandler handkes the inventiry item endpoints.
type ItemHandler struct {
	db *gorm.DB
}

// NewItemHandler creates and ItemHandler
func NewItemHandler(db *gorm.DB) *ItemHandler {
	return &ItemHandler{db: db}
}

// CreateItemRequest is the json body for POST /api/items
type CreateItemRequest struct {
	Name       string  `json:"name"`
	CategoryID uint    `json:"category_id"`
	Price      float64 `json:"price"`
	Quantity   int     `json:"quantity"`
}

// UpdateItemRequest is the json body for PUT /api/items/:id
type UpdateItemRequest struct {
	Name       *string  `json:"name"`
	CategoryID *uint    `json:"category_id"`
	Price      *float64 `json:"price"`
	Quantity   *int     `json:"quantity"`
}

// List returns all items. Use ?name=choc to search items by name.
func (h *ItemHandler) List(c *fiber.Ctx) error {
	query := h.db.Preload("Category").Order("id")

	if name := strings.TrimSpace(c.Query("name")); name != "" {
		query = query.Where("LOWER(name) LIKE ?", "%"+strings.ToLower(name)+"%")
	}

	var items []models.Item
	if err := query.Find(&items).Error; err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "database error")
	}
	return c.JSON(items)
}

// Get returns one item by id.
func (h *ItemHandler) Get(c *fiber.Ctx) error {
	id, err := parseID(c)
	if err != nil {
		return err
	}

	var item models.Item
	if err := h.db.Preload("Category").First(&item, id).Error; err != nil {
		return notFoundOr(err, "item")
	}
	return c.JSON(item)
}

// Create adds a new item.
func (h *ItemHandler) Create(c *fiber.Ctx) error {
	var req CreateItemRequest
	if err := parseBody(c, &req); err != nil {
		return err
	}

	item := models.Item{
		Name:       strings.TrimSpace(req.Name),
		CategoryID: req.CategoryID,
		Price:      req.Price,
		Quantity:   req.Quantity,
	}
	if err := h.validate(item); err != nil {
		return err
	}

	if err := h.db.Create(&item).Error; err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "could not create item")
	}
	h.db.Preload("Category").First(&item, item.ID)
	return c.Status(fiber.StatusCreated).JSON(item)
}

// Update changes an existing item.
func (h *ItemHandler) Update(c *fiber.Ctx) error {
	id, err := parseID(c)
	if err != nil {
		return err
	}

	var item models.Item
	if err := h.db.First(&item, id).Error; err != nil {
		return notFoundOr(err, "item")
	}

	var req UpdateItemRequest
	if err := parseBody(c, &req); err != nil {
		return err
	}

	// Copy only the fields that were sent.
	if req.Name != nil {
		item.Name = strings.TrimSpace(*req.Name)
	}
	if req.CategoryID != nil {
		item.CategoryID = *req.CategoryID
	}
	if req.Price != nil {
		item.Price = *req.Price
	}
	if req.Quantity != nil {
		item.Quantity = *req.Quantity
	}
	if err := h.validate(item); err != nil {
		return err
	}

	if err := h.db.Save(&item).Error; err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "could not update item")
	}
	h.db.Preload("Category").First(&item, item.ID)
	return c.JSON(item)
}

// Delete removes an item.
func (h *ItemHandler) Delete(c *fiber.Ctx) error {
	id, err := parseID(c)
	if err != nil {
		return err
	}

	result := h.db.Delete(&models.Item{}, id)
	if result.Error != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "could not delete item")
	}
	if result.RowsAffected == 0 {
		return fiber.NewError(fiber.StatusNotFound, "item not found")
	}
	return c.JSON(fiber.Map{"message": "item deleted"})
}

// validate checks the item's fields before saving.
func (h *ItemHandler) validate(item models.Item) error {
	if item.Name == "" {
		return badRequest("name is required")
	}
	if item.Price <= 0 {
		return badRequest("price must be greater than 0")
	}
	if item.Quantity < 0 {
		return badRequest("quantity cannot be negative")
	}
	if item.CategoryID == 0 {
		return badRequest("category_id is required")
	}

	var count int64
	h.db.Model(&models.Category{}).Where("id = ?", item.CategoryID).Count(&count)
	if count == 0 {
		return badRequest("category does not exist")
	}
	return nil
}
