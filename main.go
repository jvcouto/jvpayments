package main

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"jvpayments/handlers"
	redis_client "jvpayments/redis"
	"jvpayments/workers"
)

func main() {
	if err := redis_client.InitRedis(); err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	defer redis_client.CloseRedis()

	paymentWorker := workers.NewPaymentWorker()
	go paymentWorker.Start()

	// Setup graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Println("Shutting down gracefully...")
		paymentWorker.Stop()
		os.Exit(0)
	}()

	mux := http.NewServeMux()

	mux.HandleFunc("/payments", handlers.Payments)
	mux.HandleFunc("/payments-summary", handlers.PaymentsSummary)

	log.Println("Server starting on :9999")
	log.Fatal(http.ListenAndServe(":9999", mux))
}
