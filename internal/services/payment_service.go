package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"jvpayments/internal/cache"
	"jvpayments/internal/config"
	"jvpayments/internal/queue"
	"jvpayments/internal/types"
	"log"
	"net/http"
	"time"
)

type PaymentService struct {
	httpClient   *http.Client
	paymentQueue *queue.RedisPaymentQueue
	paymentCache *cache.PaymentCache
}

func NewPaymentService(paymentQueue *queue.RedisPaymentQueue, paymentCache *cache.PaymentCache) *PaymentService {
	return &PaymentService{
		httpClient: &http.Client{
			Timeout: 1 * time.Second,
		},
		paymentQueue: paymentQueue,
		paymentCache: paymentCache,
	}
}

func (ps *PaymentService) ProcessPayment(paymentReq types.PaymentRequest) error {
	log.Println("Starting processing payment")

	paymentReq.UpdateRequestedAt()
	payload := types.PaymentProcessorPayload{
		Amount:        paymentReq.Amount,
		CorrelationId: paymentReq.CorrelationId,
		RequestedAt:   paymentReq.RequestedAt.Format(time.RFC3339),
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payment request: %w", err)
	}

	httpReq, err := http.NewRequest("POST", config.CONFIG.PaymentApiUrl+"/payments", bytes.NewBuffer(jsonPayload))
	if err != nil {
		return fmt.Errorf("failed to create HTTP request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := ps.httpClient.Do(httpReq)

	if err != nil {
		if err := ps.paymentQueue.PublishPaymentJob(paymentReq, cache.PaymentDefaultKey); err != nil {
			return fmt.Errorf("failed to queue payment job: %w", err)
		}
		return fmt.Errorf("failed to make payment request: %w", err)
	}

	resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		if err := ps.paymentQueue.PublishPaymentJob(paymentReq, cache.PaymentDefaultKey); err != nil {
			return fmt.Errorf("failed to queue payment job: %w", err)
		}
		return fmt.Errorf("payment API returned status %d", resp.StatusCode)
	}

	if err := ps.paymentCache.StorePayment(cache.PaymentDefaultKey, paymentReq); err != nil {
		return fmt.Errorf("failed storing payment %w", err)
	}

	return nil
}
