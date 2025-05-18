package wac

import (
	"bytes"
	"strings"
	"testing"

	"github.com/yourusername/solid-go/internal/rdf"
)

// CreateSimpleACL creates a simple ACL for testing.
func CreateSimpleACL(resourcePath string, agent string, modes ...AccessMode) []byte {
	graph := &rdf.Graph{}
	graph.AddTriple(
		rdf.IRIRef{IRI: resourcePath + "#auth"},
		rdf.IRIRef{IRI: "http://www.w3.org/1999/02/22-rdf-syntax-ns#type"},
		rdf.IRIRef{IRI: "http://www.w3.org/ns/auth/acl#Authorization"},
	)
	graph.AddTriple(
		rdf.IRIRef{IRI: resourcePath + "#auth"},
		rdf.IRIRef{IRI: "http://www.w3.org/ns/auth/acl#agent"},
		rdf.IRIRef{IRI: agent},
	)
	graph.AddTriple(
		rdf.IRIRef{IRI: resourcePath + "#auth"},
		rdf.IRIRef{IRI: "http://www.w3.org/ns/auth/acl#accessTo"},
		rdf.IRIRef{IRI: resourcePath},
	)

	for _, mode := range modes {
		graph.AddTriple(
			rdf.IRIRef{IRI: resourcePath + "#auth"},
			rdf.IRIRef{IRI: "http://www.w3.org/ns/auth/acl#mode"},
			rdf.IRIRef{IRI: "http://www.w3.org/ns/auth/acl#" + mode.String()},
		)
	}

	var buffer bytes.Buffer
	writer := rdf.NewTurtleWriter()
	writer.AddPrefix("acl", "http://www.w3.org/ns/auth/acl#")
	writer.Write(graph, &buffer)
	return buffer.Bytes()
}

// containsTriple checks if a graph contains a triple.
func containsTriple(graph *rdf.Graph, subjectStr, predicateStr, objectStr string) bool {
	for _, triple := range graph.Triples {
		subjectMatches := triple.Subject.Type() == rdf.IRI && triple.Subject.(rdf.IRIRef).IRI == subjectStr
		predicateMatches := triple.Predicate.Type() == rdf.IRI && triple.Predicate.(rdf.IRIRef).IRI == predicateStr
		objectMatches := triple.Object.Type() == rdf.IRI && triple.Object.(rdf.IRIRef).IRI == objectStr
		if subjectMatches && predicateMatches && objectMatches {
			return true
		}
	}
	return false
}

