package types

import "time"

type PaymentRequest struct {
	Amount        int    `json:"amount"`
	CorrelationId string `json:"correlationId"`
}

type PaymentProcessorPayload struct {
	Amount        int       `json:"amount"`
	CorrelationId string    `json:"correlationId"`
	RequestedAt   time.Time `json:"requestedAt"`
}

type PaymentResponse struct {
	ID          string `json:"id"`
	Status      string `json:"status"`
	Amount      int    `json:"amount"`
	Currency    string `json:"currency"`
	Description string `json:"description"`
	CustomerID  string `json:"customer_id"`
	CreatedAt   string `json:"created_at"`
}
