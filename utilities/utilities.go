package utilities

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/rs/zerolog/log"
	"go.mongodb.org/mongo-driver/mongo"
)

// Errors
var ()

var (
	HTTP_CONTEXT_TIMEOUT          = GetEnv("HTTP_CONTEXT_TIMEOUT", "1")
	MongoDBURI                    = GetEnv("MONGODB_URI", "mongodb://mongodb-service:27017")
	DatabaseName                  = GetEnv("DATABASE_NAME", "testdb")
	UsersCollection               = GetEnv("USERS_COLLECTION", "users")
	DefaultBucketName             = GetEnv("DEFAULT_BUCKET_NAME", "midi_file_storage")
	SIGNED_URL_EXPIRATION_MINUTES = GetEnv("SIGNED_URL_EXPIRATION_MINUTES", "5")
)

func WrapError(err error, customErr error, contextInfo ...string) error {
	if err != nil {
		contextMessage := ""
		if len(contextInfo) > 0 {
			contextMessage = fmt.Sprintf(" | Context: %s", contextInfo)
		}
		return fmt.Errorf("%w: %v%s", customErr, err, contextMessage)
	}
	return nil
}

// Use a fresh timeout for each request
func WithTimeout(handler func(context.Context, http.ResponseWriter, *http.Request)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		timedContext, cancel := context.WithTimeout(r.Context(), GetSignedTimeDurationMinutes(HTTP_CONTEXT_TIMEOUT))
		defer cancel()
		handler(timedContext, w, r)
	}
}

func WithTimeoutDb(db *mongo.Database, handler func(context.Context, *mongo.Database, http.ResponseWriter, *http.Request)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		timedContext, cancel := context.WithTimeout(r.Context(), GetSignedTimeDurationMinutes(HTTP_CONTEXT_TIMEOUT))
		defer cancel()
		handler(timedContext, db, w, r)
	}
}

// withSignedUrlDuration allows passing an additional argument like time.Duration to handlers
func WithSignedUrlDuration(d time.Duration, handler func(context.Context, http.ResponseWriter, *http.Request, time.Duration)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		timedContext, cancel := context.WithTimeout(r.Context(), d)
		defer cancel()
		handler(timedContext, w, r, d)
	}
}
func LogErrorAndRespond(w http.ResponseWriter, message string, statusCode int) {
	log.Error().Int("status_code", statusCode).Msg(message)
	http.Error(w, message, statusCode)
}
func GetEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
func GetSignedTimeDurationMinutes(timeStr string) time.Duration {

	signedUrlExpirationMinutes, err := time.ParseDuration(timeStr + "m")
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to parse signed URL expiration minutes")
	}
	return signedUrlExpirationMinutes
}
