package restapi

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	mongodb "midi-file-server/mongo_db"

	"github.com/rs/zerolog/log"

	utilities "midi-file-server/utilities"

	"cloud.google.com/go/storage"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/iterator"
)

// Define error types
var (
	ErrUserExists              = fmt.Errorf("username already taken")
	ErrInvalidOTPSerial        = fmt.Errorf("invalid OTP or Serial Number")
	ErrFailedHashPassword      = fmt.Errorf("failed to hash password")
	ErrFailedRegisterUser      = fmt.Errorf("failed to register user")
	ErrInvalidCredentials      = fmt.Errorf("invalid credentials")
	ErrFailedListBucket        = fmt.Errorf("failed to list bucket contents")
	ErrFailedGenerateSignedURL = fmt.Errorf("failed to generate signed URL")
	ErrMethodNotAllowed        = fmt.Errorf("method not allowed")
)

func RegisterUser(ctx context.Context, db *mongo.Database, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utilities.LogErrorAndRespond(w, ErrMethodNotAllowed.Error(), http.StatusMethodNotAllowed)
		return
	}

	user, err := decodeUser(r)
	if err != nil {
		utilities.LogErrorAndRespond(w, utilities.WrapError(err, fmt.Errorf("invalid user data")).Error(), http.StatusBadRequest)
		return
	}

	log.Info().Str("username", user.Username).Str("otp", user.OneTimePassword).Str("serial", user.SerialNumber).Msg("Received registration request")

	if userExists(ctx, db, user.Username) {
		utilities.LogErrorAndRespond(w, ErrUserExists.Error(), http.StatusConflict)
		return
	}

	if !validateOTPAndSerial(ctx, db, user.OneTimePassword, user.SerialNumber) {
		utilities.LogErrorAndRespond(w, ErrInvalidOTPSerial.Error(), http.StatusUnauthorized)
		return
	}

	hashedPassword, err := hashPassword(user.Password)
	if err != nil {
		utilities.LogErrorAndRespond(w, utilities.WrapError(err, ErrFailedHashPassword).Error(), http.StatusInternalServerError)
		return
	}
	user.Password = hashedPassword

	if err := insertUser(ctx, db, user); err != nil {
		utilities.LogErrorAndRespond(w, utilities.WrapError(err, ErrFailedRegisterUser).Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(map[string]string{"message": "User registered successfully"}); err != nil {
		utilities.LogErrorAndRespond(w, utilities.WrapError(err, fmt.Errorf("failed to respond with success message")).Error(), http.StatusInternalServerError)
	}
}

// OnHealthSubmit returns the health status along with the last build information
func OnHealthSubmit(w http.ResponseWriter, r *http.Request) {
	log.Debug().Msg("Received health check request")

	// Retrieve the last build information from the environment variable
	lastBuild := os.Getenv("LAST_BUILD_INFO")
	if lastBuild == "" {
		lastBuild = "Unknown"
	}

	// Retrieve the digest from the environment variable
	digest := os.Getenv("DIGEST")
	if digest == "" {
		digest = "Digest: Unavailable"
	}

	// Create the health check response
	response := map[string]string{
		"health":     "Google Cloud Build!",
		"last_build": fmt.Sprintf("%s (Digest: %s)", lastBuild, digest),
	}

	// Return the health check response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Debug().Err(err).Msg("Failed to encode health response")
	}
}

func LoginUser(ctx context.Context, db *mongo.Database, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utilities.LogErrorAndRespond(w, ErrMethodNotAllowed.Error(), http.StatusMethodNotAllowed)
		return
	}

	var user User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		utilities.LogErrorAndRespond(w, utilities.WrapError(err, fmt.Errorf("invalid user data")).Error(), http.StatusBadRequest)
		return
	}

	var dbUser User
	if err := db.Collection(utilities.UsersCollection).FindOne(ctx, bson.M{"username": user.Username}).Decode(&dbUser); err != nil {
		utilities.LogErrorAndRespond(w, utilities.WrapError(err, ErrInvalidCredentials).Error(), http.StatusUnauthorized)
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(dbUser.Password), []byte(user.Password)); err != nil {
		utilities.LogErrorAndRespond(w, utilities.WrapError(err, ErrInvalidCredentials).Error(), http.StatusUnauthorized)
		return
	}

	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(map[string]string{"message": "Login successful"}); err != nil {
		utilities.LogErrorAndRespond(w, utilities.WrapError(err, fmt.Errorf("failed to respond with success message")).Error(), http.StatusInternalServerError)
	}
}

