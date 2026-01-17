package handlers

import (
	"github.com/naventro/payment-service/internal/config"
	"github.com/naventro/payment-service/internal/repository"
	"github.com/naventro/payment-service/internal/stripe"
	"github.com/naventro/payment-service/internal/webhook"
)

// Dependencies contains all dependencies needed by handlers
type Dependencies struct {
	Config        *config.Config
	SubRepo       *repository.SubscriptionRepository
	InvoiceRepo   *repository.InvoiceRepository
	StripeClient  *stripe.Client
	WebhookClient *webhook.Client
}
