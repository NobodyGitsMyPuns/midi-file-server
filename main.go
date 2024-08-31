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
	"google.golang.org/api/option"
)

const (
	HealthEp                 = "health"
	VersionEp                = "v1"
	RegisterEp               = "register"
	LoginEp                  = "login"
	GetSignedUrl             = "get-signed-url"
	ListAvailableMidiBuckets = "list-available-midi-files"
	ContextTimeout           = 60 * time.Second
	GCPProject               = "gothic-oven-433521-e1"
	MongoDBURI               = "mongodb://mongodb-service:27017"
	DatabaseName             = "testdb"
	UsersCollection          = "users"
)

func main() {
	client, err := connectMongoDB()
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer disconnectMongoDB(client)

	ensureDatabaseAndCollections(client)

	http.HandleFunc(fmt.Sprintf("/%s/%s", VersionEp, HealthEp), withTimeout(restapi.OnHealthSubmit))
	http.HandleFunc(fmt.Sprintf("/%s/%s", VersionEp, RegisterEp), withTimeout(restapi.RegisterUser))
	http.HandleFunc(fmt.Sprintf("/%s/%s", VersionEp, LoginEp), withTimeout(restapi.LoginUser))
	http.HandleFunc(fmt.Sprintf("/%s/%s", VersionEp, GetSignedUrl), withTimeout(restapi.GetSignedUrl))
	http.HandleFunc(fmt.Sprintf("/%s/%s", VersionEp, ListAvailableMidiBuckets), withTimeout(restapi.ListBucketHandler))
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func connectMongoDB() (*mongo.Client, error) {
	clientOptions := options.Client().ApplyURI(MongoDBURI)
	client, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		return nil, err
	}

	if err = client.Ping(context.TODO(), nil); err != nil {
		return nil, err
	}

	fmt.Println("Connected to MongoDB!")
	return client, nil
}

func disconnectMongoDB(client *mongo.Client) {
	if err := client.Disconnect(context.TODO()); err != nil {
		log.Fatalf("Failed to disconnect from MongoDB: %v", err)
	} else {
		fmt.Println("Disconnected from MongoDB successfully.")
	}
}

func ensureDatabaseAndCollections(client *mongo.Client) {
	database := client.Database(DatabaseName)
	collectionNames, err := database.ListCollectionNames(context.TODO(), bson.D{{Key: "name", Value: UsersCollection}})
	if err != nil {
		log.Fatalf("Failed to list collections: %v", err)
	}

	if len(collectionNames) == 0 {
		fmt.Println("Creating 'users' collection...")
		if err := database.CreateCollection(context.TODO(), UsersCollection); err != nil {
			log.Fatalf("Failed to create 'users' collection: %v", err)
		}
	} else {
		fmt.Println("'users' collection already exists.")
	}

	indexModel := mongo.IndexModel{
		Keys:    bson.D{{Key: "username", Value: 1}},
		Options: options.Index().SetUnique(true),
	}
	if _, err := database.Collection(UsersCollection).Indexes().CreateOne(context.TODO(), indexModel); err != nil {
		log.Fatalf("Failed to create index on 'username': %v", err)
	}

	fmt.Println("Ensured that the 'testdb' database and 'users' collection exist.")
}

func withTimeout(handler func(context.Context, http.ResponseWriter, *http.Request)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		timedContext, cancel := context.WithTimeout(r.Context(), ContextTimeout)
		defer cancel()
		handler(timedContext, w, r)
	}
}

func InitGCPWithServiceAccount(serviceAccountID, keyFilePath string) (*storage.Client, error) {
	ctx := context.Background()
	client, err := storage.NewClient(ctx, option.WithCredentialsFile(keyFilePath))
	if err != nil {
		return nil, fmt.Errorf("failed to create storage client: %w", err)
	}

	fmt.Println("GCP credentials initialized successfully with service account")
	return client, nil
}

func UploadFiles(bucketName, prefix string, filePaths []string) error {
	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		return fmt.Errorf("failed to create storage client: %w", err)
	}
	defer client.Close()

	for _, filePath := range filePaths {
		fileName := filepath.Base(filePath)
		objectPath := prefix + "/" + fileName
		wc := client.Bucket(bucketName).Object(objectPath).NewWriter(ctx)

		file, err := os.Open(filePath)
		if err != nil {
			return fmt.Errorf("failed to open file %s: %w", filePath, err)
		}
		defer file.Close()

		if _, err = io.Copy(wc, file); err != nil {
			return fmt.Errorf("failed to upload file %s: %w", fileName, err)
		}

		if err := wc.Close(); err != nil {
			return fmt.Errorf("failed to complete upload for file %s: %w", fileName, err)
		}

		fmt.Printf("File %s uploaded successfully to bucket %s\n", fileName, bucketName)
	}

	return nil
}
