package types

import "time"

type PaymentRequest struct {
	Amount        float64 `json:"amount"`
	CorrelationId string  `json:"correlationId"`
}

type PaymentProcessorPayload struct {
	Amount        float64   `json:"amount"`
	CorrelationId string    `json:"correlationId"`
	RequestedAt   time.Time `json:"requestedAt"`
}
