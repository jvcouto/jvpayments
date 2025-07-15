package queue

import (
	"context"
	"encoding/json"
	"fmt"
	redis_client "jvpayments/redis"
	"jvpayments/types"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	PaymentQueueName = "payment:jobs"
	JobTimeout       = 30 * time.Second
)

// PaymentJob represents a payment job in the queue
type PaymentJob struct {
	ID          string               `json:"id"`
	PaymentData types.PaymentRequest `json:"payment_data"`
	CreatedAt   time.Time            `json:"created_at"`
	RetryCount  int                  `json:"retry_count"`
	MaxRetries  int                  `json:"max_retries"`
}

// PaymentQueue handles payment job operations
type PaymentQueue struct {
	redisClient *redis.Client
}

// NewPaymentQueue creates a new payment queue instance
func NewPaymentQueue() *PaymentQueue {
	return &PaymentQueue{
		redisClient: redis_client.RedisClient,
	}
}

// PublishPaymentJob publishes a payment job to the queue
func (pq *PaymentQueue) PublishPaymentJob(paymentReq types.PaymentRequest) error {
	job := PaymentJob{
		ID:          generateJobID(),
		PaymentData: paymentReq,
		CreatedAt:   time.Now(),
		RetryCount:  0,
		MaxRetries:  3,
	}

	// Marshal the job to JSON
	jobData, err := json.Marshal(job)
	if err != nil {
		return fmt.Errorf("failed to marshal payment job: %w", err)
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), JobTimeout)
	defer cancel()

	// Push the job to the Redis queue
	err = pq.redisClient.LPush(ctx, PaymentQueueName, jobData).Err()
	if err != nil {
		return fmt.Errorf("failed to publish payment job: %w", err)
	}

	return nil
}

// ConsumePaymentJob consumes a payment job from the queue
func (pq *PaymentQueue) ConsumePaymentJob() (*PaymentJob, error) {
	ctx, cancel := context.WithTimeout(context.Background(), JobTimeout)
	defer cancel()

	// Pop a job from the queue (blocking operation)
	result, err := pq.redisClient.BRPop(ctx, 0, PaymentQueueName).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to consume payment job: %w", err)
	}

	if len(result) < 2 {
		return nil, fmt.Errorf("invalid queue result")
	}

	// Parse the job data
	var job PaymentJob
	if err := json.Unmarshal([]byte(result[1]), &job); err != nil {
		return nil, fmt.Errorf("failed to unmarshal payment job: %w", err)
	}

	return &job, nil
}

// GetQueueLength returns the current length of the payment queue
func (pq *PaymentQueue) GetQueueLength() (int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return pq.redisClient.LLen(ctx, PaymentQueueName).Result()
}

// RequeueJob puts a job back in the queue for retry
func (pq *PaymentQueue) RequeueJob(job *PaymentJob) error {
	jobData, err := json.Marshal(job)
	if err != nil {
		return fmt.Errorf("failed to marshal job for requeue: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), JobTimeout)
	defer cancel()

	return pq.redisClient.LPush(ctx, PaymentQueueName, jobData).Err()
}

// generateJobID generates a unique job ID
func generateJobID() string {
	return fmt.Sprintf("job_%d", time.Now().UnixNano())
}
