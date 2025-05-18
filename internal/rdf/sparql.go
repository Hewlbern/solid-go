package rdf

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
)

// SparqlQuery represents a SPARQL query.
type SparqlQuery struct {
	// QueryType is the type of query (SELECT, CONSTRUCT, ASK, DESCRIBE).
	QueryType string
	// Variables are the variables selected in a SELECT query.
	Variables []string
	// Where contains the where clause patterns.
	Where []TriplePattern
	// Prefixes contains the prefix declarations.
	Prefixes map[string]string
}

// TriplePattern represents a pattern in a SPARQL WHERE clause.
type TriplePattern struct {
	Subject   SparqlTerm
	Predicate SparqlTerm
	Object    SparqlTerm
}

// SparqlTerm represents a term in a SPARQL query.
type SparqlTerm interface {
	// TermType returns the type of the term.
	TermType() string
	// String returns a string representation of the term.
	String() string
	// MatchesTerm returns whether this term matches the given RDF term.
	MatchesTerm(Term) bool
}

// Variable represents a variable in a SPARQL query.
type Variable struct {
	Name string
}

// TermType returns the type of the term.
func (v Variable) TermType() string {
	return "Variable"
}

// String returns a string representation of the term.
func (v Variable) String() string {
	return "?" + v.Name
}

// MatchesTerm returns whether this term matches the given RDF term.
func (v Variable) MatchesTerm(term Term) bool {
	// Variables match any term
	return true
}

// IRIRefTerm represents an IRI reference in a SPARQL query.
type IRIRefTerm struct {
	IRI string
}

// TermType returns the type of the term.
func (i IRIRefTerm) TermType() string {
	return "IRI"
}

// String returns a string representation of the term.
func (i IRIRefTerm) String() string {
	return "<" + i.IRI + ">"
}

// MatchesTerm returns whether this term matches the given RDF term.
func (i IRIRefTerm) MatchesTerm(term Term) bool {
	if term.Type() != IRI {
		return false
	}
	iri, ok := term.(IRIRef)
	if !ok {
		return false
	}
	return iri.IRI == i.IRI
}

// LiteralTerm represents a literal in a SPARQL query.
type LiteralTerm struct {
	Value    string
	Datatype string
	Language string
}

// TermType returns the type of the term.
func (l LiteralTerm) TermType() string {
	return "Literal"
}

// String returns a string representation of the term.
func (l LiteralTerm) String() string {
	if l.Language != "" {
		return fmt.Sprintf(`"%s"@%s`, l.Value, l.Language)
	} else if l.Datatype != "" {
		return fmt.Sprintf(`"%s"^^<%s>`, l.Value, l.Datatype)
	}
	return fmt.Sprintf(`"%s"`, l.Value)
}

// MatchesTerm returns whether this term matches the given RDF term.
func (l LiteralTerm) MatchesTerm(term Term) bool {
	if term.Type() != LiteralType {
		return false
	}
	lit, ok := term.(Literal)
	if !ok {
		return false
	}
	if lit.Value != l.Value {
		return false
	}
	if l.Language != "" && lit.Language != l.Language {
		return false
	}
	if l.Datatype != "" && lit.Datatype.IRI != l.Datatype {
		return false
	}
	return true
}

// ParseSparqlQuery parses a SPARQL query string.
func ParseSparqlQuery(query string) (*SparqlQuery, error) {
	// This is a very simplified SPARQL parser that only handles basic SELECT queries
	q := &SparqlQuery{
		Prefixes: make(map[string]string),
	}

	// Extract and remove prefix declarations
	prefixPattern := regexp.MustCompile(`PREFIX\s+([a-zA-Z0-9_-]+):\s+<([^>]+)>`)
	prefixMatches := prefixPattern.FindAllStringSubmatch(query, -1)
	for _, match := range prefixMatches {
		if len(match) == 3 {
			q.Prefixes[match[1]] = match[2]
		}
	}
	query = prefixPattern.ReplaceAllString(query, "")

	// Determine query type
	if strings.HasPrefix(strings.ToUpper(strings.TrimSpace(query)), "SELECT") {
		q.QueryType = "SELECT"

		// Extract SELECT variables
		selectPattern := regexp.MustCompile(`SELECT\s+((?:\?[a-zA-Z0-9_]+\s*)+)`)
		selectMatch := selectPattern.FindStringSubmatch(query)
		if len(selectMatch) >= 2 {
			varStr := selectMatch[1]
			for _, v := range strings.Fields(varStr) {
				if v != "" && strings.HasPrefix(v, "?") {
					q.Variables = append(q.Variables, v[1:]) // Remove '?' prefix
				}
			}
		}

		// Extract WHERE clause
		wherePattern := regexp.MustCompile(`WHERE\s*\{([\s\S]*)\}`)
		whereMatch := wherePattern.FindStringSubmatch(query)
		if len(whereMatch) >= 2 {
			whereClause := whereMatch[1]
			patterns := strings.Split(whereClause, ".")

			for _, pattern := range patterns {
				pattern = strings.TrimSpace(pattern)
				if pattern == "" {
					continue
				}

				parts := strings.Fields(pattern)
				if len(parts) < 3 {
					continue
				}

				subject, err := parsePatternTerm(parts[0], q.Prefixes)
				if err != nil {
					return nil, err
				}

				predicate, err := parsePatternTerm(parts[1], q.Prefixes)
				if err != nil {
					return nil, err
				}

				object, err := parsePatternTerm(parts[2], q.Prefixes)
				if err != nil {
					return nil, err
				}

				tp := TriplePattern{
					Subject:   subject,
					Predicate: predicate,
					Object:    object,
				}

				q.Where = append(q.Where, tp)
			}
		}
	} else {
		return nil, errors.New("only SELECT queries are supported")
	}

	return q, nil
}

