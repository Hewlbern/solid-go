package server

import (
	"crypto/tls"
	"net/http"
	"os"
)

// BaseServerFactoryOptions contains options for creating a server
type BaseServerFactoryOptions struct {
	// HTTPS indicates if the server should use HTTPS
	HTTPS bool

	// TLS certificate and key files
	CertFile string
	KeyFile  string
}

// BaseServerFactory creates HTTP(S) servers
type BaseServerFactory struct {
	configurator ServerConfigurator
	options      BaseServerFactoryOptions
}

// NewBaseServerFactory creates a new BaseServerFactory
func NewBaseServerFactory(configurator ServerConfigurator, options *BaseServerFactoryOptions) *BaseServerFactory {
	opts := BaseServerFactoryOptions{
		HTTPS: false,
	}
	if options != nil {
		opts = *options
	}
	return &BaseServerFactory{
		configurator: configurator,
		options:      opts,
	}
}

// CreateServer creates a new HTTP(S) server
func (f *BaseServerFactory) CreateServer() *http.Server {
	// Create server with TLS if enabled
	var srv *http.Server
	if f.options.HTTPS {
		// Read certificate and key files
		cert, err := os.ReadFile(f.options.CertFile)
		if err != nil {
			panic(err)
		}
		key, err := os.ReadFile(f.options.KeyFile)
		if err != nil {
			panic(err)
		}

		// Create TLS config
		tlsConfig := &tls.Config{
			Certificates: []tls.Certificate{},
		}
		certificate, err := tls.X509KeyPair(cert, key)
		if err != nil {
			panic(err)
		}
		tlsConfig.Certificates = append(tlsConfig.Certificates, certificate)

		// Create HTTPS server
		srv = &http.Server{
			TLSConfig: tlsConfig,
		}
	} else {
		// Create HTTP server
		srv = &http.Server{}
	}

	// Configure server
	f.configurator.Configure(http.NewServeMux())

	return srv
}
