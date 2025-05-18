package auth

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHTTPHandler_Routes(t *testing.T) {
	// Create a mock factory with mock strategies
	factory := newMockFactory()
	storageFactory := newMockStorageFactory()

	// Create the authentication handler
	authHandler, err := NewHandler(factory, storageFactory)
	if err != nil {
		t.Fatalf("Failed to create authentication handler: %v", err)
	}

	// Create the HTTP handler
	httpHandler := NewHTTPHandler(authHandler)

	// Create a test server with the routes registered
	mux := http.NewServeMux()
	httpHandler.RegisterRoutes(mux)
	server := httptest.NewServer(mux)
	defer server.Close()

	// Test cases for endpoints
	tests := []struct {
		name           string
		method         string
		path           string
		expectedStatus int
	}{
		{
			name:           "Login GET method not allowed",
			method:         http.MethodGet,
			path:           "/login",
			expectedStatus: http.StatusMethodNotAllowed,
		},
		{
			name:           "Logout GET method not allowed",
			method:         http.MethodGet,
			path:           "/logout",
			expectedStatus: http.StatusMethodNotAllowed,
		},
		{
			name:           "Profile GET method allowed",
			method:         http.MethodGet,
			path:           "/profile",
			expectedStatus: http.StatusUnauthorized, // Not authenticated
		},
		{
			name:           "Callback GET method allowed",
			method:         http.MethodGet,
			path:           "/auth/callback",
			expectedStatus: http.StatusFound, // Should redirect
		},
		{
			name:           "Profile POST method not allowed",
			method:         http.MethodPost,
			path:           "/profile",
			expectedStatus: http.StatusMethodNotAllowed,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req, err := http.NewRequest(tc.method, server.URL+tc.path, nil)
			if err != nil {
				t.Fatalf("Failed to create request: %v", err)
			}

			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				t.Fatalf("Failed to execute request: %v", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != tc.expectedStatus {
				t.Errorf("Expected status %d, got %d", tc.expectedStatus, resp.StatusCode)
			}
		})
	}
}

func TestHTTPHandler_Login(t *testing.T) {
	// Create a mock factory with mock strategies
	factory := newMockFactory()
	storageFactory := newMockStorageFactory()

	// Create the authentication handler
	authHandler, err := NewHandler(factory, storageFactory)
	if err != nil {
		t.Fatalf("Failed to create authentication handler: %v", err)
	}

	// Create the HTTP handler
	httpHandler := NewHTTPHandler(authHandler)

	// Create a test server with the routes registered
	mux := http.NewServeMux()
	httpHandler.RegisterRoutes(mux)
	server := httptest.NewServer(mux)
	defer server.Close()

	// Add OIDC strategy to the factory
	oidcAgent := &Agent{
		ID:              "https://example.org/alice#me",
		Name:            "Alice",
		IsAuthenticated: true,
		Type:            TypeUser,
	}
	oidcStrategy := NewMockOIDCStrategy(oidcAgent)
	factory.RegisterStrategy("oidc", oidcStrategy)

	// Test cases for login
	tests := []struct {
		name           string
		requestBody    LoginRequest
		expectedStatus int
		checkToken     bool
	}{
		{
			name: "Valid OIDC login with token",
			requestBody: LoginRequest{
				Strategy: "oidc",
				Credentials: struct {
					Username string `json:"username,omitempty"`
					Password string `json:"password,omitempty"`
					Token    string `json:"token,omitempty"`
				}{
					Username: "https://example.org/alice#me",
					Token:    "test-token",
				},
			},
			expectedStatus: http.StatusOK,
			checkToken:     true,
		},
		{
			name: "Invalid strategy",
			requestBody: LoginRequest{
				Strategy: "invalid",
			},
			expectedStatus: http.StatusBadRequest,
			checkToken:     false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Marshal the request body
			body, err := json.Marshal(tc.requestBody)
			if err != nil {
				t.Fatalf("Failed to marshal request body: %v", err)
			}

			req, err := http.NewRequest(http.MethodPost, server.URL+"/login", bytes.NewBuffer(body))
			if err != nil {
				t.Fatalf("Failed to create request: %v", err)
			}
			req.Header.Set("Content-Type", "application/json")

			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				t.Fatalf("Failed to execute request: %v", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != tc.expectedStatus {
				t.Errorf("Expected status %d, got %d", tc.expectedStatus, resp.StatusCode)
			}

			if tc.checkToken {
				// Check if we got a session cookie
				var cookies []*http.Cookie
				for _, cookie := range resp.Cookies() {
					if cookie.Name == "session" {
						cookies = append(cookies, cookie)
					}
				}

				if len(cookies) == 0 {
					t.Error("No session cookie found")
				} else {
					cookie := cookies[0]
					if cookie.Value == "" {
						t.Error("Session cookie has empty value")
					}
					if !cookie.HttpOnly {
						t.Error("Session cookie is not HttpOnly")
					}
					if !cookie.Secure {
						t.Error("Session cookie is not Secure")
					}
				}

				// Check the response body
				var loginResp LoginResponse
				if err := json.NewDecoder(resp.Body).Decode(&loginResp); err != nil {
					t.Fatalf("Failed to decode response: %v", err)
				}

				if !loginResp.Success {
					t.Error("Login response indicates failure")
				}

				if loginResp.Token == "" {
					t.Error("Login response has empty token")
				}

				if loginResp.Agent == nil {
					t.Error("Login response has nil agent")
				} else if loginResp.Agent.ID != tc.requestBody.Credentials.Username {
					t.Errorf("Login response agent ID = %s, want %s", loginResp.Agent.ID, tc.requestBody.Credentials.Username)
				}
			}
		})
	}
}

