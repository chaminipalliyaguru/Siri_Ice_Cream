package handlers

import (
	"net/http"
	"net/mail"
	"strings"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"

	"siri_ice_cream/internal/models"
)

// CustomerHandler handles the customer endpoints.
type CustomerHandler struct {
	db *gorm.DB
}

// NewCustomerHandler creates a new CustomerHandler.
func NewCustomerHandler(db *gorm.DB) *CustomerHandler {
	return &CustomerHandler{db: db}
}

// CreateCustomerRequest is the json body for POST /api/customers
type CreateCustomerRequest struct {
	Name  string `json:"name"`
	Email string `json:"email"`
	phone string `json:"phone"`
}

// UpdateCustomerRequest is the json body for PUT /api/customers/:id
type UpdateCustomerRequest struct {
	Name  string `json:"name"`
	Email string `json:"email"`
	phone string `json:"phone"`
}

// List returns all customers in the database.
func (h *CustomerHandler) List(c *fiber.Ctx) error {
	query := h.db.Order("id")

	if name := strings.TrimSpace(c.Query("name")); name != "" {
		query = query.Where("LOWER(name) LIKE ?", "%"+strings.ToLower(name)+"%")
	}

	var customers []models.Customer
	if err := query.Find(&customers).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "failed to fetch customers"})
	}
	return c.JSON(customers)
}

// Get returns a single customer by ID.
func (h *CustomerHandler) Get(c *fiber.Ctx) error {
	id, err := parseID(c)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "invalid customer ID"})
	}

	var customer models.Customer
	if err := h.db.First(&customer, id).Error; err != nil {
		return notFoundOr(err, "customer")
	}
	return c.JSON(customer)
}

//Create adds a new customer

// validate checks the customer's fields before saving.
func (h *CustomerHandler) validate(customer models.Customer) error {
	if customer.Name == "" {
		return badRequest("name is required")
	}
	if customer.Email == "" {
		return badRequest("email is required")
	}
	if _, err := mail.ParseAddress(customer.Email); err != nil {
		return badRequest("email is not valid")
	}

	// Email must be unique. Ignore the customer's own record when updating.
	var count int64
	h.db.Model(&models.Customer{}).
		Where("email = ? AND id <> ?", customer.Email, customer.ID).
		Count(&count)
	if count > 0 {
		return fiber.NewError(fiber.StatusConflict, "a customer with this email already exists")
	}
	return nil
}
