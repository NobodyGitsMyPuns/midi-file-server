package main

import (
	"context"
	"fmt"
	"io"
	"log"
	restapi "midi-file-server/rest_api"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"cloud.google.com/go/storage"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

var (
	HealthEp   = "health"
	VqaEp      = "vqa"
	ValidateEp = "validate"
	VersionEp  = "v1"
)

func main() {
	healthEp := fmt.Sprintf("/%s/%s", VersionEp, HealthEp)
	fmt.Printf("Starting server on http://localhost:8080%s\n", healthEp)

	http.HandleFunc(healthEp, func(w http.ResponseWriter, r *http.Request) {
		timedContext, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()
		restapi.OnHealthSubmit(timedContext, w, r)
	})

	log.Fatal(http.ListenAndServe(":8080", nil))
}

func ReturnOne() int {
	return 1
}

type GCPStorageManager struct {
	client *storage.Client
}

const (
	GCP_project = "gothic-oven-433521-e1"
)

func ListBucketContents(bucketName string) error {
	ctx := context.Background()

	// Initialize the client
	client, err := storage.NewClient(ctx)
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}
	defer client.Close()

	// Check if the client is nil
	if client == nil {
		return fmt.Errorf("storage client is nil")
	}

	bucket := client.Bucket(bucketName)

	// Check if the bucket is nil
	if bucket == nil {
		return fmt.Errorf("bucket %s is nil", bucketName)
	}

	it := bucket.Objects(ctx, nil)
	for {
		attrs, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to list objects: %w", err)
		}
		fmt.Println(attrs.Name)
	}

	return nil
}

// InitGCPWithServiceAccount initializes the GCP client using a service account ID and key file.
func InitGCPWithServiceAccount(serviceAccountID, keyFilePath string) (*storage.Client, error) {
	ctx := context.Background()

	// Optionally, log the service account ID for debugging purposes (not generally needed for authentication)
	fmt.Printf("Initializing GCP with service account: %s\n", serviceAccountID)

	// Initialize the storage client using the service account key file
	client, err := storage.NewClient(ctx, option.WithCredentialsFile(keyFilePath))
	if err != nil {
		return nil, fmt.Errorf("failed to create storage client: %w", err)
	}

	// Optionally, list buckets or perform other verification steps to confirm credentials
	it := client.Buckets(ctx, "your-project-id")
	for {
		bucketAttrs, err := it.Next()
		if err != nil {
			break // No more buckets, exit the loop
		}
		if err != nil {
			return nil, fmt.Errorf("error iterating buckets: %w", err)
		}
		fmt.Println("Found bucket:", bucketAttrs.Name)
	}

	fmt.Println("GCP credentials initialized successfully with service account")
	return client, nil
}

// UploadFiles uploads one or more files to a Google Cloud Storage bucket using a specified prefix.
func UploadFiles(bucketName, prefix string, filePaths []string) error {
	ctx := context.Background()

	// Initialize the client
	client, err := storage.NewClient(ctx)
	if err != nil {
		return fmt.Errorf("failed to create storage client: %w", err)
	}
	defer client.Close()

	for _, filePath := range filePaths {
		// Open the file
		file, err := os.Open(filePath)
		if err != nil {
			return fmt.Errorf("failed to open file %s: %w", filePath, err)
		}
		defer file.Close()

		// Get the file name from the file path
		fileName := filepath.Base(filePath)

		// Create a handle to the destination object in the bucket
		objectPath := prefix + "/" + fileName
		wc := client.Bucket(bucketName).Object(objectPath).NewWriter(ctx)

		// Copy the file content to the GCS object
		if _, err = io.Copy(wc, file); err != nil {
			return fmt.Errorf("failed to upload file %s: %w", fileName, err)
		}

		// Close the writer to complete the upload
		if err := wc.Close(); err != nil {
			return fmt.Errorf("failed to complete upload for file %s: %w", fileName, err)
		}

		fmt.Printf("File %s uploaded successfully to bucket %s\n", fileName, bucketName)
	}

	return nil
}

// DeleteFile deletes a file from a Google Cloud Storage bucket using its name and prefix.
func DeleteFile(bucketName, prefix, fileName string) error {
	ctx := context.Background()

	// Initialize the client
	client, err := storage.NewClient(ctx)
	if err != nil {
		return fmt.Errorf("failed to create storage client: %w", err)
	}
	defer client.Close()

	// Create a handle to the file (object) in the bucket
	objectPath := prefix + "/" + fileName
	obj := client.Bucket(bucketName).Object(objectPath)

	// Delete the file
	if err := obj.Delete(ctx); err != nil {
		return fmt.Errorf("failed to delete file %s: %w", fileName, err)
	}

	fmt.Printf("File %s deleted successfully from bucket %s\n", fileName, bucketName)
	return nil
}
