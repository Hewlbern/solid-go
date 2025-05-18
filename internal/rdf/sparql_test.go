package rdf

import (
	"reflect"
	"testing"
)

func TestParseSparqlQuery(t *testing.T) {
	tests := []struct {
		name          string
		queryStr      string
		expectError   bool
		expectedType  string
		expectedVars  []string
		expectedWhere int
	}{
		{
			name:          "Basic SELECT query",
			queryStr:      "SELECT ?s ?p ?o WHERE { ?s ?p ?o . }",
			expectError:   false,
			expectedType:  "SELECT",
			expectedVars:  []string{"s", "p", "o"},
			expectedWhere: 1,
		},
		{
			name:          "SELECT query with prefixes",
			queryStr:      "PREFIX ex: <http://example.org/> SELECT ?name WHERE { ?person ex:name ?name . }",
			expectError:   false,
			expectedType:  "SELECT",
			expectedVars:  []string{"name"},
			expectedWhere: 1,
		},
		{
			name:          "Multiple triple patterns",
			queryStr:      "SELECT ?name ?email WHERE { ?person <http://xmlns.com/foaf/0.1/name> ?name . ?person <http://xmlns.com/foaf/0.1/mbox> ?email . }",
			expectError:   false,
			expectedType:  "SELECT",
			expectedVars:  []string{"name", "email"},
			expectedWhere: 2,
		},
		{
			name:        "Invalid syntax",
			queryStr:    "SELECT ?s WHERE { ?s }",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query, err := ParseSparqlQuery(tt.queryStr)
			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if query.QueryType != tt.expectedType {
				t.Errorf("Expected query type %s, got %s", tt.expectedType, query.QueryType)
			}

			if !reflect.DeepEqual(query.Variables, tt.expectedVars) {
				t.Errorf("Expected variables %v, got %v", tt.expectedVars, query.Variables)
			}

			if len(query.Where) != tt.expectedWhere {
				t.Errorf("Expected %d WHERE patterns, got %d", tt.expectedWhere, len(query.Where))
			}
		})
	}
}

func TestParsePatternTerm(t *testing.T) {
	tests := []struct {
		name        string
		termStr     string
		expectError bool
		expected    SparqlTerm
	}{
		{
			name:    "Variable",
			termStr: "?x",
			expected: Variable{
				Name: "x",
			},
		},
		{
			name:    "IRI",
			termStr: "<http://example.org/resource>",
			expected: IRIRefTerm{
				IRI: "http://example.org/resource",
			},
		},
		{
			name:    "Prefixed IRI",
			termStr: "ex:resource",
			expected: IRIRefTerm{
				IRI: "http://example.org/resource",
			},
		},
		{
			name:    "Literal",
			termStr: "\"Hello, world!\"",
			expected: LiteralTerm{
				Value: "Hello, world!",
			},
		},
		{
			name:    "Literal with language",
			termStr: "\"Hello\"@en",
			expected: LiteralTerm{
				Value:    "Hello",
				Language: "en",
			},
		},
		{
			name:    "Literal with datatype",
			termStr: "\"42\"^^<http://www.w3.org/2001/XMLSchema#integer>",
			expected: LiteralTerm{
				Value:    "42",
				Datatype: "http://www.w3.org/2001/XMLSchema#integer",
			},
		},
		{
			name:        "Invalid term",
			termStr:     "invalid",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			term, err := parsePatternTerm(tt.termStr, map[string]string{
				"ex": "http://example.org/",
			})
			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if !reflect.DeepEqual(term, tt.expected) {
				t.Errorf("Expected %v, got %v", tt.expected, term)
			}
		})
	}
}

func TestExecuteQuery(t *testing.T) {
	// Create a test graph
	graph := &Graph{
		Triples: []Triple{
			{
				Subject:   IRIRef{IRI: "http://example.org/alice"},
				Predicate: IRIRef{IRI: "http://xmlns.com/foaf/0.1/name"},
				Object:    Literal{Value: "Alice"},
			},
			{
				Subject:   IRIRef{IRI: "http://example.org/alice"},
				Predicate: IRIRef{IRI: "http://xmlns.com/foaf/0.1/mbox"},
				Object:    Literal{Value: "alice@example.org"},
			},
			{
				Subject:   IRIRef{IRI: "http://example.org/bob"},
				Predicate: IRIRef{IRI: "http://xmlns.com/foaf/0.1/name"},
				Object:    Literal{Value: "Bob"},
			},
			{
				Subject:   IRIRef{IRI: "http://example.org/bob"},
				Predicate: IRIRef{IRI: "http://xmlns.com/foaf/0.1/mbox"},
				Object:    Literal{Value: "bob@example.org"},
			},
		},
	}

	tests := []struct {
		name          string
		queryStr      string
		expectedVars  []string
		expectedRows  int
		expectedEmail string
	}{
		{
			name:         "Find all names",
			queryStr:     "SELECT ?name WHERE { ?person <http://xmlns.com/foaf/0.1/name> ?name . }",
			expectedVars: []string{"name"},
			expectedRows: 2,
		},
		{
			name:          "Find Alice's email",
			queryStr:      "SELECT ?email WHERE { <http://example.org/alice> <http://xmlns.com/foaf/0.1/mbox> ?email . }",
			expectedVars:  []string{"email"},
			expectedRows:  1,
			expectedEmail: "alice@example.org",
		},
		{
			name:         "Find person and their name",
			queryStr:     "SELECT ?person ?name WHERE { ?person <http://xmlns.com/foaf/0.1/name> ?name . }",
			expectedVars: []string{"person", "name"},
			expectedRows: 2,
		},
		{
			name:         "Find person with both properties",
			queryStr:     "SELECT ?person WHERE { ?person <http://xmlns.com/foaf/0.1/name> ?name . ?person <http://xmlns.com/foaf/0.1/mbox> ?email . }",
			expectedVars: []string{"person"},
			expectedRows: 2,
		},
		{
			name:         "No results",
			queryStr:     "SELECT ?name WHERE { ?person <http://xmlns.com/foaf/0.1/name> ?name . FILTER(?name = 'Charlie') }",
			expectedVars: []string{"name"},
			expectedRows: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query, err := ParseSparqlQuery(tt.queryStr)
			if err != nil {
				t.Errorf("Failed to parse query: %v", err)
				return
			}

			result, err := ExecuteQuery(graph, query)
			if err != nil {
				t.Errorf("Failed to execute query: %v", err)
				return
			}

			if !reflect.DeepEqual(result.Variables, tt.expectedVars) {
				t.Errorf("Expected variables %v, got %v", tt.expectedVars, result.Variables)
			}

			if len(result.Bindings) != tt.expectedRows {
				t.Errorf("Expected %d result rows, got %d", tt.expectedRows, len(result.Bindings))
			}

			if tt.expectedEmail != "" {
				found := false
				for _, binding := range result.Bindings {
					if email, ok := binding["email"].(Literal); ok && email.Value == tt.expectedEmail {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected to find email %s in results", tt.expectedEmail)
				}
			}
		})
	}
}

