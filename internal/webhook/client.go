package webhook

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

type Client struct {
	baseURL    string
	httpClient *http.Client
}

type SubscriptionWebhookPayload struct {
	UserID         string `json:"user_id"`
	Status         string `json:"status"`
	Plan           string `json:"plan"`
	SubscriptionID string `json:"subscription_id"`
}

func NewClient(baseURL string) *Client {
	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (c *Client) NotifySubscriptionChange(payload SubscriptionWebhookPayload) error {
	url := c.baseURL + "/webhooks/subscription"

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("error marshaling payload: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	log.Printf("Sending webhook to %s with payload: %+v", url, payload)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("error sending webhook: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("webhook returned non-success status code: %d", resp.StatusCode)
	}

	log.Printf("Successfully sent webhook notification for user %s", payload.UserID)
	return nil
}
