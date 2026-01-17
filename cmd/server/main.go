package main

import (
	"log"
	"net/http"

	"github.com/naventro/payment-service/internal/config"
	"github.com/naventro/payment-service/internal/database"
	"github.com/naventro/payment-service/internal/handlers"
	"github.com/naventro/payment-service/internal/middleware"
	"github.com/naventro/payment-service/internal/repository"
	"github.com/naventro/payment-service/internal/stripe"
	"github.com/naventro/payment-service/internal/webhook"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Error loading configuration: %v", err)
	}

	// Initialize database
	db, err := database.New(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Error connecting to database: %v", err)
	}
	defer db.Close()

	// Run migrations
	if err := db.RunMigrations(); err != nil {
		log.Fatalf("Error running migrations: %v", err)
	}

	// Initialize repositories
	subRepo := repository.NewSubscriptionRepository(db.DB)
	invoiceRepo := repository.NewInvoiceRepository(db.DB)

	// Initialize Stripe client
	stripeClient := stripe.NewClient(cfg.StripeSecretKey)

	// Initialize webhook client
	webhookClient := webhook.NewClient(cfg.BackendWebhookURL)

	// Initialize middleware
	authMiddleware := middleware.NewAuthMiddleware(cfg.APIKey)

	// Initialize handlers
	healthHandler := handlers.NewHealthHandler()
	checkoutHandler := handlers.NewCheckoutHandler(stripeClient)
	webhookHandler := handlers.NewStripeWebhookHandler(
		cfg.StripeWebhookSecret,
		subRepo,
		invoiceRepo,
		webhookClient,
	)
	subscriptionHandler := handlers.NewSubscriptionHandler(subRepo)
	cancelHandler := handlers.NewCancelHandler(subRepo, stripeClient, webhookClient)

	// Setup routes
	mux := http.NewServeMux()

	// Public routes (no authentication)
	mux.HandleFunc("/payments/health", healthHandler.ServeHTTP)
	mux.HandleFunc("/payments/webhook", webhookHandler.ServeHTTP)

	// Protected routes (require API key)
	mux.HandleFunc("/payments/checkout", authMiddleware.Authenticate(checkoutHandler.ServeHTTP))
	mux.HandleFunc("/payments/subscription/", authMiddleware.Authenticate(subscriptionHandler.ServeHTTP))
	mux.HandleFunc("/payments/cancel/", authMiddleware.Authenticate(cancelHandler.ServeHTTP))

	// Start server
	addr := ":" + cfg.Port
	log.Printf("Starting payment service on port %s", cfg.Port)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}
