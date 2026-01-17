package handlers

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/naventro/payment-service/internal/api/dto"
	"github.com/naventro/payment-service/internal/models"
	"github.com/naventro/payment-service/internal/webhook"
)

// NewCancelHandler creates a Fiber handler for canceling subscriptions
func NewCancelHandler(deps *Dependencies) fiber.Handler {
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

		// Check if subscription is already canceled
		if subscription.Status == models.StatusCanceled {
			return dto.SendError(c, fiber.StatusBadRequest, "Subscription is already canceled")
		}

		// Cancel subscription in Stripe
		_, err = deps.StripeClient.CancelSubscription(subscription.StripeSubscriptionID)
		if err != nil {
			log.Printf("Error canceling Stripe subscription: %v", err)
			return dto.SendError(c, fiber.StatusInternalServerError, "Error canceling subscription")
		}

		// Update subscription in database
		subscription.Status = models.StatusCanceled
		if err := deps.SubRepo.Update(subscription); err != nil {
			log.Printf("Error updating subscription: %v", err)
			return dto.SendError(c, fiber.StatusInternalServerError, "Error updating subscription")
		}

		// Notify backend
		payload := webhook.SubscriptionWebhookPayload{
			UserID:         userID,
			Status:         "canceled",
			Plan:           string(subscription.Plan),
			SubscriptionID: subscription.StripeSubscriptionID,
		}

		if err := deps.WebhookClient.NotifySubscriptionChange(payload); err != nil {
			log.Printf("Error notifying backend: %v", err)
		}

		return dto.SendSuccess(c, fiber.StatusOK, fiber.Map{
			"status":  "success",
			"message": "Subscription canceled successfully",
		})
	}
}
