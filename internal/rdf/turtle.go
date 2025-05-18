package rdf

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"regexp"
	"strings"
)

// Common namespaces used in RDF.
var commonPrefixes = map[string]string{
	"rdf":  "http://www.w3.org/1999/02/22-rdf-syntax-ns#",
	"rdfs": "http://www.w3.org/2000/01/rdf-schema#",
	"xsd":  "http://www.w3.org/2001/XMLSchema#",
	"foaf": "http://xmlns.com/foaf/0.1/",
}

// TurtleParser parses RDF in Turtle format.
type TurtleParser struct {
	prefixes map[string]string
	bnodeIDs map[string]int
	nextID   int
}

// NewTurtleParser creates a new Turtle parser.
func NewTurtleParser() *TurtleParser {
	return &TurtleParser{
		prefixes: make(map[string]string),
		bnodeIDs: make(map[string]int),
		nextID:   1,
	}
}

// ParseString parses a string containing RDF in Turtle format.
func (p *TurtleParser) ParseString(s string) (*Graph, error) {
	return p.Parse(strings.NewReader(s))
}

// Parse parses RDF in Turtle format from a reader.
func (p *TurtleParser) Parse(r io.Reader) (*Graph, error) {
	// Initialize with common prefixes
	for prefix, uri := range commonPrefixes {
		p.prefixes[prefix] = uri
	}

	graph := &Graph{}
	scanner := bufio.NewScanner(r)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Handle prefix declarations
		if strings.HasPrefix(line, "@prefix") {
			if err := p.parsePrefix(line); err != nil {
				return nil, err
			}
			continue
		}

		// Parse triples
		if err := p.parseTriple(line, graph); err != nil {
			return nil, err
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return graph, nil
}

// parsePrefix parses a prefix declaration.
func (p *TurtleParser) parsePrefix(line string) error {
	// Simple regex for @prefix declaration
	re := regexp.MustCompile(`@prefix\s+([a-zA-Z0-9_-]*):\s+<([^>]+)>\s*\.`)
	matches := re.FindStringSubmatch(line)

	if len(matches) != 3 {
		return fmt.Errorf("invalid prefix declaration: %s", line)
	}

	prefix := matches[1]
	uri := matches[2]
	p.prefixes[prefix] = uri

	return nil
}

// parseTriple parses a triple statement.
func (p *TurtleParser) parseTriple(line string, graph *Graph) error {
	// This is a simplistic parser that handles basic triple patterns
	parts := strings.Split(line, " ")
	if len(parts) < 4 { // Subject, predicate, object, and period
		return fmt.Errorf("invalid triple format: %s", line)
	}

	// Extract subject
	subject, err := p.parseTerm(parts[0])
	if err != nil {
		return err
	}

	// Extract predicate
	predicate, err := p.parseTerm(parts[1])
	if err != nil {
		return err
	}

	// Extract object - it might contain spaces if it's a literal
	objectParts := parts[2 : len(parts)-1] // Exclude the period
	objectStr := strings.Join(objectParts, " ")
	object, err := p.parseTerm(objectStr)
	if err != nil {
		return err
	}

	// Add the triple to the graph
	graph.AddTriple(subject, predicate, object)

	return nil
}

// parseTerm parses an RDF term (IRI, blank node, or literal).
func (p *TurtleParser) parseTerm(s string) (Term, error) {
	s = strings.TrimSpace(s)

	// IRI
	if strings.HasPrefix(s, "<") && strings.HasSuffix(s, ">") {
		iri := s[1 : len(s)-1]
		return IRIRef{IRI: iri}, nil
	}

	// Prefix IRI
	if strings.Contains(s, ":") {
		parts := strings.SplitN(s, ":", 2)
		prefix := parts[0]
		local := parts[1]

		if uri, ok := p.prefixes[prefix]; ok {
			return IRIRef{IRI: uri + local}, nil
		}

		return nil, fmt.Errorf("unknown prefix: %s", prefix)
	}

	// Blank node
	if strings.HasPrefix(s, "_:") {
		id := s[2:]
		return BNode{ID: id}, nil
	}

	// Literal
	if strings.HasPrefix(s, "\"") {
		// This is a simple implementation that doesn't handle all Turtle literal formats
		if strings.HasSuffix(s, "\"") {
			value := s[1 : len(s)-1]
			return Literal{Value: value}, nil
		} else if matches := regexp.MustCompile(`"([^"]*)"@([a-zA-Z-]+)`).FindStringSubmatch(s); len(matches) == 3 {
			value := matches[1]
			lang := matches[2]
			return Literal{Value: value, Language: lang}, nil
		} else if matches := regexp.MustCompile(`"([^"]*)"^^<([^>]+)>`).FindStringSubmatch(s); len(matches) == 3 {
			value := matches[1]
			datatype := matches[2]
			return Literal{Value: value, Datatype: IRIRef{IRI: datatype}}, nil
		}
	}

	// Handle bare integers and other numeric literals
	if matches := regexp.MustCompile(`^(\d+)$`).FindStringSubmatch(s); len(matches) == 2 {
		return Literal{
			Value:    matches[1],
			Datatype: IRIRef{IRI: "http://www.w3.org/2001/XMLSchema#integer"},
		}, nil
	}

	return nil, errors.New("unsupported term format: " + s)
}

// TurtleWriter writes RDF graphs in Turtle format.
type TurtleWriter struct {
	prefixes map[string]string
}

// NewTurtleWriter creates a new Turtle writer.
func NewTurtleWriter() *TurtleWriter {
	prefixes := make(map[string]string)
	for prefix, uri := range commonPrefixes {
		prefixes[prefix] = uri
	}

	return &TurtleWriter{
		prefixes: prefixes,
	}
}

// AddPrefix adds a prefix to the writer.
func (w *TurtleWriter) AddPrefix(prefix, uri string) {
	w.prefixes[prefix] = uri
}

// Write writes an RDF graph in Turtle format.
func (w *TurtleWriter) Write(g *Graph, out io.Writer) error {
	// Write prefixes
	for prefix, uri := range w.prefixes {
		if _, err := fmt.Fprintf(out, "@prefix %s: <%s> .\n", prefix, uri); err != nil {
			return err
		}
	}

	// Write a blank line after prefixes
	if _, err := fmt.Fprintln(out); err != nil {
		return err
	}

	// Write triples
	for _, triple := range g.Triples {
		if _, err := fmt.Fprintln(out, triple.String()); err != nil {
			return err
		}
	}

	return nil
}
