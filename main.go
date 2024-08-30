package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	restapi "midi-file-server/rest_api"

	"cloud.google.com/go/storage"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

var (
	HealthEp     = "health"
	VersionEp    = "v1"
	RegisterEp   = "register"
	LoginEp      = "login"
	GetSignedUrl = "get-signed-url"
)

func main() {
	// Connect to MongoDB
	client, err := connectMongoDB()
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer func() {
		if err := client.Disconnect(context.TODO()); err != nil {
			log.Fatalf("Failed to disconnect from MongoDB: %v", err)
		} else {
			fmt.Println("Disconnected from MongoDB successfully.")
		}
	}()

	// Ensure that the necessary database and collections exist
	ensureDatabaseAndCollections(client)

	healthEp := fmt.Sprintf("/%s/%s", VersionEp, HealthEp)
	registerEp := fmt.Sprintf("/%s/%s", VersionEp, RegisterEp)
	loginEp := fmt.Sprintf("/%s/%s", VersionEp, LoginEp)
	getSignedUrlEp := fmt.Sprintf("/%s/%s", VersionEp, GetSignedUrl)

	log.Println("Starting server on " + healthEp + "\n")

	http.HandleFunc(healthEp, func(w http.ResponseWriter, r *http.Request) {
		timedContext, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()
		restapi.OnHealthSubmit(timedContext, w, r)
	})

	http.HandleFunc(registerEp, func(w http.ResponseWriter, r *http.Request) {
		timedContext, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()
		restapi.RegisterUser(timedContext, w, r)
	})

	http.HandleFunc(loginEp, func(w http.ResponseWriter, r *http.Request) {
		timedContext, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()
		restapi.LoginUser(timedContext, w, r)
	})

	http.HandleFunc(getSignedUrlEp, func(w http.ResponseWriter, r *http.Request) {
		timedContext, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()
		restapi.GetSignedUrl(timedContext, w, r)
	})

	log.Fatal(http.ListenAndServe(":8080", nil))
}

// connectMongoDB connects to MongoDB using the appropriate URI.
func connectMongoDB() (*mongo.Client, error) {
	clientOptions := options.Client().ApplyURI("mongodb://mongodb-service:27017") // Replace with your MongoDB service URI
	client, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		return nil, err
	}

	// Check the connection
	err = client.Ping(context.TODO(), nil)
	if err != nil {
		return nil, err
	}

	fmt.Println("Connected to MongoDB!")
	return client, nil
}

// ensureDatabaseAndCollections ensures that the necessary database and collections exist.
func ensureDatabaseAndCollections(client *mongo.Client) {
	database := client.Database("testdb")

	// Check if the 'users' collection exists
	collectionNames, err := database.ListCollectionNames(context.TODO(), bson.D{{Key: "name", Value: "users"}})
	if err != nil {
		log.Fatalf("Failed to list collections: %v", err)
	}

	if len(collectionNames) == 0 {
		// The 'users' collection does not exist, so we create it
		fmt.Println("Creating 'users' collection...")

		// Create the 'users' collection
		err := database.CreateCollection(context.TODO(), "users")
		if err != nil {
			log.Fatalf("Failed to create 'users' collection: %v", err)
		}
	} else {
		fmt.Println("'users' collection already exists.")
	}

	// Optionally, create an index on the 'username' field to ensure uniqueness
	collection := database.Collection("users")
	indexModel := mongo.IndexModel{
		Keys:    bson.D{{Key: "username", Value: 1}},
		Options: options.Index().SetUnique(true),
	}

	_, err = collection.Indexes().CreateOne(context.TODO(), indexModel)
	if err != nil {
		log.Fatalf("Failed to create index on 'username': %v", err)
	}

	fmt.Println("Ensured that the 'testdb' database and 'users' collection exist.")
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
