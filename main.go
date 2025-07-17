package main

import (
	"log"
	"net/http"

	"jvpayments/handlers"
	"jvpayments/queue"
	redis_client "jvpayments/redis"
	"jvpayments/services"
	workers "jvpayments/workers/payment"
)

func main() {
	if err := redis_client.InitRedis(); err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	defer redis_client.CloseRedis()

	defaultBehavior := workers.NewDefaultWorkerBehavior(
		queue.NewRedisPaymentQueue(queue.PaymentQueueName),
		services.NewPaymentService(),
	)

	for i := 0; i < 10; i++ {
		go workers.NewPaymentWorker(queue.PaymentQueueName, defaultBehavior).Start()
	}

	fallbackBehavior := workers.NewDefaultWorkerBehavior(
		queue.NewRedisPaymentQueue(queue.PaymentFallabackQueueName),
		services.NewPaymentService(),
	)

	for i := 0; i < 10; i++ {
		go workers.NewPaymentWorker(queue.PaymentFallabackQueueName, fallbackBehavior).Start()
	}

	mux := http.NewServeMux()

	mux.HandleFunc("/payments", handlers.Payments)
	mux.HandleFunc("/payments-summary", handlers.PaymentsSummary)

	log.Println("Server starting on :9999")
	log.Fatal(http.ListenAndServe(":9999", mux))
}
