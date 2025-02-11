// Package handlers implements the gRPC server methods for the GophKeeper application.
package handlers

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/golang-jwt/jwt/v4"
	pb "github.com/golangTroshin/gophkeeper/grpc"
	"github.com/golangTroshin/gophkeeper/server/internal/models"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// GophKeeperServer implements the GophKeeper gRPC service.
type GophKeeperServer struct {
	pb.UnimplementedGophKeeperServiceServer
	DB *gorm.DB
}

var jwtSecret = []byte("$2a$10$1OTcy6ZovRCBv3wRLr3UseAPZgXTgGewGGTO/fctDauTR/QrCSnKu")

// GenerateJWT generates a JWT token for a given user ID.
func GenerateJWT(userID uint) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(time.Hour * 24).Unix(),
	})

	return token.SignedString(jwtSecret)
}

// VerifyToken verifies the validity of a JWT token and extracts the user ID.
func VerifyToken(tokenString string) (uint, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})

	if err != nil || !token.Valid {
		return 0, fmt.Errorf("invalid token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return 0, fmt.Errorf("invalid token claims")
	}

	userID, ok := claims["user_id"].(float64)
	if !ok {
		return 0, fmt.Errorf("invalid user ID in token")
	}

	return uint(userID), nil
}

// HashPassword securely hashes the password using bcrypt.
func HashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedPassword), nil
}

// VerifyPassword compares a hashed password with a plaintext password.
func VerifyPassword(hashedPassword, password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password)) == nil
}

func (s *GophKeeperServer) UserExists(ctx context.Context, req *pb.UserExistsRequest) (*pb.UserExistsResponse, error) {
	log.Printf("Checking if user exists: %s", req.Username)

	var existingUser models.User
	err := s.DB.Where("login = ?", req.Username).First(&existingUser).Error
	if err != nil {
		return &pb.UserExistsResponse{Success: false, Message: err.Error()}, nil
	}

	if existingUser.ID > 0 {
		return &pb.UserExistsResponse{Exists: true, Success: true}, nil
	}

	return &pb.UserExistsResponse{Exists: false, Success: true}, nil
}

// RegisterUser registers a new user, hashes the password, and generates a JWT token.
func (s *GophKeeperServer) RegisterUser(ctx context.Context, req *pb.RegisterUserRequest) (*pb.RegisterUserResponse, error) {
	log.Printf("Registering user: %s", req.Username)

	var existingUser models.User
	err := s.DB.Where("login = ?", req.Username).First(&existingUser).Error
	if err == nil {
		return &pb.RegisterUserResponse{
			Success: false,
			Message: "User already exists. Please Log in",
		}, nil
	} else if err != gorm.ErrRecordNotFound {
		log.Printf("Database error: %v", err)
		return &pb.RegisterUserResponse{
			Success: false,
			Message: "Internal server error",
		}, err
	}

	hashedPassword, err := HashPassword(req.Password)
	if err != nil {
		log.Printf("Error hashing password: %v", err)
		return &pb.RegisterUserResponse{
			Success: false,
			Message: "Failed to process password",
		}, err
	}

	newUser := models.User{
		Login:      req.Username,
		Password:   hashedPassword,
		MasterSeed: req.Seed,
	}

	if err := s.DB.Create(&newUser).Error; err != nil {
		log.Printf("Failed to create user: %v", err)
		return &pb.RegisterUserResponse{
			Success: false,
			Message: "Failed to register user",
		}, err
	}

	token, err := GenerateJWT(uint(newUser.ID))
	if err != nil {
		log.Printf("Failed to generate token: %v", err)
		return &pb.RegisterUserResponse{
			Success: false,
			Message: "Failed to generate token",
		}, err
	}

	return &pb.RegisterUserResponse{
		Success: true,
		Token:   token,
		Message: "User registered successfully",
	}, nil
}

