package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"syscall"

	"solid-go/internal/logging"
	"solid-go/internal/server"
	"solid-go/internal/storage"
)

func main() {
	// Create logger
	logger := logging.NewBasicLogger(logging.Info)

	// Parse command line flags
	port := flag.Int("port", 3000, "Port to listen on")
	https := flag.Bool("https", false, "Use HTTPS")
	certFile := flag.String("cert", "", "Path to TLS certificate file")
	keyFile := flag.String("key", "", "Path to TLS private key file")
	storagePath := flag.String("storage", "./data", "Path to storage directory")
	flag.Parse()

	// Create storage
	store, err := storage.NewFileStorage(*storagePath)
	if err != nil {
		logger.Error("Error creating storage: %v", err)
		os.Exit(1)
	}

	// Create server options
	options := &server.ServerOptions{
		Port:     *port,
		HTTPS:    *https,
		CertFile: *certFile,
		KeyFile:  *keyFile,
		Storage:  store,
		Logger:   logger,
	}

	// Create server
	srv := server.NewServer(options)

	// Create context that listens for the interrupt signal from the OS
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Start server
	go func() {
		var err error
		if *https {
			if *certFile == "" || *keyFile == "" {
				logger.Error("TLS certificate and key files are required for HTTPS")
				os.Exit(1)
			}
			err = srv.ListenAndServeTLS(*certFile, *keyFile)
		} else {
			err = srv.Start()
		}
		if err != nil {
			logger.Error("Error starting server: %v", err)
			os.Exit(1)
		}
	}()

	logger.Info("Server listening on port %d", *port)

	// Wait for interrupt signal
	<-ctx.Done()

	// Shutdown server
	if err := srv.Shutdown(context.Background()); err != nil {
		logger.Error("Error shutting down server: %v", err)
		os.Exit(1)
	}

	logger.Info("Server stopped")
}
