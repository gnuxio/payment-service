package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/naventro/payment-service/internal/api/handlers"
)

// setupPublicRoutes registers public routes (no authentication required)
func setupPublicRoutes(router fiber.Router, deps *handlers.Dependencies) {
	// Health check endpoint
	router.Get("/health", handlers.NewHealthHandler())

	// Stripe webhook endpoint (signature-verified)
	router.Post("/webhook", handlers.NewWebhookHandler(deps))
}
