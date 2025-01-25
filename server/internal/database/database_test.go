package database_test

import (
	"os"
	"testing"

	"github.com/golangTroshin/gophkeeper/server/internal/database"
	"github.com/golangTroshin/gophkeeper/server/internal/models"
	"github.com/joho/godotenv"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// TestInitDB checks if the database initializes correctly
func TestInitDB(t *testing.T) {
	err := godotenv.Load("../../.env")
	if err != nil {
		t.Fatalf("Error loading .env file: %v", err)
	}

	// Run the database initialization
	err = database.InitDB()
	if err != nil {
		t.Fatalf("Failed to initialize database: %v", err)
	}

	// Ensure the database instance is not nil
	if database.DB == nil {
		t.Fatal("Database connection is nil after InitDB")
	}
}

// TestMigrations ensures that database migrations run without errors
func TestMigrations(t *testing.T) {
	// Use an in-memory SQLite database for testing
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to create in-memory database: %v", err)
	}

	// Run migrations
	err = db.AutoMigrate(&models.User{}, &models.Vault{})
	if err != nil {
		t.Fatalf("Database migration failed: %v", err)
	}

	// Check if tables were created
	if !db.Migrator().HasTable(&models.User{}) || !db.Migrator().HasTable(&models.Vault{}) {
		t.Fatal("Expected tables were not created in the database")
	}
}

// TestInvalidDBConnection checks behavior when database connection fails
func TestInvalidDBConnection(t *testing.T) {
	// Set an invalid DSN to simulate failure
	os.Setenv("DB_HOST", "invalid_host")
	os.Setenv("DB_PORT", "invalid_port")

	err := database.InitDB()
	if err == nil {
		t.Fatal("Expected an error when initializing database with invalid connection details")
	}
}
