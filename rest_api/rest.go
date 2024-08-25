package restapi

import (
	"context"
	"encoding/json"
	"net/http"
	"time"
)

// OnHealthSubmit handles the health check request.
func OnHealthSubmit(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	select {
	case <-time.After(2 * time.Second): // Simulate some processing
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		dataF := HealthCheckResponse{Health: "OK"}
		json.NewEncoder(w).Encode(dataF)
	case <-ctx.Done(): // Handle timeout or cancellation
		http.Error(w, "Request Timeout", http.StatusRequestTimeout)
	}
}
