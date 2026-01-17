package dto

import (
	"github.com/naventro/payment-service/internal/models"
)

// CheckoutRequest represents the request body for creating a checkout session
type CheckoutRequest struct {
	UserID     string      `json:"user_id"`
	Plan       models.Plan `json:"plan"`
	SuccessURL string      `json:"success_url"`
	CancelURL  string      `json:"cancel_url"`
}

// CheckoutResponse represents the response body for a successful checkout session creation
type CheckoutResponse struct {
	SessionID  string `json:"session_id"`
	SessionURL string `json:"session_url"`
}
