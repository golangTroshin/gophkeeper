package handlers

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/golang-jwt/jwt/v4"
	pb "github.com/golangTroshin/gophkeeper/grpc"
	"github.com/golangTroshin/gophkeeper/server/internal/models"
	"github.com/golangTroshin/gophkeeper/server/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

// GophKeeperServer implements the GophKeeper gRPC service.
type GophKeeperServer struct {
	pb.UnimplementedGophKeeperServiceServer
	Repo repository.Repository
}

var jwtSecret = []byte("$2a$10$1OTcy6ZovRCBv3wRLr3UseAPZgXTgGewGGTO/fctDauTR/QrCSnKu")

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

// GenerateJWT generates a JWT token for a given user ID.
func GenerateJWT(userID uint) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(time.Hour * 24).Unix(),
	})
	return token.SignedString(jwtSecret)
}

// RegisterUser registers a new user, hashes the password, and stores the master seed.
func (s *GophKeeperServer) RegisterUser(ctx context.Context, req *pb.RegisterUserRequest) (*pb.RegisterUserResponse, error) {
	log.Printf("Registering user: %s", req.Username)

	exists, err := s.Repo.UserExists(req.Username)
	if err != nil {
		return &pb.RegisterUserResponse{Success: false, Message: "Database error"}, err
	}
	if exists {
		return &pb.RegisterUserResponse{Success: false, Message: "User already exists"}, nil
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return &pb.RegisterUserResponse{Success: false, Message: "Failed to process password"}, err
	}

	user := models.User{
		Login:      req.Username,
		Password:   string(hashedPassword),
		MasterSeed: req.Seed,
	}
	if err := s.Repo.CreateUser(&user); err != nil {
		return &pb.RegisterUserResponse{Success: false, Message: "Failed to register user"}, err
	}

	token, err := GenerateJWT(uint(user.ID))
	if err != nil {
		return &pb.RegisterUserResponse{Success: false, Message: "Failed to generate token"}, err
	}

	return &pb.RegisterUserResponse{Success: true, Token: token, Message: "User registered successfully"}, nil
}

// AuthenticateUser verifies user credentials and returns a JWT token.
func (s *GophKeeperServer) AuthenticateUser(ctx context.Context, req *pb.AuthenticateUserRequest) (*pb.AuthenticateUserResponse, error) {
	log.Printf("Authenticating user: %s", req.Username)

	user, err := s.Repo.GetUserByLogin(req.Username)
	if err != nil {
		return &pb.AuthenticateUserResponse{Success: false, Message: "Invalid username or password"}, nil
	}

	if bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)) != nil {
		return &pb.AuthenticateUserResponse{Success: false, Message: "Invalid username or password"}, nil
	}

	token, err := GenerateJWT(uint(user.ID))
	if err != nil {
		return &pb.AuthenticateUserResponse{Success: false, Message: "Failed to generate token"}, err
	}

	return &pb.AuthenticateUserResponse{Success: true, Token: token, Message: "Authentication successful"}, nil
}

// StoreData saves encrypted user data into the database.
func (s *GophKeeperServer) StoreData(ctx context.Context, req *pb.StoreDataRequest) (*pb.StoreDataResponse, error) {
	userID, err := VerifyToken(req.Token)
	if err != nil {
		return &pb.StoreDataResponse{Success: false, Message: "Unauthorized"}, nil
	}

	entry := models.Vault{
		OwnerID:  uint(userID),
		DataType: req.DataType,
		Data:     req.Data,
		Metadata: req.Metadata,
	}
	if err := s.Repo.StoreData(&entry); err != nil {
		return &pb.StoreDataResponse{Success: false, Message: "Failed to store data"}, err
	}

	return &pb.StoreDataResponse{Success: true, Message: "Data stored successfully"}, nil
}

// RetrieveData retrieves encrypted user data based on data type.
func (s *GophKeeperServer) RetrieveData(ctx context.Context, req *pb.RetrieveDataRequest) (*pb.RetrieveDataResponse, error) {
	userID, err := VerifyToken(req.Token)
	if err != nil {
		return &pb.RetrieveDataResponse{}, err
	}

	entries, err := s.Repo.RetrieveData(userID, req.Filter)
	if err != nil {
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

	masterSeed, err := s.Repo.GetMasterSeed(userID)
	if err != nil {
		return &pb.MasterSeedRetrieveResponse{Success: false, Message: "User not found"}, nil
	}

	return &pb.MasterSeedRetrieveResponse{Success: true, MasterSeed: masterSeed, Message: "Master seed retrieved successfully"}, nil
}
