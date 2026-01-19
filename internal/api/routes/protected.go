package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/naventro/payment-service/internal/api/handlers"
	"github.com/naventro/payment-service/internal/api/middleware"
)

// setupProtectedRoutes registers protected routes (require authentication and tenant header)
func setupProtectedRoutes(router fiber.Router, deps *handlers.Dependencies) {
	// Create protected group with auth and tenant middleware
	protected := router.Group("",
		middleware.NewKeyAuthMiddleware(deps.Config.APIKey),
		middleware.NewTenantMiddleware(),
	)

	// Checkout endpoint
	protected.Post("/checkout", handlers.NewCheckoutHandler(deps))

	// Subscription endpoints
	protected.Get("/subscription/:userID", handlers.NewSubscriptionHandler(deps))

	// Cancel endpoint
	protected.Post("/cancel/:userID", handlers.NewCancelHandler(deps))

	// Reactivate endpoint
	protected.Post("/reactivate/:userID", handlers.NewReactivateHandler(deps))
}
