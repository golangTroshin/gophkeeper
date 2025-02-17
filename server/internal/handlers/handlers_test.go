package handlers_test

import (
	"context"
	"testing"

	pb "github.com/golangTroshin/gophkeeper/grpc"
	"github.com/golangTroshin/gophkeeper/server/internal/handlers"
	"github.com/golangTroshin/gophkeeper/server/internal/models"
	"github.com/golangTroshin/gophkeeper/server/internal/repository"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var testRepo repository.Repository
var testServer *handlers.GophKeeperServer

// setupTestDB initializes an in-memory SQLite database for testing
func setupTestDB(t *testing.T) {
	var err error
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to create in-memory database: %v", err)
	}

	// Run migrations
	err = db.AutoMigrate(&models.User{}, &models.Vault{})
	if err != nil {
		t.Fatalf("Database migration failed: %v", err)
	}

	// Initialize repository and server with test DB
	testRepo = repository.NewRepository(db)
	testServer = &handlers.GophKeeperServer{Repo: testRepo}
}

// TestRegisterUser ensures a user is created successfully
func TestRegisterUser(t *testing.T) {
	setupTestDB(t)

	req := &pb.RegisterUserRequest{
		Username: "testuser",
		Password: "securepassword",
		Seed:     "testseed",
	}

	res, err := testServer.RegisterUser(context.Background(), req)
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
	_, _ = testServer.RegisterUser(context.Background(), req)

	// Authenticate user
	authReq := &pb.AuthenticateUserRequest{
		Username: "authuser",
		Password: "authpass",
	}

	res, err := testServer.AuthenticateUser(context.Background(), authReq)
	if err != nil {
		t.Fatalf("Failed to authenticate user: %v", err)
	}

	if !res.Success {
		t.Fatalf("Expected successful authentication, got: %s", res.Message)
	}
}

func TestVerifyToken(t *testing.T) {
	token, err := handlers.GenerateJWT(1)
	if err != nil {
		t.Fatalf("Failed to generate JWT: %v", err)
	}

	t.Logf("Generated Token: %s", token) // Log the token for debugging

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
	regRes, _ := testServer.RegisterUser(context.Background(), req)

	// Store data
	storeReq := &pb.StoreDataRequest{
		Token:    regRes.Token,
		DataType: pb.DataType_TEXT,
		Metadata: "Test Metadata",
		Data:     []byte("Encrypted data"),
	}

	storeRes, err := testServer.StoreData(context.Background(), storeReq)
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

	retrieveRes, err := testServer.RetrieveData(context.Background(), retrieveReq)
	if err != nil {
		t.Fatalf("Failed to retrieve data: %v", err)
	}

	if len(retrieveRes.Items) == 0 {
		t.Fatal("Expected retrieved data, got none")
	}
}

// TestMasterSeedRetrieve ensures the master seed is correctly retrieved
func TestMasterSeedRetrieve(t *testing.T) {
	setupTestDB(t)

	// Register user
	req := &pb.RegisterUserRequest{
		Username: "seeduser",
		Password: "seedpass",
		Seed:     "test-master-seed",
	}
	regRes, _ := testServer.RegisterUser(context.Background(), req)

	// Retrieve master seed
	seedReq := &pb.MasterSeedRetrieveRequest{Token: regRes.Token}
	seedRes, err := testServer.MasterSeedRetrieve(context.Background(), seedReq)
	if err != nil {
		t.Fatalf("Failed to retrieve master seed: %v", err)
	}

	if !seedRes.Success || seedRes.MasterSeed != "test-master-seed" {
		t.Fatalf("Master seed retrieval failed: expected 'test-master-seed', got '%s'", seedRes.MasterSeed)
	}
}
