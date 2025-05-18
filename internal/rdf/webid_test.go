package rdf

import (
	"bytes"
	"strings"
	"testing"
)

func TestParseWebID(t *testing.T) {
	// Sample WebID profile in Turtle format
	webidTurtle := `
@prefix foaf: <http://xmlns.com/foaf/0.1/> .
@prefix cert: <http://www.w3.org/ns/auth/cert#> .
@prefix xsd: <http://www.w3.org/2001/XMLSchema#> .

<#me>
    a foaf:Person ;
    foaf:name "Alice Smith" ;
    foaf:mbox <mailto:alice@example.org> ;
    foaf:knows <https://bob.example.org/profile#me> ;
    cert:key <#key-1> .

<#key-1>
    a cert:RSAPublicKey ;
    cert:modulus "00cb24facb8c8..."^^xsd:hexBinary ;
    cert:exponent 65537 .
`

	tests := []struct {
		name     string
		data     string
		baseURI  string
		wantErr  bool
		wantName string
		wantEmail string
		wantKeyCount int
		wantAgentCount int
	}{
		{
			name:     "Valid WebID profile",
			data:     webidTurtle,
			baseURI:  "https://alice.example.org/profile",
			wantErr:  false,
			wantName: "Alice Smith",
			wantEmail: "alice@example.org",
			wantKeyCount: 1,
			wantAgentCount: 1,
		},
		{
			name:     "Invalid Turtle syntax",
			data:     "invalid turtle data",
			baseURI:  "https://alice.example.org/profile",
			wantErr:  true,
		},
		{
			name:     "Empty profile",
			data:     "",
			baseURI:  "https://alice.example.org/profile",
			wantErr:  false,
			wantName: "",
			wantEmail: "",
			wantKeyCount: 0,
			wantAgentCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := strings.NewReader(tt.data)
			webid, err := ParseWebID(reader, tt.baseURI)

			if (err != nil) != tt.wantErr {
				t.Errorf("ParseWebID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				return
			}

			if webid.Name != tt.wantName {
				t.Errorf("ParseWebID() name = %v, want %v", webid.Name, tt.wantName)
			}

			if webid.Email != tt.wantEmail {
				t.Errorf("ParseWebID() email = %v, want %v", webid.Email, tt.wantEmail)
			}

			if len(webid.PublicKeys) != tt.wantKeyCount {
				t.Errorf("ParseWebID() key count = %v, want %v", len(webid.PublicKeys), tt.wantKeyCount)
			}

			if len(webid.KnownAgents) != tt.wantAgentCount {
				t.Errorf("ParseWebID() known agent count = %v, want %v", len(webid.KnownAgents), tt.wantAgentCount)
			}
		})
	}
}

