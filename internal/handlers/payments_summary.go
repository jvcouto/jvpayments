package handlers

import (
	"jvpayments/internal/cache"
	"log"
	"math"
	"strconv"
	"time"

	"github.com/bytedance/sonic"
	"github.com/valyala/fasthttp"
)

type PaymentSummaryHandler struct {
	paymentCache *cache.PaymentCache
}

func NewPaymentSummaryHandler(paymentCache *cache.PaymentCache) *PaymentSummaryHandler {
	return &PaymentSummaryHandler{
		paymentCache: paymentCache,
	}
}

func (psh *PaymentSummaryHandler) PaymentsSummary(ctx *fasthttp.RequestCtx) {
	fromStr := string(ctx.QueryArgs().Peek("from"))
	toStr := string(ctx.QueryArgs().Peek("to"))
	var from, to time.Time
	var err error

	if fromStr != "" {
		from, err = time.Parse(time.RFC3339, fromStr)
		if err != nil {
			ctx.SetStatusCode(fasthttp.StatusBadRequest)
			response := map[string]string{"error": "Invalid 'from' timestamp"}
			if jsonData, err := sonic.Marshal(response); err == nil {
				ctx.SetContentType("application/json")
				ctx.SetBody(jsonData)
			}
			return
		}
	} else {
		from = time.Unix(0, 0)
	}

	if toStr != "" {
		to, err = time.Parse(time.RFC3339, toStr)
		if err != nil {
			ctx.SetStatusCode(fasthttp.StatusBadRequest)
			response := map[string]string{"error": "Invalid 'to' timestamp"}
			if jsonData, err := sonic.Marshal(response); err == nil {
				ctx.SetContentType("application/json")
				ctx.SetBody(jsonData)
			}
			return
		}
	} else {
		to = time.Now()
	}

	result := map[string]any{}

	payments, err := psh.paymentCache.GetPaymentsByDateRange(from.UTC(), to.UTC())
	if err != nil {
		log.Println(err)
		ctx.SetStatusCode(fasthttp.StatusInternalServerError)
		response := map[string]string{"error": "Failed to query payments"}
		if jsonData, err := sonic.Marshal(response); err == nil {
			ctx.SetContentType("application/json")
			ctx.SetBody(jsonData)
		}
		return
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

	if jsonData, err := sonic.Marshal(result); err == nil {
		ctx.SetContentType("application/json")
		ctx.SetBody(jsonData)
	}
}
