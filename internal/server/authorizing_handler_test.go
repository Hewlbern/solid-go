package server

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

type mockHttpHandler struct {
	handleCalled bool
	handleErr    error
}

func (m *mockHttpHandler) Handle(w http.ResponseWriter, r *http.Request) error {
	m.handleCalled = true
	return m.handleErr
}

type mockAuthorizer struct {
	authorizeCalled bool
	authorizeErr    error
}

func (m *mockAuthorizer) Authorize(ctx context.Context, op Operation) error {
	m.authorizeCalled = true
	return m.authorizeErr
}

func TestBaseAuthorizingHandler(t *testing.T) {
	tests := []struct {
		name           string
		authorizeErr   error
		handleErr      error
		expectedStatus int
		expectHandle   bool
	}{
		{
			name:           "Authorized Request",
			authorizeErr:   nil,
			handleErr:      nil,
			expectedStatus: http.StatusOK,
			expectHandle:   true,
		},
		{
			name:           "Unauthorized Request",
			authorizeErr:   http.ErrAbortHandler,
			handleErr:      nil,
			expectedStatus: http.StatusUnauthorized,
			expectHandle:   false,
		},
		{
			name:           "Handler Error",
			authorizeErr:   nil,
			handleErr:      http.ErrAbortHandler,
			expectedStatus: http.StatusInternalServerError,
			expectHandle:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock handler and authorizer
			mockHandler := &mockHttpHandler{handleErr: tt.handleErr}
			mockAuthorizer := &mockAuthorizer{authorizeErr: tt.authorizeErr}

			// Create authorizing handler
			handler := NewBaseAuthorizingHandler(mockHandler, mockAuthorizer)

			// Create test request
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			w := httptest.NewRecorder()

			// Handle request
			err := handler.Handle(w, req)

			// Check response status
			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			// Check if handler was called
			if mockHandler.handleCalled != tt.expectHandle {
				t.Errorf("expected handle called to be %v, got %v", tt.expectHandle, mockHandler.handleCalled)
			}

			// Check if authorizer was called
			if !mockAuthorizer.authorizeCalled {
				t.Error("expected authorizer to be called")
			}

			// Check error
			if tt.handleErr != nil && err != tt.handleErr {
				t.Errorf("expected error %v, got %v", tt.handleErr, err)
			}
		})
	}
}

func TestAuthorize(t *testing.T) {
	tests := []struct {
		name         string
		authorizeErr error
		expectErr    error
	}{
		{
			name:         "Authorized Request",
			authorizeErr: nil,
			expectErr:    nil,
		},
		{
			name:         "Unauthorized Request",
			authorizeErr: http.ErrAbortHandler,
			expectErr:    http.ErrAbortHandler,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock handler and authorizer
			mockHandler := &mockHttpHandler{}
			mockAuthorizer := &mockAuthorizer{authorizeErr: tt.authorizeErr}

			// Create authorizing handler
			handler := NewBaseAuthorizingHandler(mockHandler, mockAuthorizer)

			// Create test request
			req := httptest.NewRequest(http.MethodGet, "/test", nil)

			// Authorize request
			err := handler.Authorize(req)

			// Check error
			if err != tt.expectErr {
				t.Errorf("expected error %v, got %v", tt.expectErr, err)
			}

			// Check if authorizer was called
			if !mockAuthorizer.authorizeCalled {
				t.Error("expected authorizer to be called")
			}
		})
	}
}
