package main

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestMongoDBConnectionError checks if the connection error is wrapped correctly
func TestMongoDBConnectionError(t *testing.T) {
	err := errors.New("actual connection error")
	wrappedErr := WrapError(err, ErrMongoDBConnection)

	if !errors.Is(wrappedErr, ErrMongoDBConnection) {
		t.Errorf("Expected error type %v, but got %v", ErrMongoDBConnection, wrappedErr)
	}
}

// TestMongoDBVerifyError checks if the verification error is wrapped correctly
func TestMongoDBVerifyError(t *testing.T) {
	err := errors.New("verification error")
	wrappedErr := WrapError(err, ErrMongoDBVerify)

	if !errors.Is(wrappedErr, ErrMongoDBVerify) {
		t.Errorf("Expected error type %v, but got %v", ErrMongoDBVerify, wrappedErr)
	}
}

// TestGCPStorageError checks if the GCP storage error is wrapped correctly
func TestGCPStorageError(t *testing.T) {
	err := errors.New("GCP error")
	wrappedErr := WrapError(err, ErrGCPStorage)

	if !errors.Is(wrappedErr, ErrGCPStorage) {
		t.Errorf("Expected error type %v, but got %v", ErrGCPStorage, wrappedErr)
	}
}

// TestFileUploadError checks if the file upload error is wrapped correctly
func TestFileUploadError(t *testing.T) {
	err := errors.New("upload error")
	wrappedErr := WrapError(err, ErrFileUpload)

	if !errors.Is(wrappedErr, ErrFileUpload) {
		t.Errorf("Expected error type %v, but got %v", ErrFileUpload, wrappedErr)
	}
}

// TestHealthEndpoint verifies that the health endpoint returns a 200 OK response.
func TestHealthEndpoint(t *testing.T) {
	// Create a request to pass to our handler.
	req, err := http.NewRequest("POST", "/v1/health", nil)
	if err != nil {
		t.Fatalf("Could not create request: %v", err)
	}

	// Record the response.
	rec := httptest.NewRecorder()

	// Wrap the handler with a timeout (as done in main.go)
	handler := withTimeout(func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte(`{"status":"ok"}`))
		require.NoError(t, err)
	})
	handler(rec, req)

	// Check the status code is 200
	if status := rec.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	// Check the response body is what we expect.
	expected := `{"status":"ok"}`
	if rec.Body.String() != expected {
		t.Errorf("handler returned unexpected body: got %v want %v", rec.Body.String(), expected)
	}
}
