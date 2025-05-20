package server

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestWacAllowHttpHandler(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		authorizeErr   error
		handleErr      error
		expectedStatus int
		expectHandle   bool
	}{
		{
			name:           "Authorized Request",
			method:         http.MethodGet,
			authorizeErr:   nil,
			handleErr:      nil,
			expectedStatus: http.StatusOK,
			expectHandle:   true,
		},
		{
			name:           "Unauthorized Request",
			method:         http.MethodGet,
			authorizeErr:   http.ErrAbortHandler,
			handleErr:      nil,
			expectedStatus: http.StatusUnauthorized,
			expectHandle:   false,
		},
		{
			name:           "Unauthorized HEAD Request",
			method:         http.MethodHead,
			authorizeErr:   http.ErrAbortHandler,
			handleErr:      nil,
			expectedStatus: http.StatusOK,
			expectHandle:   false,
		},
		{
			name:           "Handler Error",
			method:         http.MethodGet,
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

			// Create WAC allow handler
			handler := NewWacAllowHttpHandler(mockHandler, mockAuthorizer)

			// Create test request
			req := httptest.NewRequest(tt.method, "/test", nil)
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
