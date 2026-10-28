// Package routes connects the URLs to their handlers.
package routes

import (
	"errors"
	"log/slog"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"gorm.io/gorm"

	"siri_ice_cream/internal/auth"
	"siri_ice_cream/internal/handlers"
	"siri_ice_cream/internal/middleware"
	"siri_ice_cream/internal/models"
)

const (
	owner   = models.RoleOwner
	manager = models.RoleManager
	cashier = models.RoleCashier
)

// NewApp creates the Fiber app with all routes.
func NewApp(db *gorm.DB, jwtSecret string, log *slog.Logger) *fiber.App {
	app := fiber.New(fiber.Config{
		AppName:               "Siri Ice Cream API",
		ErrorHandler:          newErrorHandler(log),
		DisableStartupMessage: true,
	})
	app.Use(requestid.New())
	app.Use(middleware.RequestLogger(log))
	app.Use(recover.New())

	blacklist := auth.NewTokenBlacklist()
	authHandler := handlers.NewAuthHandler(db, jwtSecret, blacklist)
	categoryHandler := handlers.NewCategoryHandler(db)
	itemHandler := handlers.NewItemHandler(db)
	customerHandler := handlers.NewCustomerHandler(db)

	requireAuth := middleware.RequireAuth(jwtSecret, blacklist)
	allow := middleware.RequireRoles

	api := app.Group("/api")

	// ---------- Authentication ----------
	api.Post("/auth/login", authHandler.Login) // public: no token needed
	api.Post("/auth/logout", requireAuth, authHandler.Logout)
	api.Get("/auth/me", requireAuth, authHandler.Me)

	// Every route below this line needs a valid token.
	protected := api.Group("", requireAuth)

	// ---------- Categories ----------
	protected.Get("/categories", categoryHandler.List)

	// ---------- Inventory items ----------
	items := protected.Group("/items")
	items.Get("/", allow(owner, manager, cashier), itemHandler.List)
	items.Get("/:id", allow(owner, manager, cashier), itemHandler.Get)
	items.Post("/", allow(owner, manager), itemHandler.Create)
	items.Put("/:id", allow(owner, manager, cashier), itemHandler.Update)
	items.Patch("/:id", allow(owner, manager, cashier), itemHandler.Update)
	items.Delete("/:id", allow(owner), itemHandler.Delete)

	// ---------- Customers ----------
	customers := protected.Group("/customers")
	customers.Get("/", allow(owner, manager, cashier), customerHandler.List)
	customers.Get("/:id", allow(owner, manager, cashier), customerHandler.Get)
	customers.Post("/", allow(owner, manager, cashier), customerHandler.Create)
	customers.Put("/:id", allow(owner, manager, cashier), customerHandler.Update)
	customers.Patch("/:id", allow(owner, manager, cashier), customerHandler.Update)
	customers.Delete("/:id", allow(owner), customerHandler.Delete)

	return app
}

// errorHandler sends every error back as JSON: {"error": "message"}.
func newErrorHandler(log *slog.Logger) fiber.ErrorHandler {
	return func(c *fiber.Ctx, err error) error {
		code := fiber.StatusInternalServerError
		message := "internal server error"

		var fiberErr *fiber.Error
		if errors.As(err, &fiberErr) {
			code = fiberErr.Code
			message = fiberErr.Message
		} else {
			log.Error("unexpected error",
				"request_id", c.Locals("requestid"),
				"path", c.Path(),
				"error", err,
			)
		}
		return c.Status(code).JSON(fiber.Map{"error": message})
	}
}
