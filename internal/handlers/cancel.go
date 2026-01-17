package handlers

import (
	"log"
	"net/http"
	"strings"

	"github.com/naventro/payment-service/internal/models"
	"github.com/naventro/payment-service/internal/repository"
	"github.com/naventro/payment-service/internal/stripe"
	"github.com/naventro/payment-service/internal/webhook"
)

type CancelHandler struct {
	subRepo       *repository.SubscriptionRepository
	stripeClient  *stripe.Client
	webhookClient *webhook.Client
}

func NewCancelHandler(
	subRepo *repository.SubscriptionRepository,
	stripeClient *stripe.Client,
	webhookClient *webhook.Client,
) *CancelHandler {
	return &CancelHandler{
		subRepo:       subRepo,
		stripeClient:  stripeClient,
		webhookClient: webhookClient,
	}
}

func (h *CancelHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		respondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Get tenant from header
	tenant := r.Header.Get("X-Tenant-ID")
	if tenant == "" {
		respondWithError(w, http.StatusBadRequest, "X-Tenant-ID header is required")
		return
	}

	// Extract userID from URL path
	// Expected path: /payments/cancel/{userID}
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/payments/cancel/"), "/")
	if len(parts) == 0 || parts[0] == "" {
		respondWithError(w, http.StatusBadRequest, "User ID is required")
		return
	}
	userID := parts[0]

	// Get subscription from database
	subscription, err := h.subRepo.GetByUserID(userID, tenant)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error fetching subscription")
		return
	}

	if subscription == nil {
		respondWithError(w, http.StatusNotFound, "Subscription not found")
		return
	}

	// Check if subscription is already canceled
	if subscription.Status == models.StatusCanceled {
		respondWithError(w, http.StatusBadRequest, "Subscription is already canceled")
		return
	}

	// Cancel subscription in Stripe
	_, err = h.stripeClient.CancelSubscription(subscription.StripeSubscriptionID)
	if err != nil {
		log.Printf("Error canceling Stripe subscription: %v", err)
		respondWithError(w, http.StatusInternalServerError, "Error canceling subscription")
		return
	}

	// Update subscription in database
	subscription.Status = models.StatusCanceled
	if err := h.subRepo.Update(subscription); err != nil {
		log.Printf("Error updating subscription: %v", err)
		respondWithError(w, http.StatusInternalServerError, "Error updating subscription")
		return
	}

	// Notify backend
	payload := webhook.SubscriptionWebhookPayload{
		UserID:         userID,
		Status:         "canceled",
		Plan:           string(subscription.Plan),
		SubscriptionID: subscription.StripeSubscriptionID,
	}

	if err := h.webhookClient.NotifySubscriptionChange(payload); err != nil {
		log.Printf("Error notifying backend: %v", err)
	}

	respondWithJSON(w, http.StatusOK, map[string]string{
		"status":  "success",
		"message": "Subscription canceled successfully",
	})
}
