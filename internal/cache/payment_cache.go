package cache

import (
	"fmt"
	redis_client "jvpayments/internal/redis"
	"jvpayments/internal/types"
	"time"

	"context"

	"github.com/redis/go-redis/v9"
)

const (
	PaymentDefaultKey  = "payments:default"
	PaymentFallbackKey = "payments:fallback"
	PaymentsByDateKey  = "payments:byDate"

	DefaultPaymentApiStatus  = "default:api:status"
	FallbackPaymentApiStatus = "fallback:api:status"
)

type PaymentCache struct {
	redisClient *redis.Client
}

func NewPaymentCache() *PaymentCache {
	return &PaymentCache{
		redisClient: redis_client.RedisClient,
	}
}

func (pc *PaymentCache) StorePayment(paymentService string, paymentData types.PaymentRequest) error {

	ctx := context.Background()

	// hashKey := fmt.Sprintf("%s:%s", paymentService, correlationId)
	// _, err := pc.redisClient.HSet(ctx, hashKey, map[string]interface{}{
	// 	"amount":      amount,
	// 	"requestedAt": requestedAt.Format(time.RFC3339),
	// }).Result()
	// if err != nil {
	// 	return err
	// }

	scoreMember := fmt.Sprintf("%s:%s:%f", paymentService, paymentData.CorrelationId, paymentData.Amount)

	_, err := pc.redisClient.ZAdd(ctx, PaymentsByDateKey, redis.Z{
		Score:  float64(paymentData.RequestedAt.Unix()),
		Member: scoreMember,
	}).Result()

	if err != nil {
		return fmt.Errorf("error persisting payment: %w", err)
	}

	return nil

}

func (pc *PaymentCache) GetPaymentsByDateRange(start, end time.Time) ([]string, error) {
	ctx := context.Background()

	startScore := float64(start.Unix())
	endScore := float64(end.Unix())
	return pc.redisClient.ZRangeByScore(ctx, PaymentsByDateKey, &redis.ZRangeBy{
		Min: fmt.Sprintf("(%f", startScore),
		Max: fmt.Sprintf("%f", endScore),
	}).Result()
}

// func (pc *PaymentCache) GetPayment(paymentService string, correlationId string) (map[string]string, error) {
// 	ctx := context.Background()
// 	hashKey := fmt.Sprintf("%s:%s", paymentService, correlationId)
// 	return pc.redisClient.HGetAll(ctx, hashKey).Result()
// }

func (pc *PaymentCache) DeleteAllData() error {
	ctx := context.Background()
	keys := []string{
		"payments:default",
		"payments:fallback",
		"payments:byDate",
		"payment:jobs",
		"payment:jobs:fallback",
	}

	if err := pc.redisClient.Del(ctx, keys...).Err(); err != nil {
		return fmt.Errorf("error deleting all data: %w", err)
	}

	return nil
}

func (pc *PaymentCache) GetApisStatus() (map[string]any, error) {
	ctx := context.Background()
	apisStatus, err := pc.redisClient.MGet(ctx, DefaultPaymentApiStatus, FallbackPaymentApiStatus).Result()

	if err != nil {
		return nil, fmt.Errorf("error getting apis status: %w", err)
	}

	result := make(map[string]any)
	apiKeys := []string{"default", "fallback"}

	for i, status := range apisStatus {

		if status == nil {
			result[apiKeys[i]] = ""
			continue
		}

		var failing bool
		var minResponseTime int

		if statusStr, ok := status.(string); ok {
			fmt.Sscanf(statusStr, "%t:%d", &failing, &minResponseTime)
		}

		result[apiKeys[i]] = map[string]any{
			"failing":         failing,
			"minResponseTime": minResponseTime,
		}
	}

	return result, nil
}

func (pc *PaymentCache) UpdateApisStatus(paymentService string, healthResponse types.HealthResponse) error {
	ctx := context.Background()

	var statusKey string
	switch paymentService {
	case "default":
		statusKey = DefaultPaymentApiStatus
	case "fallback":
		statusKey = FallbackPaymentApiStatus
	default:
		return fmt.Errorf("unknown payment service: %s", paymentService)
	}

	statusValue := fmt.Sprintf("%t:%d", healthResponse.Failing, healthResponse.MinResponseTime)

	_, err := pc.redisClient.Set(ctx, statusKey, statusValue, 0).Result()
	if err != nil {
		return fmt.Errorf("error updating api status for %s: %w", paymentService, err)
	}

	return nil
}
