package stripe

import (
	"fmt"

	"github.com/naventro/payment-service/internal/models"
	"github.com/stripe/stripe-go/v84"
	"github.com/stripe/stripe-go/v84/checkout/session"
	"github.com/stripe/stripe-go/v84/customer"
	"github.com/stripe/stripe-go/v84/subscription"
)

type Client struct {
	secretKey string
}

func NewClient(secretKey string) *Client {
	stripe.Key = secretKey
	return &Client{secretKey: secretKey}
}

// GetPriceID returns the Stripe Price ID for a given plan
func (c *Client) GetPriceID(plan models.Plan) (string, error) {
	// These should be set in environment variables or configuration
	// For now, we'll return placeholder IDs that need to be replaced
	switch plan {
	case models.PlanPremiumMonthly:
		// Replace with actual Stripe Price ID for monthly plan
		return "price_1SqQiMEOzQkrhqSSgh3KVRps", nil
	case models.PlanPremiumYearly:
		// Replace with actual Stripe Price ID for yearly plan
		return "price_1SqQjcEOzQkrhqSSLqO5rDdb", nil
	default:
		return "", fmt.Errorf("invalid plan: %s", plan)
	}
}

// CreateCheckoutSession creates a Stripe Checkout Session
func (c *Client) CreateCheckoutSession(userID, tenant string, plan models.Plan, successURL, cancelURL string) (*stripe.CheckoutSession, error) {
	priceID, err := c.GetPriceID(plan)
	if err != nil {
		return nil, err
	}

	// Create or retrieve customer
	customerID, err := c.getOrCreateCustomer(userID, tenant)
	if err != nil {
		return nil, fmt.Errorf("error getting/creating customer: %w", err)
	}

	params := &stripe.CheckoutSessionParams{
		Customer: stripe.String(customerID),
		Mode:     stripe.String(string(stripe.CheckoutSessionModeSubscription)),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				Price:    stripe.String(priceID),
				Quantity: stripe.Int64(1),
			},
		},
		SuccessURL: stripe.String(successURL),
		CancelURL:  stripe.String(cancelURL),
		// Metadata for the checkout session
		Metadata: map[string]string{
			"user_id": userID,
			"tenant":  tenant,
			"plan":    string(plan),
		},
		// IMPORTANT: Pass metadata to the subscription that will be created
		SubscriptionData: &stripe.CheckoutSessionSubscriptionDataParams{
			Metadata: map[string]string{
				"user_id": userID,
				"tenant":  tenant,
				"plan":    string(plan),
			},
		},
	}

	sess, err := session.New(params)
	if err != nil {
		return nil, fmt.Errorf("error creating checkout session: %w", err)
	}

	return sess, nil
}

// getOrCreateCustomer gets or creates a Stripe customer
func (c *Client) getOrCreateCustomer(userID, tenant string) (string, error) {
	// Search for existing customer with this user_id
	params := &stripe.CustomerSearchParams{
		SearchParams: stripe.SearchParams{
			Query: fmt.Sprintf("metadata['user_id']:'%s' AND metadata['tenant']:'%s'", userID, tenant),
		},
	}

	iter := customer.Search(params)
	if iter.Next() {
		return iter.Customer().ID, nil
	}

	if err := iter.Err(); err != nil {
		return "", fmt.Errorf("error searching for customer: %w", err)
	}

	// Customer doesn't exist, create new one
	customerParams := &stripe.CustomerParams{
		Metadata: map[string]string{
			"user_id": userID,
			"tenant":  tenant,
		},
	}

	cust, err := customer.New(customerParams)
	if err != nil {
		return "", fmt.Errorf("error creating customer: %w", err)
	}

	return cust.ID, nil
}

// CancelSubscription schedules a Stripe subscription to cancel at period end
// This allows the user to keep access until the end of their billing period
func (c *Client) CancelSubscription(subscriptionID string) (*stripe.Subscription, error) {
	params := &stripe.SubscriptionParams{
		CancelAtPeriodEnd: stripe.Bool(true),
	}
	sub, err := subscription.Update(subscriptionID, params)
	if err != nil {
		return nil, fmt.Errorf("error scheduling subscription cancellation: %w", err)
	}

	return sub, nil
}

// ReactivateSubscription removes the scheduled cancellation for a subscription
func (c *Client) ReactivateSubscription(subscriptionID string) (*stripe.Subscription, error) {
	params := &stripe.SubscriptionParams{
		CancelAtPeriodEnd: stripe.Bool(false),
	}
	sub, err := subscription.Update(subscriptionID, params)
	if err != nil {
		return nil, fmt.Errorf("error reactivating subscription: %w", err)
	}

	return sub, nil
}

// GetSubscription retrieves a Stripe subscription
func (c *Client) GetSubscription(subscriptionID string) (*stripe.Subscription, error) {
	sub, err := subscription.Get(subscriptionID, nil)
	if err != nil {
		return nil, fmt.Errorf("error getting subscription: %w", err)
	}

	return sub, nil
}
