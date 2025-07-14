package handlers

import (
	"net/http"
)

func Payments(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, `{"error": "Method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
