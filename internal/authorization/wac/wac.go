package wac

import (
	"errors"
	"fmt"
	"io"
	"path"
	"strings"

	"github.com/yourusername/solid-go/internal/rdf"
)

// AccessMode represents the access modes for a resource.
type AccessMode int

const (
	// Read represents read access.
	Read AccessMode = 1 << iota
	// Write represents write access.
	Write
	// Append represents append access.
	Append
	// Control represents control access.
	Control
)

// String returns a string representation of the access mode.
func (mode AccessMode) String() string {
	var modes []string
	if mode&Read != 0 {
		modes = append(modes, "Read")
	}
	if mode&Write != 0 {
		modes = append(modes, "Write")
	}
	if mode&Append != 0 {
		modes = append(modes, "Append")
	}
	if mode&Control != 0 {
		modes = append(modes, "Control")
	}
	return strings.Join(modes, ", ")
}

// Contains checks if the mode contains all the modes in the other mode.
func (mode AccessMode) Contains(other AccessMode) bool {
	return mode&other == other
}

// Agent represents an agent (user or group) in the WAC system.
type Agent struct {
	// WebID is the WebID of the agent.
	WebID string
	// Type is the type of the agent.
	Type AgentType
}

// AgentType represents the type of an agent.
type AgentType int

const (
	// User represents a user agent.
	User AgentType = iota
	// Group represents a group agent.
	Group
	// Public represents the public (everyone).
	Public
	// Authenticated represents authenticated users.
	Authenticated
)

// String returns a string representation of the agent type.
func (t AgentType) String() string {
	switch t {
	case User:
		return "User"
	case Group:
		return "Group"
	case Public:
		return "Public"
	case Authenticated:
		return "Authenticated"
	default:
		return "Unknown"
	}
}

// URIForType returns the WAC URI for an agent type.
func URIForType(t AgentType) string {
	switch t {
	case Public:
		return "http://xmlns.com/foaf/0.1/Agent"
	case Authenticated:
		return "http://www.w3.org/ns/auth/acl#AuthenticatedAgent"
	default:
		return ""
	}
}

// ACL represents an Access Control List for a resource.
type ACL struct {
	// Resource is the resource that this ACL applies to.
	Resource string
	// IsDefault indicates if this ACL is the default ACL for a container.
	IsDefault bool
	// Authorizations are the authorizations in this ACL.
	Authorizations []Authorization
}

// Authorization represents an authorization in an ACL.
type Authorization struct {
	// ID is the identifier for this authorization.
	ID string
	// Agents are the agents this authorization applies to.
	Agents []Agent
	// Modes are the access modes granted by this authorization.
	Modes AccessMode
	// Resources are the resources this authorization applies to.
	Resources []string
}

// AccessChecker checks access to resources.
type AccessChecker interface {
	// CheckAccess checks if the agent has the specified access to the resource.
	CheckAccess(resourcePath string, agent Agent, mode AccessMode) (bool, error)
	// GetACL gets the ACL for a resource.
	GetACL(resourcePath string) (*ACL, error)
}

// StorageAccessor provides access to resources in storage.
type StorageAccessor interface {
	// GetResource gets a resource from storage.
	GetResource(path string) ([]byte, string, error)
	// ResourceExists checks if a resource exists.
	ResourceExists(path string) (bool, error)
}

// Handler handles WAC operations.
type Handler struct {
	// storage is the storage accessor.
	storage StorageAccessor
}

// NewHandler creates a new WAC handler.
func NewHandler(storage StorageAccessor) (*Handler, error) {
	return &Handler{
		storage: storage,
	}, nil
}

