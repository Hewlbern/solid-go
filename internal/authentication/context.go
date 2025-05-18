package auth

import (
	"context"
)

// contextKey is a private type for context keys to avoid collisions with other packages
// that might also be using context.WithValue.
type contextKey int

// agentContextKey is the key for agent values in Contexts. It is used to store and
// retrieve agent information from the context.
const agentContextKey contextKey = iota

// NewAgentContext creates a new Context with the given agent value. The agent will be
// available to all functions that receive this context. This is typically used at the
// beginning of a request to store the authenticated agent.
//
// Example:
//
//	agent := &Agent{ID: "https://example.org/alice#me", IsAuthenticated: true}
//	ctx := NewAgentContext(context.Background(), agent)
func NewAgentContext(ctx context.Context, agent *Agent) context.Context {
	return context.WithValue(ctx, agentContextKey, agent)
}

// AgentFromContext extracts the agent from the given Context. It returns the agent and
// a boolean indicating whether an agent was found in the context. If no agent is found,
// it returns (nil, false).
//
// Example:
//
//	agent, ok := AgentFromContext(ctx)
//	if !ok {
//	    // No agent in context
//	}
func AgentFromContext(ctx context.Context) (*Agent, bool) {
	agent, ok := ctx.Value(agentContextKey).(*Agent)
	return agent, ok
}

// GetAgent extracts the agent from the context and returns an anonymous agent if none is found.
// This is useful when you always want to have an agent, even if it's just an anonymous one.
// The anonymous agent will have IsAuthenticated set to false and Type set to TypeAnonymous.
//
// Example:
//
//	agent := GetAgent(ctx)
//	if agent.IsAuthenticated {
//	    // Handle authenticated agent
//	}
func GetAgent(ctx context.Context) *Agent {
	agent, ok := AgentFromContext(ctx)
	if !ok {
		return NewAnonymousAgent()
	}
	return agent
}

// IsAuthenticated checks if there is an authenticated agent in the context. It returns
// true if there is an agent in the context and that agent is authenticated, false otherwise.
//
// Example:
//
//	if IsAuthenticated(ctx) {
//	    // Handle authenticated request
//	}
func IsAuthenticated(ctx context.Context) bool {
	agent, ok := AgentFromContext(ctx)
	return ok && agent.IsAuthenticated
}

// RequireAuthentication checks if there is an authenticated agent and returns an error if not.
// It returns the authenticated agent if one is found, or ErrAuthenticationRequired if no
// authenticated agent is found. This is useful for protecting routes that require authentication.
//
// Example:
//
//	agent, err := RequireAuthentication(ctx)
//	if err != nil {
//	    // Handle authentication error
//	}
func RequireAuthentication(ctx context.Context) (*Agent, error) {
	agent, ok := AgentFromContext(ctx)
	if !ok || !agent.IsAuthenticated {
		return nil, ErrAuthenticationRequired
	}
	return agent, nil
}
