package handlers

import (
	"context"
	"net/http"
	"time"
)

// HandlerUtil provides utility functions for HTTP handlers
type HandlerUtil struct{}

// NewHandlerUtil creates a new HandlerUtil
func NewHandlerUtil() *HandlerUtil {
	return &HandlerUtil{}
}

// Handler is a function that handles HTTP requests
type Handler func(http.ResponseWriter, *http.Request) error

// Middleware is a function that wraps a handler
type Middleware func(Handler) Handler

// ServeHTTP implements the http.Handler interface
func (h Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := h(w, r); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// WithMiddleware wraps a handler with one or more middleware functions
func (h *HandlerUtil) WithMiddleware(handler Handler, middleware ...Middleware) Handler {
	for i := len(middleware) - 1; i >= 0; i-- {
		handler = middleware[i](handler)
	}
	return handler
}

// WithContext wraps a handler with a context
func (h *HandlerUtil) WithContext(ctx context.Context) Middleware {
	return func(next Handler) Handler {
		return func(w http.ResponseWriter, r *http.Request) error {
			return next(w, r.WithContext(ctx))
		}
	}
}

// WithTimeout wraps a handler with a timeout
func (h *HandlerUtil) WithTimeout(timeout time.Duration) Middleware {
	return func(next Handler) Handler {
		return func(w http.ResponseWriter, r *http.Request) error {
			ctx, cancel := context.WithTimeout(r.Context(), timeout)
			defer cancel()
			return next(w, r.WithContext(ctx))
		}
	}
}

// WithRecovery wraps a handler to recover from panics
func (h *HandlerUtil) WithRecovery() Middleware {
	return func(next Handler) Handler {
		return func(w http.ResponseWriter, r *http.Request) (err error) {
			defer func() {
				if r := recover(); r != nil {
					err = http.ErrAbortHandler
				}
			}()
			return next(w, r)
		}
	}
}

// WithCORS wraps a handler to handle CORS
func (h *HandlerUtil) WithCORS() Middleware {
	return func(next Handler) Handler {
		return func(w http.ResponseWriter, r *http.Request) error {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return nil
			}
			return next(w, r)
		}
	}
}

// WithErrorHandling wraps a handler to handle errors
func (h *HandlerUtil) WithErrorHandling() Middleware {
	return func(next Handler) Handler {
		return func(w http.ResponseWriter, r *http.Request) error {
			err := next(w, r)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			return nil
		}
	}
}
