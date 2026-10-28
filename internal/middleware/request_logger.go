package middleware

import (
	"log/slog"
	"time"

	"github.com/gofiber/fiber/v2"
)

// RequestLogger writes one structured log line for every request.
func RequestLogger(log *slog.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()

		err := c.Next()
		if err != nil {
			// Let the error handler write the response now,
			// so we can log the real status code below.
			if handlerErr := c.App().ErrorHandler(c, err); handlerErr != nil {
				_ = c.SendStatus(fiber.StatusInternalServerError)
			}
		}

		status := c.Response().StatusCode()

		// Pick the log level from the status code.
		level := slog.LevelInfo
		switch {
		case status >= 500:
			level = slog.LevelError
		case status >= 400:
			level = slog.LevelWarn
		}

		attrs := []any{
			"request_id", c.Locals("requestid"),
			"method", c.Method(),
			"path", c.Path(),
			"status", status,
			"duration_ms", time.Since(start).Milliseconds(),
			"ip", c.IP(),
		}
		// RequireAuth saves the user id; it is only there for logged-in requests.
		if userID, ok := c.Locals(LocalUserID).(uint); ok {
			attrs = append(attrs, "user_id", userID)
		}

		log.Log(c.UserContext(), level, "request", attrs...)
		return nil // the response has already been written
	}
}
