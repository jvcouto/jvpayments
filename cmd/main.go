package main

import (
	"log"
	"os"
	"time"

	"jvpayments/internal/cache"
	"jvpayments/internal/config"
	"jvpayments/internal/handlers"
	"jvpayments/internal/queue"
	redis_client "jvpayments/internal/redis"
	"jvpayments/internal/services"
	workers "jvpayments/internal/workers/payment"

	"github.com/valyala/fasthttp"
)

func main() {
	socketPath := os.Getenv("SOCKET_PATH")

	config.LoadConfig()
	if err := redis_client.InitRedis(); err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	defer redis_client.CloseRedis()

	paymentCache := cache.NewPaymentCache()
	paymentQueue := queue.NewRedisPaymentQueue()
	paymentService := services.NewPaymentService(paymentQueue, paymentCache)

	healthCheckService := services.NewHeathCheckService(paymentCache)
	healthCheckService.Start()
	defer healthCheckService.Stop()

	paymentWorkers := workers.NewPaymentWorker(paymentQueue, paymentService)
	for range 5 {
		go paymentWorkers.Start()
	}

	paymentHandler := handlers.NewPaymentHandler(paymentService, paymentQueue)
	paymentSummaryHandler := handlers.NewPaymentSummaryHandler(paymentCache)
	dbPurgeHandler := handlers.NewDbPurgeHandler(paymentCache)

	requestHandler := func(ctx *fasthttp.RequestCtx) {
		path := string(ctx.Path())
		method := string(ctx.Method())

		switch {
		case method == "POST" && path == "/payments":
			paymentHandler.Payments(ctx)
		case method == "GET" && path == "/payments-summary":
			paymentSummaryHandler.PaymentsSummary(ctx)
		case method == "POST" && path == "/purge-payments":
			dbPurgeHandler.DbPurge(ctx)
		default:
			ctx.SetStatusCode(fasthttp.StatusNotFound)
			ctx.SetBodyString("Not Found")
		}
	}

	server := &fasthttp.Server{
		Handler:           requestHandler,
		TCPKeepalive:      true,
		DisableKeepalive:  false,
		ReduceMemoryUsage: true,
		IdleTimeout:       60 * time.Second,
	}

	log.Printf("Server starting on Unix socket: %s", socketPath)
	log.Fatal(server.ListenAndServeUNIX(socketPath, 0666))
}