func TestCreateWebIDProfile(t *testing.T) {
	tests := []struct {
		name     string
		webid    *WebID
		wantErr  bool
		wantTriples int
	}{
		{
			name: "Basic WebID",
			webid: &WebID{
				URI:  "https://alice.example.org/profile#me",
				Name: "Alice Smith",
				Email: "alice@example.org",
			},
			wantErr: false,
			wantTriples: 3, // type, name, email
		},
		{
			name: "WebID with public key",
			webid: &WebID{
				URI:  "https://alice.example.org/profile#me",
				Name: "Alice Smith",
				PublicKeys: []WebIDPublicKey{
					{
						ModulusHex: "00cb24facb8c8...",
						Exponent:   65537,
					},
				},
			},
			wantErr: false,
			wantTriples: 6, // type, name, key reference, key type, modulus, exponent
		},
		{
			name: "WebID with known agent",
			webid: &WebID{
				URI:  "https://alice.example.org/profile#me",
				KnownAgents: []string{"https://bob.example.org/profile#me"},
			},
			wantErr: false,
			wantTriples: 2, // type, knows
		},
		{
			name: "WebID without URI",
			webid: &WebID{
				Name: "Alice Smith",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			graph, err := CreateWebIDProfile(tt.webid)

			if (err != nil) != tt.wantErr {
				t.Errorf("CreateWebIDProfile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				return
			}

			if len(graph.Triples) != tt.wantTriples {
				t.Errorf("CreateWebIDProfile() triple count = %v, want %v", len(graph.Triples), tt.wantTriples)
			}
		})
	}
}

func TestSerializeWebID(t *testing.T) {
	webid := &WebID{
		URI:  "https://alice.example.org/profile#me",
		Name: "Alice Smith",
		Email: "alice@example.org",
		PublicKeys: []WebIDPublicKey{
			{
				ModulusHex: "00cb24facb8c8...",
				Exponent:   65537,
			},
		},
		KnownAgents: []string{"https://bob.example.org/profile#me"},
	}

	var buf bytes.Buffer
	err := SerializeWebID(webid, &buf)
	if err != nil {
		t.Errorf("SerializeWebID() unexpected error = %v", err)
	}

	output := buf.String()

	// Check for essential elements in the output
	essentialElements := []string{
		"foaf:Person",
		"foaf:name \"Alice Smith\"",
		"foaf:mbox <mailto:alice@example.org>",
		"foaf:knows <https://bob.example.org/profile#me>",
		"cert:RSAPublicKey",
		"cert:modulus \"00cb24facb8c8...\"",
		"cert:exponent \"65537\"",
	}

	for _, element := range essentialElements {
		if !strings.Contains(output, element) {
			t.Errorf("SerializeWebID() output missing element: %s", element)
		}
	}
}

func TestFindTriples(t *testing.T) {
	// Create a test graph
	graph := &Graph{}
	
	graph.AddTriple(
		IRIRef{IRI: "http://example.org/alice"},
		IRIRef{IRI: "http://xmlns.com/foaf/0.1/name"},
		Literal{Value: "Alice"},
	)
	
	graph.AddTriple(
		IRIRef{IRI: "http://example.org/alice"},
		IRIRef{IRI: "http://xmlns.com/foaf/0.1/mbox"},
		IRIRef{IRI: "mailto:alice@example.org"},
	)
	
	graph.AddTriple(
		IRIRef{IRI: "http://example.org/bob"},
		IRIRef{IRI: "http://xmlns.com/foaf/0.1/name"},
		Literal{Value: "Bob"},
	)

	tests := []struct {
		name     string
		subject  string
		predicate string
		object   string
		wantCount int
	}{
		{
			name:     "Match by subject",
			subject:  "http://example.org/alice",
			predicate: "",
			object:   "",
			wantCount: 2,
		},
		{
			name:     "Match by predicate",
			subject:  "",
			predicate: "http://xmlns.com/foaf/0.1/name",
			object:   "",
			wantCount: 2,
		},
		{
			name:     "Match by object",
			subject:  "",
			predicate: "",
			object:   "Alice",
			wantCount: 1,
		},
		{
			name:     "Match by subject and predicate",
			subject:  "http://example.org/alice",
			predicate: "http://xmlns.com/foaf/0.1/name",
			object:   "",
			wantCount: 1,
		},
		{
			name:     "No matches",
			subject:  "http://example.org/charlie",
			predicate: "",
			object:   "",
			wantCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			triples := findTriples(graph, tt.subject, tt.predicate, tt.object)

			if len(triples) != tt.wantCount {
				t.Errorf("findTriples() count = %v, want %v", len(triples), tt.wantCount)
			}
		})
	}
}

func TestExtractPublicKey(t *testing.T) {
	// Create a test graph with a public key
	graph := &Graph{}
	
	keyID := "http://example.org/alice#key-1"
	
	graph.AddTriple(
		IRIRef{IRI: keyID},
		IRIRef{IRI: "http://www.w3.org/1999/02/22-rdf-syntax-ns#type"},
		IRIRef{IRI: "http://www.w3.org/ns/auth/cert#RSAPublicKey"},
	)
	
	graph.AddTriple(
		IRIRef{IRI: keyID},
		IRIRef{IRI: "http://www.w3.org/ns/auth/cert#modulus"},
		Literal{Value: "00cb24facb8c8..."},
	)
	
	graph.AddTriple(
		IRIRef{IRI: keyID},
		IRIRef{IRI: "http://www.w3.org/ns/auth/cert#exponent"},
		Literal{Value: "65537"},
	)

	tests := []struct {
		name     string
		keyID    string
		wantErr  bool
		wantModulus string
		wantExponent int
	}{
		{
			name:     "Valid key",
			keyID:    keyID,
			wantErr:  false,
			wantModulus: "00cb24facb8c8...",
			wantExponent: 65537,
		},
		{
			name:     "Non-existent key",
			keyID:    "http://example.org/alice#key-2",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key, err := extractPublicKey(graph, tt.keyID)

			if (err != nil) != tt.wantErr {
				t.Errorf("extractPublicKey() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				return
			}

			if key.ModulusHex != tt.wantModulus {
				t.Errorf("extractPublicKey() modulus = %v, want %v", key.ModulusHex, tt.wantModulus)
			}

			if key.Exponent != tt.wantExponent {
				t.Errorf("extractPublicKey() exponent = %v, want %v", key.Exponent, tt.wantExponent)
			}
		})
	}
} 