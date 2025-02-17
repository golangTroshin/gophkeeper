// Package main is the entry point for the GophKeeper client application.
//
// GophKeeper is a secure password and data manager using gRPC for communication
// with the backend server. The client provides a TUI (Terminal User Interface)
// for users to authenticate and manage their encrypted data.
//
// Features:
// - User Authentication (Login / Sign-up)
// - Secure Data Storage with Encryption
// - Retrieve and Manage Data via gRPC
// - Interactive TUI using `tview`
package main

import (
	"log"

	"github.com/golangTroshin/gophkeeper/client/internal/forms"
	pb "github.com/golangTroshin/gophkeeper/grpc"
	"github.com/rivo/tview"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// client is a global gRPC client instance used to communicate with the GophKeeper server.
var client pb.GophKeeperServiceClient

// These variables will be set at build time
var (
	Version   = "dev"
	BuildDate = "unknown"
)

// main initializes the gRPC connection and starts the TUI application.
func main() {
	conn, err := grpc.NewClient("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect to gRPC server: %v", err)
	}
	defer conn.Close()
	client = pb.NewGophKeeperServiceClient(conn)

	app := tview.NewApplication()

	forms.ShowVersionInfo(app, client, Version, BuildDate)

	if err := app.Run(); err != nil {
		log.Fatalf("Error running TUI application: %v", err)
	}
}
