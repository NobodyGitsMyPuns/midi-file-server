package restapi

// import (
// 	"bytes"
// 	"context"
// 	"encoding/json"
// 	"net/http"
// 	"net/http/httptest"
// 	"testing"

// 	"github.com/stretchr/testify/mock"
// 	"go.mongodb.org/mongo-driver/bson"
// 	"go.mongodb.org/mongo-driver/mongo"
// )

// // MockSingleResult is a mock for the SingleResult returned by FindOne
// type MockSingleResult struct {
// 	mock.Mock
// }

// // Define an interface for MongoDB operations
// type MongoDBInterface interface {
// 	FindOne(ctx context.Context, filter interface{}) *mongo.SingleResult
// 	InsertOne(ctx context.Context, document interface{}) (*mongo.InsertOneResult, error)
// }

// // Decode simulates decoding the result of FindOne into the provided interface
// func (m *MockSingleResult) Decode(v interface{}) error {
// 	args := m.Called(v)
// 	return args.Error(0)
// }

// // MockMongoDB is a mock for the MongoDB operations
// type MockMongoDB struct {
// 	mock.Mock
// }

// // FindOne mocks the MongoDB FindOne operation
// func (m *MockMongoDB) FindOne(ctx context.Context, filter interface{}) *MockSingleResult {
// 	args := m.Called(ctx, filter)
// 	result := args.Get(0).(*MockSingleResult)
// 	return result
// }

// // InsertOne mocks the MongoDB InsertOne operation
// func (m *MockMongoDB) InsertOne(ctx context.Context, document interface{}) (*mongo.InsertOneResult, error) {
// 	args := m.Called(ctx, document)
// 	return args.Get(0).(*mongo.InsertOneResult), args.Error(1)
// }

// // MockMongoDB instance used for testing
// var mockMongoDB *MockMongoDB

// // Set up mock database in the handler
// func setupMockDB() {
// 	mockMongoDB = new(MockMongoDB)
// 	client = mockMongoDB // Replace the global client with mockMongoDB
// 	db = mockMongoDB     // Replace the global db with mockMongoDB
// }

// // Test RegisterUser handler with valid input
// func TestRegisterUser_Success(t *testing.T) {
// 	setupMockDB()

// 	// Create a valid user for registration
// 	user := User{
// 		Username:        "testuser",
// 		Password:        "password",
// 		OneTimePassword: "D4:8A:FC:9E:77:E0",
// 		SerialNumber:    "ESP32-SN-001",
// 	}

// 	// Create a mock SingleResult for FindOne that returns ErrNoDocuments
// 	mockResult := new(MockSingleResult)
// 	mockResult.On("Decode", mock.Anything).Return(mongo.ErrNoDocuments)

// 	// Mock MongoDB FindOne to return no existing user
// 	mockMongoDB.On("FindOne", mock.Anything, bson.M{"username": user.Username}).Return(mockResult)

// 	// Mock MongoDB InsertOne to succeed
// 	mockMongoDB.On("InsertOne", mock.Anything, mock.Anything).Return(&mongo.InsertOneResult{}, nil)

// 	// Prepare the request
// 	body, _ := json.Marshal(user)
// 	req := httptest.NewRequest("POST", "/v1/register", bytes.NewBuffer(body))
// 	req.Header.Set("Content-Type", "application/json")

// 	// Record the response
// 	rec := httptest.NewRecorder()

// 	// Call the handler
// 	RegisterUser(context.TODO(), rec, req)

// 	// Check response code
// 	if status := rec.Code; status != http.StatusCreated {
// 		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusCreated)
// 	}

// 	// Check response body
// 	expected := `{"message":"User registered successfully"}`
// 	if rec.Body.String() != expected {
// 		t.Errorf("handler returned unexpected body: got %v want %v", rec.Body.String(), expected)
// 	}

// 	// Assert expectations for MongoDB operations
// 	mockMongoDB.AssertExpectations(t)
// 	mockResult.AssertExpectations(t)
// }

// // // Test RegisterUser handler with invalid input
// // func TestRegisterUser_InvalidInput(t *testing.T) {
// // 	req := httptest.NewRequest("POST", "/v1/register", bytes.NewBuffer([]byte("invalid data")))
// // 	req.Header.Set("Content-Type", "application/json")
// // 	rec := httptest.NewRecorder()

// // 	RegisterUser(context.TODO(), rec, req)

// // 	if status := rec.Code; status != http.StatusBadRequest {
// // 		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusBadRequest)
// // 	}
// // }

// // // Test LoginUser handler with valid credentials
// // func TestLoginUser_Success(t *testing.T) {
// // 	mockMongoDB := new(MockMongoDB)
// // 	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)
// // 	dbUser := User{Username: "testuser", Password: string(hashedPassword)}

// // 	mockMongoDB.On("FindOne", mock.Anything, bson.M{"username": "testuser"}).Return(&mongo.SingleResult{}).Run(func(args mock.Arguments) {
// // 		result := args.Get(0).(*mongo.SingleResult)
// // 		result.Decode(&dbUser)
// // 	})

// // 	reqBody := `{"username":"testuser","password":"password"}`
// // 	req := httptest.NewRequest("POST", "/v1/login", bytes.NewBuffer([]byte(reqBody)))
// // 	req.Header.Set("Content-Type", "application/json")
// // 	rec := httptest.NewRecorder()

// // 	LoginUser(context.TODO(), rec, req)

// // 	if status := rec.Code; status != http.StatusOK {
// // 		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
// // 	}

// // 	expected := `{"message":"Login successful"}`
// // 	if rec.Body.String() != expected {
// // 		t.Errorf("handler returned unexpected body: got %v want %v", rec.Body.String(), expected)
// // 	}

// // 	mockMongoDB.AssertExpectations(t)
// // }

// // // Test GenerateSignedURL function
// // func TestGenerateSignedURL_Success(t *testing.T) {
// // 	mockStorage := new(MockGoogleCloudStorage)
// // 	mockStorage.On("GenerateSignedURL", mock.Anything, "test-bucket", "test-object").Return("http://signedurl.com", nil)

// // 	url, err := mockStorage.GenerateSignedURL(context.TODO(), "test-bucket", "test-object")
// // 	if err != nil {
// // 		t.Errorf("Expected no error, but got %v", err)
// // 	}

// // 	expectedURL := "http://signedurl.com"
// // 	if url != expectedURL {
// // 		t.Errorf("Expected URL %v, but got %v", expectedURL, url)
// // 	}

// // 	mockStorage.AssertExpectations(t)
// // }

// // // Test GetSignedUrl handler with valid input
// // func TestGetSignedUrl_Success(t *testing.T) {
// // 	mockStorage := new(MockGoogleCloudStorage)
// // 	mockStorage.On("GenerateSignedURL", mock.Anything, "midi_file_storage", "test-file").Return("http://signedurl.com", nil)

// // 	reqBody := `{"objectName":["test-file"]}`
// // 	req := httptest.NewRequest("POST", "/v1/get-signed-url", bytes.NewBuffer([]byte(reqBody)))
// // 	req.Header.Set("Content-Type", "application/json")
// // 	rec := httptest.NewRecorder()

// // 	GetSignedUrl(context.TODO(), rec, req)

// // 	if status := rec.Code; status != http.StatusOK {
// // 		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
// // 	}

// // 	expected := `[{"signedURL":"http://signedurl.com","objectName":"test-file"}]`
// // 	if rec.Body.String() != expected {
// // 		t.Errorf("handler returned unexpected body: got %v want %v", rec.Body.String(), expected)
// // 	}

// // 	mockStorage.AssertExpectations(t)
// // }
