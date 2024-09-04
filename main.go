package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	mongodb "midi-file-server/mongo_db"
	restapi "midi-file-server/rest_api"

	"cloud.google.com/go/storage"
	"google.golang.org/api/option"
)

var (
	// Define custom errors
	ErrMongoDBConnection     = errors.New("failed to connect to MongoDB")
	ErrMongoDBVerify         = errors.New("failed to verify MongoDB")
	ErrGCPStorage            = errors.New("failed to initialize Google Cloud Storage")
	ErrFileUpload            = errors.New("failed to upload file")
	ErrFileOpen              = errors.New("failed to open file")
	ErrFileClose             = errors.New("failed to close file")
	VersionEp                = "v1"
	HealthEp                 = "health"
	RegisterEp               = "register"
	LoginEp                  = "login"
	GetSignedUrl             = "get-signed-url"
	ListAvailableMidiBuckets = "list-available-midi-files"
	ContextTimeout           = 60 * time.Second
)

func main() {
	timedContext, cancel := context.WithTimeout(context.Background(), ContextTimeout)
	defer cancel()

	mongoDB := mongodb.NewMongoDBClient(timedContext)

	err := WrapError(mongoDB.Connect(), ErrMongoDBConnection)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}
	defer mongoDB.Disconnect()

	err = WrapError(mongoDB.VerifyDB(), ErrMongoDBVerify)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}

	// Register handlers with the shared context
	http.HandleFunc(fmt.Sprintf("/%s/%s", VersionEp, HealthEp), withTimeout(timedContext, restapi.OnHealthSubmit))
	http.HandleFunc(fmt.Sprintf("/%s/%s", VersionEp, GetSignedUrl), withTimeout(timedContext, restapi.GetSignedUrl))
	http.HandleFunc(fmt.Sprintf("/%s/%s", VersionEp, ListAvailableMidiBuckets), withTimeout(timedContext, restapi.ListBucketHandler))
	http.HandleFunc(fmt.Sprintf("/%s/%s", VersionEp, RegisterEp), withTimeout(timedContext, restapi.RegisterUser))
	http.HandleFunc(fmt.Sprintf("/%s/%s", VersionEp, LoginEp), withTimeout(timedContext, restapi.LoginUser))

	log.Fatal(http.ListenAndServe(":8080", nil))
}

// Modified withTimeout to accept an external context
func withTimeout(parentCtx context.Context, handler func(context.Context, http.ResponseWriter, *http.Request)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		timedContext, cancel := context.WithTimeout(parentCtx, ContextTimeout)
		defer cancel()
		handler(timedContext, w, r)
	}
}

// Modified InitGCPWithServiceAccount to accept a context
func InitGCPWithServiceAccount(ctx context.Context, serviceAccountID, keyFilePath string) (*storage.Client, error) {
	client, err := storage.NewClient(ctx, option.WithCredentialsFile(keyFilePath))
	if err != nil {
		return nil, WrapError(err, ErrGCPStorage)
	}

	fmt.Println("GCP credentials initialized successfully with service account")
	return client, nil
}

// UploadFiles modified to accept a context and pass it through
func UploadFiles(ctx context.Context, bucketName, prefix string, filePaths []string) error {
	client, err := storage.NewClient(ctx)
	if err != nil {
		return WrapError(err, ErrGCPStorage)
	}
	defer client.Close()

	for _, filePath := range filePaths {
		fileName := filepath.Base(filePath)
		objectPath := prefix + "/" + fileName
		wc := client.Bucket(bucketName).Object(objectPath).NewWriter(ctx)

		file, err := os.Open(filePath)
		if err != nil {
			return WrapError(fmt.Errorf("file: %s", filePath), ErrFileOpen)
		}
		defer file.Close()

		if _, err = io.Copy(wc, file); err != nil {
			return WrapError(err, ErrFileUpload)
		}

		if err := wc.Close(); err != nil {
			return WrapError(err, ErrFileClose)
		}

		fmt.Printf("File %s uploaded successfully to bucket %s\n", fileName, bucketName)
	}

	return nil
}

// WrapError wraps a custom error around an original error, preserving the custom error type for testing
func WrapError(err error, customErr error) error {
	if err != nil {
		return fmt.Errorf("%w: %v", customErr, err)
	}
	return nil
}
