package main

import (
	"log"

	"jvpayments/internal/cache"
	"jvpayments/internal/config"
	"jvpayments/internal/handlers"
	"jvpayments/internal/queue"
	redis_client "jvpayments/internal/redis"
	"jvpayments/internal/services"
	workers "jvpayments/internal/workers/payment"

	"github.com/bytedance/sonic"
	"github.com/gofiber/fiber/v2"
)

func main() {
	config.LoadConfig()
	if err := redis_client.InitRedis(); err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	defer redis_client.CloseRedis()

	paymentCache := cache.NewPaymentCache()
	paymentQueue := queue.NewRedisPaymentQueue()
	paymentService := services.NewPaymentService(paymentQueue, paymentCache)

	for range 1000 {
		go workers.NewPaymentWorker(paymentQueue, paymentService)
	}

	app := fiber.New(fiber.Config{
		Prefork:     false,
		JSONEncoder: sonic.Marshal,
		JSONDecoder: sonic.Unmarshal,
	})

	app.Post("/payments", handlers.NewPaymentHandler(paymentService).Payments)
	app.Get("/payments-summary", handlers.NewPaymentSummaryHandler(paymentCache).PaymentsSummary)
	app.Post("/purge-payments", handlers.NewDbPurgeHandler(paymentCache).DbPurge)

	log.Println("Server starting on :3001")
	log.Fatal(app.Listen(":3001"))
}
