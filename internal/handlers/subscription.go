package handlers

import (
	"net/http"
	"strings"

	"github.com/naventro/payment-service/internal/repository"
)

type SubscriptionHandler struct {
	subRepo *repository.SubscriptionRepository
}

func NewSubscriptionHandler(subRepo *repository.SubscriptionRepository) *SubscriptionHandler {
	return &SubscriptionHandler{
		subRepo: subRepo,
	}
}

func (h *SubscriptionHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		respondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Get tenant from header
	tenant := r.Header.Get("X-Tenant-ID")
	if tenant == "" {
		respondWithError(w, http.StatusBadRequest, "X-Tenant-ID header is required")
		return
	}

	// Extract userID from URL path
	// Expected path: /payments/subscription/{userID}
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/payments/subscription/"), "/")
	if len(parts) == 0 || parts[0] == "" {
		respondWithError(w, http.StatusBadRequest, "User ID is required")
		return
	}
	userID := parts[0]

	// Get subscription from database
	subscription, err := h.subRepo.GetByUserID(userID, tenant)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error fetching subscription")
		return
	}

	if subscription == nil {
		respondWithError(w, http.StatusNotFound, "Subscription not found")
		return
	}

	respondWithJSON(w, http.StatusOK, subscription)
}
