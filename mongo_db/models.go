package mongodb

import (
	"context"

	"go.mongodb.org/mongo-driver/mongo"
)

type ValidOTPSerial struct {
	OTP          string `bson:"otp"`
	SerialNumber string `bson:"serial_number"`
}
type MongoDBClient struct {
	Client          *mongo.Client
	DatabaseName    string
	UsersCollection string
	Context         context.Context
}
