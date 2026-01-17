package handlers

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/naventro/payment-service/internal/models"
	"github.com/naventro/payment-service/internal/repository"
	"github.com/naventro/payment-service/internal/webhook"
	"github.com/stripe/stripe-go/v84"
	stripewebhook "github.com/stripe/stripe-go/v84/webhook"
)

type StripeWebhookHandler struct {
	webhookSecret      string
	subRepo            *repository.SubscriptionRepository
	invoiceRepo        *repository.InvoiceRepository
	webhookClient      *webhook.Client
}

func NewStripeWebhookHandler(
	webhookSecret string,
	subRepo *repository.SubscriptionRepository,
	invoiceRepo *repository.InvoiceRepository,
	webhookClient *webhook.Client,
) *StripeWebhookHandler {
	return &StripeWebhookHandler{
		webhookSecret: webhookSecret,
		subRepo:       subRepo,
		invoiceRepo:   invoiceRepo,
		webhookClient: webhookClient,
	}
}

func (h *StripeWebhookHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		respondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Error reading request body")
		return
	}

	signature := r.Header.Get("Stripe-Signature")
	if signature == "" {
		respondWithError(w, http.StatusBadRequest, "Missing Stripe-Signature header")
		return
	}

	event, err := stripewebhook.ConstructEvent(body, signature, h.webhookSecret)
	if err != nil {
		log.Printf("Error verifying webhook signature: %v", err)
		respondWithError(w, http.StatusBadRequest, "Invalid signature")
		return
	}

	log.Printf("Received Stripe webhook event: %s", event.Type)

	// Handle different event types
	switch event.Type {
	case "checkout.session.completed":
		h.handleCheckoutSessionCompleted(event)
	case "customer.subscription.created":
		h.handleSubscriptionCreated(event)
	case "customer.subscription.updated":
		h.handleSubscriptionUpdated(event)
	case "customer.subscription.deleted":
		h.handleSubscriptionDeleted(event)
	case "invoice.paid":
		h.handleInvoicePaid(event)
	case "invoice.payment_failed":
		h.handleInvoicePaymentFailed(event)
	default:
		log.Printf("Unhandled event type: %s", event.Type)
	}

	respondWithJSON(w, http.StatusOK, map[string]string{"status": "success"})
}

func (h *StripeWebhookHandler) handleCheckoutSessionCompleted(event stripe.Event) {
	var session stripe.CheckoutSession
	if err := json.Unmarshal(event.Data.Raw, &session); err != nil {
		log.Printf("Error unmarshaling checkout session: %v", err)
		return
	}

	log.Printf("Checkout session completed: %s", session.ID)
}

func (h *StripeWebhookHandler) handleSubscriptionCreated(event stripe.Event) {
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

	if err := h.subRepo.Create(subscription); err != nil {
		log.Printf("Error creating subscription in database: %v", err)
		return
	}

	// Notify backend
	h.notifyBackend(userID, string(subscription.Status), plan, sub.ID)

	log.Printf("Subscription created successfully for user %s", userID)
}

func (h *StripeWebhookHandler) handleSubscriptionUpdated(event stripe.Event) {
	var sub stripe.Subscription
	if err := json.Unmarshal(event.Data.Raw, &sub); err != nil {
		log.Printf("Error unmarshaling subscription: %v", err)
		return
	}

	// Get existing subscription from database
	existingSub, err := h.subRepo.GetByStripeSubscriptionID(sub.ID)
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

	if err := h.subRepo.Update(existingSub); err != nil {
		log.Printf("Error updating subscription: %v", err)
		return
	}

	// Notify backend
	h.notifyBackend(existingSub.UserID, string(existingSub.Status), string(existingSub.Plan), sub.ID)

	log.Printf("Subscription updated successfully: %s", sub.ID)
}

func (h *StripeWebhookHandler) handleSubscriptionDeleted(event stripe.Event) {
	var sub stripe.Subscription
	if err := json.Unmarshal(event.Data.Raw, &sub); err != nil {
		log.Printf("Error unmarshaling subscription: %v", err)
		return
	}

	// Get existing subscription from database
	existingSub, err := h.subRepo.GetByStripeSubscriptionID(sub.ID)
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

	if err := h.subRepo.Update(existingSub); err != nil {
		log.Printf("Error updating subscription: %v", err)
		return
	}

	// Notify backend
	h.notifyBackend(existingSub.UserID, "canceled", string(existingSub.Plan), sub.ID)

	log.Printf("Subscription deleted successfully: %s", sub.ID)
}

func (h *StripeWebhookHandler) handleInvoicePaid(event stripe.Event) {
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
	sub, err := h.subRepo.GetByStripeSubscriptionID(subscriptionID)
	if err != nil {
		log.Printf("Error fetching subscription: %v", err)
		return
	}

	if sub == nil {
		log.Printf("Subscription not found for invoice: %s", invoice.ID)
		return
	}

	// Check if invoice already exists
	existingInvoice, err := h.invoiceRepo.GetByStripeInvoiceID(invoice.ID)
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

	if err := h.invoiceRepo.Create(newInvoice); err != nil {
		log.Printf("Error creating invoice: %v", err)
		return
	}

	log.Printf("Invoice saved successfully: %s", invoice.ID)
}

func (h *StripeWebhookHandler) handleInvoicePaymentFailed(event stripe.Event) {
	var invoice stripe.Invoice
	if err := json.Unmarshal(event.Data.Raw, &invoice); err != nil {
		log.Printf("Error unmarshaling invoice: %v", err)
		return
	}

	log.Printf("Invoice payment failed: %s", invoice.ID)
}

func (h *StripeWebhookHandler) notifyBackend(userID, status, plan, subscriptionID string) {
	payload := webhook.SubscriptionWebhookPayload{
		UserID:         userID,
		Status:         status,
		Plan:           plan,
		SubscriptionID: subscriptionID,
	}

	if err := h.webhookClient.NotifySubscriptionChange(payload); err != nil {
		log.Printf("Error notifying backend: %v", err)
	}
}
