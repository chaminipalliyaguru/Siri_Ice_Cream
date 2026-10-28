package handlers

import (
	"errors"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// parseBody reads the JSON request body into out.
func parseBody(c *fiber.Ctx, out interface{}) error {
	if err := c.BodyParser(out); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid JSON body")
	}
	return nil
}

// badRequest is a short way to return a 400 validation error.
func badRequest(message string) error {
	return fiber.NewError(fiber.StatusBadRequest, message)
}

func notFoundOr(err error, what string) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return fiber.NewError(fiber.StatusNotFound, what+" not found")
	}
	return fiber.NewError(fiber.StatusInternalServerError, "database error")
}

// parseID reads the ":id" part of the URL and checks it is a positive number.
func parseID(c *fiber.Ctx) (uint, error) {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil || id == 0 {
		return 0, fiber.NewError(fiber.StatusBadRequest, "id must be a positive number")
	}
	return uint(id), nil
}
