package handlers

import (
	"encoding/json"
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/naventro/payment-service/internal/api/dto"
	"github.com/naventro/payment-service/internal/models"
	"github.com/naventro/payment-service/internal/webhook"
	"github.com/stripe/stripe-go/v84"
	stripewebhook "github.com/stripe/stripe-go/v84/webhook"
)

// NewWebhookHandler creates a Fiber handler for processing Stripe webhooks
func NewWebhookHandler(deps *Dependencies) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get raw body for signature verification
		// IMPORTANT: Must use c.Body() to get the exact bytes Stripe sent
		body := c.Body()

		// Get Stripe signature from header
		signature := c.Get("Stripe-Signature")
		if signature == "" {
			log.Printf("Missing Stripe-Signature header")
			return dto.SendError(c, fiber.StatusBadRequest, "Missing Stripe-Signature header")
		}

		// Verify webhook signature
		event, err := stripewebhook.ConstructEvent(body, signature, deps.Config.StripeWebhookSecret)
		if err != nil {
			log.Printf("Error verifying webhook signature: %v", err)
			return dto.SendError(c, fiber.StatusBadRequest, "Invalid signature")
		}

		log.Printf("Received Stripe webhook event: %s", event.Type)

		// Handle different event types
		switch event.Type {
		case "checkout.session.completed":
			handleCheckoutSessionCompleted(event)
		case "customer.subscription.created":
			handleSubscriptionCreated(deps, event)
		case "customer.subscription.updated":
			handleSubscriptionUpdated(deps, event)
		case "customer.subscription.deleted":
			handleSubscriptionDeleted(deps, event)
		case "invoice.paid":
			handleInvoicePaid(deps, event)
		case "invoice.payment_failed":
			handleInvoicePaymentFailed(event)
		default:
			log.Printf("Unhandled event type: %s", event.Type)
		}

		return dto.SendSuccess(c, fiber.StatusOK, fiber.Map{"status": "success"})
	}
}

func handleCheckoutSessionCompleted(event stripe.Event) {
	var session stripe.CheckoutSession
	if err := json.Unmarshal(event.Data.Raw, &session); err != nil {
		log.Printf("Error unmarshaling checkout session: %v", err)
		return
	}

	log.Printf("Checkout session completed: %s", session.ID)
}

func handleSubscriptionCreated(deps *Dependencies, event stripe.Event) {
	var sub stripe.Subscription
	if err := json.Unmarshal(event.Data.Raw, &sub); err != nil {
		log.Printf("Error unmarshaling subscription: %v", err)
		return
	}

	userID, ok := sub.Metadata["user_id"]
	if !ok {
		log.Printf("Missing user_id in subscription metadata")
		return
	}

	tenant, ok := sub.Metadata["tenant"]
	if !ok {
		log.Printf("Missing tenant in subscription metadata")
		return
	}

	plan, ok := sub.Metadata["plan"]
	if !ok {
		log.Printf("Missing plan in subscription metadata")
		return
	}

	// Save subscription to database
	// In API v84+, period dates are at subscription item level
	var periodStart, periodEnd time.Time
	if len(sub.Items.Data) > 0 {
		periodStart = time.Unix(sub.Items.Data[0].CurrentPeriodStart, 0)
		periodEnd = time.Unix(sub.Items.Data[0].CurrentPeriodEnd, 0)
	}

	subscription := &models.Subscription{
		UserID:               userID,
		Tenant:               tenant,
		StripeCustomerID:     sub.Customer.ID,
		StripeSubscriptionID: sub.ID,
		Status:               models.SubscriptionStatus(sub.Status),
		Plan:                 models.Plan(plan),
		CurrentPeriodStart:   &periodStart,
		CurrentPeriodEnd:     &periodEnd,
		CancelAtPeriodEnd:    sub.CancelAtPeriodEnd,
	}

	if err := deps.SubRepo.Create(subscription); err != nil {
		log.Printf("Error creating subscription in database: %v", err)
		return
	}

	// Notify backend
	notifyBackend(deps, userID, string(subscription.Status), plan, sub.ID)

	log.Printf("Subscription created successfully for user %s", userID)
}

func handleSubscriptionUpdated(deps *Dependencies, event stripe.Event) {
	var sub stripe.Subscription
	if err := json.Unmarshal(event.Data.Raw, &sub); err != nil {
		log.Printf("Error unmarshaling subscription: %v", err)
		return
	}

	// Get existing subscription from database
	existingSub, err := deps.SubRepo.GetByStripeSubscriptionID(sub.ID)
	if err != nil {
		log.Printf("Error fetching subscription: %v", err)
		return
	}

	if existingSub == nil {
		log.Printf("Subscription not found in database: %s", sub.ID)
		return
	}

	// Update subscription
	// In API v84+, period dates are at subscription item level
	var periodStart, periodEnd time.Time
	if len(sub.Items.Data) > 0 {
		periodStart = time.Unix(sub.Items.Data[0].CurrentPeriodStart, 0)
		periodEnd = time.Unix(sub.Items.Data[0].CurrentPeriodEnd, 0)
	}

	existingSub.Status = models.SubscriptionStatus(sub.Status)
	existingSub.CurrentPeriodStart = &periodStart
	existingSub.CurrentPeriodEnd = &periodEnd
	existingSub.CancelAtPeriodEnd = sub.CancelAtPeriodEnd

	if err := deps.SubRepo.Update(existingSub); err != nil {
		log.Printf("Error updating subscription: %v", err)
		return
	}

	// Notify backend
	notifyBackend(deps, existingSub.UserID, string(existingSub.Status), string(existingSub.Plan), sub.ID)

	log.Printf("Subscription updated successfully: %s", sub.ID)
}

