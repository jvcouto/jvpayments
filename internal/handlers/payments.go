package handlers

import (
	"jvpayments/internal/queue"
	"jvpayments/internal/services"
	"jvpayments/internal/types"
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
)

type WorkerPool struct {
	NumWorkers int
	Jobs       chan types.PaymentRequest
}

func (wp *WorkerPool) Start(ps *services.PaymentService, paymentQueue *queue.RedisPaymentQueue) {
	for i := 0; i < wp.NumWorkers; i++ {
		go func(workerID int) {
			for job := range wp.Jobs {
				err := ps.ProcessPayment(job)
				if err != nil {
					log.Printf("error processing payment: %v", err)
					queueErr := paymentQueue.PublishPaymentJob(job)
					if queueErr != nil {
						log.Printf("failed to make payment request: %v", queueErr)
					}
				}
			}
		}(i)
	}
}

type PaymentHandler struct {
	workerPool WorkerPool
}

func NewPaymentHandler(paymentService *services.PaymentService, paymentQueue *queue.RedisPaymentQueue) *PaymentHandler {
	wp := WorkerPool{NumWorkers: 350, Jobs: make(chan types.PaymentRequest, 1750)}
	wp.Start(paymentService, paymentQueue)
	return &PaymentHandler{
		workerPool: wp,
	}
}

func (ph *PaymentHandler) Payments(c *fiber.Ctx) error {
	start := time.Now()
	defer func() {
		elapsed := time.Since(start)
		log.Printf("[Payments]Execution took %s", elapsed)
	}()

	log.Println("New payment request received")

	var paymentReq types.PaymentRequest

	if err := c.BodyParser(&paymentReq); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	// if err := validatePaymentRequest(paymentReq); err != nil {
	// 	return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid payment data"})
	// }

	ph.workerPool.Jobs <- paymentReq
	log.Printf("[Payments] Worker pool jobs queue size: %d", len(ph.workerPool.Jobs))

	return c.SendStatus(fiber.StatusNoContent)
}

// func validatePaymentRequest(req types.PaymentRequest) error {
// 	// TODO: Implement validation logic
// 	return nil
// }
