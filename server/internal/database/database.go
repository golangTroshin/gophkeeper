// Package database handles database connection and initialization for the GophKeeper server.
package database

import (
	"fmt"
	"log"
	"os"

	"github.com/golangTroshin/gophkeeper/server/internal/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// DB is a global variable that holds the database connection.
var DB *gorm.DB

// InitDB initializes the database connection using environment variables.
// It sets up the PostgreSQL connection and performs automatic migrations.
//
// Environment Variables:
//   - DB_HOST: Database host address
//   - DB_PORT: Database port
//   - DB_USER: Database username
//   - DB_PASSWORD: Database password
//   - DB_NAME: Database name
//   - DB_SSLMODE: SSL mode for the connection (e.g., "disable", "require")
//
// Returns an error if the connection or migration fails.
func InitDB() error {
	var err error
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")
	dbSSLMode := os.Getenv("DB_SSLMODE")

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		dbHost, dbPort, dbUser, dbPassword, dbName, dbSSLMode)
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Printf("Failed to connect to the database: %v", err)
		return err
	}

	if err := DB.AutoMigrate(&models.User{}, &models.Vault{}); err != nil {
		log.Printf("Failed to migrate database: %v", err)
		return err
	}

	log.Println("Database connected and migrated successfully.")
	return nil
}
