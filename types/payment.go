package types

type PaymentRequest struct {
	Amount        float64 `json:"amount"`
	CorrelationId string  `json:"correlationId"`
}

type PaymentProcessorPayload struct {
	Amount        float64 `json:"amount"`
	CorrelationId string  `json:"correlationId"`
	RequestedAt   string  `json:"requestedAt"`
}
