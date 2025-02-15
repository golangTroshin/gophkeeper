// Package repository provides an abstraction layer for database operations in GophKeeper.
// It defines an interface for user and data storage operations and implements these methods using GORM.
package repository

import (
	pb "github.com/golangTroshin/gophkeeper/grpc"
	"github.com/golangTroshin/gophkeeper/server/internal/models"

	"gorm.io/gorm"
)

// Repository defines the database operations for GophKeeper.
type Repository interface {
	UserExists(username string) (bool, error)
	CreateUser(user *models.User) error
	GetUserByLogin(username string) (*models.User, error)
	StoreData(entry *models.Vault) error
	RetrieveData(userID uint, dataType pb.DataType) ([]models.Vault, error)
	GetMasterSeed(userID uint) (string, error)
}

// repositoryImpl is the concrete implementation of Repository using GORM.
type repositoryImpl struct {
	db *gorm.DB
}

// NewRepository creates a new repository instance.
func NewRepository(db *gorm.DB) Repository {
	return &repositoryImpl{db: db}
}

// UserExists checks if a user exists in the database.
func (r *repositoryImpl) UserExists(username string) (bool, error) {
	var count int64
	err := r.db.Model(&models.User{}).Where("login = ?", username).Count(&count).Error
	return count > 0, err
}

// CreateUser saves a new user to the database.
func (r *repositoryImpl) CreateUser(user *models.User) error {
	return r.db.Create(user).Error
}

// GetUserByLogin retrieves a user by their username.
func (r *repositoryImpl) GetUserByLogin(username string) (*models.User, error) {
	var user models.User
	err := r.db.Where("login = ?", username).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// StoreData saves encrypted user data into the database.
func (r *repositoryImpl) StoreData(entry *models.Vault) error {
	return r.db.Create(entry).Error
}

// RetrieveData fetches stored data of a given type for a user.
func (r *repositoryImpl) RetrieveData(userID uint, dataType pb.DataType) ([]models.Vault, error) {
	var entries []models.Vault
	err := r.db.Where("owner_id = ? AND data_type = ?", userID, dataType).Find(&entries).Error
	if err != nil {
		return nil, err
	}
	return entries, nil
}

// GetMasterSeed retrieves the encrypted master seed for a user.
func (r *repositoryImpl) GetMasterSeed(userID uint) (string, error) {
	var user models.User
	err := r.db.Where("id = ?", userID).First(&user).Error
	if err != nil {
		return "", err
	}
	return user.MasterSeed, nil
}