func TestHTTPHandler_Logout(t *testing.T) {
	// Create a mock factory with mock strategies
	factory := newMockFactory()
	storageFactory := newMockStorageFactory()

	// Create the authentication handler
	authHandler, err := NewHandler(factory, storageFactory)
	if err != nil {
		t.Fatalf("Failed to create authentication handler: %v", err)
	}

	// Create the HTTP handler
	httpHandler := NewHTTPHandler(authHandler)

	// Create a test server with the routes registered
	mux := http.NewServeMux()
	httpHandler.RegisterRoutes(mux)
	server := httptest.NewServer(mux)
	defer server.Close()

	// Test logout
	req, err := http.NewRequest(http.MethodPost, server.URL+"/logout", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	// Set a session cookie
	req.AddCookie(&http.Cookie{
		Name:  "session",
		Value: "test-session",
	})

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("Failed to execute request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, resp.StatusCode)
	}

	// Check if the session cookie was cleared
	var cookies []*http.Cookie
	for _, cookie := range resp.Cookies() {
		if cookie.Name == "session" {
			cookies = append(cookies, cookie)
		}
	}

	if len(cookies) == 0 {
		t.Error("No session cookie found in response")
	} else {
		cookie := cookies[0]
		if cookie.Value != "" {
			t.Error("Session cookie value not cleared")
		}
		if cookie.MaxAge >= 0 {
			t.Error("Session cookie MaxAge not negative")
		}
	}

	// Check the response body
	var logoutResp map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&logoutResp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	success, ok := logoutResp["success"].(bool)
	if !ok || !success {
		t.Error("Logout response does not indicate success")
	}
}

func TestHTTPHandler_Profile(t *testing.T) {
	// Create a mock factory with mock strategies
	factory := newMockFactory()
	storageFactory := newMockStorageFactory()

	// Create the authentication handler
	authHandler, err := NewHandler(factory, storageFactory)
	if err != nil {
		t.Fatalf("Failed to create authentication handler: %v", err)
	}

	// Create the HTTP handler
	httpHandler := NewHTTPHandler(authHandler)

	// Create a test server with the routes registered
	mux := http.NewServeMux()
	httpHandler.RegisterRoutes(mux)
	server := httptest.NewServer(mux)
	defer server.Close()

	// Add OIDC strategy to the factory
	oidcAgent := &Agent{
		ID:              "https://example.org/alice#me",
		Name:            "Alice",
		IsAuthenticated: true,
		Type:            TypeUser,
	}
	oidcStrategy := &mockOIDCStrategy{
		agent:   oidcAgent,
		enabled: true,
	}
	factory.RegisterStrategy("oidc", oidcStrategy)

	// Test cases for profile
	tests := []struct {
		name           string
		authenticated  bool
		expectedStatus int
	}{
		{
			name:           "Authenticated user",
			authenticated:  true,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Unauthenticated user",
			authenticated:  false,
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodGet, server.URL+"/profile", nil)
			if err != nil {
				t.Fatalf("Failed to create request: %v", err)
			}

			if tc.authenticated {
				// Register a session for the user
				oidcStrategy.agent.IsAuthenticated = true
				req.Header.Set("Authorization", "Bearer test-token")
			}

			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				t.Fatalf("Failed to execute request: %v", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != tc.expectedStatus {
				t.Errorf("Expected status %d, got %d", tc.expectedStatus, resp.StatusCode)
			}

			if tc.expectedStatus == http.StatusOK {
				// Check the response body
				var agent Agent
				if err := json.NewDecoder(resp.Body).Decode(&agent); err != nil {
					t.Fatalf("Failed to decode response: %v", err)
				}

				if agent.ID != oidcAgent.ID {
					t.Errorf("Agent ID = %s, want %s", agent.ID, oidcAgent.ID)
				}

				if agent.Name != oidcAgent.Name {
					t.Errorf("Agent Name = %s, want %s", agent.Name, oidcAgent.Name)
				}

				if !agent.IsAuthenticated {
					t.Error("Agent is not authenticated")
				}
			}
		})
	}
}

