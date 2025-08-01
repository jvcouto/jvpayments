package services

import (
	"bytes"
	"fmt"
	"jvpayments/internal/cache"
	"jvpayments/internal/config"
	"jvpayments/internal/queue"
	"jvpayments/internal/types"
	"log"
	"net/http"
	"time"

	"github.com/bytedance/sonic"
)

const MAX_ACCEPTABLE_REPONSETIME = 100

type PaymentService struct {
	httpClient   *http.Client
	paymentQueue *queue.RedisPaymentQueue
	paymentCache *cache.PaymentCache
}

func NewPaymentService(paymentQueue *queue.RedisPaymentQueue, paymentCache *cache.PaymentCache) *PaymentService {
	return &PaymentService{
		httpClient:   &http.Client{},
		paymentQueue: paymentQueue,
		paymentCache: paymentCache,
	}
}

func (ps *PaymentService) pickPaymentServiceToUse() (string, error) {
	apisStatus, err := ps.paymentCache.GetApisStatus()
	if err != nil {
		return "", fmt.Errorf("failed to get apis status: %w", err)
	}

	defaultStatus := apisStatus["default"]
	fallbackStatus := apisStatus["fallback"]

	getStatus := func(status any) (failing bool, minResponseTime int) {
		if status == nil || status == "" {
			return true, -1
		}
		statusMap := status.(map[string]any)
		failing = statusMap["failing"].(bool)
		minResponseTime = statusMap["minResponseTime"].(int)
		return failing, minResponseTime
	}

	defaultFailing, defaultRT := getStatus(defaultStatus)
	fallbackFailing, fallbackRT := getStatus(fallbackStatus)

	if !defaultFailing {
		return "default", nil
	}
	if defaultFailing && !fallbackFailing {
		return "fallback", nil
	}

	if !defaultFailing && !fallbackFailing {
		if defaultRT < MAX_ACCEPTABLE_REPONSETIME && fallbackRT >= MAX_ACCEPTABLE_REPONSETIME {
			return "default", nil
		}
		if fallbackRT < MAX_ACCEPTABLE_REPONSETIME && defaultRT >= MAX_ACCEPTABLE_REPONSETIME {
			return "fallback", nil
		}
		if defaultRT < fallbackRT {
			return "default", nil
		}
		return "fallback", nil
	}
	return "default", nil
}

func (ps *PaymentService) ProcessPayment(paymentReq types.PaymentRequest) error {
	start := time.Now()
	defer func() {
		elapsed := time.Since(start)
		log.Printf("[ProcessPayment]Execution took %s", elapsed)
	}()

	log.Println("Starting processing payment")

	paymentServiceToUse, err := ps.pickPaymentServiceToUse()

	if err != nil {
		return fmt.Errorf("failed to pick payment service: %w", err)
	}

	paymentReq.UpdateRequestedAt()
	payload := types.PaymentProcessorPayload{
		Amount:        paymentReq.Amount,
		CorrelationId: paymentReq.CorrelationId,
		RequestedAt:   paymentReq.RequestedAt.Format(time.RFC3339),
	}

	jsonPayload, err := sonic.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payment request: %w", err)
	}

	httpReq, err := http.NewRequest("POST", func() string {
		if paymentServiceToUse == "fallback" {
			return config.CONFIG.PaymentApiFallbackUrl
		}
		return config.CONFIG.PaymentApiUrl
	}()+"/payments", bytes.NewBuffer(jsonPayload))

	if err != nil {
		return fmt.Errorf("failed to create HTTP request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := ps.httpClient.Do(httpReq)

	if err != nil {
		queueErr := ps.paymentQueue.PublishPaymentJob(paymentReq)
		if queueErr != nil {
			return fmt.Errorf("failed to make payment request: %w; additionally failed to queue payment job: %w", err, queueErr)
		}
		return fmt.Errorf("failed to make payment request: %w", err)
	}

	resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		queueErr := ps.paymentQueue.PublishPaymentJob(paymentReq)
		if queueErr != nil {
			return fmt.Errorf(
				"payment API returned status %d; additionally failed to queue payment job: %w",
				resp.StatusCode, queueErr,
			)
		}
		return fmt.Errorf("payment API returned status %d", resp.StatusCode)
	}

	if err := ps.paymentCache.StorePayment(
		func() string {
			if paymentServiceToUse == "fallback" {
				return cache.PaymentFallbackKey
			}
			return cache.PaymentDefaultKey
		}(),
		paymentReq,
	); err != nil {
		return fmt.Errorf("failed storing payment %w", err)
	}

	return nil
}
