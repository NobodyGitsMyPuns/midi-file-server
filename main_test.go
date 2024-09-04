package main

import (
	"errors"
	"midi-file-server/utilities"
	"testing"
)

// TestMongoDBConnectionError checks if the connection error is wrapped correctly
func TestMongoDBConnectionError(t *testing.T) {
	err := errors.New("actual connection error")
	wrappedErr := utilities.WrapError(err, ErrMongoDBConnection)

	if !errors.Is(wrappedErr, ErrMongoDBConnection) {
		t.Errorf("Expected error type %v, but got %v", ErrMongoDBConnection, wrappedErr)
	}
}

// TestMongoDBVerifyError checks if the verification error is wrapped correctly
func TestMongoDBVerifyError(t *testing.T) {
	err := errors.New("verification error")
	wrappedErr := utilities.WrapError(err, ErrMongoDBVerify)

	if !errors.Is(wrappedErr, ErrMongoDBVerify) {
		t.Errorf("Expected error type %v, but got %v", ErrMongoDBVerify, wrappedErr)
	}
}

// TestGCPStorageError checks if the GCP storage error is wrapped correctly
func TestGCPStorageError(t *testing.T) {
	err := errors.New("GCP error")
	wrappedErr := utilities.WrapError(err, ErrGCPStorage)

	if !errors.Is(wrappedErr, ErrGCPStorage) {
		t.Errorf("Expected error type %v, but got %v", ErrGCPStorage, wrappedErr)
	}
}

// TestFileUploadError checks if the file upload error is wrapped correctly
func TestFileUploadError(t *testing.T) {
	err := errors.New("upload error")
	wrappedErr := utilities.WrapError(err, ErrFileUpload)

	if !errors.Is(wrappedErr, ErrFileUpload) {
		t.Errorf("Expected error type %v, but got %v", ErrFileUpload, wrappedErr)
	}
}

// TestHealthEndpoint verifies that the health endpoint returns a 200 OK response.
// func TestHealthEndpoint(t *testing.T) {
// 	// Create a request to pass to our handler.
// 	req, err := http.NewRequest("POST", "/v1/health", nil)
// 	require.NoError(t, err, "Could not create request")

// 	// Record the response.
// 	rec := httptest.NewRecorder()

// 	// Wrap the handler with a timeout (as done in main.go)
// 	ctx := context.TODO()
// 	handler := withTimeout(ctx, func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
// 		w.WriteHeader(http.StatusOK)
// 		_, err := w.Write([]byte(`{"status":"ok"}`))
// 		require.NoError(t, err, "Could not write response")
// 	})
// 	handler(rec, req)

// 	// Check the status code is 200
// 	require.Equal(t, http.StatusOK, rec.Code, "handler returned wrong status code")

// 	// Check the response body is what we expect.
// 	expected := `{"status":"ok"}`
// 	require.Equal(t, expected, rec.Body.String(), "handler returned unexpected body")
// }
