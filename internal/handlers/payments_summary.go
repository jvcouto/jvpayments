package handlers

import (
	"jvpayments/internal/cache"
	"log"
	"math"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
)

type PaymentSummaryHandler struct {
	paymentCache *cache.PaymentCache
}

func NewPaymentSummaryHandler(paymentCache *cache.PaymentCache) *PaymentSummaryHandler {
	return &PaymentSummaryHandler{
		paymentCache: paymentCache,
	}
}

func (psh *PaymentSummaryHandler) PaymentsSummary(c *fiber.Ctx) error {
	start := time.Now()
	defer func() {
		elapsed := time.Since(start)
		log.Printf("[PaymentSummary] Execution took %s", elapsed)
	}()

	fromStr := c.Query("from", "")
	toStr := c.Query("to", "")
	var from, to time.Time
	var err error

	if fromStr != "" {
		from, err = time.Parse(time.RFC3339, fromStr)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid 'from' timestamp"})
		}
	} else {
		from = time.Unix(0, 0)
	}

	if toStr != "" {
		to, err = time.Parse(time.RFC3339, toStr)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid 'to' timestamp"})
		}
	} else {
		to = time.Now()
	}

	result := map[string]any{}

	payments, err := psh.paymentCache.GetPaymentsByDateRange(from.UTC(), to.UTC())
	if err != nil {
		log.Println(err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to query payments"})
	}

	for _, paymentService := range []string{cache.PaymentDefaultKey, cache.PaymentFallbackKey} {
		totalRequests := 0
		totalAmount := 0.0
		for _, payment := range payments {
			if len(payment) < len(paymentService)+1 || payment[:len(paymentService)] != paymentService {
				continue
			}
			paymentValue := payment[len(paymentService)+1+36+1:]
			amount, _ := strconv.ParseFloat(paymentValue, 64)
			totalAmount += amount
			totalRequests++
		}
		key := "default"
		if paymentService == cache.PaymentFallbackKey {
			key = "fallback"
		}
		result[key] = map[string]any{
			"totalRequests": totalRequests,
			"totalAmount":   math.Round(totalAmount*100) / 100,
		}
	}

	return c.JSON(result)
}
