package main

import (
	"flag"
	"log"
	"net/http"
	"os"

	"solid-go/internal/server"
	"solid-go/internal/storage"
)

func main() {
	// Parse command line flags
	port := flag.String("port", "8080", "Port to listen on")
	dataDir := flag.String("data", "./data", "Data directory")
	flag.Parse()

	// Create data directory if it doesn't exist
	if err := os.MkdirAll(*dataDir, 0755); err != nil {
		log.Fatalf("Failed to create data directory: %v", err)
	}

	// Initialize storage
	store := storage.NewFileStorage(*dataDir)

	// Initialize server
	srv := server.NewServer(store)

	// Start server
	addr := ":" + *port
	log.Printf("Starting server on %s", addr)
	if err := http.ListenAndServe(addr, srv); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
