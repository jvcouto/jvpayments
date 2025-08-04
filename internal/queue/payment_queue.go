package queue

import (
	"context"
	"fmt"
	redis_client "jvpayments/internal/redis"
	"jvpayments/internal/types"
	"time"

	"github.com/bytedance/sonic"
	"github.com/redis/go-redis/v9"
)

const (
	PaymentQueueName = "payment:jobs"
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

func NewRedisPaymentQueue() *RedisPaymentQueue {
	return &RedisPaymentQueue{
		redisClient: redis_client.RedisClient,
	}
}

func (pq *RedisPaymentQueue) PublishPaymentJob(paymentReq types.PaymentRequest) error {

	job := PaymentJob{
		ID:          generateJobID(),
		PaymentData: paymentReq,
		RetryCount:  0,
		MaxRetries:  12,
	}

	jobData, err := sonic.Marshal(job)
	if err != nil {
		return fmt.Errorf("failed to marshal payment job: %w", err)
	}

	ctx := context.Background()

	err = pq.redisClient.LPush(ctx, PaymentQueueName, jobData).Err()
	if err != nil {
		return fmt.Errorf("failed to publish payment job to queue: %w", err)
	}

	return nil
}

func (pq *RedisPaymentQueue) ConsumePaymentJob() (*PaymentJob, error) {
	ctx := context.Background()

	result, err := pq.redisClient.BRPop(ctx, 0, PaymentQueueName).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to consume payment job from queue: %w", err)
	}

	var job PaymentJob
	if err := sonic.Unmarshal([]byte(result[1]), &job); err != nil {
		return nil, fmt.Errorf("failed to unmarshal payment job from queue: %w", err)
	}

	return &job, nil
}

func (pq *RedisPaymentQueue) RequeueJob(job *PaymentJob) error {
	jobData, err := sonic.Marshal(job)
	if err != nil {
		return fmt.Errorf("failed to marshal job for requeue: %w", err)
	}

	ctx := context.Background()

	return pq.redisClient.LPush(ctx, PaymentQueueName, jobData).Err()
}

func generateJobID() string {
	return fmt.Sprintf("job_%d", time.Now().UnixNano())
}
