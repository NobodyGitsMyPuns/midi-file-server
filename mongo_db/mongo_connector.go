package mongodb

import (
	"context"
	"fmt"
	"log"
	"os"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	mongoDBURI = getEnv("MONGODB_URI", "mongodb://mongodb-service:27017")
)

// getEnv is a helper function to get environment variables with a default fallback.
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

// NewMongoDBClient creates a new instance of MongoDBClient.
func NewMongoDBClient(ctx context.Context) *MongoDBClient {
	return &MongoDBClient{
		DatabaseName:    getEnv("DATABASE_NAME", "testdb"),
		UsersCollection: getEnv("USERS_COLLECTION", "users"),
		Context:         ctx,
	}
}

// Connect establishes a connection to MongoDB and assigns it to the MongoDBClient.
func (m *MongoDBClient) Connect() error {
	clientOptions := options.Client().ApplyURI(mongoDBURI)
	client, err := mongo.Connect(m.Context, clientOptions)
	if err != nil {
		return err
	}

	if err = client.Ping(m.Context, nil); err != nil {
		return err
	}

	m.Client = client
	fmt.Println("Connected to MongoDB!")
	err = m.AddDemoData()
	if err != nil {
		return fmt.Errorf("failed to add demo data: %w", err)
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

// VerifyDB checks if the necessary collections exist in the database
func (m *MongoDBClient) VerifyDB() error {
	database := m.Client.Database(m.DatabaseName)
	collectionNames, err := database.ListCollectionNames(m.Context, bson.D{{Key: "name", Value: m.UsersCollection}})
	if err != nil {
		return fmt.Errorf("failed to list collections: %w", err)
	}

	if len(collectionNames) == 0 {
		fmt.Printf("Creating '%s' collection...\n", m.UsersCollection)
		if err := database.CreateCollection(m.Context, m.UsersCollection); err != nil {
			return fmt.Errorf("failed to create '%s' collection: %w", m.UsersCollection, err)
		}
	} else {
		fmt.Printf("'%s' collection already exists.\n", m.UsersCollection)
	}

	indexModel := mongo.IndexModel{
		Keys:    bson.D{{Key: "username", Value: 1}},
		Options: options.Index().SetUnique(true),
	}
	if _, err := database.Collection(m.UsersCollection).Indexes().CreateOne(m.Context, indexModel); err != nil {
		return fmt.Errorf("failed to create index on 'username' in collection '%s': %w", m.UsersCollection, err)
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
		return fmt.Errorf("failed to insert demo data: %w", err)
	}
	return nil
}
