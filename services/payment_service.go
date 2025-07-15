package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"jvpayments/types"
	"net/http"
	"time"
)

type PaymentService struct {
	httpClient *http.Client
}

func NewPaymentService() *PaymentService {
	return &PaymentService{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (ps *PaymentService) ProcessPayment(req types.PaymentRequest) (*types.PaymentResponse, error) {
	payload := map[string]interface{}{
		"amount": req.Amount,
	}

	// Marshal the payload to JSON
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payment request: %w", err)
	}

	// Create HTTP request
	httpReq, err := http.NewRequest("POST", ps.apiURL, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	// Make the request
	resp, err := ps.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to make payment request: %w", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("payment API returned status %d", resp.StatusCode)
	}

	// Parse the response
	var paymentResp types.PaymentResponse
	if err := json.NewDecoder(resp.Body).Decode(&paymentResp); err != nil {
		return nil, fmt.Errorf("failed to decode payment response: %w", err)
	}

	return &paymentResp, nil
}
