package wac

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// Test the required access mode based on HTTP method.
func TestRequiredMode(t *testing.T) {
	tests := []struct {
		method string
		path   string
		webID  string
		want   AccessMode
	}{
		{http.MethodGet, "/resource", "https://example.com/profile/user#me", Read},
		{http.MethodHead, "/resource", "https://example.com/profile/user#me", Read},
		{http.MethodOptions, "/resource", "https://example.com/profile/user#me", Read},
		{http.MethodPut, "/resource", "https://example.com/profile/user#me", Write},
		{http.MethodDelete, "/resource", "https://example.com/profile/user#me", Write},
		{http.MethodPost, "/resource", "https://example.com/profile/user#me", Append},
		{http.MethodPatch, "/resource", "https://example.com/profile/user#me", Write},
		{"UNKNOWN", "/resource", "https://example.com/profile/user#me", Read},
	}

	for _, tt := range tests {
		t.Run(tt.method, func(t *testing.T) {
			req := &ResourceRequest{
				Method: tt.method,
				Path:   tt.path,
				WebID:  tt.webID,
			}
			if got := req.RequiredMode(); got != tt.want {
				t.Errorf("RequiredMode() = %v, want %v", got, tt.want)
			}
		})
	}
}

// Test the authorization of requests.
func TestAuthorizeRequest(t *testing.T) {
	storage := NewMockStorage()
	wacHandler, err := NewHandler(storage)
	if err != nil {
		t.Fatalf("Failed to create WAC handler: %v", err)
	}

	authHandler := NewMockAuthHandler()
	httpHandler := NewHTTPHandler(wacHandler, authHandler)

	// Add some ACLs to the storage
	resourcePath := "/resource"
	resourceACLPath := "/resource.acl"
	storage.AddResource(resourceACLPath, CreateSimpleACL(resourcePath, "https://example.com/profile/owner#me", Write, Control))

	// Create test requests
	tests := []struct {
		name        string
		method      string
		path        string
		userID      string
		wantAuth    bool
		wantErrMsg  string
		setupCookie bool
	}{
		{
			name:        "Public read access to resource",
			method:      http.MethodGet,
			path:        resourcePath,
			userID:      "",
			wantAuth:    true,
			setupCookie: false,
		},
		{
			name:        "Public write access to resource",
			method:      http.MethodPut,
			path:        resourcePath,
			userID:      "",
			wantAuth:    false,
			setupCookie: false,
		},
		{
			name:        "Owner write access to resource",
			method:      http.MethodPut,
			path:        resourcePath,
			userID:      "https://example.com/profile/owner#me",
			wantAuth:    true,
			setupCookie: true,
		},
		{
			name:        "Owner control access to ACL",
			method:      http.MethodGet,
			path:        resourceACLPath,
			userID:      "https://example.com/profile/owner#me",
			wantAuth:    true,
			setupCookie: true,
		},
		{
			name:        "Public control access to ACL",
			method:      http.MethodGet,
			path:        resourceACLPath,
			userID:      "",
			wantAuth:    false,
			setupCookie: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)

			// Set up authentication if needed
			if tt.setupCookie {
				w := httptest.NewRecorder()
				_, err := authHandler.CreateSession(w, tt.userID)
				if err != nil {
					t.Fatalf("Failed to create session: %v", err)
				}
				cookie := w.Result().Cookies()[0]
				req.AddCookie(cookie)
			}

			authorized, err := httpHandler.AuthorizeRequest(req)
			if err != nil {
				if tt.wantErrMsg == "" {
					t.Fatalf("Unexpected error: %v", err)
				}
				if !strings.Contains(err.Error(), tt.wantErrMsg) {
					t.Errorf("Expected error containing %q, got %v", tt.wantErrMsg, err)
				}
				return
			}
			if tt.wantErrMsg != "" {
				t.Fatalf("Expected error containing %q, got nil", tt.wantErrMsg)
			}
			if authorized != tt.wantAuth {
				t.Errorf("AuthorizeRequest() = %v, want %v", authorized, tt.wantAuth)
			}
		})
	}
}

