package wac

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strings"

	"github.com/yourusername/solid-go/internal/auth"
	"github.com/yourusername/solid-go/internal/rdf"
)

// HTTPHandler handles HTTP requests for WAC operations.
type HTTPHandler struct {
	// wacHandler is the WAC handler.
	wacHandler *Handler
	// authHandler is the authentication handler.
	authHandler auth.AuthHandler
}

// NewHTTPHandler creates a new HTTP handler for WAC operations.
func NewHTTPHandler(wacHandler *Handler, authHandler auth.AuthHandler) *HTTPHandler {
	return &HTTPHandler{
		wacHandler:  wacHandler,
		authHandler: authHandler,
	}
}

// RegisterRoutes registers the WAC routes with the given mux.
func (h *HTTPHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.Handle("/acl/", h.authRequired(h.handleACLRequest))
}

// ResourceRequest contains information about a request for a resource.
type ResourceRequest struct {
	// Method is the HTTP method.
	Method string
	// Path is the path to the resource.
	Path string
	// WebID is the WebID of the agent making the request.
	WebID string
}

// RequiredMode determines the access mode required for a request.
func (r *ResourceRequest) RequiredMode() AccessMode {
	switch r.Method {
	case http.MethodGet, http.MethodHead, http.MethodOptions:
		return Read
	case http.MethodPut, http.MethodDelete:
		return Write
	case http.MethodPost:
		return Append
	case http.MethodPatch:
		return Write
	default:
		return Read
	}
}

// AuthorizeRequest checks if a request is authorized.
func (h *HTTPHandler) AuthorizeRequest(req *http.Request) (bool, error) {
	// Extract the resource path from the request
	resourcePath := req.URL.Path

	// Check if this is an ACL resource
	if strings.HasSuffix(resourcePath, ".acl") || path.Base(resourcePath) == ".acl" {
		// ACL resources require Control access to the associated resource
		resourcePath = strings.TrimSuffix(resourcePath, ".acl")
		if resourcePath == "" {
			resourcePath = "/"
		}

		// Get the agent from the request
		agent, err := h.getAgentFromRequest(req)
		if err != nil {
			return false, err
		}

		// Check if the agent has Control access
		return h.wacHandler.CheckAccess(resourcePath, agent, Control)
	}

	// Get the agent from the request
	agent, err := h.getAgentFromRequest(req)
	if err != nil {
		return false, err
	}

	// Determine the required access mode
	requiredMode := (&ResourceRequest{
		Method: req.Method,
		Path:   resourcePath,
		WebID:  agent.WebID,
	}).RequiredMode()

	// Check if the agent has the required access
	return h.wacHandler.CheckAccess(resourcePath, agent, requiredMode)
}

// getAgentFromRequest extracts the agent from a request.
func (h *HTTPHandler) getAgentFromRequest(req *http.Request) (Agent, error) {
	// Get the agent from the request context
	agent := auth.GetAgent(req.Context())
	if agent.IsAuthenticated {
		// This is an authenticated user
		return Agent{
			WebID: agent.ID,
			Type:  User,
		}, nil
	}

	// This is a public agent
	return Agent{
		Type: Public,
	}, nil
}

// handleACLRequest handles requests for ACL resources.
func (h *HTTPHandler) handleACLRequest(w http.ResponseWriter, req *http.Request) {
	// The path to the ACL resource
	aclPath := req.URL.Path
	// The path to the associated resource
	resourcePath := strings.TrimPrefix(aclPath, "/acl")

	// Get the agent from the request
	agent, err := h.getAgentFromRequest(req)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get agent: %v", err), http.StatusInternalServerError)
		return
	}

	// Check if the agent has Control access to the resource
	hasControl, err := h.wacHandler.CheckAccess(resourcePath, agent, Control)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to check access: %v", err), http.StatusInternalServerError)
		return
	}
	if !hasControl {
		http.Error(w, "Forbidden: You do not have Control access to this resource", http.StatusForbidden)
		return
	}

	// Handle the request based on the method
	switch req.Method {
	case http.MethodGet:
		// Get the ACL resource
		acl, err := h.wacHandler.findACL(resourcePath)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to find ACL: %v", err), http.StatusInternalServerError)
			return
		}

		// Set the content type
		w.Header().Set("Content-Type", "text/turtle")

		// Serialize the ACL to Turtle
		if err := SerializeACL(acl, w); err != nil {
			http.Error(w, fmt.Sprintf("Failed to serialize ACL: %v", err), http.StatusInternalServerError)
			return
		}

	case http.MethodPut:
		// Read the request body
		body, err := io.ReadAll(req.Body)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to read request body: %v", err), http.StatusBadRequest)
			return
		}

		// Parse the ACL
		contentType := req.Header.Get("Content-Type")
		if contentType == "" {
			contentType = "text/turtle"
		}

		// Use the correct parser based on the content type
		var graph *rdf.Graph
		var parseErr error
		if strings.Contains(contentType, "turtle") || strings.Contains(contentType, "text/n3") {
			parser := rdf.NewTurtleParser()
			graph, parseErr = parser.Parse(strings.NewReader(string(body)))
		} else if strings.Contains(contentType, "json") {
			// In a real implementation, this would parse JSON-LD
			parseErr = fmt.Errorf("JSON-LD parsing not implemented")
		} else {
			parseErr = fmt.Errorf("unsupported content type: %s", contentType)
		}

		if parseErr != nil {
			http.Error(w, fmt.Sprintf("Failed to parse ACL: %v", parseErr), http.StatusBadRequest)
			return
		}

		// Convert the graph to an ACL
		acl, err := graphToACL(graph, resourcePath)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to convert graph to ACL: %v", err), http.StatusBadRequest)
			return
		}

		// Validate the ACL
		if err := validateACL(acl); err != nil {
			http.Error(w, fmt.Sprintf("Invalid ACL: %v", err), http.StatusBadRequest)
			return
		}

		// In a real implementation, we would store the ACL
		// For now, just return a success response
		w.WriteHeader(http.StatusCreated)

	case http.MethodDelete:
		// In a real implementation, we would delete the ACL
		// For now, just return a success response
		w.WriteHeader(http.StatusNoContent)

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// authRequired is a middleware that requires authentication for a request.
func (h *HTTPHandler) authRequired(handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		// Get the agent from the request
		agent, err := h.getAgentFromRequest(req)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to get agent: %v", err), http.StatusInternalServerError)
			return
		}

		// Check if the agent is authenticated
		if agent.Type != User {
			// Redirect to the login page
			redirectURL := "/login?redirect=" + url.QueryEscape(req.URL.String())
			http.Redirect(w, req, redirectURL, http.StatusFound)
			return
		}

		// Call the handler
		handler(w, req)
	}
}

