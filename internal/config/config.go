package config

import (
	"fmt"
	"os"
)

type Config struct {
	Port                string
	DatabaseURL         string
	StripeSecretKey     string
	StripeWebhookSecret string
	APIKey              string
	BackendWebhookURL   string
}

func Load() (*Config, error) {
	port := getEnv("PORT", "8081")
	databaseURL := getEnv("DATABASE_URL", "")
	if databaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}

	stripeSecretKey := getEnv("STRIPE_SECRET_KEY", "")
	if stripeSecretKey == "" {
		return nil, fmt.Errorf("STRIPE_SECRET_KEY is required")
	}

	stripeWebhookSecret := getEnv("STRIPE_WEBHOOK_SECRET", "")
	if stripeWebhookSecret == "" {
		return nil, fmt.Errorf("STRIPE_WEBHOOK_SECRET is required")
	}

	apiKey := getEnv("API_KEY", "")
	if apiKey == "" {
		return nil, fmt.Errorf("API_KEY is required")
	}

	backendWebhookURL := getEnv("BACKEND_WEBHOOK_URL", "")
	if backendWebhookURL == "" {
		return nil, fmt.Errorf("BACKEND_WEBHOOK_URL is required")
	}

	return &Config{
		Port:                port,
		DatabaseURL:         databaseURL,
		StripeSecretKey:     stripeSecretKey,
		StripeWebhookSecret: stripeWebhookSecret,
		APIKey:              apiKey,
		BackendWebhookURL:   backendWebhookURL,
	}, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
