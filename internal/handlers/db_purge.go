package handlers

import (
	"context"
	"jvpayments/internal/cache"
	redis_client "jvpayments/internal/redis"
	"net/http"
)

type DbPurgeHandler struct {
	paymentCache *cache.PaymentCache
}

func NewDbPurgeHandler() *DbPurgeHandler {
	return &DbPurgeHandler{}
}

func (dph *DbPurgeHandler) DbPurge(w http.ResponseWriter, _ *http.Request) {
	ctx := context.Background()
	keys := []string{
		"payments:default",
		"payments:fallback",
		"payments:byDate",
		"payment:jobs",
		"payment:jobs:fallback",
	}
	if err := redis_client.RedisClient.Del(ctx, keys...).Err(); err != nil {
		http.Error(w, `{"error": "Failed to purge payment keys"}`, http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message": "Payment keys purged"}`))
}