// Test the ACL middleware.
func TestACLMiddleware(t *testing.T) {
	storage := NewMockStorage()
	wacHandler, err := NewHandler(storage)
	if err != nil {
		t.Fatalf("Failed to create WAC handler: %v", err)
	}

	authHandler := NewMockAuthHandler()
	httpHandler := NewHTTPHandler(wacHandler, authHandler)

	// Add some ACLs to the storage
	resourcePath := "/resource"
	resourceACLPath := "/resource.acl"
	storage.AddResource(resourceACLPath, CreateSimpleACL(resourcePath, "https://example.com/profile/owner#me", Write, Control))

	// Create a test handler that will be wrapped by the middleware
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})
	middlewareHandler := httpHandler.ACLMiddleware(testHandler)

	// Create test requests
	tests := []struct {
		name         string
		method       string
		path         string
		userID       string
		wantStatus   int
		wantLocation string
		setupCookie  bool
	}{
		{
			name:        "Public read access to resource",
			method:      http.MethodGet,
			path:        resourcePath,
			userID:      "",
			wantStatus:  http.StatusOK,
			setupCookie: false,
		},
		{
			name:         "Public write access to resource",
			method:       http.MethodPut,
			path:         resourcePath,
			userID:       "",
			wantStatus:   http.StatusFound,
			wantLocation: "/login?redirect=%2Fresource",
			setupCookie:  false,
		},
		{
			name:        "Owner write access to resource",
			method:      http.MethodPut,
			path:        resourcePath,
			userID:      "https://example.com/profile/owner#me",
			wantStatus:  http.StatusOK,
			setupCookie: true,
		},
		{
			name:        "Other user write access to resource",
			method:      http.MethodPut,
			path:        resourcePath,
			userID:      "https://example.com/profile/other#me",
			wantStatus:  http.StatusForbidden,
			setupCookie: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			w := httptest.NewRecorder()

			// Set up authentication if needed
			if tt.setupCookie {
				_, err := authHandler.CreateSession(w, tt.userID)
				if err != nil {
					t.Fatalf("Failed to create session: %v", err)
				}
				cookie := w.Result().Cookies()[0]
				req.AddCookie(cookie)
			}

			middlewareHandler.ServeHTTP(w, req)

			resp := w.Result()
			if resp.StatusCode != tt.wantStatus {
				t.Errorf("Status code = %v, want %v", resp.StatusCode, tt.wantStatus)
			}

			if tt.wantLocation != "" {
				location := resp.Header.Get("Location")
				if location != tt.wantLocation {
					t.Errorf("Location = %v, want %v", location, tt.wantLocation)
				}
			}
		})
	}
}

// Test the authentication required middleware.
func TestAuthRequired(t *testing.T) {
	storage := NewMockStorage()
	wacHandler, err := NewHandler(storage)
	if err != nil {
		t.Fatalf("Failed to create WAC handler: %v", err)
	}

	authHandler := NewMockAuthHandler()
	httpHandler := NewHTTPHandler(wacHandler, authHandler)

	// Create a test handler that will be wrapped by the middleware
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})
	middlewareHandler := httpHandler.authRequired(testHandler)

	// Create test requests
	tests := []struct {
		name         string
		userID       string
		wantStatus   int
		wantLocation string
		setupCookie  bool
	}{
		{
			name:         "Unauthenticated user",
			userID:       "",
			wantStatus:   http.StatusFound,
			wantLocation: "/login?redirect=%2Fresource",
			setupCookie:  false,
		},
		{
			name:        "Authenticated user",
			userID:      "https://example.com/profile/owner#me",
			wantStatus:  http.StatusOK,
			setupCookie: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/resource", nil)
			w := httptest.NewRecorder()

			// Set up authentication if needed
			if tt.setupCookie {
				_, err := authHandler.CreateSession(w, tt.userID)
				if err != nil {
					t.Fatalf("Failed to create session: %v", err)
				}
				cookie := w.Result().Cookies()[0]
				req.AddCookie(cookie)
			}

			middlewareHandler.ServeHTTP(w, req)

			resp := w.Result()
			if resp.StatusCode != tt.wantStatus {
				t.Errorf("Status code = %v, want %v", resp.StatusCode, tt.wantStatus)
			}

			if tt.wantLocation != "" {
				location := resp.Header.Get("Location")
				if location != tt.wantLocation {
					t.Errorf("Location = %v, want %v", location, tt.wantLocation)
				}
			}
		})
	}
}

