package main

import (
	"log"
	"net/http"

	"jvpayments/internal/cache"
	"jvpayments/internal/config"
	"jvpayments/internal/handlers"
	"jvpayments/internal/queue"
	redis_client "jvpayments/internal/redis"
	"jvpayments/internal/services"
)

func main() {
	config.LoadConfig()
	if err := redis_client.InitRedis(); err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	defer redis_client.CloseRedis()

	// defaultBehavior := workers.NewDefaultWorkerBehavior(
	// 	queue.NewRedisPaymentQueue(),
	// 	services.NewPaymentService(),
	// 	cache.NewPaymentCache(),
	// )

	// for range 25 {
	// 	go workers.NewPaymentWorker(queue.PaymentQueueName, defaultBehavior).Start()
	// }

	// fallbackBehavior := workers.NewDefaultWorkerBehavior(
	// 	queue.NewRedisPaymentQueue(queue.PaymentFallabackQueueName),
	// 	services.NewPaymentService(),
	// 	cache.NewPaymentCache(),
	// )

	// for range 10 {
	// 	go workers.NewPaymentWorker(queue.PaymentFallabackQueueName, fallbackBehavior).Start()
	// }

	paymentCache := cache.NewPaymentCache()
	paymentQueue := queue.NewRedisPaymentQueue()
	paymentService := services.NewPaymentService(paymentQueue, paymentCache)

	paymentHandler := handlers.NewPaymentHandler(paymentService)
	paymentSummaryHandler := handlers.NewPaymentSummaryHandler(paymentCache)
	dbPurgeHandler := handlers.NewDbPurgeHandler()

	mux := http.NewServeMux()
	mux.HandleFunc("/payments", paymentHandler.Payments)
	mux.HandleFunc("/payments-summary", paymentSummaryHandler.PaymentsSummary)
	mux.HandleFunc("/db-purge", dbPurgeHandler.DbPurge)

	log.Println("Server starting on :3001")
	log.Fatal(http.ListenAndServe(":3001", mux))
}
