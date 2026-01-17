package models

import "time"

type InvoiceStatus string

const (
	InvoiceStatusDraft         InvoiceStatus = "draft"
	InvoiceStatusOpen          InvoiceStatus = "open"
	InvoiceStatusPaid          InvoiceStatus = "paid"
	InvoiceStatusUncollectible InvoiceStatus = "uncollectible"
	InvoiceStatusVoid          InvoiceStatus = "void"
)

type Invoice struct {
	ID               int           `json:"id"`
	SubscriptionID   int           `json:"subscription_id"`
	StripeInvoiceID  string        `json:"stripe_invoice_id"`
	UserID           string        `json:"user_id"`
	Tenant           string        `json:"tenant"`
	AmountPaid       int64         `json:"amount_paid"`
	Currency         string        `json:"currency"`
	Status           InvoiceStatus `json:"status"`
	InvoicePDF       *string       `json:"invoice_pdf,omitempty"`
	HostedInvoiceURL *string       `json:"hosted_invoice_url,omitempty"`
	PeriodStart      *time.Time    `json:"period_start,omitempty"`
	PeriodEnd        *time.Time    `json:"period_end,omitempty"`
	CreatedAt        time.Time     `json:"created_at"`
}

func (i InvoiceStatus) String() string {
	return string(i)
}
