package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/yourusername/solid-go/internal/server"
)

func main() {
	// Define command-line flags
	port := flag.Int("port", 8080, "Port to run the server on")
	storagePath := flag.String("storage", "./data", "Path to store data")
	logLevel := flag.String("log-level", "info", "Logging level (debug, info, warn, error)")
	flag.Parse()

	// Override with environment variables if present
	if envPort := os.Getenv("SOLID_PORT"); envPort != "" {
		if p, err := strconv.Atoi(envPort); err == nil {
			*port = p
		}
	}

	if envStoragePath := os.Getenv("SOLID_STORAGE_PATH"); envStoragePath != "" {
		*storagePath = envStoragePath
	}

	if envLogLevel := os.Getenv("SOLID_LOG_LEVEL"); envLogLevel != "" {
		*logLevel = envLogLevel
	}

	// Set up logging
	switch *logLevel {
	case "debug":
		log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	case "info":
		log.SetFlags(log.Ldate | log.Ltime)
	default:
		log.SetFlags(log.Ldate | log.Ltime)
	}

	log.Printf("Starting Solid server on port %d with storage at %s", *port, *storagePath)

	// Create server facade
	solidServer, err := server.NewServerFacade(&server.Config{
		Port:        *port,
		StoragePath: *storagePath,
		LogLevel:    *logLevel,
	})

	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}

	// Start the server
	if err := solidServer.Start(); err != nil {
		log.Fatalf("Server error: %v", err)
	}

	fmt.Println("Solid server stopped")
} 