// Test handling ACL requests.
func TestHandleACLRequest(t *testing.T) {
	storage := NewMockStorage()
	wacHandler, err := NewHandler(storage)
	if err != nil {
		t.Fatalf("Failed to create WAC handler: %v", err)
	}

	authHandler := NewMockAuthHandler()
	httpHandler := NewHTTPHandler(wacHandler, authHandler)

	// Add some ACLs to the storage
	resourcePath := "/resource"
	resourceACLPath := "/resource.acl"
	storage.AddResource(resourceACLPath, CreateSimpleACL(resourcePath, "https://example.com/profile/owner#me", Write, Control))

	// Create test requests
	tests := []struct {
		name        string
		method      string
		path        string
		userID      string
		wantStatus  int
		setupCookie bool
		requestBody string
	}{
		{
			name:        "Get ACL as owner",
			method:      http.MethodGet,
			path:        "/acl/resource",
			userID:      "https://example.com/profile/owner#me",
			wantStatus:  http.StatusOK,
			setupCookie: true,
		},
		{
			name:        "Get ACL as public",
			method:      http.MethodGet,
			path:        "/acl/resource",
			userID:      "",
			wantStatus:  http.StatusForbidden,
			setupCookie: false,
		},
		{
			name:        "Put ACL as owner",
			method:      http.MethodPut,
			path:        "/acl/resource",
			userID:      "https://example.com/profile/owner#me",
			wantStatus:  http.StatusCreated,
			setupCookie: true,
			requestBody: `
@prefix acl: <http://www.w3.org/ns/auth/acl#> .
@prefix foaf: <http://xmlns.com/foaf/0.1/> .

</resource#public> a acl:Authorization ;
    acl:mode acl:Read ;
    acl:accessTo </resource> ;
    acl:agentClass foaf:Agent .

</resource#owner> a acl:Authorization ;
    acl:mode acl:Write, acl:Control ;
    acl:accessTo </resource> ;
    acl:agent <https://example.com/profile/owner#me> .
`,
		},
		{
			name:        "Delete ACL as owner",
			method:      http.MethodDelete,
			path:        "/acl/resource",
			userID:      "https://example.com/profile/owner#me",
			wantStatus:  http.StatusNoContent,
			setupCookie: true,
		},
		{
			name:        "Unsupported method",
			method:      http.MethodPatch,
			path:        "/acl/resource",
			userID:      "https://example.com/profile/owner#me",
			wantStatus:  http.StatusMethodNotAllowed,
			setupCookie: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var reqBody io.Reader
			if tt.requestBody != "" {
				reqBody = strings.NewReader(tt.requestBody)
			}
			req := httptest.NewRequest(tt.method, tt.path, reqBody)
			w := httptest.NewRecorder()

			// Set up authentication if needed
			if tt.setupCookie {
				_, err := authHandler.CreateSession(w, tt.userID)
				if err != nil {
					t.Fatalf("Failed to create session: %v", err)
				}
				cookie := w.Result().Cookies()[0]
				req.AddCookie(cookie)
			}

			if tt.requestBody != "" {
				req.Header.Set("Content-Type", "text/turtle")
			}

			httpHandler.handleACLRequest(w, req)

			resp := w.Result()
			if resp.StatusCode != tt.wantStatus {
				body, _ := io.ReadAll(resp.Body)
				t.Errorf("Status code = %v, want %v, body: %s", resp.StatusCode, tt.wantStatus, body)
			}

			if tt.method == http.MethodGet && resp.StatusCode == http.StatusOK {
				contentType := resp.Header.Get("Content-Type")
				if contentType != "text/turtle" {
					t.Errorf("Content-Type = %v, want text/turtle", contentType)
				}

				body, err := io.ReadAll(resp.Body)
				if err != nil {
					t.Fatalf("Failed to read response body: %v", err)
				}
				if !bytes.Contains(body, []byte("@prefix acl:")) {
					t.Errorf("Response body does not contain ACL prefixes")
				}
			}
		})
	}
}

// Test getting the agent from the request.
func TestGetAgentFromRequest(t *testing.T) {
	storage := NewMockStorage()
	wacHandler, err := NewHandler(storage)
	if err != nil {
		t.Fatalf("Failed to create WAC handler: %v", err)
	}

	authHandler := NewMockAuthHandler()
	httpHandler := NewHTTPHandler(wacHandler, authHandler)

	// Create test requests
	tests := []struct {
		name        string
		userID      string
		wantType    AgentType
		wantWebID   string
		setupCookie bool
	}{
		{
			name:        "Unauthenticated user",
			userID:      "",
			wantType:    Public,
			wantWebID:   "",
			setupCookie: false,
		},
		{
			name:        "Authenticated user",
			userID:      "https://example.com/profile/owner#me",
			wantType:    User,
			wantWebID:   "https://example.com/profile/owner#me",
			setupCookie: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/resource", nil)
			w := httptest.NewRecorder()

			// Set up authentication if needed
			if tt.setupCookie {
				_, err := authHandler.CreateSession(w, tt.userID)
				if err != nil {
					t.Fatalf("Failed to create session: %v", err)
				}
				cookie := w.Result().Cookies()[0]
				req.AddCookie(cookie)
			}

			agent, err := httpHandler.getAgentFromRequest(req)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if agent.Type != tt.wantType {
				t.Errorf("Agent type = %v, want %v", agent.Type, tt.wantType)
			}
			if agent.WebID != tt.wantWebID {
				t.Errorf("Agent WebID = %v, want %v", agent.WebID, tt.wantWebID)
			}
		})
	}
}

