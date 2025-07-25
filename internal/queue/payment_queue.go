package queue

import (
	"context"
	"encoding/json"
	"fmt"
	redis_client "jvpayments/internal/redis"
	"jvpayments/internal/types"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	PaymentQueueName          = "payment:jobs"
	PaymentFallabackQueueName = "payment:jobs:fallback"
	JobTimeout                = 30 * time.Second
)

type PaymentJob struct {
	ID          string               `json:"id"`
	PaymentData types.PaymentRequest `json:"payment_data"`
	RetryCount  int                  `json:"retry_count"`
	MaxRetries  int                  `json:"max_retries"`
}

type RedisPaymentQueue struct {
	redisClient *redis.Client
}

func getMaxRetries(queueName string) int {
	switch queueName {
	case PaymentQueueName:
		return 3
	case PaymentFallabackQueueName:
		return 1
	default:
		return 1
	}
}

func NewRedisPaymentQueue() *RedisPaymentQueue {
	return &RedisPaymentQueue{
		redisClient: redis_client.RedisClient,
	}
}

func (pq *RedisPaymentQueue) PublishPaymentJob(paymentReq types.PaymentRequest, queueName string) error {
	job := PaymentJob{
		ID:          generateJobID(),
		PaymentData: paymentReq,
		RetryCount:  0,
		MaxRetries:  getMaxRetries(queueName),
	}

	jobData, err := json.Marshal(job)
	if err != nil {
		return fmt.Errorf("failed to marshal payment job: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), JobTimeout)
	defer cancel()

	err = pq.redisClient.LPush(ctx, queueName, jobData).Err()
	if err != nil {
		return fmt.Errorf("failed to publish payment job to queue %s: %w", queueName, err)
	}

	return nil
}

func (pq *RedisPaymentQueue) ConsumePaymentJob(queueName string) (*PaymentJob, error) {
	ctx, cancel := context.WithTimeout(context.Background(), JobTimeout)
	defer cancel()

	result, err := pq.redisClient.BRPop(ctx, 0, queueName).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to consume payment job from queue %s: %w", queueName, err)
	}

	var job PaymentJob
	if err := json.Unmarshal([]byte(result[1]), &job); err != nil {
		return nil, fmt.Errorf("failed to unmarshal payment job from queue %s: %w", queueName, err)
	}

	return &job, nil
}

func (pq *RedisPaymentQueue) RequeueJob(job *PaymentJob, queueName string) error {
	jobData, err := json.Marshal(job)
	if err != nil {
		return fmt.Errorf("failed to marshal job for requeue: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), JobTimeout)
	defer cancel()

	return pq.redisClient.LPush(ctx, queueName, jobData).Err()
}

func generateJobID() string {
	return fmt.Sprintf("job_%d", time.Now().UnixNano())
}
