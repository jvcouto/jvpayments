package handlers

import (
	"encoding/json"
	"net/http"
)

func PaymentsSummary(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != "GET" {
		http.Error(w, `{"error": "Method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	response := map[string]any{
		"summary": map[string]any{
			"default": map[string]any{
				"totalRequests": 43236,
				"totalAmount":   415542345.98,
			},
			"fallback": map[string]any{
				"totalRequests": 423545,
				"totalAmount":   329347.34,
			},
		},
	}

	json.NewEncoder(w).Encode(response)
}
