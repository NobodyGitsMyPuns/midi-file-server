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
	"github.com/dgrijalva/jwt-go"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/oauth2/google"
)

var client *mongo.Client
var db *mongo.Database

func init() {
	// Connect to MongoDB
	var err error
	ctx := context.Background()

	// Read the MongoDB URI from the environment variable
	mongoURI := os.Getenv("MONGODB_URI")
	if mongoURI == "" {
		// Fallback to a default value if the environment variable is not set
		mongoURI = "mongodb://mongodb-service:27017"
	}

	client, err = mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
	if err != nil {
		log.Fatal(err)
	}

	db = client.Database("testdb")
}

func OnHealthSubmit(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	dataF := HealthCheckResponse{Health: "OK"}
	err := json.NewEncoder(w).Encode(dataF)
	if err != nil {
		log.Printf("Failed to encode health response: %v", err)
		http.Error(w, "Internal Server Error: Failed to encode health response", http.StatusInternalServerError)
		return
	}
}

func RegisterUser(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	var user User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		log.Printf("Failed to decode user data: %v", err)
		http.Error(w, "Bad Request: Invalid user data", http.StatusBadRequest)
		return
	}

	// Check if the username is already taken
	var existingUser User
	err = db.Collection("users").FindOne(ctx, bson.M{"username": user.Username}).Decode(&existingUser)
	if err == nil {
		http.Error(w, "Username already taken", http.StatusConflict)
		return
	}

	if err != mongo.ErrNoDocuments {
		log.Printf("Failed to check existing user: %v", err)
		http.Error(w, "Internal Server Error: Failed to check existing user", http.StatusInternalServerError)
		return
	}

	// Validate OTP and Serial Number
	if user.OneTimePassword != "valid_otp" || user.SerialNumber != "valid_serial" {
		http.Error(w, "Invalid OTP or Serial Number", http.StatusUnauthorized)
		return
	}

	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("Failed to hash password: %v", err)
		http.Error(w, "Internal Server Error: Failed to hash password", http.StatusInternalServerError)
		return
	}
	user.Password = string(hashedPassword)

	// Insert into MongoDB
	_, err = db.Collection("users").InsertOne(ctx, user)
	if err != nil {
		log.Printf("Failed to insert user into MongoDB: %v", err)
		http.Error(w, "Internal Server Error: Failed to register user", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(map[string]string{"message": "User registered successfully"})
	if err != nil {
		log.Printf("Failed to encode registration success message: %v", err)
		http.Error(w, "Internal Server Error: Failed to respond with success message", http.StatusInternalServerError)
	}
}

func LoginUser(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	var user User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		log.Printf("Failed to decode user data: %v", err)
		http.Error(w, "Bad Request: Invalid user data", http.StatusBadRequest)
		return
	}

	var dbUser User
	err = db.Collection("users").FindOne(ctx, bson.M{"username": user.Username}).Decode(&dbUser)
	if err != nil {
		log.Printf("Failed to find user in MongoDB: %v", err)
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(dbUser.Password), []byte(user.Password))
	if err != nil {
		log.Printf("Password comparison failed: %v", err)
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"username": user.Username,
		"exp":      time.Now().Add(time.Hour * 72).Unix(),
	})

	tokenString, err := token.SignedString([]byte("your-secret-key"))
	if err != nil {
		log.Printf("Failed to sign JWT token: %v", err)
		http.Error(w, "Internal Server Error: Failed to generate token", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(map[string]string{"token": tokenString})
	if err != nil {
		log.Printf("Failed to encode token response: %v", err)
		http.Error(w, "Internal Server Error: Failed to respond with token", http.StatusInternalServerError)
	}
}

func GetSignedUrl(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	var req SignedUrlRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		log.Printf("Failed to decode download request: %v", err)
		http.Error(w, "Bad Request: Invalid download request", http.StatusBadRequest)
		return
	}

	if req.BucketName == "" {
		http.Error(w, "Missing midi bucket", http.StatusInternalServerError)
		return
	}

	if req.ObjectName == "" {
		http.Error(w, "Missing midi object", http.StatusInternalServerError)
		return
	}

	signedURL, err := generateSignedURL(ctx, req.BucketName, req.ObjectName)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to generate signed URL: %v", err), http.StatusInternalServerError)
		return
	}

	response := struct {
		SignedURL  string `json:"signedUrl"`
		ObjectName string `json:"objectName"`
	}{
		SignedURL:  signedURL,
		ObjectName: req.ObjectName,
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		log.Printf("Failed to encode signed URL response: %v", err)
		http.Error(w, "Internal Server Error: Failed to respond with signed URL", http.StatusInternalServerError)
	}
}

func generateSignedURL(ctx context.Context, bucketName, objectName string) (string, error) {
	// Load credentials from the environment variable or file
	creds, err := google.FindDefaultCredentials(ctx, storage.ScopeReadOnly)
	if err != nil {
		return "", fmt.Errorf("failed to find default credentials: %w", err)
	}

	// Parse the credentials JSON to extract the private key
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
		Expires:        time.Now().Add(15 * time.Minute),
		PrivateKey:     []byte(parsedCreds.PrivateKey),
	}

	url, err := storage.SignedURL(bucketName, objectName, opts)
	if err != nil {
		return "", fmt.Errorf("failed to create signed URL: %w", err)
	}

	return url, nil
}

// func getGoogleAccessID(ctx context.Context, secretName string) (string, error) {
// 	client, err := secretmanager.NewClient(ctx)
// 	if err != nil {
// 		return "", fmt.Errorf("failed to create secretmanager client: %w", err)
// 	}
// 	defer client.Close()

// 	req := &secretmanagerpb.AccessSecretVersionRequest{
// 		Name: secretName,
// 	}

// 	result, err := client.AccessSecretVersion(ctx, req)
// 	if err != nil {
// 		return "", fmt.Errorf("failed to access secret version: %w", err)
// 	}

// 	var serviceAccount struct {
// 		ClientEmail string `json:"client_email"`
// 	}

// 	if err := json.Unmarshal(result.Payload.Data, &serviceAccount); err != nil {
// 		return "", fmt.Errorf("failed to unmarshal secret data: %w", err)
// 	}

// 	return serviceAccount.ClientEmail, nil
// }
