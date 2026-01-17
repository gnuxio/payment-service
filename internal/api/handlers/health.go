package handlers

import (
	"github.com/gofiber/fiber/v2"
)

// NewHealthHandler creates a Fiber handler for the health check endpoint
func NewHealthHandler() fiber.Handler {
	return func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status": "healthy",
		})
	}
}
