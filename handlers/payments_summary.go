package handlers

import (
	"encoding/json"
	"jvpayments/cache"
	"log"
	"net/http"
	"time"
)

func PaymentsSummary(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != "GET" {
		http.Error(w, `{"error": "Method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	q := r.URL.Query()
	var from, to string

	if fromStr := q.Get("from"); fromStr != "" {
		from = fromStr
	} else {
		from = time.Unix(0, 0).Format(time.RFC3339)
	}

	if toStr := q.Get("to"); toStr != "" {
		to = toStr
	} else {
		to = time.Now().Format(time.RFC3339)
	}

	paymentCacheService := cache.NewPaymentCache()
	result := map[string]any{}

	for _, paymentService := range []string{cache.PaymentDefaultKey, cache.PaymentFallbackKey} {
		payments, err := paymentCacheService.GetPaymentsByDateRange(from, to)
		if err != nil {
			log.Println(err)
			http.Error(w, `{"error": "Failed to query payments"}`, http.StatusInternalServerError)
			return
		}
		log.Println(payments)
		log.Println(paymentService)

		// totalRequests := 0
		// totalAmount := 0.0
		// for _, id := range payments {
		// 	if len(id) < len(svc)+1 || id[:len(svc)] != svc {
		// 		continue // skip if not this service
		// 	}
		// 	payment, err := paymentCacheService.GetPayment(svc, id[len(svc)+1:])
		// 	if err != nil || len(payment) == 0 {
		// 		continue
		// 	}
		// 	amount, _ := strconv.ParseFloat(payment["amount"], 64)
		// 	totalAmount += amount
		// 	totalRequests++
		// }
		// key := "default"
		// if svc == cache.PaymentFallbackKey {
		// 	key = "fallback"
		// }
		// result[key] = map[string]any{
		// 	"totalRequests": totalRequests,
		// 	"totalAmount":   totalAmount,
		// }
	}

	json.NewEncoder(w).Encode(result)
}
