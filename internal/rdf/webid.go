package rdf

import (
	"errors"
	"fmt"
	"io"
	"strings"
)

// WebID represents a WebID profile.
type WebID struct {
	URI         string
	Name        string
	Email       string
	PublicKeys  []WebIDPublicKey
	KnownAgents []string
	TypeURIs    []string
	Graph       *Graph
}

// WebIDPublicKey represents a public key in a WebID profile.
type WebIDPublicKey struct {
	ID         string
	ModulusHex string
	Exponent   int
}

// ParseWebID parses a WebID profile from RDF data.
func ParseWebID(r io.Reader, baseURI string) (*WebID, error) {
	parser := NewTurtleParser()
	graph, err := parser.Parse(r)
	if err != nil {
		return nil, fmt.Errorf("failed to parse WebID profile: %w", err)
	}

	webid := &WebID{
		URI:   baseURI,
		Graph: graph,
	}

	// Find the WebID subject - it's typically the fragment in the URI
	webidRef := baseURI
	if !strings.Contains(baseURI, "#") {
		webidRef = baseURI + "#me"
	}

	// Extract name
	nameTriples := findTriples(graph, webidRef, "http://xmlns.com/foaf/0.1/name", "")
	if len(nameTriples) > 0 {
		if lit, ok := nameTriples[0].Object.(Literal); ok {
			webid.Name = lit.Value
		}
	}

	// Extract email
	emailTriples := findTriples(graph, webidRef, "http://xmlns.com/foaf/0.1/mbox", "")
	for _, triple := range emailTriples {
		if iri, ok := triple.Object.(IRIRef); ok && strings.HasPrefix(iri.IRI, "mailto:") {
			webid.Email = iri.IRI[7:] // Remove "mailto:" prefix
			break
		}
	}

	// Extract type URIs
	typeTriples := findTriples(graph, webidRef, "http://www.w3.org/1999/02/22-rdf-syntax-ns#type", "")
	for _, triple := range typeTriples {
		if iri, ok := triple.Object.(IRIRef); ok {
			webid.TypeURIs = append(webid.TypeURIs, iri.IRI)
		}
	}

	// Extract known agents
	knownTriples := findTriples(graph, webidRef, "http://xmlns.com/foaf/0.1/knows", "")
	for _, triple := range knownTriples {
		if iri, ok := triple.Object.(IRIRef); ok {
			webid.KnownAgents = append(webid.KnownAgents, iri.IRI)
		}
	}

	// Extract public keys
	keyTriples := findTriples(graph, webidRef, "http://www.w3.org/ns/auth/cert#key", "")
	for _, keyTriple := range keyTriples {
		if bnode, ok := keyTriple.Object.(BNode); ok {
			publicKey, err := extractPublicKey(graph, bnode.ID)
			if err == nil {
				webid.PublicKeys = append(webid.PublicKeys, publicKey)
			}
		} else if iri, ok := keyTriple.Object.(IRIRef); ok {
			publicKey, err := extractPublicKey(graph, iri.IRI)
			if err == nil {
				webid.PublicKeys = append(webid.PublicKeys, publicKey)
			}
		}
	}

	return webid, nil
}

// extractPublicKey extracts a public key from the graph.
func extractPublicKey(graph *Graph, keyID string) (WebIDPublicKey, error) {
	var publicKey WebIDPublicKey
	publicKey.ID = keyID

	// Check if it's an RSA public key
	typeTriples := findTriples(graph, keyID, "http://www.w3.org/1999/02/22-rdf-syntax-ns#type", "http://www.w3.org/ns/auth/cert#RSAPublicKey")
	if len(typeTriples) == 0 {
		return publicKey, errors.New("not an RSA public key")
	}

	// Extract modulus
	modulusTriples := findTriples(graph, keyID, "http://www.w3.org/ns/auth/cert#modulus", "")
	if len(modulusTriples) > 0 {
		if lit, ok := modulusTriples[0].Object.(Literal); ok {
			publicKey.ModulusHex = lit.Value
		}
	}

	// Extract exponent
	exponentTriples := findTriples(graph, keyID, "http://www.w3.org/ns/auth/cert#exponent", "")
	if len(exponentTriples) > 0 {
		if lit, ok := exponentTriples[0].Object.(Literal); ok {
			fmt.Sscanf(lit.Value, "%d", &publicKey.Exponent)
		}
	}

	if publicKey.ModulusHex == "" || publicKey.Exponent == 0 {
		return publicKey, errors.New("incomplete RSA public key")
	}

	return publicKey, nil
}

