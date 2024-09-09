package restapi

import "go.mongodb.org/mongo-driver/bson/primitive"

type HealthCheckResponse struct {
	Health    string `json:"health"`
	LastBuild string `json:"last_build"`
}

type User struct {
	ID              primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	Username        string             `json:"username" bson:"username"`
	Password        string             `json:"password" bson:"password"`
	OneTimePassword string             `json:"otp" bson:"otp"`
	SerialNumber    string             `json:"serialNumber" bson:"serialNumber"`
}

type SignedUrlRequest struct {
	ObjectName []string `json:"objectName"`
}

type DownloadRespones struct {
	DownloadResponse []DownloadResponse `json:"downloadResponse"`
}
type DownloadResponse struct {
	SignedURL  string `json:"signedUrl"`
	ObjectName string `json:"objectName"`
}

var UserCredentials struct {
	PrivateKey  string `json:"private_key"`
	ClientEmail string `json:"client_email"`
}