func TestHTTPHandler_Middleware(t *testing.T) {
	// Create a mock factory with mock strategies
	factory := newMockFactory()
	storageFactory := newMockStorageFactory()

	// Create the authentication handler
	authHandler, err := NewHandler(factory, storageFactory)
	if err != nil {
		t.Fatalf("Failed to create authentication handler: %v", err)
	}

	// Create the HTTP handler
	httpHandler := NewHTTPHandler(authHandler)

	// Add OIDC strategy to the factory
	oidcAgent := &Agent{
		ID:              "https://example.org/alice#me",
		Name:            "Alice",
		IsAuthenticated: true,
		Type:            TypeUser,
	}
	oidcStrategy := &mockOIDCStrategy{
		agent:   oidcAgent,
		enabled: true,
	}
	factory.RegisterStrategy("oidc", oidcStrategy)

	// Test handler that checks the agent context
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		agent, ok := AgentFromContext(r.Context())
		if !ok {
			t.Error("No agent found in context")
			http.Error(w, "No agent in context", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(agent)
	})

	// Test cases for middleware
	tests := []struct {
		name           string
		middleware     func(http.Handler) http.Handler
		authenticated  bool
		expectedStatus int
	}{
		{
			name:           "RequireAuth with authenticated user",
			middleware:     httpHandler.RequireAuth,
			authenticated:  true,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "RequireAuth with unauthenticated user",
			middleware:     httpHandler.RequireAuth,
			authenticated:  false,
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "OptionalAuth with authenticated user",
			middleware:     httpHandler.OptionalAuth,
			authenticated:  true,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "OptionalAuth with unauthenticated user",
			middleware:     httpHandler.OptionalAuth,
			authenticated:  false,
			expectedStatus: http.StatusOK, // Should still pass through
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Create a test server with the middleware
			server := httptest.NewServer(tc.middleware(testHandler))
			defer server.Close()

			req, err := http.NewRequest(http.MethodGet, server.URL, nil)
			if err != nil {
				t.Fatalf("Failed to create request: %v", err)
			}

			if tc.authenticated {
				oidcStrategy.agent.IsAuthenticated = true
				req.Header.Set("Authorization", "Bearer test-token")
			} else {
				oidcStrategy.agent.IsAuthenticated = false
			}

			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				t.Fatalf("Failed to execute request: %v", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != tc.expectedStatus {
				t.Errorf("Expected status %d, got %d", tc.expectedStatus, resp.StatusCode)
			}

			if resp.StatusCode == http.StatusOK {
				// Check the response body
				var agent Agent
				if err := json.NewDecoder(resp.Body).Decode(&agent); err != nil {
					t.Fatalf("Failed to decode response: %v", err)
				}

				if tc.authenticated {
					if agent.ID != oidcAgent.ID {
						t.Errorf("Agent ID = %s, want %s", agent.ID, oidcAgent.ID)
					}
					if !agent.IsAuthenticated {
						t.Error("Agent is not authenticated")
					}
				} else if tc.middleware != nil {
					// For OptionalAuth without authentication, we should get an anonymous agent
					if agent.IsAuthenticated {
						t.Error("Anonymous agent should not be authenticated")
					}
					if agent.Type != TypeAnonymous {
						t.Errorf("Agent type = %v, want %v", agent.Type, TypeAnonymous)
					}
				}
			}
		})
	}
}

// mockOIDCStrategy wrapper to make testing easier
type mockOIDCStrategy struct {
	agent   *Agent
	enabled bool
}

func (s *mockOIDCStrategy) Name() string {
	return "oidc"
}

func (s *mockOIDCStrategy) IsEnabled() bool {
	return s.enabled
}

func (s *mockOIDCStrategy) Authenticate(r *http.Request) (*Agent, error) {
	if !s.enabled {
		return nil, ErrAuthenticationFailed
	}
	if !s.agent.IsAuthenticated {
		return nil, ErrAuthenticationFailed
	}
	return s.agent, nil
}
