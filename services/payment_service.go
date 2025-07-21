package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"jvpayments/types"
	"log"
	"net/http"
	"time"
)

type PaymentService struct {
	httpClient *http.Client
}

func NewPaymentService() *PaymentService {
	return &PaymentService{
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

func (ps *PaymentService) ProcessPayment(req types.PaymentRequest, apiUrl string, requestTime time.Time) error {
	log.Println("Requesting payment processor api")

	payload := types.PaymentProcessorPayload{
		Amount:        req.Amount,
		CorrelationId: req.CorrelationId,
		RequestedAt:   requestTime.Format(time.RFC3339),
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payment request: %w", err)
	}

	httpReq, err := http.NewRequest("POST", apiUrl+"/payments", bytes.NewBuffer(jsonPayload))
	if err != nil {
		return fmt.Errorf("failed to create HTTP request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := ps.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("failed to make payment request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("payment API returned status %d", resp.StatusCode)
	}

	return nil
}
