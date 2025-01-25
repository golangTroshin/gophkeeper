package handlers_test

import (
	"context"
	"testing"

	pb "github.com/golangTroshin/gophkeeper/grpc"
	"github.com/golangTroshin/gophkeeper/server/internal/handlers"
	"github.com/golangTroshin/gophkeeper/server/internal/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var testDB *gorm.DB
var server *handlers.GophKeeperServer

// setupTestDB initializes an in-memory SQLite database for testing
func setupTestDB(t *testing.T) {
	var err error
	testDB, err = gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to create in-memory database: %v", err)
	}

	// Run migrations
	err = testDB.AutoMigrate(&models.User{}, &models.Vault{})
	if err != nil {
		t.Fatalf("Database migration failed: %v", err)
	}

	// Create the test server instance
	server = &handlers.GophKeeperServer{DB: testDB}
}

// TestRegisterUser ensures a user is created successfully
func TestRegisterUser(t *testing.T) {
	setupTestDB(t)

	req := &pb.RegisterUserRequest{
		Username: "testuser",
		Password: "securepassword",
		Seed:     "testseed",
	}

	res, err := server.RegisterUser(context.Background(), req)
	if err != nil {
		t.Fatalf("Failed to register user: %v", err)
	}

	if !res.Success {
		t.Fatalf("Expected successful registration, got: %s", res.Message)
	}
}

// TestAuthenticateUser validates user authentication
func TestAuthenticateUser(t *testing.T) {
	setupTestDB(t)

	// Register user first
	req := &pb.RegisterUserRequest{
		Username: "authuser",
		Password: "authpass",
		Seed:     "authseed",
	}
	_, _ = server.RegisterUser(context.Background(), req)

	// Authenticate user
	authReq := &pb.AuthenticateUserRequest{
		Username: "authuser",
		Password: "authpass",
	}

	res, err := server.AuthenticateUser(context.Background(), authReq)
	if err != nil {
		t.Fatalf("Failed to authenticate user: %v", err)
	}

	if !res.Success {
		t.Fatalf("Expected successful authentication, got: %s", res.Message)
	}
}

// TestVerifyToken ensures token verification works correctly
func TestVerifyToken(t *testing.T) {
	token, err := handlers.GenerateJWT(1)
	if err != nil {
		t.Fatalf("Failed to generate JWT: %v", err)
	}

	userID, err := handlers.VerifyToken(token)
	if err != nil {
		t.Fatalf("Failed to verify token: %v", err)
	}

	if userID != 1 {
		t.Fatalf("Expected user ID 1, got %d", userID)
	}
}

// TestStoreAndRetrieveData checks if encrypted data can be stored and retrieved
func TestStoreAndRetrieveData(t *testing.T) {
	setupTestDB(t)

	// Register user and get token
	req := &pb.RegisterUserRequest{
		Username: "datauser",
		Password: "datapass",
		Seed:     "dataseed",
	}
	regRes, _ := server.RegisterUser(context.Background(), req)

	// Store data
	storeReq := &pb.StoreDataRequest{
		Token:    regRes.Token,
		DataType: pb.DataType_TEXT,
		Metadata: "Test Metadata",
		Data:     []byte("Encrypted data"),
	}

	storeRes, err := server.StoreData(context.Background(), storeReq)
	if err != nil {
		t.Fatalf("Failed to store data: %v", err)
	}

	if !storeRes.Success {
		t.Fatalf("Expected successful storage, got: %s", storeRes.Message)
	}

	// Retrieve data
	retrieveReq := &pb.RetrieveDataRequest{
		Token:  regRes.Token,
		Filter: pb.DataType_TEXT,
	}

	retrieveRes, err := server.RetrieveData(context.Background(), retrieveReq)
	if err != nil {
		t.Fatalf("Failed to retrieve data: %v", err)
	}

	if len(retrieveRes.Items) == 0 {
		t.Fatal("Expected retrieved data, got none")
	}
}
