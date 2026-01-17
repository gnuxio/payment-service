package models

import "time"

type SubscriptionStatus string

const (
	StatusActive            SubscriptionStatus = "active"
	StatusCanceled          SubscriptionStatus = "canceled"
	StatusIncomplete        SubscriptionStatus = "incomplete"
	StatusIncompleteExpired SubscriptionStatus = "incomplete_expired"
	StatusPastDue           SubscriptionStatus = "past_due"
	StatusTrialing          SubscriptionStatus = "trialing"
	StatusUnpaid            SubscriptionStatus = "unpaid"
)

type Plan string

const (
	PlanPremiumMonthly Plan = "premium_monthly"
	PlanPremiumYearly  Plan = "premium_yearly"
)

type Subscription struct {
	ID                   int                `json:"id"`
	UserID               string             `json:"user_id"`
	Tenant               string             `json:"tenant"`
	StripeCustomerID     string             `json:"stripe_customer_id"`
	StripeSubscriptionID string             `json:"stripe_subscription_id"`
	Status               SubscriptionStatus `json:"status"`
	Plan                 Plan               `json:"plan"`
	CurrentPeriodStart   *time.Time         `json:"current_period_start,omitempty"`
	CurrentPeriodEnd     *time.Time         `json:"current_period_end,omitempty"`
	CancelAtPeriodEnd    bool               `json:"cancel_at_period_end"`
	CreatedAt            time.Time          `json:"created_at"`
	UpdatedAt            time.Time          `json:"updated_at"`
}

func (s SubscriptionStatus) String() string {
	return string(s)
}

func (p Plan) String() string {
	return string(p)
}

func (p Plan) IsValid() bool {
	return p == PlanPremiumMonthly || p == PlanPremiumYearly
}
