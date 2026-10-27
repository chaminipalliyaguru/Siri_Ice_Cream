// Package routes connects the URLs to their handlers.
package routes

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"gorm.io/gorm"

	"siri_ice_cream/internal/handlers"
)

// NewApp creates the Fiber app with all routes.
func NewApp(db *gorm.DB) *fiber.App {
	app := fiber.New(fiber.Config{
		AppName:      "Siri Ice Cream API",
		ErrorHandler: errorHandler,
	})
	app.Use(recover.New()) // turns a crash (panic) into a 500 error
	app.Use(logger.New())  // prints each request in the console

	categoryHandler := handlers.NewCategoryHandler(db)

	api := app.Group("/api")
	api.Get("/categories", categoryHandler.List)

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
