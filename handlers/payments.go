package handlers

import (
	"encoding/json"
	"jvpayments/queue"
	"jvpayments/types"
	"log"
	"net/http"
)

func Payments(w http.ResponseWriter, r *http.Request) {
	log.Println("Starting processing new payment")

	if r.Method != "POST" {
		http.Error(w, `{"error": "Method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	var paymentReq types.PaymentRequest
	if err := json.NewDecoder(r.Body).Decode(&paymentReq); err != nil {
		http.Error(w, `{"error": "Invalid request body"}`, http.StatusBadRequest)
		return
	}

	// if err := validatePaymentRequest(paymentReq); err != nil {
	// 	http.Error(w, `{"error": "Invalid payment data"}`, http.StatusBadRequest)
	// 	return
	// }

	/*
		se serviço default ok = serviço defalut
		se servico default false e fallback true = fallback
		se os 2 serviços estiverem false
	*/
	defaultPaymentQueue := queue.NewRedisPaymentQueue(queue.PaymentQueueName)

	if err := defaultPaymentQueue.PublishPaymentJob(paymentReq); err != nil {
		http.Error(w, `{"error": "Failed to queue payment job"}`, http.StatusInternalServerError)
		return
	}
}

// func validatePaymentRequest(req types.PaymentRequest) error {
// 	// TODO: Implement validation logic
// 	return nil
// }
