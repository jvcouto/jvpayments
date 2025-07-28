package handlers

import (
	"io"
	"jvpayments/internal/services"
	"jvpayments/internal/types"
	"log"
	"net/http"
	"time"

	"github.com/bytedance/sonic"
)

type WorkerPool struct {
	NumWorkers int
	Jobs       chan types.PaymentRequest
}

func (wp *WorkerPool) Start(ps *services.PaymentService) {
	for i := 0; i < wp.NumWorkers; i++ {
		go func(workerID int) {
			for job := range wp.Jobs {
				err := ps.ProcessPayment(job)
				if err != nil {
					log.Printf("error processing payment: %v", err)
				}
			}
		}(i)
	}
}

type PaymentHandler struct {
	workerPool WorkerPool
}

func NewPaymentHandler(paymentService *services.PaymentService) *PaymentHandler {
	wp := WorkerPool{NumWorkers: 800, Jobs: make(chan types.PaymentRequest, 100)}
	wp.Start(paymentService)
	return &PaymentHandler{
		workerPool: wp,
	}
}

func (ph *PaymentHandler) Payments(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	defer func() {
		elapsed := time.Since(start)
		log.Printf("Execution took %s", elapsed)
	}()

	log.Println("New payment request received")

	if r.Method != "POST" {
		http.Error(w, `{"error": "Method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	var paymentReq types.PaymentRequest

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, `{"error": "Error reading request body"}`, http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	if err := sonic.Unmarshal(body, &paymentReq); err != nil {
		http.Error(w, `{"error": "Invalid request body"}`, http.StatusBadRequest)
		return
	}

	// if err := validatePaymentRequest(paymentReq); err != nil {
	// 	http.Error(w, `{"error": "Invalid payment data"}`, http.StatusBadRequest)
	// 	return
	// }

	ph.workerPool.Jobs <- paymentReq
}

// func validatePaymentRequest(req types.PaymentRequest) error {
// 	// TODO: Implement validation logic
// 	return nil
// }