func handleSubscriptionDeleted(deps *Dependencies, event stripe.Event) {
	var sub stripe.Subscription
	if err := json.Unmarshal(event.Data.Raw, &sub); err != nil {
		log.Printf("Error unmarshaling subscription: %v", err)
		return
	}

	// Get existing subscription from database
	existingSub, err := deps.SubRepo.GetByStripeSubscriptionID(sub.ID)
	if err != nil {
		log.Printf("Error fetching subscription: %v", err)
		return
	}

	if existingSub == nil {
		log.Printf("Subscription not found in database: %s", sub.ID)
		return
	}

	// Update status to canceled
	existingSub.Status = models.StatusCanceled

	if err := deps.SubRepo.Update(existingSub); err != nil {
		log.Printf("Error updating subscription: %v", err)
		return
	}

	// Notify backend
	notifyBackend(deps, existingSub.UserID, "canceled", string(existingSub.Plan), sub.ID)

	log.Printf("Subscription deleted successfully: %s", sub.ID)
}

func handleInvoicePaid(deps *Dependencies, event stripe.Event) {
	var invoice stripe.Invoice
	if err := json.Unmarshal(event.Data.Raw, &invoice); err != nil {
		log.Printf("Error unmarshaling invoice: %v", err)
		return
	}

	// In API v84+, subscription is in invoice.Parent.SubscriptionDetails.Subscription
	if invoice.Parent == nil || invoice.Parent.SubscriptionDetails == nil || invoice.Parent.SubscriptionDetails.Subscription == nil {
		log.Printf("Invoice %s is not associated with a subscription", invoice.ID)
		return
	}

	subscriptionID := invoice.Parent.SubscriptionDetails.Subscription.ID
	if subscriptionID == "" {
		log.Printf("No subscription ID found for invoice: %s", invoice.ID)
		return
	}

	// Get subscription from database
	sub, err := deps.SubRepo.GetByStripeSubscriptionID(subscriptionID)
	if err != nil {
		log.Printf("Error fetching subscription: %v", err)
		return
	}

	if sub == nil {
		log.Printf("Subscription not found for invoice: %s", invoice.ID)
		return
	}

	// Check if invoice already exists
	existingInvoice, err := deps.InvoiceRepo.GetByStripeInvoiceID(invoice.ID)
	if err != nil {
		log.Printf("Error checking existing invoice: %v", err)
		return
	}

	if existingInvoice != nil {
		log.Printf("Invoice already exists: %s", invoice.ID)
		return
	}

	// Save invoice to database
	var periodStart, periodEnd *time.Time
	if invoice.PeriodStart > 0 {
		t := time.Unix(invoice.PeriodStart, 0)
		periodStart = &t
	}
	if invoice.PeriodEnd > 0 {
		t := time.Unix(invoice.PeriodEnd, 0)
		periodEnd = &t
	}

	newInvoice := &models.Invoice{
		SubscriptionID:   sub.ID,
		StripeInvoiceID:  invoice.ID,
		UserID:           sub.UserID,
		Tenant:           sub.Tenant,
		AmountPaid:       invoice.AmountPaid,
		Currency:         string(invoice.Currency),
		Status:           models.InvoiceStatus(invoice.Status),
		InvoicePDF:       &invoice.InvoicePDF,
		HostedInvoiceURL: &invoice.HostedInvoiceURL,
		PeriodStart:      periodStart,
		PeriodEnd:        periodEnd,
	}

	if err := deps.InvoiceRepo.Create(newInvoice); err != nil {
		log.Printf("Error creating invoice: %v", err)
		return
	}

	log.Printf("Invoice saved successfully: %s", invoice.ID)
}

func handleInvoicePaymentFailed(event stripe.Event) {
	var invoice stripe.Invoice
	if err := json.Unmarshal(event.Data.Raw, &invoice); err != nil {
		log.Printf("Error unmarshaling invoice: %v", err)
		return
	}

	log.Printf("Invoice payment failed: %s", invoice.ID)
}

func notifyBackend(deps *Dependencies, userID, status, plan, subscriptionID string) {
	// Obtener subscription completa de la DB para enviar todos los datos
	sub, err := deps.SubRepo.GetByStripeSubscriptionID(subscriptionID)
	if err != nil || sub == nil {
		log.Printf("Error fetching subscription for webhook: %v", err)
		// Fallback: enviar sin period data
		payload := webhook.SubscriptionWebhookPayload{
			UserID:         userID,
			Status:         status,
			Plan:           plan,
			SubscriptionID: subscriptionID,
		}
		_ = deps.WebhookClient.NotifySubscriptionChange(payload)
		return
	}

	// Enviar payload completo con todos los datos
	payload := webhook.SubscriptionWebhookPayload{
		UserID:             userID,
		Status:             status,
		Plan:               plan,
		SubscriptionID:     subscriptionID,
		CurrentPeriodStart: sub.CurrentPeriodStart,
		CurrentPeriodEnd:   sub.CurrentPeriodEnd,
	}

	if err := deps.WebhookClient.NotifySubscriptionChange(payload); err != nil {
		log.Printf("Error notifying backend: %v", err)
	}
}