func TestMatchPatternWithBinding(t *testing.T) {
	// Sample triple and pattern
	triple := Triple{
		Subject:   IRIRef{IRI: "http://example.org/alice"},
		Predicate: IRIRef{IRI: "http://xmlns.com/foaf/0.1/name"},
		Object:    Literal{Value: "Alice"},
	}

	tests := []struct {
		name        string
		pattern     TriplePattern
		binding     map[string]Term
		shouldMatch bool
		newBindings map[string]Term
	}{
		{
			name: "All variables - empty binding",
			pattern: TriplePattern{
				Subject:   Variable{Name: "s"},
				Predicate: Variable{Name: "p"},
				Object:    Variable{Name: "o"},
			},
			binding:     map[string]Term{},
			shouldMatch: true,
			newBindings: map[string]Term{
				"s": IRIRef{IRI: "http://example.org/alice"},
				"p": IRIRef{IRI: "http://xmlns.com/foaf/0.1/name"},
				"o": Literal{Value: "Alice"},
			},
		},
		{
			name: "Specific subject, variable predicate and object",
			pattern: TriplePattern{
				Subject:   IRIRefTerm{IRI: "http://example.org/alice"},
				Predicate: Variable{Name: "p"},
				Object:    Variable{Name: "o"},
			},
			binding:     map[string]Term{},
			shouldMatch: true,
			newBindings: map[string]Term{
				"p": IRIRef{IRI: "http://xmlns.com/foaf/0.1/name"},
				"o": Literal{Value: "Alice"},
			},
		},
		{
			name: "Non-matching subject",
			pattern: TriplePattern{
				Subject:   IRIRefTerm{IRI: "http://example.org/bob"},
				Predicate: Variable{Name: "p"},
				Object:    Variable{Name: "o"},
			},
			binding:     map[string]Term{},
			shouldMatch: false,
		},
		{
			name: "Variable with existing binding",
			pattern: TriplePattern{
				Subject:   Variable{Name: "s"},
				Predicate: Variable{Name: "p"},
				Object:    Variable{Name: "o"},
			},
			binding: map[string]Term{
				"s": IRIRef{IRI: "http://example.org/alice"},
			},
			shouldMatch: true,
			newBindings: map[string]Term{
				"s": IRIRef{IRI: "http://example.org/alice"},
				"p": IRIRef{IRI: "http://xmlns.com/foaf/0.1/name"},
				"o": Literal{Value: "Alice"},
			},
		},
		{
			name: "Variable with non-matching binding",
			pattern: TriplePattern{
				Subject:   Variable{Name: "s"},
				Predicate: Variable{Name: "p"},
				Object:    Variable{Name: "o"},
			},
			binding: map[string]Term{
				"s": IRIRef{IRI: "http://example.org/bob"},
			},
			shouldMatch: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			newBinding, matches := matchPatternWithBinding(tt.pattern, triple, tt.binding)

			if matches != tt.shouldMatch {
				t.Errorf("Expected match = %v, got %v", tt.shouldMatch, matches)
			}

			if !matches {
				return
			}

			// Check that all expected bindings are present
			for varName, expectedTerm := range tt.newBindings {
				actualTerm, exists := newBinding[varName]
				if !exists {
					t.Errorf("Expected binding for %s, but not found", varName)
					continue
				}

				if actualTerm.String() != expectedTerm.String() {
					t.Errorf("Expected %s = %s, got %s", varName, expectedTerm.String(), actualTerm.String())
				}
			}

			// Check that no unexpected bindings are present
			for varName := range newBinding {
				if _, exists := tt.newBindings[varName]; !exists {
					t.Errorf("Unexpected binding for %s", varName)
				}
			}
		})
	}
}
