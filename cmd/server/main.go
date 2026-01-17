package main

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/naventro/payment-service/internal/api/handlers"
	"github.com/naventro/payment-service/internal/api/routes"
	"github.com/naventro/payment-service/internal/config"
	"github.com/naventro/payment-service/internal/database"
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

	// Create dependencies container
	deps := &handlers.Dependencies{
		Config:        cfg,
		DB:            db,
		SubRepo:       subRepo,
		InvoiceRepo:   invoiceRepo,
		StripeClient:  stripeClient,
		WebhookClient: webhookClient,
	}

	// Create Fiber app
	app := fiber.New(fiber.Config{
		// Disable body limit to allow large webhook payloads
		BodyLimit: 4 * 1024 * 1024, // 4MB
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}
			return c.Status(code).JSON(fiber.Map{
				"error": err.Error(),
			})
		},
	})

	// Global middleware
	app.Use(logger.New())
	app.Use(recover.New())

	// Setup routes
	routes.Setup(app, deps)

	// Start server
	addr := ":" + cfg.Port
	log.Printf("Starting payment service on port %s", cfg.Port)
	if err := app.Listen(addr); err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}
