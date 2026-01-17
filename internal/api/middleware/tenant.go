package middleware

import (
	"github.com/gofiber/fiber/v2"
)

// NewTenantMiddleware creates a middleware that validates X-Tenant-ID header
// and stores the tenant value in context for use by handlers
func NewTenantMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		tenant := c.Get("X-Tenant-ID")
		if tenant == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "X-Tenant-ID header is required",
			})
		}

		// Store tenant in locals for handler access
		c.Locals("tenant", tenant)

		return c.Next()
	}
}
