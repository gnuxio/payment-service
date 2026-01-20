package handlers

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/naventro/payment-service/internal/api/dto"
	"github.com/naventro/payment-service/internal/models"
	"github.com/naventro/payment-service/internal/webhook"
	"github.com/stripe/stripe-go/v84/customer"
)

// NewReactivateHandler creates a Fiber handler for reactivating canceled subscriptions
func NewReactivateHandler(deps *Dependencies) fiber.Handler {
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

		// Check if subscription is active and scheduled for cancellation
		if subscription.Status != models.StatusActive {
			return dto.SendError(c, fiber.StatusBadRequest, "Subscription is not active")
		}

		if !subscription.CancelAtPeriodEnd {
			return dto.SendError(c, fiber.StatusBadRequest, "Subscription is not scheduled for cancellation")
		}

		// Reactivate subscription in Stripe
		_, err = deps.StripeClient.ReactivateSubscription(subscription.StripeSubscriptionID)
		if err != nil {
			log.Printf("Error reactivating Stripe subscription: %v", err)
			return dto.SendError(c, fiber.StatusInternalServerError, "Error reactivating subscription")
		}

		// Update subscription in database
		subscription.CancelAtPeriodEnd = false
		if err := deps.SubRepo.Update(subscription); err != nil {
			log.Printf("Error updating subscription: %v", err)
			return dto.SendError(c, fiber.StatusInternalServerError, "Error updating subscription")
		}

		// Get customer email from Stripe
		email := getCustomerEmail(subscription.StripeCustomerID)

		// Notify backend with complete payload
		payload := webhook.SubscriptionWebhookPayload{
			UserID:             userID,
			Email:              email,
			Status:             string(subscription.Status),
			Plan:               string(subscription.Plan),
			SubscriptionID:     subscription.StripeSubscriptionID,
			CurrentPeriodStart: subscription.CurrentPeriodStart,
			CurrentPeriodEnd:   subscription.CurrentPeriodEnd,
			CancelAtPeriodEnd:  subscription.CancelAtPeriodEnd,
		}

		if err := deps.WebhookClient.NotifySubscriptionChange(payload); err != nil {
			log.Printf("Error notifying backend: %v", err)
		}

		return dto.SendSuccess(c, fiber.StatusOK, fiber.Map{
			"status":  "success",
			"message": "Subscription reactivated successfully",
		})
	}
}