// Test validating an ACL.
func TestValidateACL(t *testing.T) {
	tests := []struct {
		name      string
		acl       *ACL
		wantError bool
	}{
		{
			name: "Valid ACL",
			acl: &ACL{
				Resource: "/resource",
				Authorizations: []Authorization{
					{
						ID:        "/resource#public",
						Modes:     Read,
						Agents:    []Agent{{Type: Public}},
						Resources: []string{"/resource"},
					},
				},
			},
			wantError: false,
		},
		{
			name: "No authorizations",
			acl: &ACL{
				Resource:       "/resource",
				Authorizations: []Authorization{},
			},
			wantError: true,
		},
		{
			name: "No agents",
			acl: &ACL{
				Resource: "/resource",
				Authorizations: []Authorization{
					{
						ID:        "/resource#public",
						Modes:     Read,
						Agents:    []Agent{},
						Resources: []string{"/resource"},
					},
				},
			},
			wantError: true,
		},
		{
			name: "No resources",
			acl: &ACL{
				Resource: "/resource",
				Authorizations: []Authorization{
					{
						ID:        "/resource#public",
						Modes:     Read,
						Agents:    []Agent{{Type: Public}},
						Resources: []string{},
					},
				},
			},
			wantError: true,
		},
		{
			name: "No modes",
			acl: &ACL{
				Resource: "/resource",
				Authorizations: []Authorization{
					{
						ID:        "/resource#public",
						Modes:     0,
						Agents:    []Agent{{Type: Public}},
						Resources: []string{"/resource"},
					},
				},
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateACL(tt.acl)
			if (err != nil) != tt.wantError {
				t.Errorf("validateACL() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

// Test the GetACLAsJSON method.
func TestGetACLAsJSON(t *testing.T) {
	storage := NewMockStorage()
	wacHandler, err := NewHandler(storage)
	if err != nil {
		t.Fatalf("Failed to create WAC handler: %v", err)
	}

	authHandler := NewMockAuthHandler()
	httpHandler := NewHTTPHandler(wacHandler, authHandler)

	// Add some ACLs to the storage
	resourcePath := "/resource"
	resourceACLPath := "/resource.acl"
	storage.AddResource(resourceACLPath, CreateSimpleACL(resourcePath, "https://example.com/profile/owner#me", Write, Control))

	// Create test request
	req := httptest.NewRequest(http.MethodGet, resourcePath, nil)
	w := httptest.NewRecorder()

	// Call the method
	httpHandler.GetACLAsJSON(w, req)

	// Check the response
	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Status code = %v, want %v", resp.StatusCode, http.StatusOK)
	}

	contentType := resp.Header.Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("Content-Type = %v, want application/json", contentType)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}

	// Check that the JSON is well-formed and contains expected fields
	if !bytes.Contains(body, []byte(`"resource":"/resource"`)) {
		t.Errorf("Response body does not contain resource path")
	}
	if !bytes.Contains(body, []byte(`"modes":["read"]`)) {
		t.Errorf("Response body does not contain read mode")
	}
	if !bytes.Contains(body, []byte(`"type":"Public"`)) {
		t.Errorf("Response body does not contain Public agent type")
	}
	if !bytes.Contains(body, []byte(`"webId":"https://example.com/profile/owner#me"`)) {
		t.Errorf("Response body does not contain owner WebID")
	}
}

// Test the WAC HTTP handler.
func TestHTTPHandler(t *testing.T) {
	// Set up WAC components
	storage := NewMockStorage()
	wacHandler, err := NewHandler(storage)
	if err != nil {
		t.Fatalf("Failed to create WAC handler: %v", err)
	}

	authHandler := NewMockAuthHandler()
	wacHTTPHandler := NewHTTPHandler(wacHandler, authHandler)

	// Add a public ACL for testing
	resourcePath := "/resource"
	resourceACLPath := "/resource.acl"
	containerPath := "/containers/test"
	containerACLPath := "/containers/test.acl"
	profilePath := "/profiles/user"
	profileACLPath := "/profiles/user.acl"

	// Public read access for all
	storage.AddResource(resourceACLPath, CreateSimpleACL(resourcePath, "http://xmlns.com/foaf/0.1/Agent", Read))
	storage.AddResource(containerACLPath, CreateSimpleACL(containerPath, "http://xmlns.com/foaf/0.1/Agent", Read))
	storage.AddResource(profileACLPath, CreateSimpleACL(profilePath, "http://xmlns.com/foaf/0.1/Agent", Read))

	// Create a test request
	req := httptest.NewRequest(http.MethodGet, "/acl/resource", nil)
	req.Header.Set("Authorization", "Bearer test-token")

	// Create a test response writer
	rw := httptest.NewRecorder()

	// Test the ACL request handler
	wacHTTPHandler.handleACLRequest(rw, req)

	// Check the response
	if rw.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, rw.Code)
	}
}