// CheckAccess checks if the agent has the specified access to the resource.
func (h *Handler) CheckAccess(resourcePath string, agent Agent, mode AccessMode) (bool, error) {
	// Find the ACL for the resource
	acl, err := h.findACL(resourcePath)
	if err != nil {
		return false, fmt.Errorf("failed to find ACL: %w", err)
	}

	// Check each authorization
	for _, auth := range acl.Authorizations {
		// Check if the authorization applies to the resource
		resourceMatches := false
		for _, authResource := range auth.Resources {
			if authResource == resourcePath {
				resourceMatches = true
				break
			}
			// Check if this is a default ACL for a container
			if acl.IsDefault && strings.HasPrefix(resourcePath, authResource) {
				resourceMatches = true
				break
			}
		}
		if !resourceMatches {
			continue
		}

		// Check if the authorization applies to the agent
		agentMatches := false
		for _, authAgent := range auth.Agents {
			if authAgent.Type == Public {
				// Public access applies to everyone
				agentMatches = true
				break
			} else if authAgent.Type == Authenticated && agent.WebID != "" {
				// Authenticated access applies to any authenticated agent
				agentMatches = true
				break
			} else if authAgent.Type == User && authAgent.WebID == agent.WebID {
				// User access applies to a specific user
				agentMatches = true
				break
			} else if authAgent.Type == Group {
				// Group access - would need to check group membership
				// For simplicity, we skip this for now
			}
		}
		if !agentMatches {
			continue
		}

		// Check if the authorization grants the requested mode
		if auth.Modes.Contains(mode) {
			return true, nil
		}
	}

	return false, nil
}

// findACL finds the ACL for a resource.
func (h *Handler) findACL(resourcePath string) (*ACL, error) {
	// Try to find the ACL for the resource
	aclPath := resourcePath + ".acl"
	exists, err := h.storage.ResourceExists(aclPath)
	if err != nil {
		return nil, err
	}
	if exists {
		return h.parseACL(aclPath, resourcePath)
	}

	// If not found, look for the closest parent container with a default ACL
	parentPath := path.Dir(resourcePath)
	if parentPath == "." || parentPath == "/" {
		return nil, errors.New("no ACL found")
	}

	parentACLPath := parentPath + "/.acl"
	exists, err = h.storage.ResourceExists(parentACLPath)
	if err != nil {
		return nil, err
	}
	if exists {
		acl, err := h.parseACL(parentACLPath, parentPath)
		if err != nil {
			return nil, err
		}
		acl.IsDefault = true
		return acl, nil
	}

	// Recursively check parent containers
	return h.findACL(parentPath)
}