// AuthenticateUser verifies user credentials and returns a JWT token.
func (s *GophKeeperServer) AuthenticateUser(ctx context.Context, req *pb.AuthenticateUserRequest) (*pb.AuthenticateUserResponse, error) {
	log.Printf("Authenticating user: %s", req.Username)

	var user models.User
	if err := s.DB.Where("login = ?", req.Username).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return &pb.AuthenticateUserResponse{
				Success: false,
				Message: "Invalid username or password",
			}, nil
		}
		log.Printf("Database error: %v", err)
		return &pb.AuthenticateUserResponse{
			Success: false,
			Message: "Internal server error",
		}, err
	}

	if bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)) != nil {
		log.Printf("Invalid username or password")
		return &pb.AuthenticateUserResponse{
			Success: false,
			Message: "Invalid username or password",
		}, nil
	}

	token, err := GenerateJWT(uint(user.ID))
	if err != nil {
		log.Printf("Failed to generate token: %v", err)
		return &pb.AuthenticateUserResponse{
			Success: false,
			Message: "Failed to generate token",
		}, err
	}

	log.Printf("Authenticating user: %v successful", user.ID)
	return &pb.AuthenticateUserResponse{
		Success: true,
		Token:   token,
		Message: "Authentication successful",
	}, nil
}

// StoreData saves encrypted user data into the database.
func (s *GophKeeperServer) StoreData(ctx context.Context, req *pb.StoreDataRequest) (*pb.StoreDataResponse, error) {
	userID, err := VerifyToken(req.Token)
	if err != nil {
		log.Printf("Wrong userID: %v, token: %v, error: %v", userID, req.Token, err.Error())
		return &pb.StoreDataResponse{
			Success: false,
			Message: "Unauthorized",
		}, nil
	}

	newEntry := models.Vault{
		OwnerID:  uint(userID),
		DataType: req.DataType,
		Data:     req.Data,
		Metadata: req.Metadata,
	}

	if err := s.DB.Create(&newEntry).Error; err != nil {
		log.Printf("Error while creating: %s", err.Error())
		return &pb.StoreDataResponse{
			Success: false,
			Message: "Failed to store data",
		}, err
	}

	log.Printf("Data stored successfully: %v", newEntry)
	return &pb.StoreDataResponse{
		Success: true,
		Message: "Data stored successfully",
	}, nil
}

// RetrieveData retrieves encrypted user data based on data type.
func (s *GophKeeperServer) RetrieveData(ctx context.Context, req *pb.RetrieveDataRequest) (*pb.RetrieveDataResponse, error) {
	userID, err := VerifyToken(req.Token)
	if err != nil {
		return &pb.RetrieveDataResponse{}, err
	}

	var entries []models.Vault
	err = s.DB.Where("owner_id = ? AND data_type = ?", userID, req.Filter).Find(&entries).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return &pb.RetrieveDataResponse{Items: []*pb.DataItem{}}, nil
		}
		return &pb.RetrieveDataResponse{}, err
	}

	var items []*pb.DataItem
	for _, entry := range entries {
		items = append(items, &pb.DataItem{
			DataType: entry.DataType,
			Metadata: entry.Metadata,
			Data:     entry.Data,
		})
	}

	return &pb.RetrieveDataResponse{Items: items}, nil
}

// MasterSeedRetrieve retrieves the encrypted master seed for a user.
func (s *GophKeeperServer) MasterSeedRetrieve(ctx context.Context, req *pb.MasterSeedRetrieveRequest) (*pb.MasterSeedRetrieveResponse, error) {
	userID, err := VerifyToken(req.Token)
	if err != nil {
		return &pb.MasterSeedRetrieveResponse{Success: false, Message: "Unauthorized"}, nil
	}

	var user models.User
	if err := s.DB.Where("id = ?", userID).First(&user).Error; err != nil {
		return &pb.MasterSeedRetrieveResponse{Success: false, Message: "User not found"}, nil
	}

	return &pb.MasterSeedRetrieveResponse{Success: true, MasterSeed: user.MasterSeed, Message: "Master seed retrieved successfully"}, nil
}