// findTriples finds triples in the graph matching the given subject, predicate, and object.
// If subject, predicate, or object is empty, it matches any value.
func findTriples(graph *Graph, subject, predicate, object string) []Triple {
	var matches []Triple

	for _, triple := range graph.Triples {
		subjectMatches := subject == "" || (triple.Subject.Type() == IRI && triple.Subject.(IRIRef).IRI == subject) ||
			(triple.Subject.Type() == BlankNode && "_:"+triple.Subject.(BNode).ID == subject)

		predicateMatches := predicate == "" || (triple.Predicate.Type() == IRI && triple.Predicate.(IRIRef).IRI == predicate)

		objectMatches := object == "" ||
			(triple.Object.Type() == IRI && triple.Object.(IRIRef).IRI == object) ||
			(triple.Object.Type() == LiteralType && triple.Object.(Literal).Value == object)

		if subjectMatches && predicateMatches && objectMatches {
			matches = append(matches, triple)
		}
	}

	return matches
}

// CreateWebIDProfile creates a new WebID profile.
func CreateWebIDProfile(webid *WebID) (*Graph, error) {
	if webid.URI == "" {
		return nil, errors.New("WebID URI is required")
	}

	graph := &Graph{}
	webidRef := webid.URI
	if !strings.Contains(webidRef, "#") {
		webidRef += "#me"
	}

	// Add type
	hasType := false
	for _, typeURI := range webid.TypeURIs {
		graph.AddTriple(
			IRIRef{IRI: webidRef},
			IRIRef{IRI: "http://www.w3.org/1999/02/22-rdf-syntax-ns#type"},
			IRIRef{IRI: typeURI},
		)
		hasType = true
	}

	// Add default type if none provided
	if !hasType {
		graph.AddTriple(
			IRIRef{IRI: webidRef},
			IRIRef{IRI: "http://www.w3.org/1999/02/22-rdf-syntax-ns#type"},
			IRIRef{IRI: "http://xmlns.com/foaf/0.1/Person"},
		)
	}

	// Add name
	if webid.Name != "" {
		graph.AddTriple(
			IRIRef{IRI: webidRef},
			IRIRef{IRI: "http://xmlns.com/foaf/0.1/name"},
			Literal{Value: webid.Name},
		)
	}

	// Add email
	if webid.Email != "" {
		graph.AddTriple(
			IRIRef{IRI: webidRef},
			IRIRef{IRI: "http://xmlns.com/foaf/0.1/mbox"},
			IRIRef{IRI: "mailto:" + webid.Email},
		)
	}

	// Add known agents
	for _, agent := range webid.KnownAgents {
		graph.AddTriple(
			IRIRef{IRI: webidRef},
			IRIRef{IRI: "http://xmlns.com/foaf/0.1/knows"},
			IRIRef{IRI: agent},
		)
	}

	// Add public keys
	for i, key := range webid.PublicKeys {
		keyID := ""
		if key.ID != "" {
			keyID = key.ID
		} else {
			keyID = fmt.Sprintf("%s#key-%d", webidRef, i+1)
		}

		graph.AddTriple(
			IRIRef{IRI: webidRef},
			IRIRef{IRI: "http://www.w3.org/ns/auth/cert#key"},
			IRIRef{IRI: keyID},
		)

		graph.AddTriple(
			IRIRef{IRI: keyID},
			IRIRef{IRI: "http://www.w3.org/1999/02/22-rdf-syntax-ns#type"},
			IRIRef{IRI: "http://www.w3.org/ns/auth/cert#RSAPublicKey"},
		)

		if key.ModulusHex != "" {
			graph.AddTriple(
				IRIRef{IRI: keyID},
				IRIRef{IRI: "http://www.w3.org/ns/auth/cert#modulus"},
				Literal{
					Value:    key.ModulusHex,
					Datatype: IRIRef{IRI: "http://www.w3.org/2001/XMLSchema#hexBinary"},
				},
			)
		}

		if key.Exponent != 0 {
			graph.AddTriple(
				IRIRef{IRI: keyID},
				IRIRef{IRI: "http://www.w3.org/ns/auth/cert#exponent"},
				Literal{
					Value:    fmt.Sprintf("%d", key.Exponent),
					Datatype: IRIRef{IRI: "http://www.w3.org/2001/XMLSchema#integer"},
				},
			)
		}
	}

	return graph, nil
}

// SerializeWebID serializes a WebID profile to Turtle format.
func SerializeWebID(webid *WebID, w io.Writer) error {
	graph, err := CreateWebIDProfile(webid)
	if err != nil {
		return err
	}

	writer := NewTurtleWriter()

	// Add common prefixes for WebID profiles
	writer.AddPrefix("foaf", "http://xmlns.com/foaf/0.1/")
	writer.AddPrefix("cert", "http://www.w3.org/ns/auth/cert#")
	writer.AddPrefix("xsd", "http://www.w3.org/2001/XMLSchema#")

	return writer.Write(graph, w)
}
