package cache

import (
	"fmt"
	redis_client "jvpayments/redis"
	"time"

	"context"

	"github.com/redis/go-redis/v9"
)

const (
	PaymentDefaultKey  = "payments:default"
	PaymentFallbackKey = "payments:fallback"
	PaymentsByDateKey  = "payments:byDate"
)

type PaymentCache struct {
	redisClient *redis.Client
}

func NewPaymentCache() *PaymentCache {
	return &PaymentCache{
		redisClient: redis_client.RedisClient,
	}
}

func (pc *PaymentCache) StorePayment(paymentService string, correlationId string, amount float64, requestedAt time.Time) error {

	if paymentService != PaymentDefaultKey && paymentService != PaymentFallbackKey {
		panic("Invalid payment service")
	}

	ctx := context.Background()
	hashKey := fmt.Sprintf("%s:%s", paymentService, correlationId)

	_, err := pc.redisClient.HSet(ctx, hashKey, map[string]interface{}{
		"amount":      amount,
		"requestedAt": requestedAt.Format(time.RFC3339),
	}).Result()
	if err != nil {
		return err
	}

	_, err = pc.redisClient.ZAdd(ctx, PaymentsByDateKey, redis.Z{
		Score:  float64(requestedAt.Unix()),
		Member: hashKey,
	}).Result()
	return err
}

func (pc *PaymentCache) GetPaymentsByDateRange(start, end time.Time) ([]string, error) {
	ctx := context.Background()
	startScore := float64(start.Unix())
	endScore := float64(end.Unix())
	return pc.redisClient.ZRangeByScore(ctx, PaymentsByDateKey, &redis.ZRangeBy{
		Min: fmt.Sprintf("%f", startScore),
		Max: fmt.Sprintf("%f", endScore),
	}).Result()
}

func (pc *PaymentCache) GetPayment(paymentService string, correlationId string) (map[string]string, error) {
	ctx := context.Background()
	hashKey := fmt.Sprintf("%s:%s", paymentService, correlationId)
	return pc.redisClient.HGetAll(ctx, hashKey).Result()
}
