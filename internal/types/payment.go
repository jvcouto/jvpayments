package types

import "time"

type PaymentRequest struct {
	Amount        float64   `json:"amount"`
	CorrelationId string    `json:"correlationId"`
	RequestedAt   time.Time `json:"requestedAt"`
}

func (pr *PaymentRequest) UpdateRequestedAt() {
	pr.RequestedAt = time.Now().UTC()
}

type PaymentProcessorPayload struct {
	Amount        float64 `json:"amount"`
	CorrelationId string  `json:"correlationId"`
	RequestedAt   string  `json:"requestedAt"`
}
