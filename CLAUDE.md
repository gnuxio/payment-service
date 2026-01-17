# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Payment microservice built with pure Go (net/http) for managing Stripe subscriptions across multiple SaaS applications. Uses PostgreSQL for persistence and Docker for containerization.

## Tech Stack

- Go 1.25.3 with pure net/http (no frameworks)
- Stripe API (stripe-go/v81)
- PostgreSQL with lib/pq driver
- Docker and Docker Compose

## Development Commands

### Running the Service

```bash
# With Docker (recommended)
docker-compose up -d

# Without Docker
docker run -d --name payment-postgres \
  -e POSTGRES_DB=payment_service \
  -e POSTGRES_USER=postgres \
  -e POSTGRES_PASSWORD=postgres \
  -p 5433:5432 \
  postgres:16-alpine

export $(cat .env | xargs)
go run cmd/server/main.go
```

### Testing

```bash
# Health check
curl http://localhost:8081/payments/health

# Test all endpoints
./examples/curl-examples.sh

# Test webhooks with Stripe CLI
stripe listen --forward-to localhost:8081/payments/webhook
stripe trigger checkout.session.completed
```

### Database

```bash
# View logs
docker logs -f payment-service

# Connect to database
psql postgres://postgres:postgres@localhost:5433/payment_service?sslmode=disable
```

## Architecture

### Request Flow

```
Frontend → menuum-backend → payment-service → Stripe
                ↑                    ↓
                └──── webhook ───────┘
```

The frontend NEVER calls payment-service directly. Only menuum-backend communicates with this service.

### Core Components

**cmd/server/main.go**: Application entry point that wires together all dependencies. Initialize components in this order: config → database → repositories → clients → middleware → handlers → routes.

**internal/config/config.go**: Loads environment variables. All secrets (Stripe keys, API keys) come from .env file.

**internal/database/**: PostgreSQL connection and migrations. Migrations run automatically on startup from `migrations/` directory in numbered order (001_*, 002_*).

**internal/models/**: Domain models for Subscription and Invoice. Subscriptions track recurring payments; Invoices track individual payment records.

**internal/repository/**: Data access layer. Uses raw SQL with lib/pq for PostgreSQL operations.

**internal/stripe/client.go**: Stripe API wrapper. IMPORTANT: Price IDs in `GetPriceID()` function are placeholders and MUST be updated with real Stripe Price IDs before production use.

**internal/handlers/**: HTTP handlers implementing business logic:
- health.go: Health check endpoint (public)
- checkout.go: Create Stripe checkout sessions (protected)
- webhook.go: Receive and process Stripe webhooks (public, signature-verified)
- subscription.go: Get subscription status (protected)
- cancel.go: Cancel subscriptions (protected)

**internal/middleware/auth.go**: API key authentication. Validates X-API-Key header for protected endpoints.

**internal/webhook/client.go**: HTTP client to notify menuum-backend of subscription changes.

### Multi-tenancy

All protected requests require X-Tenant-ID header (e.g., "menuum") to identify which SaaS application the request belongs to. Tenant ID is stored with subscriptions and customers in both Stripe metadata and local database.

### Webhook Flow

1. Stripe sends webhook to /payments/webhook
2. Signature is validated using STRIPE_WEBHOOK_SECRET
3. Event type is processed (checkout.session.completed, customer.subscription.*, invoice.*)
4. Database is updated (subscriptions/invoices tables)
5. Backend webhook notification is sent to BACKEND_WEBHOOK_URL

### Authentication

- Public endpoints: /payments/health, /payments/webhook (signature-verified)
- Protected endpoints: All others require X-API-Key header matching API_KEY env var

## Critical Configuration

Before first use, you MUST:

1. Update Stripe Price IDs in `internal/stripe/client.go:GetPriceID()` with actual Price IDs from Stripe dashboard
2. Configure STRIPE_SECRET_KEY in .env from Stripe dashboard → Developers → API Keys
3. Configure STRIPE_WEBHOOK_SECRET in .env from Stripe dashboard → Developers → Webhooks
4. Generate secure API_KEY with `openssl rand -hex 32` and share with menuum-backend

## Database Schema

**subscriptions table**:
- Tracks user subscription state
- Links to Stripe via stripe_customer_id and stripe_subscription_id
- Status values: active, canceled, past_due, etc.
- Plan values: premium_monthly, premium_yearly

**invoices table**:
- Stores payment history
- Links to subscriptions table via subscription_id
- Populated automatically by Stripe invoice.paid webhooks

Migrations are in `internal/database/migrations/` and run automatically on startup.

## Code Patterns

**Handler initialization**: All handlers are initialized in main.go with required dependencies injected via constructor functions (New*Handler).

**Error handling**: Use log.Printf for all errors and return appropriate HTTP status codes. Never expose internal errors to clients.

**Stripe integration**: Always pass user_id, tenant, and plan in metadata when creating Stripe resources. This metadata flows back in webhooks and enables proper tracking.

**Repository pattern**: All database operations go through repository layer. Use prepared statements implicitly via lib/pq parameterized queries.

## Common Tasks

**Adding a new endpoint**:
1. Create handler in internal/handlers/
2. Initialize in cmd/server/main.go
3. Add route in main.go (use authMiddleware.Authenticate() for protected routes)

**Adding a new webhook event**:
1. Add case to switch statement in internal/handlers/webhook.go ServeHTTP
2. Create handler method following existing pattern (handleEventType)
3. Update Stripe webhook configuration to include new event type

**Updating Price IDs**: Edit internal/stripe/client.go GetPriceID() function with Price IDs from Stripe dashboard.
