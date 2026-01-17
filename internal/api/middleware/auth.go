package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/keyauth"
)

// NewKeyAuthMiddleware creates a Fiber keyauth middleware that validates X-API-Key header
func NewKeyAuthMiddleware(apiKey string) fiber.Handler {
	return keyauth.New(keyauth.Config{
		KeyLookup: "header:X-API-Key",
		Validator: func(c *fiber.Ctx, key string) (bool, error) {
			return key == apiKey, nil
		},
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Missing or invalid API key",
			})
		},
	})
}
