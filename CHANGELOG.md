# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.0.0] - 2026-01-16

### Added

- Initial release of payment-service
- Stripe integration for subscription management
- Multi-tenant support via X-Tenant-ID header
- API Key authentication for backend-to-backend communication
- PostgreSQL database with automatic migrations
- RESTful API endpoints:
  - `GET /payments/health` - Health check
  - `POST /payments/checkout` - Create Stripe checkout session
  - `GET /payments/subscription/:userId` - Get subscription status
  - `POST /payments/cancel/:userId` - Cancel subscription
  - `POST /payments/webhook` - Stripe webhook handler
- Webhook client to notify backend (menuum-backend) on subscription changes
- Database models for subscriptions and invoices
- Comprehensive logging for debugging
- Docker support with docker-compose
- Environment-based configuration
- README with detailed setup instructions
- Example integration code for Python/FastAPI backends
- curl examples for API testing

### Features

- Two subscription plans:
  - Premium Monthly: $9.99/month
  - Premium Yearly: $99/year
- Automatic subscription status synchronization with Stripe
- Invoice history tracking
- Subscription lifecycle management (create, update, cancel)
- Webhook notifications to backend services
- Secure Stripe webhook signature verification

### Infrastructure

- Go 1.23+ with net/http (no frameworks)
- PostgreSQL 16 for data persistence
- Stripe API v81
- Docker containerization
- Embedded SQL migrations

### Documentation

- Comprehensive README with setup guide
- Stripe configuration instructions
- API usage examples
- Integration examples for Python/FastAPI
- curl command examples
- Architecture diagrams
- Troubleshooting guide

### Security

- API Key authentication for protected endpoints
- Stripe webhook signature verification
- Environment variable based secrets management
- Input validation on all endpoints
- Tenant isolation

## [Unreleased]

### Planned Features

- Unit and integration tests
- Rate limiting
- Prometheus metrics
- Retry logic for failed webhook notifications
- Coupon/discount code support
- Plan upgrade/downgrade functionality
- Customer portal integration
- Detailed invoice PDF generation
- Email notifications for subscription events
- Admin API for subscription management
- GraphQL API option
- Multiple payment method support
- Proration support for plan changes
- Subscription pause/resume functionality