func GetSignedUrl(ctx context.Context, w http.ResponseWriter, r *http.Request, d time.Duration) {
	var reqs SignedUrlRequest
	if err := json.NewDecoder(r.Body).Decode(&reqs); err != nil {
		utilities.LogErrorAndRespond(w, utilities.WrapError(err, fmt.Errorf("invalid download request")).Error(), http.StatusBadRequest)
		return
	}

	responsePayload := []DownloadResponse{}

	for _, currentObjectName := range reqs.ObjectName {
		if currentObjectName == "" {
			utilities.LogErrorAndRespond(w, "Missing midi object", http.StatusBadRequest)
			return
		}

		signedURL, err := generateSignedURL(ctx, utilities.DefaultBucketName, currentObjectName, d)
		if err != nil {
			utilities.LogErrorAndRespond(w, utilities.WrapError(err, ErrFailedGenerateSignedURL).Error(), http.StatusInternalServerError)
			return
		}

		responsePayload = append(responsePayload, DownloadResponse{
			SignedURL:  signedURL,
			ObjectName: currentObjectName,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(responsePayload); err != nil {
		utilities.LogErrorAndRespond(w, utilities.WrapError(err, fmt.Errorf("failed to respond with signed URL")).Error(), http.StatusInternalServerError)
	}
}

func ListBucketHandler(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	objectNames, err := ListBucketContents(ctx, utilities.DefaultBucketName)
	if err != nil {
		utilities.LogErrorAndRespond(w, utilities.WrapError(err, ErrFailedListBucket).Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(objectNames); err != nil {
		utilities.LogErrorAndRespond(w, utilities.WrapError(err, fmt.Errorf("failed to encode bucket contents")).Error(), http.StatusInternalServerError)
	}
}

func generateSignedURL(ctx context.Context, bucketName, objectName string, d time.Duration) (string, error) {
	creds, err := google.FindDefaultCredentials(ctx, storage.ScopeReadOnly)
	if err != nil {
		return "", utilities.WrapError(err, fmt.Errorf("failed to find default credentials"))
	}

	if err := json.Unmarshal(creds.JSON, &UserCredentials); err != nil {
		return "", utilities.WrapError(err, fmt.Errorf("failed to parse credentials JSON"))
	}

	opts := &storage.SignedURLOptions{
		GoogleAccessID: UserCredentials.ClientEmail,
		Scheme:         storage.SigningSchemeV4,
		Method:         "GET",
		Expires:        time.Now().Add(d),
		PrivateKey:     []byte(UserCredentials.PrivateKey),
	}

	url, err := storage.SignedURL(bucketName, objectName, opts)
	if err != nil {
		return "", utilities.WrapError(err, fmt.Errorf("failed to create signed URL"))
	}

	return url, nil
}

func ListBucketContents(ctx context.Context, bucketName string) ([]string, error) {
	client, err := storage.NewClient(ctx)
	if err != nil {
		return nil, utilities.WrapError(err, fmt.Errorf("failed to create client"))
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
			return nil, utilities.WrapError(err, fmt.Errorf("failed to list objects"))
		}
		objectNames = append(objectNames, attrs.Name)
	}

	return objectNames, nil
}

// decodeUser decodes the incoming request into a User struct
func decodeUser(r *http.Request) (User, error) {
	var user User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		return user, utilities.WrapError(err, fmt.Errorf("failed to decode user data"))
	}
	return user, nil
}

// userExists checks if the username already exists in the database
func userExists(ctx context.Context, db *mongo.Database, username string) bool {
	var existingUser User
	err := db.Collection(utilities.UsersCollection).FindOne(ctx, bson.M{"username": username}).Decode(&existingUser)
	if err == nil {
		return true
	} else if err != mongo.ErrNoDocuments {
		log.Error().Err(err).Str("username", username).Msg("Failed to check existing user")
		return true // Assume user exists in case of an error to avoid duplicates
	}
	return false
}

// validateOTPAndSerial checks if the OTP and serial number are valid
func validateOTPAndSerial(ctx context.Context, db *mongo.Database, otp string, serialNumber string) bool {
	var validEntry mongodb.ValidOTPSerial
	err := db.Collection("valid_otp_serials").FindOne(ctx, bson.M{"otp": otp, "serial_number": serialNumber}).Decode(&validEntry)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			log.Info().Str("otp", otp).Str("serial", serialNumber).Msg("OTP and Serial Number not found in the database")
		} else {
			log.Error().Err(err).Str("otp", otp).Str("serial", serialNumber).Msg("Failed to validate OTP and Serial Number")
		}
		return false
	}
	return true
}

// hashPassword hashes the user's password
func hashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Error().Err(err).Msg("Failed to hash password")
		return "", err
	}
	return string(hashedPassword), nil
}

// insertUser inserts a new user into the database
func insertUser(ctx context.Context, db *mongo.Database, user User) error {
	_, err := db.Collection(utilities.UsersCollection).InsertOne(ctx, user)
	if err != nil {
		log.Error().Err(err).Str("username", user.Username).Msg("Failed to insert user into MongoDB")
		return err
	}
	return nil
}
