// Package models defines the database models used in the GophKeeper application.
package models

import (
	"time"

	pb "github.com/golangTroshin/gophkeeper/grpc"
)

// User represents a registered user in the system.
type User struct {
	ID         int32  `gorm:"primaryKey"`      // Unique identifier for the user
	Login      string `gorm:"unique;not null"` // User's login username (must be unique)
	Password   string `gorm:"not null"`        // Hashed password for authentication
	MasterSeed string `gorm:"not null"`        // Encrypted master seed for data encryption
}

// Vault represents a secure storage for user data.
type Vault struct {
	ID         uint        `gorm:"primaryKey"`     // Unique identifier for the stored data entry
	Data       []byte      `gorm:"not null"`       // Encrypted user data
	DataType   pb.DataType `gorm:"not null"`       // Type of data (e.g., credentials, text, binary, card)
	Metadata   string      `gorm:"not null"`       // Nullable Metadata describing the stored data
	OwnerID    uint        `gorm:"not null"`       // ID of the user who owns this data
	ModifiedAt time.Time   `gorm:"autoUpdateTime"` // Timestamp of last modification
	CreatedAt  time.Time   `gorm:"autoCreateTime"` // Timestamp of when the data was created
}
