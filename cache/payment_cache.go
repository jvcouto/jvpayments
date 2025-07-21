package cache

import (
	"fmt"
	redis_client "jvpayments/redis"

	"context"
	"strconv"

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

func (pc *PaymentCache) StorePayment(paymentService string, correlationId string, amount float64, requestedAt string) error {

	if paymentService != PaymentDefaultKey && paymentService != PaymentFallbackKey {
		panic("Invalid payment service")
	}

	ctx := context.Background()
	hashKey := fmt.Sprintf("%s:%s", paymentService, correlationId)

	_, err := pc.redisClient.HSet(ctx, hashKey, map[string]interface{}{
		"amount":      amount,
		"requestedAt": requestedAt,
	}).Result()
	if err != nil {
		return err
	}
	score, err := strconv.ParseFloat(requestedAt, 64)
	if err != nil {
		return err
	}
	_, err = pc.redisClient.ZAdd(ctx, PaymentsByDateKey, redis.Z{
		Score:  score,
		Member: hashKey,
	}).Result()
	return err
}

func (pc *PaymentCache) GetPaymentsByDateRange(start, end string) ([]string, error) {
	ctx := context.Background()
	return pc.redisClient.ZRangeByScore(ctx, PaymentsByDateKey, &redis.ZRangeBy{
		Min: start,
		Max: end,
	}).Result()
}

func (pc *PaymentCache) GetPayment(paymentService string, correlationId string) (map[string]string, error) {
	ctx := context.Background()
	hashKey := fmt.Sprintf("%s:%s", paymentService, correlationId)
	return pc.redisClient.HGetAll(ctx, hashKey).Result()
}
