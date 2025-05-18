package rdf

import (
	"fmt"
	"strings"

	"github.com/yourusername/solid-go/internal/auth"
)

// Triple represents an RDF triple with subject, predicate, and object.
type Triple struct {
	Subject   Term
	Predicate Term
	Object    Term
}

// String returns a string representation of the triple.
func (t Triple) String() string {
	return fmt.Sprintf("%s %s %s .", t.Subject, t.Predicate, t.Object)
}

// TermType represents the type of an RDF term.
type TermType int

const (
	// IRI represents an IRI term.
	IRI TermType = iota
	// BlankNode represents a blank node term.
	BlankNode
	// LiteralType represents a literal term.
	LiteralType
)

// Term represents an RDF term (IRI, blank node, or literal).
type Term interface {
	// Type returns the type of the term.
	Type() TermType
	// String returns a string representation of the term.
	String() string
}

// IRIRef represents an IRI reference.
type IRIRef struct {
	IRI string
}

// Type returns the type of the term.
func (i IRIRef) Type() TermType {
	return IRI
}

// String returns a string representation of the IRI.
func (i IRIRef) String() string {
	return fmt.Sprintf("<%s>", i.IRI)
}

// BNode represents a blank node.
type BNode struct {
	ID string
}

// Type returns the type of the term.
func (b BNode) Type() TermType {
	return BlankNode
}

// String returns a string representation of the blank node.
func (b BNode) String() string {
	return "_:" + b.ID
}

// Literal represents an RDF literal.
type Literal struct {
	Value    string
	Datatype IRIRef
	Language string
}

// Type returns the type of the term.
func (l Literal) Type() TermType {
	return LiteralType
}

// String returns a string representation of the literal.
func (l Literal) String() string {
	if l.Language != "" {
		return fmt.Sprintf(`"%s"@%s`, escapeLiteral(l.Value), l.Language)
	} else if l.Datatype.IRI != "" {
		return fmt.Sprintf(`"%s"^^%s`, escapeLiteral(l.Value), l.Datatype)
	}
	return fmt.Sprintf(`"%s"`, escapeLiteral(l.Value))
}

// escapeLiteral escapes a literal string for Turtle serialization.
func escapeLiteral(s string) string {
	return strings.NewReplacer(
		"\\", "\\\\",
		"\"", "\\\"",
		"\n", "\\n",
		"\r", "\\r",
		"\t", "\\t",
	).Replace(s)
}

// Graph represents an RDF graph, which is a set of triples.
type Graph struct {
	Triples []Triple
}

// AddTriple adds a triple to the graph.
func (g *Graph) AddTriple(subject, predicate, object Term) {
	g.Triples = append(g.Triples, Triple{
		Subject:   subject,
		Predicate: predicate,
		Object:    object,
	})
}

// String returns a string representation of the graph in Turtle format.
func (g *Graph) String() string {
	var sb strings.Builder
	for _, triple := range g.Triples {
		sb.WriteString(triple.String())
		sb.WriteString("\n")
	}
	return sb.String()
}

// StorageAccessor provides access to resources in storage.
type StorageAccessor interface {
	GetResource(path string) ([]byte, string, error)
	ResourceExists(path string) (bool, error)
}

// Handler handles RDF operations.
type Handler struct {
	storage StorageAccessor
}

type HTTPHandler struct {
	wacHandler  *Handler
	authHandler auth.Handler
}
