package handlers

import (
	"encoding/json"
	"jvpayments/internal/services"
	"jvpayments/internal/types"
	"log"
	"net/http"
	"time"
)

type PaymentHandler struct {
	paymentService *services.PaymentService
}

func NewPaymentHandler(paymentService *services.PaymentService) *PaymentHandler {
	return &PaymentHandler{
		paymentService: paymentService,
	}
}

func (ph *PaymentHandler) Payments(w http.ResponseWriter, r *http.Request) {
	log.Println("New payment request received")
	start := time.Now()

	if r.Method != "POST" {
		http.Error(w, `{"error": "Method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	var paymentReq types.PaymentRequest
	if err := json.NewDecoder(r.Body).Decode(&paymentReq); err != nil {
		http.Error(w, `{"error": "Invalid request body"}`, http.StatusBadRequest)
		return
	}

	defer func() {
		elapsed := time.Since(start)
		log.Printf("Execution took %s ----- %v", elapsed, paymentReq)
	}()

	// if err := validatePaymentRequest(paymentReq); err != nil {
	// 	http.Error(w, `{"error": "Invalid payment data"}`, http.StatusBadRequest)
	// 	return
	// }

	go ph.paymentService.ProcessPayment(paymentReq)
}

// func validatePaymentRequest(req types.PaymentRequest) error {
// 	// TODO: Implement validation logic
// 	return nil
// }
