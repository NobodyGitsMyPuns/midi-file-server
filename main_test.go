package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

const (
	Acct = "112168818644504200034"
)

func TestReturnOne(t *testing.T) {
	if ReturnOne() != 1 {
		t.Error("Expected 1")
	}
}

func TestGetBucketLs(t *testing.T) {
	err := ListBucketContents("midi_file_storage")
	require.NoError(t, err)

}

func TestInitGCP(t *testing.T) {
	_, err := InitGCPWithServiceAccount(GCP_project, "/Users/jesselopez/Documents/repos/midi-file-server/gothic_key.json")
	require.NoError(t, err)
}

func TestConnectBluetooth(t *testing.T) {
	ip := "192.168.1.43"
	message := "Hello from Mac!"

	err := ConnectToESP32(ip, message)

	require.NoError(t, err)
}

func TestUploadListDelete(t *testing.T) {
	esp32IP := "192.168.1.43" // Replace with your ESP32 IP address
	filename := "Requiem_for_a_dream_mansell.mid"
	filePath := "/Users/jesselopez/Desktop/midi/Requiem_for_a_dream_mansell.mid" // Path to the local file you want to upload

	// 1. Upload the file
	t.Run("Upload File", func(t *testing.T) {
		url := fmt.Sprintf("http://%s/upload?filename=%s", esp32IP, filename)

		fileData, err := ioutil.ReadFile(filePath)
		if err != nil {
			t.Fatalf("Failed to read file: %v", err)
		}

		req, err := http.NewRequest("POST", url, bytes.NewBuffer(fileData))
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}
		req.Header.Set("Content-Type", "application/octet-stream")

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("Failed to send request: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("Unexpected status code: %d", resp.StatusCode)
		}

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("Failed to read response body: %v", err)
		}

		fmt.Printf("Upload Response: %s\n", body)
	})

	// 2. List files
	t.Run("List Files After Upload", func(t *testing.T) {
		url := fmt.Sprintf("http://%s/files", esp32IP)

		resp, err := http.Get(url)
		if err != nil {
			t.Fatalf("Failed to send request: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("Unexpected status code: %d", resp.StatusCode)
		}

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("Failed to read response body: %v", err)
		}

		fmt.Printf("Files After Upload:\n%s\n", body)

		if !strings.Contains(string(body), filename) {
			t.Fatalf("Uploaded file not found in file list")
		}
	})

	// 3. Delete the uploaded file
	t.Run("Delete File", func(t *testing.T) {
		url := fmt.Sprintf("http://%s/delete?name=%s", esp32IP, filename)

		req, err := http.NewRequest("DELETE", url, nil)
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("Failed to send request: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("Unexpected status code: %d", resp.StatusCode)
		}

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("Failed to read response body: %v", err)
		}

		fmt.Printf("Delete Response: %s\n", body)
	})

	// 4. List files again to confirm deletion
	t.Run("List Files After Deletion", func(t *testing.T) {
		url := fmt.Sprintf("http://%s/files", esp32IP)

		resp, err := http.Get(url)
		if err != nil {
			t.Fatalf("Failed to send request: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("Unexpected status code: %d", resp.StatusCode)
		}

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("Failed to read response body: %v", err)
		}

		fmt.Printf("Files After Deletion:\n%s\n", body)

		if strings.Contains(string(body), filename) {
			t.Fatalf("Deleted file still found in file list")
		}
	})
}
