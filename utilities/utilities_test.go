package utilities

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/mongo"
)

func TestWrapError_WithContextInfo(t *testing.T) {
	originalErr := errors.New("original error")
	customErr := errors.New("custom error")

	resultErr := WrapError(originalErr, customErr, "additional context")
	expectedErr := fmt.Errorf("custom error: original error | Context: [additional context]")

	assert.EqualError(t, resultErr, expectedErr.Error())
}

func TestWrapError_NoContextInfo(t *testing.T) {
	originalErr := errors.New("original error")
	customErr := errors.New("custom error")

	resultErr := WrapError(originalErr, customErr)
	expectedErr := fmt.Errorf("custom error: original error")

	assert.EqualError(t, resultErr, expectedErr.Error())
}

func TestWrapError_NoError(t *testing.T) {
	resultErr := WrapError(nil, errors.New("custom error"))
	assert.Nil(t, resultErr)
}

// Test GetSignedTimeDurationMinutes
func TestGetSignedTimeDurationMinutes(t *testing.T) {
	os.Setenv("SIGNED_URL_EXPIRATION_MINUTES", "5")
	duration := GetSignedTimeDurationMinutes(os.Getenv("SIGNED_URL_EXPIRATION_MINUTES"))
	assert.Equal(t, 5*time.Minute, duration)
}

func TestWithTimeout(t *testing.T) {
	// Set an environment variable for testing
	os.Setenv("HTTP_CONTEXT_TIMEOUT", "1")

	handler := func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	WithTimeout(handler)(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestWithTimeoutDb(t *testing.T) {
	mockDB := new(mongo.Database)

	// Set environment variable for context timeout
	os.Setenv("HTTP_CONTEXT_TIMEOUT", "1")

	handler := func(ctx context.Context, db *mongo.Database, w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, mockDB, db)
		w.WriteHeader(http.StatusOK)
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	WithTimeoutDb(mockDB, handler)(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestWithSignedUrlDuration(t *testing.T) {
	handler := func(ctx context.Context, w http.ResponseWriter, r *http.Request, d time.Duration) {
		assert.Equal(t, 5*time.Second, d) // Check if duration is passed correctly
		w.WriteHeader(http.StatusOK)
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	WithSignedUrlDuration(5*time.Second, handler)(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestLogErrorAndRespond(t *testing.T) {
	w := httptest.NewRecorder()
	message := "Test error message"
	statusCode := http.StatusBadRequest

	LogErrorAndRespond(w, message, statusCode)

	resp := w.Result()
	defer resp.Body.Close()

	// Read the actual response body and trim the newline
	actualMessage := w.Body.String()
	expectedMessage := message + "\n"

	assert.Equal(t, statusCode, resp.StatusCode)
	assert.Equal(t, expectedMessage, actualMessage)
}
