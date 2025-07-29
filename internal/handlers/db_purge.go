package handlers

import (
	"jvpayments/internal/cache"

	"github.com/gofiber/fiber/v2"
)

type DbPurgeHandler struct {
	paymentCache *cache.PaymentCache
}

func NewDbPurgeHandler(paymentCache *cache.PaymentCache) *DbPurgeHandler {
	return &DbPurgeHandler{
		paymentCache: paymentCache,
	}
}

func (dph *DbPurgeHandler) DbPurge(c *fiber.Ctx) error {

	if err := dph.paymentCache.DeleteAllData(); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to purge payment keys"})

	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "Payment keys purged"})
}