// authOptional is a middleware that allows both authenticated and unauthenticated access.
func (h *HTTPHandler) authOptional(handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		// Call the handler directly
		handler(w, req)
	}
}

// ACLMiddleware is a middleware that checks if a request is authorized.
func (h *HTTPHandler) ACLMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		// Check if the request is authorized
		authorized, err := h.AuthorizeRequest(req)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to authorize request: %v", err), http.StatusInternalServerError)
			return
		}
		if !authorized {
			// Get the agent from the request
			agent, err := h.getAgentFromRequest(req)
			if err != nil {
				http.Error(w, fmt.Sprintf("Failed to get agent: %v", err), http.StatusInternalServerError)
				return
			}

			// If the agent is not authenticated, redirect to the login page
			if agent.Type != User {
				redirectURL := "/login?redirect=" + url.QueryEscape(req.URL.String())
				http.Redirect(w, req, redirectURL, http.StatusFound)
				return
			}

			// Otherwise, return a forbidden error
			http.Error(w, "Forbidden: You do not have the required access to this resource", http.StatusForbidden)
			return
		}

		// Call the next handler
		next.ServeHTTP(w, req)
	})
}

// graphToACL converts an RDF graph to an ACL.
func graphToACL(graph *rdf.Graph, resourcePath string) (*ACL, error) {
	acl := &ACL{
		Resource: resourcePath,
	}

	// Find authorization subjects
	for _, triple := range graph.Triples {
		if triple.Predicate.Type() != rdf.IRI {
			continue
		}
		predIRI := triple.Predicate.(rdf.IRIRef).IRI
		if predIRI != "http://www.w3.org/1999/02/22-rdf-syntax-ns#type" {
			continue
		}
		if triple.Object.Type() != rdf.IRI {
			continue
		}
		objIRI := triple.Object.(rdf.IRIRef).IRI
		if objIRI != "http://www.w3.org/ns/auth/acl#Authorization" {
			continue
		}
		if triple.Subject.Type() != rdf.IRI {
			continue
		}
		authURI := triple.Subject.(rdf.IRIRef).IRI

		// Parse the authorization
		auth := Authorization{
			ID: authURI,
		}

		// Find the modes
		modeTriples := findTriples(graph, authURI, "http://www.w3.org/ns/auth/acl#mode", "")
		for _, modeTriple := range modeTriples {
			if modeTriple.Object.Type() != rdf.IRI {
				continue
			}
			modeIRI := modeTriple.Object.(rdf.IRIRef).IRI
			switch modeIRI {
			case "http://www.w3.org/ns/auth/acl#Read":
				auth.Modes |= Read
			case "http://www.w3.org/ns/auth/acl#Write":
				auth.Modes |= Write
			case "http://www.w3.org/ns/auth/acl#Append":
				auth.Modes |= Append
			case "http://www.w3.org/ns/auth/acl#Control":
				auth.Modes |= Control
			}
		}

		// Find the resources
		resourceTriples := findTriples(graph, authURI, "http://www.w3.org/ns/auth/acl#accessTo", "")
		for _, resourceTriple := range resourceTriples {
			if resourceTriple.Object.Type() != rdf.IRI {
				continue
			}
			resourceIRI := resourceTriple.Object.(rdf.IRIRef).IRI
			auth.Resources = append(auth.Resources, resourceIRI)
		}

		// Find default resources
		defaultTriples := findTriples(graph, authURI, "http://www.w3.org/ns/auth/acl#default", "")
		for _, defaultTriple := range defaultTriples {
			if defaultTriple.Object.Type() != rdf.IRI {
				continue
			}
			defaultIRI := defaultTriple.Object.(rdf.IRIRef).IRI
			auth.Resources = append(auth.Resources, defaultIRI)
			acl.IsDefault = true
		}

		// Find the agents
		agentTriples := findTriples(graph, authURI, "http://www.w3.org/ns/auth/acl#agent", "")
		for _, agentTriple := range agentTriples {
			if agentTriple.Object.Type() != rdf.IRI {
				continue
			}
			agentIRI := agentTriple.Object.(rdf.IRIRef).IRI
			auth.Agents = append(auth.Agents, Agent{
				WebID: agentIRI,
				Type:  User,
			})
		}

		// Find the agent classes
		agentClassTriples := findTriples(graph, authURI, "http://www.w3.org/ns/auth/acl#agentClass", "")
		for _, agentClassTriple := range agentClassTriples {
			if agentClassTriple.Object.Type() != rdf.IRI {
				continue
			}
			agentClassIRI := agentClassTriple.Object.(rdf.IRIRef).IRI
			switch agentClassIRI {
			case "http://xmlns.com/foaf/0.1/Agent":
				auth.Agents = append(auth.Agents, Agent{
					Type: Public,
				})
			case "http://www.w3.org/ns/auth/acl#AuthenticatedAgent":
				auth.Agents = append(auth.Agents, Agent{
					Type: Authenticated,
				})
			default:
				// Assume it's a group
				auth.Agents = append(auth.Agents, Agent{
					WebID: agentClassIRI,
					Type:  Group,
				})
			}
		}

		// Add the authorization to the ACL
		acl.Authorizations = append(acl.Authorizations, auth)
	}

	return acl, nil
}

