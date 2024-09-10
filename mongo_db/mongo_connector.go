package mongodb

import (
	"context"
	"fmt"
	"log"

	utilities "midi-file-server/utilities" // Import utilities for WrapError

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	ErrMongoDBConnection = fmt.Errorf("failed to connect to MongoDB")
	ErrMongoDBPing       = fmt.Errorf("failed to ping MongoDB")
	ErrMongoDBAddDemo    = fmt.Errorf("failed to add demo data")
	ErrMongoDBDisconnect = fmt.Errorf("failed to disconnect from MongoDB")
	ErrMongoDBListColls  = fmt.Errorf("failed to list collections")
	ErrMongoDBCreateColl = fmt.Errorf("failed to create collection")
	ErrMongoDBCreateIdx  = fmt.Errorf("failed to create index on collection")
	ErrMongoDBInsertDemo = fmt.Errorf("failed to insert demo data")
)

// NewMongoDBClient creates a new instance of MongoDBClient.
func NewMongoDBClient(ctx context.Context) *MongoDBClient {
	return &MongoDBClient{
		DatabaseName:    utilities.DatabaseName,
		UsersCollection: utilities.UsersCollection,
		Context:         ctx,
	}
}

// Connect establishes a connection to MongoDB and assigns it to the MongoDBClient.
func (m *MongoDBClient) Connect() error {
	clientOptions := options.Client().ApplyURI(utilities.MongoDBURI)
	client, err := mongo.Connect(m.Context, clientOptions)
	if err != nil {
		return utilities.WrapError(err, ErrMongoDBConnection, "Connecting to MongoDB")
	}

	if err = client.Ping(m.Context, nil); err != nil {
		return utilities.WrapError(err, ErrMongoDBPing, "Pinging MongoDB")
	}

	m.Client = client
	fmt.Println("Connected to MongoDB!")
	err = m.AddDemoData()
	if err != nil {
		return utilities.WrapError(err, ErrMongoDBAddDemo, "Adding demo data to MongoDB")
	}
	fmt.Println("Added demo data to MongoDB!")
	return nil
}

// Disconnect closes the MongoDB connection.
func (m *MongoDBClient) Disconnect() {
	if m.Client != nil {
		if err := m.Client.Disconnect(m.Context); err != nil {
			log.Fatalf("Failed to disconnect from MongoDB: %v", err)
		} else {
			fmt.Println("Disconnected from MongoDB successfully.")
		}
	}
}

// VerifyDB  checks if the necessary collections exist in the database
func (m *MongoDBClient) VerifyDB() error {
	database := m.Client.Database(m.DatabaseName)
	collectionNames, err := database.ListCollectionNames(m.Context, bson.D{{Key: "name", Value: m.UsersCollection}})
	if err != nil {
		return utilities.WrapError(err, ErrMongoDBListColls, fmt.Sprintf("Database: %s", m.DatabaseName))
	}

	if len(collectionNames) == 0 {
		fmt.Printf("Creating '%s' collection...\n", m.UsersCollection)
		if err := database.CreateCollection(m.Context, m.UsersCollection); err != nil {
			return utilities.WrapError(err, ErrMongoDBCreateColl, fmt.Sprintf("Database: %s, Collection: %s", m.DatabaseName, m.UsersCollection))
		}
	} else {
		fmt.Printf("'%s' collection already exists.\n", m.UsersCollection)
	}

	indexModel := mongo.IndexModel{
		Keys:    bson.D{{Key: "username", Value: 1}},
		Options: options.Index().SetUnique(true),
	}
	if _, err := database.Collection(m.UsersCollection).Indexes().CreateOne(m.Context, indexModel); err != nil {
		return utilities.WrapError(err, ErrMongoDBCreateIdx, fmt.Sprintf("Database: %s, Collection: %s", m.DatabaseName, m.UsersCollection))
	}

	fmt.Printf("Ensured that the '%s' database and '%s' collection exist.\n", m.DatabaseName, m.UsersCollection)
	return nil
}

func (m *MongoDBClient) AddDemoData() error {
	// Get the database instance from the client
	db := m.Client.Database(m.DatabaseName)

	// Insert valid OTP and Serial Numbers
	otpSerials := []interface{}{
		ValidOTPSerial{OTP: "D4:8A:FC:9E:77:E0", SerialNumber: "ESP32-SN-001"},
		ValidOTPSerial{OTP: "D4:8A:FC:9E:77:E1", SerialNumber: "ESP32-SN-002"},
		ValidOTPSerial{OTP: "D4:8A:FC:9E:77:E2", SerialNumber: "ESP32-SN-003"},
		// Add more as needed
	}

	_, err := db.Collection("valid_otp_serials").InsertMany(m.Context, otpSerials)
	if err != nil {
		return utilities.WrapError(err, ErrMongoDBInsertDemo, "Inserting demo data into MongoDB")
	}
	return nil
}
