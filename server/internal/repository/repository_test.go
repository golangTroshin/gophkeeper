package repository_test

import (
	"testing"

	pb "github.com/golangTroshin/gophkeeper/grpc"
	"github.com/golangTroshin/gophkeeper/server/internal/models"
	"github.com/golangTroshin/gophkeeper/server/internal/repository"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var repo repository.Repository
var testDB *gorm.DB

// setupTestDB initializes an in-memory SQLite database for testing.
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

	// Create the repository instance
	repo = repository.NewRepository(testDB)
}

// TestUserExists ensures checking user existence works.
func TestUserExists(t *testing.T) {
	setupTestDB(t)

	// Insert test user
	testUser := models.User{
		Login:      "testuser",
		Password:   "hashedpassword",
		MasterSeed: "testseed",
	}
	if err := repo.CreateUser(&testUser); err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// Check if the user exists
	exists, err := repo.UserExists("testuser")
	if err != nil {
		t.Fatalf("Error checking user existence: %v", err)
	}
	if !exists {
		t.Fatalf("Expected user to exist, but it does not")
	}
}

// TestCreateUser ensures a user is created successfully.
func TestCreateUser(t *testing.T) {
	setupTestDB(t)

	testUser := models.User{
		Login:      "newuser",
		Password:   "hashedpassword",
		MasterSeed: "newseed",
	}

	err := repo.CreateUser(&testUser)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// Verify user exists
	exists, err := repo.UserExists("newuser")
	if err != nil || !exists {
		t.Fatalf("User creation failed: %v", err)
	}
}

// TestGetUserByLogin ensures retrieving a user by login works.
func TestGetUserByLogin(t *testing.T) {
	setupTestDB(t)

	testUser := models.User{
		Login:      "lookupuser",
		Password:   "hashedpassword",
		MasterSeed: "seed123",
	}

	err := repo.CreateUser(&testUser)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	user, err := repo.GetUserByLogin("lookupuser")
	if err != nil {
		t.Fatalf("Failed to get user by login: %v", err)
	}

	if user.Login != "lookupuser" {
		t.Fatalf("Expected username 'lookupuser', got '%s'", user.Login)
	}
}

// TestStoreAndRetrieveData ensures data can be stored and retrieved correctly.
func TestStoreAndRetrieveData(t *testing.T) {
	setupTestDB(t)

	// Create a test user
	testUser := models.User{
		Login:      "datauser",
		Password:   "hashedpassword",
		MasterSeed: "dataseed",
	}
	if err := repo.CreateUser(&testUser); err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// Store test data
	testEntry := models.Vault{
		OwnerID:  uint(testUser.ID),
		DataType: pb.DataType_TEXT,
		Metadata: "Test Data",
		Data:     []byte("Encrypted text"),
	}
	if err := repo.StoreData(&testEntry); err != nil {
		t.Fatalf("Failed to store data: %v", err)
	}

	// Retrieve data
	dataEntries, err := repo.RetrieveData(uint(testUser.ID), pb.DataType_TEXT)
	if err != nil {
		t.Fatalf("Failed to retrieve data: %v", err)
	}
	if len(dataEntries) == 0 {
		t.Fatal("Expected at least one data entry, got none")
	}

	// Verify the stored data
	if string(dataEntries[0].Data) != "Encrypted text" {
		t.Fatalf("Stored data does not match, expected 'Encrypted text', got '%s'", string(dataEntries[0].Data))
	}
}

// TestGetMasterSeed ensures retrieving a user's master seed works.
func TestGetMasterSeed(t *testing.T) {
	setupTestDB(t)

	testUser := models.User{
		Login:      "seeduser",
		Password:   "hashedpassword",
		MasterSeed: "supersecretseed",
	}

	if err := repo.CreateUser(&testUser); err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	seed, err := repo.GetMasterSeed(uint(testUser.ID))
	if err != nil {
		t.Fatalf("Failed to retrieve master seed: %v", err)
	}

	if seed != "supersecretseed" {
		t.Fatalf("Expected master seed 'supersecretseed', got '%s'", seed)
	}
}
