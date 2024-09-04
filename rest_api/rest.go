package restapi

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"cloud.google.com/go/storage"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/iterator"
)

const (
	URLExpiration = 5 * time.Minute
)

var (
	client            *mongo.Client
	db                *mongo.Database
	mongoDBURI        = getEnv("MONGODB_URI", "mongodb://mongodb-service:27017")
	databaseName      = getEnv("DATABASE_NAME", "testdb")
	usersCollection   = getEnv("USERS_COLLECTION", "users")
	defaultBucketName = getEnv("DEFAULT_BUCKET_NAME", "midi_file_storage")
)

func RegisterUser(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	var user User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		log.Printf("Failed to decode user data: %v", err)
		http.Error(w, "Bad Request: Invalid user data", http.StatusBadRequest)
		return
	}

	log.Printf("Received registration request for username: %s with OTP: %s and Serial: %s", user.Username, user.OneTimePassword, user.SerialNumber)

	var existingUser User
	err := db.Collection(usersCollection).FindOne(ctx, bson.M{"username": user.Username}).Decode(&existingUser)
	if err == nil {
		http.Error(w, "Username already taken", http.StatusConflict)
		return
	} else if err != mongo.ErrNoDocuments {
		log.Printf("Failed to check existing user: %v", err)
		http.Error(w, "Internal Server Error: Failed to check existing user", http.StatusInternalServerError)
		return
	}
	type ValidOTPSerial struct {
		OTP          string `bson:"otp"`
		SerialNumber string `bson:"serial_number"`
	}
	// Check OTP and Serial Number
	var validEntry ValidOTPSerial
	err = db.Collection("valid_otp_serials").FindOne(ctx, bson.M{"otp": user.OneTimePassword, "serial_number": user.SerialNumber}).Decode(&validEntry)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			log.Printf("OTP and Serial Number not found in the database. OTP: %s, Serial: %s", user.OneTimePassword, user.SerialNumber)
			http.Error(w, "Invalid OTP or Serial Number", http.StatusUnauthorized)
		} else {
			log.Printf("Failed to validate OTP and Serial Number: %v", err)
			http.Error(w, "Internal Server Error: Failed to validate OTP and Serial Number", http.StatusInternalServerError)
		}
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("Failed to hash password: %v", err)
		http.Error(w, "Internal Server Error: Failed to hash password", http.StatusInternalServerError)
		return
	}
	user.Password = string(hashedPassword)

	if _, err := db.Collection(usersCollection).InsertOne(ctx, user); err != nil {
		log.Printf("Failed to insert user into MongoDB: %v", err)
		http.Error(w, "Internal Server Error: Failed to register user", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(map[string]string{"message": "User registered successfully"}); err != nil {
		log.Printf("Failed to encode registration success message: %v", err)
		http.Error(w, "Internal Server Error: Failed to respond with success message", http.StatusInternalServerError)
	}
}

func OnHealthSubmit(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(HealthCheckResponse{Health: "OK"}); err != nil {
		log.Printf("Failed to encode health response: %v", err)
		http.Error(w, "Internal Server Error: Failed to encode health response", http.StatusInternalServerError)
	}
}

func LoginUser(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	var user User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		log.Printf("Failed to decode user data: %v", err)
		http.Error(w, "Bad Request: Invalid user data", http.StatusBadRequest)
		return
	}

	var dbUser User
	if err := db.Collection(usersCollection).FindOne(ctx, bson.M{"username": user.Username}).Decode(&dbUser); err != nil {
		log.Printf("Failed to find user in MongoDB: %v", err)
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(dbUser.Password), []byte(user.Password)); err != nil {
		log.Printf("Password comparison failed: %v", err)
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(map[string]string{"message": "Login successful"}); err != nil {
		log.Printf("Failed to encode login success response: %v", err)
		http.Error(w, "Internal Server Error: Failed to respond with success message", http.StatusInternalServerError)
	}
}

func GetSignedUrl(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	var reqs SignedUrlRequest
	if err := json.NewDecoder(r.Body).Decode(&reqs); err != nil {
		log.Printf("Failed to decode download request: %v", err)
		http.Error(w, "Bad Request: Invalid download request", http.StatusBadRequest)
		return
	}
	responsePayload := []DownloadResponse{}

	for _, currentObjectName := range reqs.ObjectName {

		if currentObjectName == "" {
			http.Error(w, "Missing midi object", http.StatusBadRequest)
			return
		}

		signedURL, err := generateSignedURL(ctx, defaultBucketName, currentObjectName)
		if err != nil {
			log.Printf("Failed to generate signed URL for object %s: %v", currentObjectName, err)
			http.Error(w, fmt.Sprintf("Failed to generate signed URL: %v", err), http.StatusInternalServerError)
			return
		}

		responsePayload = append(responsePayload, DownloadResponse{
			SignedURL:  signedURL,
			ObjectName: currentObjectName,
		})

	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(responsePayload); err != nil {
		log.Printf("Failed to encode signed URL response: %v", err)
		http.Error(w, "Internal Server Error: Failed to respond with signed URL", http.StatusInternalServerError)
	}
}

func generateSignedURL(ctx context.Context, bucketName, objectName string) (string, error) {
	creds, err := google.FindDefaultCredentials(ctx, storage.ScopeReadOnly)
	if err != nil {
		return "", fmt.Errorf("failed to find default credentials: %w", err)
	}

	var parsedCreds struct {
		PrivateKey  string `json:"private_key"`
		ClientEmail string `json:"client_email"`
	}

	if err := json.Unmarshal(creds.JSON, &parsedCreds); err != nil {
		return "", fmt.Errorf("failed to parse credentials JSON: %w", err)
	}

	opts := &storage.SignedURLOptions{
		GoogleAccessID: parsedCreds.ClientEmail,
		Scheme:         storage.SigningSchemeV4,
		Method:         "GET",
		Expires:        time.Now().Add(URLExpiration),
		PrivateKey:     []byte(parsedCreds.PrivateKey),
	}

	url, err := storage.SignedURL(bucketName, objectName, opts)
	if err != nil {
		return "", fmt.Errorf("failed to create signed URL: %w", err)
	}

	return url, nil
}

func ListBucketHandler(ctx context.Context, w http.ResponseWriter, r *http.Request) {

	objectNames, err := ListBucketContents(ctx, defaultBucketName)
	if err != nil {
		log.Printf("Failed to list bucket contents: %v", err)
		http.Error(w, fmt.Sprintf("Failed to list bucket contents: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(objectNames); err != nil {
		log.Printf("Failed to encode bucket contents: %v", err)
		http.Error(w, fmt.Sprintf("Failed to encode bucket contents: %v", err), http.StatusInternalServerError)
	}
}

func ListBucketContents(ctx context.Context, bucketName string) ([]string, error) {

	client, err := storage.NewClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create client: %w", err)
	}
	defer client.Close()

	it := client.Bucket(bucketName).Objects(ctx, nil)
	var objectNames []string
	for {
		attrs, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to list objects: %w", err)
		}
		objectNames = append(objectNames, attrs.Name)
	}

	return objectNames, nil
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
