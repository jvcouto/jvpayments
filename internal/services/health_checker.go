package services

import (
	"context"
	"fmt"
	"jvpayments/internal/cache"
	"jvpayments/internal/config"
	redis_client "jvpayments/internal/redis"
	"jvpayments/internal/types"
	"log"
	"net/http"
	"time"

	"github.com/bsm/redislock"
	"github.com/bytedance/sonic"
	"github.com/redis/go-redis/v9"
)

const (
	PaymentServiceHealthCheckLockKey = "payment-service:health-check:lock"
)

type HeathCheckService struct {
	httpClient      *http.Client
	paymentCache    *cache.PaymentCache
	redisClient     *redis.Client
	ctx             context.Context
	cancel          context.CancelFunc
	serviceCheckers map[string]string
}

func NewHeathCheckService(paymentCache *cache.PaymentCache) *HeathCheckService {
	ctx, cancel := context.WithCancel(context.Background())
	return &HeathCheckService{
		httpClient:   &http.Client{},
		redisClient:  redis_client.RedisClient,
		paymentCache: paymentCache,
		ctx:          ctx,
		cancel:       cancel,
		serviceCheckers: map[string]string{
			"default":  config.CONFIG.PaymentApiUrl + "/payments/service-health",
			"fallback": config.CONFIG.PaymentApiFallbackUrl + "/payments/service-health",
		},
	}
}

func (hcs *HeathCheckService) Stop() {
	if hcs.cancel != nil {
		hcs.cancel()
	}
}

func (hcs *HeathCheckService) Start() error {
	locker := redislock.New(hcs.redisClient)

	_, err := locker.Obtain(hcs.ctx, PaymentServiceHealthCheckLockKey, 1*time.Minute, nil)
	if err == redislock.ErrNotObtained {
		return fmt.Errorf("another instance holds the lock")
	} else if err != nil {
		return err
	}

	hcs.startHealthCheckScheduler()

	return nil
}

func (hcs *HeathCheckService) startHealthCheckScheduler() {
	ticker := time.NewTicker(5 * time.Second)

	go func() {
		for {
			select {
			case <-ticker.C:
				log.Println("Processing heath check")
				err := hcs.runHealthCheck()
				if err != nil {
					log.Printf("error running heath check: %v", err)
				}
			case <-hcs.ctx.Done():
				ticker.Stop()
				return
			}
		}
	}()
}

func (hcs *HeathCheckService) runHealthCheck() error {
	for name, url := range hcs.serviceCheckers {
		resp, err := hcs.httpClient.Get(url)
		if err != nil {
			log.Printf("Health check failed for %s: %v", name, err)
			continue
		}

		var healthResponse types.HealthResponse
		if err := sonic.ConfigDefault.NewDecoder(resp.Body).Decode(&healthResponse); err != nil {
			log.Printf("Error decoding health for %s: %v", name, err)
			resp.Body.Close()
			continue
		}
		resp.Body.Close()

		err = hcs.paymentCache.UpdateApisStatus(name, healthResponse)
		if err != nil {
			log.Printf("Error updating api status for %s: %v", name, err)
		}
	}

	return nil
}
