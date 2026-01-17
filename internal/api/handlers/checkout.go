package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/naventro/payment-service/internal/api/dto"
)

// NewCheckoutHandler creates a Fiber handler for creating checkout sessions
func NewCheckoutHandler(deps *Dependencies) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get tenant from locals (set by middleware)
		tenant := c.Locals("tenant").(string)

		var req dto.CheckoutRequest
		if err := c.BodyParser(&req); err != nil {
			return dto.SendError(c, fiber.StatusBadRequest, "Invalid request body")
		}

		// Validate request
		if req.UserID == "" {
			return dto.SendError(c, fiber.StatusBadRequest, "user_id is required")
		}

		if !req.Plan.IsValid() {
			return dto.SendError(c, fiber.StatusBadRequest, "Invalid plan")
		}

		if req.SuccessURL == "" || req.CancelURL == "" {
			return dto.SendError(c, fiber.StatusBadRequest, "success_url and cancel_url are required")
		}

		// Create Stripe checkout session
		session, err := deps.StripeClient.CreateCheckoutSession(
			req.UserID,
			tenant,
			req.Plan,
			req.SuccessURL,
			req.CancelURL,
		)
		if err != nil {
			return dto.SendError(c, fiber.StatusInternalServerError, "Error creating checkout session: "+err.Error())
		}

		response := dto.CheckoutResponse{
			SessionID:  session.ID,
			SessionURL: session.URL,
		}

		return dto.SendSuccess(c, fiber.StatusOK, response)
	}
}
