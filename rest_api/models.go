package restapi

import "go.mongodb.org/mongo-driver/bson/primitive"

type HealthCheckResponse struct {
	Health string `json:"health"`
}
type User struct {
	ID              primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	Username        string             `json:"username" bson:"username"`
	Password        string             `json:"password" bson:"password"`
	OneTimePassword string             `json:"otp" bson:"otp"`
	SerialNumber    string             `json:"serialNumber" bson:"serialNumber"`
}
