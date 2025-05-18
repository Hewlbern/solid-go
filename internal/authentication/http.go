package auth

import (
	"encoding/json"
	"net/http"
	"time"
)

// HTTPHandler provides HTTP endpoints for authentication related operations
type HTTPHandler struct {
	authHandler *Handler
}

// NewHTTPHandler creates a new HTTPHandler
func NewHTTPHandler(authHandler *Handler) *HTTPHandler {
	return &HTTPHandler{
		authHandler: authHandler,
	}
}

// RegisterRoutes registers the authentication routes on the given router
func (h *HTTPHandler) RegisterRoutes(mux *http.ServeMux) {
	// Login endpoint
	mux.HandleFunc("/login", h.handleLogin)

	// Logout endpoint
	mux.HandleFunc("/logout", h.handleLogout)

	// User profile endpoint
	mux.HandleFunc("/profile", h.handleProfile)

	// OIDC callback endpoint
	mux.HandleFunc("/auth/callback", h.handleOIDCCallback)
}

// LoginRequest represents a login request
type LoginRequest struct {
	Strategy    string `json:"strategy"`
	Credentials struct {
		Username string `json:"username,omitempty"`
		Password string `json:"password,omitempty"`
		Token    string `json:"token,omitempty"`
	} `json:"credentials,omitempty"`
}

// LoginResponse represents a login response
type LoginResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
	Token   string `json:"token,omitempty"`
	Agent   *Agent `json:"agent,omitempty"`
}

// handleLogin handles login requests
func (h *HTTPHandler) handleLogin(w http.ResponseWriter, r *http.Request) {
	// Only allow POST requests
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse the login request
	var loginReq LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&loginReq); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Get the strategy
	strategy, err := h.authHandler.factory.GetStrategy(loginReq.Strategy)
	if err != nil {
		http.Error(w, "Authentication strategy not available", http.StatusBadRequest)
		return
	}

	// For OIDC strategy, we need to handle token-based authentication
	if loginReq.Strategy == "oidc" {
		oidcStrategy, ok := strategy.(*OIDCStrategy)
		if !ok {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		// If a token is provided, validate it
		if loginReq.Credentials.Token != "" {
			// Validate the token
			agent, err := oidcStrategy.validateToken(loginReq.Credentials.Token)
			if err != nil {
				http.Error(w, "Invalid token", http.StatusUnauthorized)
				return
			}

			// Create the response
			resp := LoginResponse{
				Success: true,
				Message: "Login successful",
				Token:   loginReq.Credentials.Token,
				Agent:   agent,
			}

			// Set the session cookie
			http.SetCookie(w, &http.Cookie{
				Name:     "session",
				Value:    loginReq.Credentials.Token,
				Path:     "/",
				HttpOnly: true,
				Secure:   true,
				SameSite: http.SameSiteStrictMode,
			})

			// Write the response
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
			return
		}
	}

	// For other strategies, authenticate using credentials
	agent, err := strategy.Authenticate(r)
	if err != nil {
		http.Error(w, "Authentication failed", http.StatusUnauthorized)
		return
	}

	// Create the response
	resp := LoginResponse{
		Success: true,
		Message: "Login successful",
		Agent:   agent,
	}

	// Write the response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// handleLogout handles logout requests
func (h *HTTPHandler) handleLogout(w http.ResponseWriter, r *http.Request) {
	// Only allow POST requests
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Clear the session cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   -1,
	})

	// Write the response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"success": true})
}

// handleProfile handles profile requests
func (h *HTTPHandler) handleProfile(w http.ResponseWriter, r *http.Request) {
	// Only allow GET requests
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get the agent from the request context
	agent := GetAgent(r.Context())
	if !agent.IsAuthenticated {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Write the response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(agent)
}

// handleOIDCCallback handles OIDC callback requests
func (h *HTTPHandler) handleOIDCCallback(w http.ResponseWriter, r *http.Request) {
	// Only allow GET requests
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get the OIDC strategy
	strategy, err := h.authHandler.factory.GetStrategy("oidc")
	if err != nil {
		http.Error(w, "Authentication strategy not available", http.StatusBadRequest)
		return
	}

	oidcStrategy, ok := strategy.(*OIDCStrategy)
	if !ok {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Authenticate using the OIDC strategy
	agent, err := oidcStrategy.Authenticate(r)
	if err != nil {
		http.Error(w, "Authentication failed", http.StatusUnauthorized)
		return
	}

	// Create a session token
	token, err := oidcStrategy.RegisterSession(agent.ID, time.Hour)
	if err != nil {
		http.Error(w, "Failed to create session", http.StatusInternalServerError)
		return
	}

	// Set the session cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
	})

	// Redirect to the profile page
	http.Redirect(w, r, "/profile", http.StatusFound)
}

// MiddlewareFunc is a type for middleware functions
type MiddlewareFunc func(http.Handler) http.Handler

// AuthMiddleware creates a middleware that authenticates requests
func (h *HTTPHandler) AuthMiddleware(required bool) MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Authenticate the request
			agent, err := h.authHandler.Authenticate(r)

			// If authentication is required and it failed, return an error
			if required && (err != nil || !agent.IsAuthenticated) {
				http.Error(w, "Authentication required", http.StatusUnauthorized)
				return
			}

			// Store the agent in the request context
			ctx := r.Context()
			ctx = NewAgentContext(ctx, agent)
			r = r.WithContext(ctx)

			// Call the next handler
			next.ServeHTTP(w, r)
		})
	}
}

// RequireAuth is a middleware that requires authentication
func (h *HTTPHandler) RequireAuth(next http.Handler) http.Handler {
	return h.AuthMiddleware(true)(next)
}

// OptionalAuth is a middleware that makes authentication optional
func (h *HTTPHandler) OptionalAuth(next http.Handler) http.Handler {
	return h.AuthMiddleware(false)(next)
}
