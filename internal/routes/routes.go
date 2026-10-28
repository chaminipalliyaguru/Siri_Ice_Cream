// Package routes connects the URLs to their handlers.
package routes

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"gorm.io/gorm"

	"siri_ice_cream/internal/auth"
	"siri_ice_cream/internal/handlers"
	"siri_ice_cream/internal/middleware"
)

// NewApp creates the Fiber app with all routes.
func NewApp(db *gorm.DB, jwtSecret string) *fiber.App {
	app := fiber.New(fiber.Config{
		AppName:      "Siri Ice Cream API",
		ErrorHandler: errorHandler,
	})
	app.Use(recover.New())
	app.Use(logger.New())

	blacklist := auth.NewTokenBlacklist()
	authHandler := handlers.NewAuthHandler(db, jwtSecret, blacklist)
	categoryHandler := handlers.NewCategoryHandler(db)

	requireAuth := middleware.RequireAuth(jwtSecret, blacklist)

	api := app.Group("/api")

	// ---------- Authentication ----------
	api.Post("/auth/login", authHandler.Login) // public: no token needed
	api.Post("/auth/logout", requireAuth, authHandler.Logout)
	api.Get("/auth/me", requireAuth, authHandler.Me)

	// Every route below this line needs a valid token.
	protected := api.Group("", requireAuth)

	// ---------- Categories ----------
	protected.Get("/categories", categoryHandler.List)

	return app
}

// errorHandler sends every error back as JSON: {"error": "message"}.
func errorHandler(c *fiber.Ctx, err error) error {
	code := fiber.StatusInternalServerError
	message := "internal server error"

	var fiberErr *fiber.Error
	if errors.As(err, &fiberErr) {
		code = fiberErr.Code
		message = fiberErr.Message
	}
	return c.Status(code).JSON(fiber.Map{"error": message})
}
