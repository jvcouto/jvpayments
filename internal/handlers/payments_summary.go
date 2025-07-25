package handlers

import (
	"encoding/json"
	"jvpayments/internal/cache"
	"log"
	"net/http"
	"strconv"
	"time"
)

type PaymentSummaryHandler struct {
	paymentCache *cache.PaymentCache
}

func NewPaymentSummaryHandler(paymentCache *cache.PaymentCache) *PaymentSummaryHandler {
	return &PaymentSummaryHandler{
		paymentCache: paymentCache,
	}
}

func (psh *PaymentSummaryHandler) PaymentsSummary(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != "GET" {
		http.Error(w, `{"error": "Method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	q := r.URL.Query()
	var from, to time.Time

	if fromStr := q.Get("from"); fromStr != "" {
		fromTimeStamp, err := time.Parse(time.RFC3339, fromStr)
		if err != nil {
			http.Error(w, `{"error": "Invalid 'from' timestamp"}`, http.StatusBadRequest)
			return
		}
		from = fromTimeStamp
	} else {
		from = time.Unix(0, 0)
	}

	if toStr := q.Get("to"); toStr != "" {
		toTimeStamp, err := time.Parse(time.RFC3339, toStr)
		if err != nil {
			http.Error(w, `{"error": "Invalid 'from' timestamp"}`, http.StatusBadRequest)
			return
		}
		to = toTimeStamp
	} else {
		to = time.Now()
	}

	result := map[string]any{}

	payments, err := psh.paymentCache.GetPaymentsByDateRange(from, to)
	if err != nil {
		log.Println(err)
		http.Error(w, `{"error": "Failed to query payments"}`, http.StatusInternalServerError)
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
			"totalAmount":   totalAmount,
		}
	}

	json.NewEncoder(w).Encode(result)
}