// parseACL parses an ACL from a resource.
func (h *Handler) parseACL(aclPath, resourcePath string) (*ACL, error) {
	// Get the ACL resource
	data, _, err := h.storage.GetResource(aclPath)
	if err != nil {
		return nil, err
	}

	// Parse the ACL as RDF
	parser := rdf.NewTurtleParser()
	graph, err := parser.Parse(strings.NewReader(string(data)))
	if err != nil {
		return nil, fmt.Errorf("failed to parse ACL: %w", err)
	}

	// Create the ACL
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

// findTriples finds triples in the graph matching the given subject, predicate, and object.
// This is a simplified version of the function in the RDF package.
func findTriples(graph *rdf.Graph, subject, predicate, object string) []rdf.Triple {
	var matches []rdf.Triple

	for _, triple := range graph.Triples {
		subjectMatches := subject == "" || (triple.Subject.Type() == rdf.IRI && triple.Subject.(rdf.IRIRef).IRI == subject)
		predicateMatches := predicate == "" || (triple.Predicate.Type() == rdf.IRI && triple.Predicate.(rdf.IRIRef).IRI == predicate)
		objectMatches := object == "" ||
			(triple.Object.Type() == rdf.IRI && triple.Object.(rdf.IRIRef).IRI == object) ||
			(triple.Object.Type() == rdf.LiteralType && triple.Object.(rdf.Literal).Value == object)

		if subjectMatches && predicateMatches && objectMatches {
			matches = append(matches, triple)
		}
	}

	return matches
}

// CreateACL creates a new ACL for a resource.
func CreateACL(resource string, authorizations []Authorization) (*rdf.Graph, error) {
	graph := &rdf.Graph{}

	for i, auth := range authorizations {
		// Create an ID for the authorization if not provided
		authID := auth.ID
		if authID == "" {
			authID = fmt.Sprintf("%s#authorization-%d", resource, i+1)
		}

		// Add the type triple
		graph.AddTriple(
			rdf.IRIRef{IRI: authID},
			rdf.IRIRef{IRI: "http://www.w3.org/1999/02/22-rdf-syntax-ns#type"},
			rdf.IRIRef{IRI: "http://www.w3.org/ns/auth/acl#Authorization"},
		)

		// Add the access modes
		if auth.Modes&Read != 0 {
			graph.AddTriple(
				rdf.IRIRef{IRI: authID},
				rdf.IRIRef{IRI: "http://www.w3.org/ns/auth/acl#mode"},
				rdf.IRIRef{IRI: "http://www.w3.org/ns/auth/acl#Read"},
			)
		}
		if auth.Modes&Write != 0 {
			graph.AddTriple(
				rdf.IRIRef{IRI: authID},
				rdf.IRIRef{IRI: "http://www.w3.org/ns/auth/acl#mode"},
				rdf.IRIRef{IRI: "http://www.w3.org/ns/auth/acl#Write"},
			)
		}
		if auth.Modes&Append != 0 {
			graph.AddTriple(
				rdf.IRIRef{IRI: authID},
				rdf.IRIRef{IRI: "http://www.w3.org/ns/auth/acl#mode"},
				rdf.IRIRef{IRI: "http://www.w3.org/ns/auth/acl#Append"},
			)
		}
		if auth.Modes&Control != 0 {
			graph.AddTriple(
				rdf.IRIRef{IRI: authID},
				rdf.IRIRef{IRI: "http://www.w3.org/ns/auth/acl#mode"},
				rdf.IRIRef{IRI: "http://www.w3.org/ns/auth/acl#Control"},
			)
		}

		// Add the resources
		for _, resourceURI := range auth.Resources {
			graph.AddTriple(
				rdf.IRIRef{IRI: authID},
				rdf.IRIRef{IRI: "http://www.w3.org/ns/auth/acl#accessTo"},
				rdf.IRIRef{IRI: resourceURI},
			)
		}

		// Add the agents
		for _, agent := range auth.Agents {
			switch agent.Type {
			case User:
				graph.AddTriple(
					rdf.IRIRef{IRI: authID},
					rdf.IRIRef{IRI: "http://www.w3.org/ns/auth/acl#agent"},
					rdf.IRIRef{IRI: agent.WebID},
				)
			case Public:
				graph.AddTriple(
					rdf.IRIRef{IRI: authID},
					rdf.IRIRef{IRI: "http://www.w3.org/ns/auth/acl#agentClass"},
					rdf.IRIRef{IRI: "http://xmlns.com/foaf/0.1/Agent"},
				)
			case Authenticated:
				graph.AddTriple(
					rdf.IRIRef{IRI: authID},
					rdf.IRIRef{IRI: "http://www.w3.org/ns/auth/acl#agentClass"},
					rdf.IRIRef{IRI: "http://www.w3.org/ns/auth/acl#AuthenticatedAgent"},
				)
			case Group:
				graph.AddTriple(
					rdf.IRIRef{IRI: authID},
					rdf.IRIRef{IRI: "http://www.w3.org/ns/auth/acl#agentClass"},
					rdf.IRIRef{IRI: agent.WebID},
				)
			}
		}
	}

	return graph, nil
}

// SerializeACL serializes an ACL to Turtle format.
func SerializeACL(acl *ACL, w io.Writer) error {
	// Convert the ACL to an RDF graph
	graph, err := CreateACL(acl.Resource, acl.Authorizations)
	if err != nil {
		return err
	}

	// Serialize the graph to Turtle
	writer := rdf.NewTurtleWriter()

	// Add common prefixes for WAC
	writer.AddPrefix("acl", "http://www.w3.org/ns/auth/acl#")
	writer.AddPrefix("foaf", "http://xmlns.com/foaf/0.1/")

	return writer.Write(graph, w)
}
