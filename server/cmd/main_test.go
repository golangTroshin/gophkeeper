package main

import (
	"net"
	"os"
	"os/signal"
	"syscall"
	"testing"
	"time"

	"google.golang.org/grpc"

	pb "github.com/golangTroshin/gophkeeper/grpc"
	"github.com/golangTroshin/gophkeeper/server/internal/database"
	"github.com/golangTroshin/gophkeeper/server/internal/handlers"
)

// TestGRPCServerStartup ensures the gRPC server starts and handles shutdown correctly.
func TestGRPCServerStartup(t *testing.T) {
	// Listen on a random port
	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		t.Fatalf("Failed to create listener: %v", err)
	}
	defer listener.Close()

	grpcServer := grpc.NewServer()
	gophKeeperServer := &handlers.GophKeeperServer{DB: database.DB}
	pb.RegisterGophKeeperServiceServer(grpcServer, gophKeeperServer)

	// Channel to listen for shutdown signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Start the server in a separate goroutine
	go func() {
		if err := grpcServer.Serve(listener); err != nil {
			t.Errorf("Failed to start gRPC server: %v", err)
		}
	}()

	// Allow the server to start
	time.Sleep(1 * time.Second)

	// Send a shutdown signal
	sigChan <- os.Interrupt

	// Allow time for cleanup
	time.Sleep(1 * time.Second)

	// Gracefully stop the server
	grpcServer.GracefulStop()
}