// TestAccessModeString tests the String method of AccessMode.
func TestAccessModeString(t *testing.T) {
	tests := []struct {
		mode AccessMode
		want string
	}{
		{Read, "Read"},
		{Write, "Write"},
		{Append, "Append"},
		{Control, "Control"},
		{Read | Write, "Read, Write"},
		{Read | Write | Append, "Read, Write, Append"},
		{Read | Write | Append | Control, "Read, Write, Append, Control"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := tt.mode.String()
			if got != tt.want {
				t.Errorf("AccessMode.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAccessModeContains(t *testing.T) {
	tests := []struct {
		mode     AccessMode
		other    AccessMode
		expected bool
	}{
		{Read, Read, true},
		{Read | Write, Read, true},
		{Read, Read | Write, false},
		{Read | Write | Control, Read | Control, true},
		{Read | Write | Control, Read | Write | Control | Append, false},
	}

	for _, test := range tests {
		t.Run(test.mode.String()+"_contains_"+test.other.String(), func(t *testing.T) {
			result := test.mode.Contains(test.other)
			if result != test.expected {
				t.Errorf("Expected %v, got %v", test.expected, result)
			}
		})
	}
}

func TestAgentTypeString(t *testing.T) {
	tests := []struct {
		agentType AgentType
		expected  string
	}{
		{User, "User"},
		{Group, "Group"},
		{Public, "Public"},
		{Authenticated, "Authenticated"},
		{AgentType(99), "Unknown"},
	}

	for _, test := range tests {
		t.Run(test.expected, func(t *testing.T) {
			result := test.agentType.String()
			if result != test.expected {
				t.Errorf("Expected %q, got %q", test.expected, result)
			}
		})
	}
}

func TestURIForType(t *testing.T) {
	tests := []struct {
		agentType AgentType
		expected  string
	}{
		{Public, "http://xmlns.com/foaf/0.1/Agent"},
		{Authenticated, "http://www.w3.org/ns/auth/acl#AuthenticatedAgent"},
		{User, ""},
		{Group, ""},
	}

	for _, test := range tests {
		t.Run(test.agentType.String(), func(t *testing.T) {
			result := URIForType(test.agentType)
			if result != test.expected {
				t.Errorf("Expected %q, got %q", test.expected, result)
			}
		})
	}
}

func TestCheckAccess(t *testing.T) {
	storage := NewMockStorage()

	// Add some ACLs to the storage
	resourcePath := "/resource"
	resourceACLPath := "/resource.acl"
	storage.AddResource(resourceACLPath, CreateSimpleACL(resourcePath, "https://example.com/profile/owner#me", Write, Control))

	containerPath := "/container/"
	containerACLPath := "/container/.acl"
	storage.AddResource(containerACLPath, CreateSimpleACL(containerPath, "https://example.com/profile/owner#me", Read, Write, Control))

	handler, err := NewHandler(storage)
	if err != nil {
		t.Fatalf("Failed to create handler: %v", err)
	}

	tests := []struct {
		name         string
		resourcePath string
		agent        Agent
		mode         AccessMode
		expected     bool
	}{
		{
			name:         "Public read access to resource",
			resourcePath: resourcePath,
			agent:        Agent{Type: Public},
			mode:         Read,
			expected:     true,
		},
		{
			name:         "Public write access to resource",
			resourcePath: resourcePath,
			agent:        Agent{Type: Public},
			mode:         Write,
			expected:     false,
		},
		{
			name:         "Owner write access to resource",
			resourcePath: resourcePath,
			agent:        Agent{WebID: "https://example.com/profile/owner#me", Type: User},
			mode:         Write,
			expected:     true,
		},
		{
			name:         "Owner control access to resource",
			resourcePath: resourcePath,
			agent:        Agent{WebID: "https://example.com/profile/owner#me", Type: User},
			mode:         Control,
			expected:     true,
		},
		{
			name:         "Other user write access to resource",
			resourcePath: resourcePath,
			agent:        Agent{WebID: "https://example.com/profile/other#me", Type: User},
			mode:         Write,
			expected:     false,
		},
		{
			name:         "Public read access to container",
			resourcePath: containerPath,
			agent:        Agent{Type: Public},
			mode:         Read,
			expected:     true,
		},
		{
			name:         "Public read access to resource in container",
			resourcePath: containerPath + "resource",
			agent:        Agent{Type: Public},
			mode:         Read,
			expected:     true,
		},
		{
			name:         "Owner write access to resource in container",
			resourcePath: containerPath + "resource",
			agent:        Agent{WebID: "https://example.com/profile/owner#me", Type: User},
			mode:         Write,
			expected:     true,
		},
		{
			name:         "Other user write access to resource in container",
			resourcePath: containerPath + "resource",
			agent:        Agent{WebID: "https://example.com/profile/other#me", Type: User},
			mode:         Write,
			expected:     false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := handler.CheckAccess(test.resourcePath, test.agent, test.mode)
			if err != nil {
				t.Fatalf("Failed to check access: %v", err)
			}
			if result != test.expected {
				t.Errorf("Expected %v, got %v", test.expected, result)
			}
		})
	}
}

func TestCreateACL(t *testing.T) {
	// Test creating an ACL for a resource
	resourcePath := "/resource"
	auths := []Authorization{
		{
			ID:        resourcePath + "#public",
			Modes:     Read,
			Agents:    []Agent{{Type: Public}},
			Resources: []string{resourcePath},
		},
		{
			ID:        resourcePath + "#owner",
			Modes:     Write | Control,
			Agents:    []Agent{{WebID: "https://example.com/profile/owner#me", Type: User}},
			Resources: []string{resourcePath},
		},
	}

	graph, err := CreateACL(resourcePath, auths)
	if err != nil {
		t.Fatalf("Failed to create ACL: %v", err)
	}

	// Check that the graph contains the expected triples
	// Public read access
	if !containsTriple(graph, resourcePath+"#public", "http://www.w3.org/1999/02/22-rdf-syntax-ns#type", "http://www.w3.org/ns/auth/acl#Authorization") {
		t.Errorf("Missing public authorization type triple")
	}
	if !containsTriple(graph, resourcePath+"#public", "http://www.w3.org/ns/auth/acl#mode", "http://www.w3.org/ns/auth/acl#Read") {
		t.Errorf("Missing public read mode triple")
	}
	if !containsTriple(graph, resourcePath+"#public", "http://www.w3.org/ns/auth/acl#accessTo", resourcePath) {
		t.Errorf("Missing public accessTo triple")
	}
	if !containsTriple(graph, resourcePath+"#public", "http://www.w3.org/ns/auth/acl#agentClass", "http://xmlns.com/foaf/0.1/Agent") {
		t.Errorf("Missing public agentClass triple")
	}

	// Owner write and control access
	if !containsTriple(graph, resourcePath+"#owner", "http://www.w3.org/1999/02/22-rdf-syntax-ns#type", "http://www.w3.org/ns/auth/acl#Authorization") {
		t.Errorf("Missing owner authorization type triple")
	}
	if !containsTriple(graph, resourcePath+"#owner", "http://www.w3.org/ns/auth/acl#mode", "http://www.w3.org/ns/auth/acl#Write") {
		t.Errorf("Missing owner write mode triple")
	}
	if !containsTriple(graph, resourcePath+"#owner", "http://www.w3.org/ns/auth/acl#mode", "http://www.w3.org/ns/auth/acl#Control") {
		t.Errorf("Missing owner control mode triple")
	}
	if !containsTriple(graph, resourcePath+"#owner", "http://www.w3.org/ns/auth/acl#accessTo", resourcePath) {
		t.Errorf("Missing owner accessTo triple")
	}
	if !containsTriple(graph, resourcePath+"#owner", "http://www.w3.org/ns/auth/acl#agent", "https://example.com/profile/owner#me") {
		t.Errorf("Missing owner agent triple")
	}
}

func TestSerializeACL(t *testing.T) {
	resourcePath := "/resource"
	acl := &ACL{
		Resource: resourcePath,
		Authorizations: []Authorization{
			{
				ID:        resourcePath + "#public",
				Modes:     Read,
				Agents:    []Agent{{Type: Public}},
				Resources: []string{resourcePath},
			},
			{
				ID:        resourcePath + "#owner",
				Modes:     Write | Control,
				Agents:    []Agent{{WebID: "https://example.com/profile/owner#me", Type: User}},
				Resources: []string{resourcePath},
			},
		},
	}

	var buffer bytes.Buffer
	err := SerializeACL(acl, &buffer)
	if err != nil {
		t.Fatalf("Failed to serialize ACL: %v", err)
	}

	// Check that the serialized ACL contains the expected content
	serialized := buffer.String()
	if !strings.Contains(serialized, "@prefix acl: <http://www.w3.org/ns/auth/acl#>") {
		t.Errorf("Missing acl prefix in serialized ACL")
	}
	if !strings.Contains(serialized, "@prefix foaf: <http://xmlns.com/foaf/0.1/>") {
		t.Errorf("Missing foaf prefix in serialized ACL")
	}
	if !strings.Contains(serialized, "<"+resourcePath+"#public>") {
		t.Errorf("Missing public authorization in serialized ACL")
	}
	if !strings.Contains(serialized, "<"+resourcePath+"#owner>") {
		t.Errorf("Missing owner authorization in serialized ACL")
	}
	if !strings.Contains(serialized, "acl:Read") {
		t.Errorf("Missing read mode in serialized ACL")
	}
	if !strings.Contains(serialized, "acl:Write") {
		t.Errorf("Missing write mode in serialized ACL")
	}
	if !strings.Contains(serialized, "acl:Control") {
		t.Errorf("Missing control mode in serialized ACL")
	}
	if !strings.Contains(serialized, "foaf:Agent") {
		t.Errorf("Missing foaf:Agent in serialized ACL")
	}
	if !strings.Contains(serialized, "<https://example.com/profile/owner#me>") {
		t.Errorf("Missing owner WebID in serialized ACL")
	}
}
