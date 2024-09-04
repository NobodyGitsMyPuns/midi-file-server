package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"time"

	mongodb "midi-file-server/mongo_db"
	restapi "midi-file-server/rest_api"
	"midi-file-server/utilities"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"go.mongodb.org/mongo-driver/mongo"
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
	SignedURLDuration        = time.Minute * 5
)

func main() {
	// Initialize zerolog to use human-readable output in the console
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	// Use a background context for MongoDB connection to avoid it timing out with HTTP requests
	backgroundContext := context.Background()

	mongoDB := mongodb.NewMongoDBClient(backgroundContext)

	err := utilities.WrapError(mongoDB.Connect(), ErrMongoDBConnection)
	if err != nil {
		log.Fatal().Err(err).Msg("MongoDB connection error")
	}
	defer mongoDB.Disconnect()

	err = utilities.WrapError(mongoDB.VerifyDB(), ErrMongoDBVerify)
	if err != nil {
		log.Fatal().Err(err).Msg("MongoDB verification error")
	}
	db := mongoDB.Client.Database(mongoDB.DatabaseName)

	// Register handlers with the shared context
	http.HandleFunc(fmt.Sprintf("/%s/%s", VersionEp, HealthEp), withTimeout(restapi.OnHealthSubmit))
	http.HandleFunc(fmt.Sprintf("/%s/%s", VersionEp, GetSignedUrl), withSignedUrlDuration(SignedURLDuration, restapi.GetSignedUrl))
	http.HandleFunc(fmt.Sprintf("/%s/%s", VersionEp, ListAvailableMidiBuckets), withTimeout(restapi.ListBucketHandler))
	http.HandleFunc(fmt.Sprintf("/%s/%s", VersionEp, RegisterEp), withTimeoutDb(db, restapi.RegisterUser))
	http.HandleFunc(fmt.Sprintf("/%s/%s", VersionEp, LoginEp), withTimeoutDb(db, restapi.LoginUser))

	log.Fatal().Err(http.ListenAndServe(":8080", nil)).Msg("Server failed")
}

// Use a fresh timeout for each request
func withTimeout(handler func(context.Context, http.ResponseWriter, *http.Request)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		timedContext, cancel := context.WithTimeout(r.Context(), ContextTimeout)
		defer cancel()
		handler(timedContext, w, r)
	}
}

func withTimeoutDb(db *mongo.Database, handler func(context.Context, *mongo.Database, http.ResponseWriter, *http.Request)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		timedContext, cancel := context.WithTimeout(r.Context(), ContextTimeout)
		defer cancel()
		handler(timedContext, db, w, r)
	}
}

// withSignedUrlDuration allows passing an additional argument like time.Duration to handlers
func withSignedUrlDuration(d time.Duration, handler func(context.Context, http.ResponseWriter, *http.Request, time.Duration)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		timedContext, cancel := context.WithTimeout(r.Context(), d)
		defer cancel()
		handler(timedContext, w, r, d)
	}
}
