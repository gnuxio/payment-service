package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/naventro/payment-service/internal/api/dto"
)

// NewSubscriptionHandler creates a Fiber handler for getting subscription status
func NewSubscriptionHandler(deps *Dependencies) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get tenant from locals (set by middleware)
		tenant := c.Locals("tenant").(string)

		// Extract userID from URL path parameter
		userID := c.Params("userID")
		if userID == "" {
			return dto.SendError(c, fiber.StatusBadRequest, "User ID is required")
		}

		// Get subscription from database
		subscription, err := deps.SubRepo.GetByUserID(userID, tenant)
		if err != nil {
			return dto.SendError(c, fiber.StatusInternalServerError, "Error fetching subscription")
		}

		if subscription == nil {
			return dto.SendError(c, fiber.StatusNotFound, "Subscription not found")
		}

		return dto.SendSuccess(c, fiber.StatusOK, subscription)
	}
}
