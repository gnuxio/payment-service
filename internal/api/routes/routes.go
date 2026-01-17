package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/naventro/payment-service/internal/api/handlers"
)

// Setup configures all application routes
func Setup(app *fiber.App, deps *handlers.Dependencies) {
	// Create /payments group
	payments := app.Group("/payments")

	// Register public routes
	setupPublicRoutes(payments, deps)

	// Register protected routes
	setupProtectedRoutes(payments, deps)
}