// validateACL validates an ACL.
func validateACL(acl *ACL) error {
	// Check that the ACL has at least one authorization
	if len(acl.Authorizations) == 0 {
		return fmt.Errorf("ACL must have at least one authorization")
	}

	// Check that each authorization has at least one agent
	for i, auth := range acl.Authorizations {
		if len(auth.Agents) == 0 {
			return fmt.Errorf("authorization %d has no agents", i+1)
		}
		if len(auth.Resources) == 0 {
			return fmt.Errorf("authorization %d has no resources", i+1)
		}
		if auth.Modes == 0 {
			return fmt.Errorf("authorization %d has no modes", i+1)
		}
	}

	return nil
}

// GetEffectiveACL gets the effective ACL for a resource.
func (h *HTTPHandler) GetEffectiveACL(resourcePath string) (*ACL, error) {
	return h.wacHandler.findACL(resourcePath)
}

// GetACLAsJSON returns the ACL as JSON.
func (h *HTTPHandler) GetACLAsJSON(w http.ResponseWriter, req *http.Request) {
	// Get the ACL
	acl, err := h.GetEffectiveACL(req.URL.Path)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get ACL: %v", err), http.StatusInternalServerError)
		return
	}

	// Convert to JSON
	jsonACL := struct {
		Resource       string `json:"resource"`
		IsDefault      bool   `json:"isDefault"`
		Authorizations []struct {
			ID     string   `json:"id"`
			Modes  []string `json:"modes"`
			Agents []struct {
				WebID string `json:"webId,omitempty"`
				Type  string `json:"type"`
			} `json:"agents"`
			Resources []string `json:"resources"`
		} `json:"authorizations"`
	}{
		Resource:  acl.Resource,
		IsDefault: acl.IsDefault,
	}

	// Convert authorizations
	for _, auth := range acl.Authorizations {
		jsonAuth := struct {
			ID     string   `json:"id"`
			Modes  []string `json:"modes"`
			Agents []struct {
				WebID string `json:"webId,omitempty"`
				Type  string `json:"type"`
			} `json:"agents"`
			Resources []string `json:"resources"`
		}{
			ID:        auth.ID,
			Resources: auth.Resources,
		}

		// Convert modes
		if auth.Modes&Read != 0 {
			jsonAuth.Modes = append(jsonAuth.Modes, "read")
		}
		if auth.Modes&Write != 0 {
			jsonAuth.Modes = append(jsonAuth.Modes, "write")
		}
		if auth.Modes&Append != 0 {
			jsonAuth.Modes = append(jsonAuth.Modes, "append")
		}
		if auth.Modes&Control != 0 {
			jsonAuth.Modes = append(jsonAuth.Modes, "control")
		}

		// Convert agents
		for _, agent := range auth.Agents {
			jsonAuth.Agents = append(jsonAuth.Agents, struct {
				WebID string `json:"webId,omitempty"`
				Type  string `json:"type"`
			}{
				WebID: agent.WebID,
				Type:  string(agent.Type),
			})
		}

		jsonACL.Authorizations = append(jsonACL.Authorizations, jsonAuth)
	}

	// Write JSON response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(jsonACL)
}
