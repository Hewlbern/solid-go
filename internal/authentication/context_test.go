package auth

import (
	"context"
	"testing"
)

func TestAgentContext(t *testing.T) {
	// Create some test agents
	agent1 := &Agent{
		ID:              "https://example.org/alice#me",
		Name:            "Alice",
		IsAuthenticated: true,
		Type:            TypeUser,
	}
	
	agent2 := &Agent{
		ID:              "https://example.org/bob#me",
		Name:            "Bob",
		IsAuthenticated: false,
		Type:            TypeUser,
	}

	// Create a base context
	baseCtx := context.Background()

	// Test cases
	tests := []struct {
		name                  string
		agent                 *Agent
		expectAuthenticated   bool
		expectRequireAuthErr  bool
	}{
		{
			name:                  "Authenticated user",
			agent:                 agent1,
			expectAuthenticated:   true,
			expectRequireAuthErr:  false,
		},
		{
			name:                  "Unauthenticated user",
			agent:                 agent2,
			expectAuthenticated:   false,
			expectRequireAuthErr:  true,
		},
		{
			name:                  "No agent",
			agent:                 nil,
			expectAuthenticated:   false,
			expectRequireAuthErr:  true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var ctx context.Context
			
			if tc.agent != nil {
				// Create a context with the agent
				ctx = NewAgentContext(baseCtx, tc.agent)
				
				// Test AgentFromContext
				retrievedAgent, ok := AgentFromContext(ctx)
				if !ok {
					t.Error("Failed to retrieve agent from context")
				}
				if retrievedAgent != tc.agent {
					t.Errorf("Retrieved agent = %v, want %v", retrievedAgent, tc.agent)
				}
			} else {
				ctx = baseCtx
				
				// Test AgentFromContext with no agent
				_, ok := AgentFromContext(ctx)
				if ok {
					t.Error("AgentFromContext reported agent found when none was set")
				}
			}
			
			// Test IsAuthenticated
			if got := IsAuthenticated(ctx); got != tc.expectAuthenticated {
				t.Errorf("IsAuthenticated() = %v, want %v", got, tc.expectAuthenticated)
			}
			
			// Test RequireAuthentication
			agent, err := RequireAuthentication(ctx)
			if tc.expectRequireAuthErr {
				if err == nil {
					t.Error("RequireAuthentication() did not return error when expected")
				}
				if err != ErrAuthenticationRequired {
					t.Errorf("RequireAuthentication() returned %v, want %v", err, ErrAuthenticationRequired)
				}
			} else {
				if err != nil {
					t.Errorf("RequireAuthentication() returned unexpected error: %v", err)
				}
				if agent != tc.agent {
					t.Errorf("RequireAuthentication() returned %v, want %v", agent, tc.agent)
				}
			}
			
			// Test GetAgent
			retrievedAgent := GetAgent(ctx)
			if tc.agent != nil {
				if retrievedAgent != tc.agent {
					t.Errorf("GetAgent() = %v, want %v", retrievedAgent, tc.agent)
				}
			} else {
				if retrievedAgent.IsAuthenticated {
					t.Error("GetAgent() returned authenticated agent when none was in context")
				}
				if retrievedAgent.Type != TypeAnonymous {
					t.Errorf("GetAgent() returned agent type = %v, want %v", retrievedAgent.Type, TypeAnonymous)
				}
			}
		})
	}
} 