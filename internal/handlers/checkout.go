package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/naventro/payment-service/internal/models"
	"github.com/naventro/payment-service/internal/stripe"
)

type CheckoutHandler struct {
	stripeClient *stripe.Client
}

type CheckoutRequest struct {
	UserID     string      `json:"user_id"`
	Plan       models.Plan `json:"plan"`
	SuccessURL string      `json:"success_url"`
	CancelURL  string      `json:"cancel_url"`
}

type CheckoutResponse struct {
	SessionID  string `json:"session_id"`
	SessionURL string `json:"session_url"`
}

func NewCheckoutHandler(stripeClient *stripe.Client) *CheckoutHandler {
	return &CheckoutHandler{
		stripeClient: stripeClient,
	}
}

func (h *CheckoutHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		respondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Get tenant from header
	tenant := r.Header.Get("X-Tenant-ID")
	if tenant == "" {
		respondWithError(w, http.StatusBadRequest, "X-Tenant-ID header is required")
		return
	}

	var req CheckoutRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate request
	if req.UserID == "" {
		respondWithError(w, http.StatusBadRequest, "user_id is required")
		return
	}

	if !req.Plan.IsValid() {
		respondWithError(w, http.StatusBadRequest, "Invalid plan")
		return
	}

	if req.SuccessURL == "" || req.CancelURL == "" {
		respondWithError(w, http.StatusBadRequest, "success_url and cancel_url are required")
		return
	}

	// Create Stripe checkout session
	session, err := h.stripeClient.CreateCheckoutSession(
		req.UserID,
		tenant,
		req.Plan,
		req.SuccessURL,
		req.CancelURL,
	)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error creating checkout session: "+err.Error())
		return
	}

	response := CheckoutResponse{
		SessionID:  session.ID,
		SessionURL: session.URL,
	}

	respondWithJSON(w, http.StatusOK, response)
}
