package restapi

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/dgrijalva/jwt-go"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/crypto/bcrypt"
)

var client *mongo.Client
var db *mongo.Database

func init() {
	// Connect to MongoDB
	var err error
	ctx := context.Background()
	client, err = mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		log.Fatal(err)
	}
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

func DownloadMIDI(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		log.Default().Println("Method Not Allowed")
		return
	}

	tokenString := r.Header.Get("Authorization")
	if tokenString == "" {
		http.Error(w, "No token provided", http.StatusUnauthorized)
		log.Default().Println("No token provided")
		return
	}

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte("your-secret-key"), nil
	})
	log.Default().Println(token)

	if err != nil || !token.Valid {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		log.Fatal("Invalid token")
		return
	}

	w.Header().Set("Content-Type", "audio/midi")
	w.Header().Set("Content-Disposition", "attachment; filename=midi-file.mid")
	http.ServeFile(w, r, "path/to/your/midi/file.mid")
}
