package types

type PaymentRequest struct {
	Amount         int    `json:"amount"`
	Correlation_Id string `json:"correlation_id"`
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
