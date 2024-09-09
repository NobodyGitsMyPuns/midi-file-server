package restapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOnHealthSubmit_Success(t *testing.T) {
	// Simulate an HTTP POST request
	req := httptest.NewRequest(http.MethodPost, "/health", nil)
	w := httptest.NewRecorder()

	// Call the function
	OnHealthSubmit(req.Context(), w, req)

	// Validate the response
	resp := w.Result()
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var response HealthCheckResponse
	_ = json.NewDecoder(resp.Body).Decode(&response)

	assert.Equal(t, "OK", response.Health)
}

func TestOnHealthSubmit_MethodNotAllowed(t *testing.T) {
	// Simulate an HTTP GET request (invalid method)
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	// Call the function
	OnHealthSubmit(req.Context(), w, req)

	// Validate the response
	resp := w.Result()
	defer resp.Body.Close()

	assert.Equal(t, http.StatusMethodNotAllowed, resp.StatusCode)
}
