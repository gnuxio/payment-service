package handlers

import (
	"github.com/gofiber/fiber/v2"
)

// NewHealthHandler creates a Fiber handler for the health check endpoint
func NewHealthHandler(deps *Dependencies) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Check database connection
		dbStatus := "healthy"
		if err := deps.DB.Ping(); err != nil {
			dbStatus = "unhealthy"
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
				"status": "unhealthy",
				"database": fiber.Map{
					"status": dbStatus,
					"error":  err.Error(),
				},
			})
		}

		return c.JSON(fiber.Map{
			"status": "healthy",
			"database": fiber.Map{
				"status": dbStatus,
			},
		})
	}
}
