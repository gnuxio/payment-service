#!/bin/bash

# Example curl commands to test payment-service endpoints
# Make sure to replace the API_KEY and USER_ID with actual values

API_KEY="your_api_key_here"
PAYMENT_SERVICE_URL="http://localhost:8081"
TENANT_ID="menuum"
USER_ID="cognito_user_id_here"

echo "==================================================="
echo "Payment Service API Examples"
echo "==================================================="
echo ""

# 1. Health Check
echo "1. Health Check"
echo "   GET /payments/health"
curl -X GET "${PAYMENT_SERVICE_URL}/payments/health"
echo -e "\n"

# 2. Create Checkout Session
echo "2. Create Checkout Session"
echo "   POST /payments/checkout"
curl -X POST "${PAYMENT_SERVICE_URL}/payments/checkout" \
  -H "X-API-Key: ${API_KEY}" \
  -H "X-Tenant-ID: ${TENANT_ID}" \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "'"${USER_ID}"'",
    "plan": "premium_monthly",
    "success_url": "https://menuum.com/success",
    "cancel_url": "https://menuum.com/cancel"
  }'
echo -e "\n"

# 3. Get Subscription Status
echo "3. Get Subscription Status"
echo "   GET /payments/subscription/{userId}"
curl -X GET "${PAYMENT_SERVICE_URL}/payments/subscription/${USER_ID}" \
  -H "X-API-Key: ${API_KEY}" \
  -H "X-Tenant-ID: ${TENANT_ID}"
echo -e "\n"

# 4. Cancel Subscription
echo "4. Cancel Subscription"
echo "   POST /payments/cancel/{userId}"
curl -X POST "${PAYMENT_SERVICE_URL}/payments/cancel/${USER_ID}" \
  -H "X-API-Key: ${API_KEY}" \
  -H "X-Tenant-ID: ${TENANT_ID}"
echo -e "\n"

echo "==================================================="
echo "Test Stripe Webhook (Local Testing)"
echo "==================================================="
echo ""

# For local testing, you can trigger a test webhook using Stripe CLI:
# stripe trigger checkout.session.completed

# Or manually send a test webhook:
echo "5. Test Webhook (Manual - requires Stripe webhook secret)"
echo "   POST /payments/webhook"
echo "   Note: In production, this is called by Stripe automatically"
echo "   Use 'stripe listen --forward-to localhost:8081/payments/webhook' for local testing"
echo ""

echo "==================================================="
echo "Testing with Different Plans"
echo "==================================================="
echo ""

# Premium Monthly
echo "6. Create Checkout for Premium Monthly ($9.99/month)"
curl -X POST "${PAYMENT_SERVICE_URL}/payments/checkout" \
  -H "X-API-Key: ${API_KEY}" \
  -H "X-Tenant-ID: ${TENANT_ID}" \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "'"${USER_ID}"'",
    "plan": "premium_monthly",
    "success_url": "https://menuum.com/success",
    "cancel_url": "https://menuum.com/cancel"
  }'
echo -e "\n"

# Premium Yearly
echo "7. Create Checkout for Premium Yearly ($99/year)"
curl -X POST "${PAYMENT_SERVICE_URL}/payments/checkout" \
  -H "X-API-Key: ${API_KEY}" \
  -H "X-Tenant-ID: ${TENANT_ID}" \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "'"${USER_ID}"'",
    "plan": "premium_yearly",
    "success_url": "https://menuum.com/success",
    "cancel_url": "https://menuum.com/cancel"
  }'
echo -e "\n"

echo "==================================================="
echo "Error Cases"
echo "==================================================="
echo ""

# Missing API Key
echo "8. Missing API Key (should return 401)"
curl -X POST "${PAYMENT_SERVICE_URL}/payments/checkout" \
  -H "X-Tenant-ID: ${TENANT_ID}" \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "'"${USER_ID}"'",
    "plan": "premium_monthly",
    "success_url": "https://menuum.com/success",
    "cancel_url": "https://menuum.com/cancel"
  }'
echo -e "\n"

# Invalid API Key
echo "9. Invalid API Key (should return 401)"
curl -X POST "${PAYMENT_SERVICE_URL}/payments/checkout" \
  -H "X-API-Key: invalid_key" \
  -H "X-Tenant-ID: ${TENANT_ID}" \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "'"${USER_ID}"'",
    "plan": "premium_monthly",
    "success_url": "https://menuum.com/success",
    "cancel_url": "https://menuum.com/cancel"
  }'
echo -e "\n"

# Missing Tenant ID
echo "10. Missing Tenant ID (should return 400)"
curl -X POST "${PAYMENT_SERVICE_URL}/payments/checkout" \
  -H "X-API-Key: ${API_KEY}" \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "'"${USER_ID}"'",
    "plan": "premium_monthly",
    "success_url": "https://menuum.com/success",
    "cancel_url": "https://menuum.com/cancel"
  }'
echo -e "\n"

# Invalid Plan
echo "11. Invalid Plan (should return 400)"
curl -X POST "${PAYMENT_SERVICE_URL}/payments/checkout" \
  -H "X-API-Key: ${API_KEY}" \
  -H "X-Tenant-ID: ${TENANT_ID}" \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "'"${USER_ID}"'",
    "plan": "invalid_plan",
    "success_url": "https://menuum.com/success",
    "cancel_url": "https://menuum.com/cancel"
  }'
echo -e "\n"

# User Not Found
echo "12. Get Subscription for Non-Existent User (should return 404)"
curl -X GET "${PAYMENT_SERVICE_URL}/payments/subscription/non_existent_user" \
  -H "X-API-Key: ${API_KEY}" \
  -H "X-Tenant-ID: ${TENANT_ID}"
echo -e "\n"

echo "==================================================="
echo "Done!"
echo "==================================================="
