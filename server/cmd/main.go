// Package main initializes and runs the GophKeeper gRPC server.
package main

import (
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	pb "github.com/golangTroshin/gophkeeper/grpc"
	"github.com/golangTroshin/gophkeeper/server/internal/database"
	"github.com/golangTroshin/gophkeeper/server/internal/handlers"
	"github.com/joho/godotenv"
)

// main is the entry point of the GophKeeper gRPC server.
// It initializes the environment, sets up the database, and starts the server.
func main() {
	// Load environment variables from .env file.
	err := godotenv.Load("../.env")
	if err != nil {
		log.Printf("Error loading .env file: %v", err)
	}

	// Initialize the database connection.
	err = database.InitDB()
	if err != nil {
		log.Printf("Failed to initialize database: %v", err)
	}
	defer func() {
		log.Println("Closing database connection...")
		sqlDB, err := database.DB.DB()
		if err != nil {
			log.Printf("Error getting underlying DB: %v", err)
		}
		sqlDB.Close()
		log.Println("Database connection closed.")
	}()

	grpcServer := grpc.NewServer()

	gophKeeperServer := &handlers.GophKeeperServer{DB: database.DB}
	pb.RegisterGophKeeperServiceServer(grpcServer, gophKeeperServer)

	reflection.Register(grpcServer)

	listener, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Printf("Failed to listen: %v", err)
	}

	log.Println("Server is listening on port 50051")

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-c
		log.Println("Shutting down gracefully...")
		grpcServer.GracefulStop()
		log.Println("gRPC server stopped.")
		os.Exit(0)
	}()

	if err := grpcServer.Serve(listener); err != nil {
		log.Printf("Failed to serve: %v", err)
	}
}
