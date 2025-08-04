package handlers

import (
	"jvpayments/internal/cache"

	"github.com/bytedance/sonic"
	"github.com/valyala/fasthttp"
)

type DbPurgeHandler struct {
	paymentCache *cache.PaymentCache
}

func NewDbPurgeHandler(paymentCache *cache.PaymentCache) *DbPurgeHandler {
	return &DbPurgeHandler{
		paymentCache: paymentCache,
	}
}

func (dph *DbPurgeHandler) DbPurge(ctx *fasthttp.RequestCtx) {
	if err := dph.paymentCache.DeleteAllData(); err != nil {
		ctx.SetStatusCode(fasthttp.StatusInternalServerError)
		response := map[string]string{"error": "Failed to purge payment keys"}
		if jsonData, err := sonic.Marshal(response); err == nil {
			ctx.SetContentType("application/json")
			ctx.SetBody(jsonData)
		}
		return
	}

	ctx.SetStatusCode(fasthttp.StatusOK)
	response := map[string]string{"message": "Payment keys purged"}
	if jsonData, err := sonic.Marshal(response); err == nil {
		ctx.SetContentType("application/json")
		ctx.SetBody(jsonData)
	}
}
