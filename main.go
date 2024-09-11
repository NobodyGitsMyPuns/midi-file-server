package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"

	mongodb "midi-file-server/mongo_db"
	restapi "midi-file-server/rest_api"
	"midi-file-server/utilities"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var (
	// Define custom errors
	ErrMongoDBConnection = errors.New("failed to connect to MongoDB")
	ErrMongoDBVerify     = errors.New("failed to verify MongoDB")
	ErrGCPStorage        = errors.New("failed to initialize Google Cloud Storage")
	ErrFileUpload        = errors.New("failed to upload file")
	ErrFileOpen          = errors.New("failed to open file")
	ErrFileClose         = errors.New("failed to close file")
)

const (
	VersionEp                = "v1"
	HealthEp                 = "health"
	RegisterEp               = "register"
	LoginEp                  = "login"
	GetSignedUrl             = "get-signed-url"
	ListAvailableMidiBuckets = "list-available-midi-files"
)

func main() {
	// Initialize zerolog to use human-readable output in the console
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	// Use a background context for MongoDB connection to avoid it timing out with HTTP requests,
	backgroundContext := context.Background()

	mongoDB := mongodb.NewMongoDBClient(backgroundContext)

	err := utilities.WrapError(mongoDB.Connect(), ErrMongoDBConnection)
	if err != nil {
		utilities.LogErrorAndRespond(nil, "MongoDB connection error", http.StatusInternalServerError)
		log.Fatal().Err(err).Msg("MongoDB connection error")
	}
	defer mongoDB.Disconnect()

	err = utilities.WrapError(mongoDB.VerifyDB(), ErrMongoDBVerify)
	if err != nil {
		utilities.LogErrorAndRespond(nil, "MongoDB verification error", http.StatusInternalServerError)
		log.Fatal().Err(err).Msg("MongoDB verification error")
	}
	db := mongoDB.Client.Database(mongoDB.DatabaseName)

	// Register handlers with the shared context
	http.HandleFunc(fmt.Sprintf("/%s/%s", VersionEp, HealthEp), utilities.WithTimeout(restapi.OnHealthSubmit))
	http.HandleFunc(fmt.Sprintf("/%s/%s", VersionEp, GetSignedUrl), utilities.WithSignedUrlDuration(utilities.GetSignedTimeDurationMinutes(utilities.SIGNED_URL_EXPIRATION_MINUTES), restapi.GetSignedUrl))
	http.HandleFunc(fmt.Sprintf("/%s/%s", VersionEp, ListAvailableMidiBuckets), utilities.WithTimeout(restapi.ListBucketHandler))
	http.HandleFunc(fmt.Sprintf("/%s/%s", VersionEp, RegisterEp), utilities.WithTimeoutDb(db, restapi.RegisterUser))
	http.HandleFunc(fmt.Sprintf("/%s/%s", VersionEp, LoginEp), utilities.WithTimeoutDb(db, restapi.LoginUser))

	log.Fatal().Err(http.ListenAndServe(":8080", nil)).Msg("Server failed")
}
