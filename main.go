package main

import (
	"log"
	"net/http"

	"jvpayments/cache"
	"jvpayments/handlers"
)

func main() {
	if err := cache.InitRedis(); err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	defer cache.CloseRedis()

	mux := http.NewServeMux()

	mux.HandleFunc("/payments", handlers.Payments)
	mux.HandleFunc("/payments-summary", handlers.PaymentsSummary)

	log.Println("Server starting on :9999")
	log.Fatal(http.ListenAndServe(":9999", mux))
}
