package main

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/integration/mtest"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	Acct = "112168818644504200034"
)

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

// mongoDB
// Connect to MongoDB
func connectMongoDB() (*mongo.Client, error) {
	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")
	client, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		return nil, err
	}
	return client, nil
}

// Register a user
func RegisterUser(client *mongo.Client, serialNumber, username, password string) error {
	collection := client.Database("testdb").Collection("users")
	user := bson.D{
		{"serial_number", serialNumber},
		{"username", username},
		{"password", password},
	}
	_, err := collection.InsertOne(context.TODO(), user)
	return err
}

// Login a user
func LoginUser(client *mongo.Client, username, password string) (bool, error) {
	collection := client.Database("testdb").Collection("users")
	filter := bson.D{{"username", username}, {"password", password}}
	var result bson.D
	err := collection.FindOne(context.TODO(), filter).Decode(&result)
	if err == mongo.ErrNoDocuments {
		return false, nil
	}
	return err == nil, err
}

func TestRegisterUser(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	mt.Run("register user", func(mt *mtest.T) {
		client, err := connectMongoDB()
		if err != nil {
			t.Fatalf("Failed to connect to MongoDB: %v", err)
		}

		err = RegisterUser(client, "12345", "testuser", "testpass")
		if err != nil {
			t.Errorf("Failed to register user: %v", err)
		}
	})
}

func TestLoginUser(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	mt.Run("login user", func(mt *mtest.T) {
		client, err := connectMongoDB()
		if err != nil {
			t.Fatalf("Failed to connect to MongoDB: %v", err)
		}

		// First, register a user
		err = RegisterUser(client, "12345", "testuser", "testpass")
		if err != nil {
			t.Fatalf("Failed to register user: %v", err)
		}

		// Now, try to log in
		success, err := LoginUser(client, "testuser", "testpass")
		if err != nil {
			t.Errorf("Failed to log in user: %v", err)
		}

		if !success {
			t.Errorf("Expected login to succeed, but it failed")
		}
	})
}