// parsePatternTerm parses a term in a SPARQL pattern.
func parsePatternTerm(s string, prefixes map[string]string) (SparqlTerm, error) {
	s = strings.TrimSpace(s)

	// Variable
	if strings.HasPrefix(s, "?") {
		return Variable{Name: s[1:]}, nil
	}

	// IRI
	if strings.HasPrefix(s, "<") && strings.HasSuffix(s, ">") {
		return IRIRefTerm{IRI: s[1 : len(s)-1]}, nil
	}

	// Prefixed IRI
	if strings.Contains(s, ":") {
		parts := strings.SplitN(s, ":", 2)
		prefix := parts[0]
		local := parts[1]

		if uri, ok := prefixes[prefix]; ok {
			return IRIRefTerm{IRI: uri + local}, nil
		}

		return nil, fmt.Errorf("unknown prefix: %s", prefix)
	}

	// Literal
	if strings.HasPrefix(s, "\"") {
		if strings.HasSuffix(s, "\"") {
			return LiteralTerm{Value: s[1 : len(s)-1]}, nil
		} else if matches := regexp.MustCompile(`"([^"]*)"@([a-zA-Z-]+)`).FindStringSubmatch(s); len(matches) == 3 {
			return LiteralTerm{Value: matches[1], Language: matches[2]}, nil
		} else if matches := regexp.MustCompile(`"([^"]*)"^^<([^>]+)>`).FindStringSubmatch(s); len(matches) == 3 {
			return LiteralTerm{Value: matches[1], Datatype: matches[2]}, nil
		}
	}

	return nil, errors.New("unsupported term format: " + s)
}

// SparqlResult represents the result of a SPARQL query.
type SparqlResult struct {
	// Variables are the variables in the result.
	Variables []string
	// Bindings are the bindings for each result row.
	Bindings []map[string]Term
}

// ExecuteQuery executes a SPARQL query on an RDF graph.
func ExecuteQuery(graph *Graph, query *SparqlQuery) (*SparqlResult, error) {
	if query.QueryType != "SELECT" {
		return nil, errors.New("only SELECT queries are supported")
	}

	result := &SparqlResult{
		Variables: query.Variables,
	}

	// For each pattern in the query
	patterns := query.Where

	// Start with an empty binding
	bindings := []map[string]Term{{}}

	// For each pattern
	for _, pattern := range patterns {
		var newBindings []map[string]Term

		// For each existing binding
		for _, binding := range bindings {
			// For each triple in the graph
			for _, triple := range graph.Triples {
				// Check if the triple matches the pattern with the current binding
				newBinding, matches := matchPatternWithBinding(pattern, triple, binding)
				if matches {
					newBindings = append(newBindings, newBinding)
				}
			}
		}

		bindings = newBindings
	}

	// Filter the bindings to include only the requested variables
	for _, binding := range bindings {
		resultBinding := make(map[string]Term)
		for _, v := range query.Variables {
			if term, ok := binding[v]; ok {
				resultBinding[v] = term
			}
		}
		result.Bindings = append(result.Bindings, resultBinding)
	}

	return result, nil
}

// matchPatternWithBinding checks if a triple matches a pattern with a given binding.
func matchPatternWithBinding(pattern TriplePattern, triple Triple, binding map[string]Term) (map[string]Term, bool) {
	// Clone the binding
	newBinding := make(map[string]Term)
	for k, v := range binding {
		newBinding[k] = v
	}

	// Check subject
	if !matchTermWithBinding(pattern.Subject, triple.Subject, newBinding) {
		return nil, false
	}

	// Check predicate
	if !matchTermWithBinding(pattern.Predicate, triple.Predicate, newBinding) {
		return nil, false
	}

	// Check object
	if !matchTermWithBinding(pattern.Object, triple.Object, newBinding) {
		return nil, false
	}

	return newBinding, true
}

// matchTermWithBinding checks if a term matches a pattern term with a given binding.
func matchTermWithBinding(patternTerm SparqlTerm, term Term, binding map[string]Term) bool {
	// If the pattern term is a variable
	if v, ok := patternTerm.(Variable); ok {
		// If the variable is already bound
		if boundTerm, exists := binding[v.Name]; exists {
			// The bound term must match the current term
			return termEqual(boundTerm, term)
		}

		// Otherwise, bind the variable to the term
		binding[v.Name] = term
		return true
	}

	// Otherwise, check if the pattern term matches the RDF term
	return patternTerm.MatchesTerm(term)
}

// termEqual checks if two RDF terms are equal.
func termEqual(t1, t2 Term) bool {
	if t1.Type() != t2.Type() {
		return false
	}

	return t1.String() == t2.String()
}
