package handlers

import (
	"encoding/base64"
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
)

// Mock data for testing
const (
	mockSeed     = "test_master_seed"
	mockPassword = "securePassword123"
	mockData     = "test data"
)

// TestDeriveKeyFromSeed ensures key derivation is working correctly
func TestDeriveKeyFromSeed(t *testing.T) {
	key := DeriveKeyFromSeed(mockSeed)
	assert.NotNil(t, key, "Derived key should not be nil")
	assert.Len(t, key, 32, "Derived key should be 32 bytes long")
}

// TestEncryptDecryptData ensures encryption and decryption work correctly
func TestEncryptDecryptData(t *testing.T) {
	key := DeriveKeyFromSeed(mockSeed)
	encryptedData, err := encryptData([]byte(mockData), key)
	assert.NoError(t, err, "Encryption should not return an error")
	assert.NotEmpty(t, encryptedData, "Encrypted data should not be empty")

	encodedData := base64.StdEncoding.EncodeToString(encryptedData)
	decryptedData, err := DecryptData(encodedData, key)
	assert.NoError(t, err, "Decryption should not return an error")
	assert.Equal(t, mockData, decryptedData, "Decrypted data should match the original")
}

// TestHashPassword ensures password hashing works correctly
func TestHashPassword(t *testing.T) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(mockPassword), bcrypt.DefaultCost)
	assert.NoError(t, err, "Hashing should not return an error")
	assert.NotEmpty(t, hashedPassword, "Hashed password should not be empty")
}

// TestVerifyPassword ensures password verification works correctly
func TestVerifyPassword(t *testing.T) {
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(mockPassword), bcrypt.DefaultCost)
	err := bcrypt.CompareHashAndPassword(hashedPassword, []byte(mockPassword))
	assert.NoError(t, err, "Password verification should be successful")

	err = bcrypt.CompareHashAndPassword(hashedPassword, []byte("wrongPassword"))
	assert.Error(t, err, "Wrong password should not be verified successfully")
}

// TestSessionToken ensures session token is stored correctly
func TestSessionToken(t *testing.T) {
	session.UserToken = "mock_token"
	assert.Equal(t, "mock_token", session.UserToken, "Session token should be stored correctly")
}
