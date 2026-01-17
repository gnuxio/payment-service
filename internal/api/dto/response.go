package dto

import (
	"github.com/gofiber/fiber/v2"
)

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error string `json:"error"`
}

// SendError sends an error response with the given status code and message
func SendError(c *fiber.Ctx, status int, message string) error {
	return c.Status(status).JSON(ErrorResponse{
		Error: message,
	})
}

// SendSuccess sends a success response with the given status code and data
func SendSuccess(c *fiber.Ctx, status int, data interface{}) error {
	return c.Status(status).JSON(data)
}
