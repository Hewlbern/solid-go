package rdf

import (
	"bytes"
	"strings"
	"testing"
)

func TestTurtleParserParsePrefix(t *testing.T) {
	tests := []struct {
		name        string
		prefixLine  string
		expectError bool
		prefix      string
		uri         string
	}{
		{
			name:        "Valid prefix",
			prefixLine:  "@prefix ex: <http://example.org/> .",
			expectError: false,
			prefix:      "ex",
			uri:         "http://example.org/",
		},
		{
			name:        "Valid prefix with whitespace",
			prefixLine:  "@prefix  test:  <http://test.org/>  .",
			expectError: false,
			prefix:      "test",
			uri:         "http://test.org/",
		},
		{
			name:        "Invalid prefix format",
			prefixLine:  "@prefix ex <http://example.org/> .",
			expectError: true,
		},
		{
			name:        "Missing period",
			prefixLine:  "@prefix ex: <http://example.org/>",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewTurtleParser()
			err := parser.parsePrefix(tt.prefixLine)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}

				if uri, ok := parser.prefixes[tt.prefix]; !ok || uri != tt.uri {
					t.Errorf("Expected prefix %s to be %s, got %s", tt.prefix, tt.uri, uri)
				}
			}
		})
	}
}

func TestTurtleParserParseTerm(t *testing.T) {
	tests := []struct {
		name         string
		termString   string
		prefixes     map[string]string
		expectError  bool
		expectedTerm Term
	}{
		{
			name:         "IRI",
			termString:   "<http://example.org/resource>",
			expectError:  false,
			expectedTerm: IRIRef{IRI: "http://example.org/resource"},
		},
		{
			name:         "Blank node",
			termString:   "_:b1",
			expectError:  false,
			expectedTerm: BNode{ID: "b1"},
		},
		{
			name:         "Simple literal",
			termString:   "\"test\"",
			expectError:  false,
			expectedTerm: Literal{Value: "test"},
		},
		{
			name:         "Literal with language",
			termString:   "\"test\"@en",
			expectError:  false,
			expectedTerm: Literal{Value: "test", Language: "en"},
		},
		{
			name:         "Literal with datatype",
			termString:   "\"42\"^^<http://www.w3.org/2001/XMLSchema#integer>",
			expectError:  false,
			expectedTerm: Literal{Value: "42", Datatype: IRIRef{IRI: "http://www.w3.org/2001/XMLSchema#integer"}},
		},
		{
			name:         "Prefixed IRI with known prefix",
			termString:   "ex:resource",
			prefixes:     map[string]string{"ex": "http://example.org/"},
			expectError:  false,
			expectedTerm: IRIRef{IRI: "http://example.org/resource"},
		},
		{
			name:        "Prefixed IRI with unknown prefix",
			termString:  "unknown:resource",
			expectError: true,
		},
		{
			name:        "Invalid term format",
			termString:  "invalid",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewTurtleParser()
			if tt.prefixes != nil {
				parser.prefixes = tt.prefixes
			}

			term, err := parser.parseTerm(tt.termString)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}

				// Check type
				if term.Type() != tt.expectedTerm.Type() {
					t.Errorf("Expected term type %v, got %v", tt.expectedTerm.Type(), term.Type())
				}

				// Check string representation
				if term.String() != tt.expectedTerm.String() {
					t.Errorf("Expected term string %v, got %v", tt.expectedTerm.String(), term.String())
				}
			}
		})
	}
}

func TestTurtleParserParse(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectError   bool
		expectedCount int
	}{
		{
			name:          "Simple triple",
			input:         "<http://example.org/subject> <http://example.org/predicate> <http://example.org/object> .",
			expectError:   false,
			expectedCount: 1,
		},
		{
			name: "Multiple triples",
			input: `<http://example.org/subject1> <http://example.org/predicate1> <http://example.org/object1> .
				<http://example.org/subject2> <http://example.org/predicate2> <http://example.org/object2> .`,
			expectError:   false,
			expectedCount: 2,
		},
		{
			name:          "Triple with literal",
			input:         `<http://example.org/subject> <http://example.org/predicate> "literal value" .`,
			expectError:   false,
			expectedCount: 1,
		},
		{
			name: "Triple with prefix declaration",
			input: `@prefix ex: <http://example.org/> .
				ex:subject ex:predicate ex:object .`,
			expectError:   false,
			expectedCount: 1,
		},
		{
			name: "Comment and empty lines",
			input: `# This is a comment
				
				<http://example.org/subject> <http://example.org/predicate> <http://example.org/object> .`,
			expectError:   false,
			expectedCount: 1,
		},
		{
			name:        "Invalid triple format",
			input:       `<http://example.org/subject> <http://example.org/predicate>`,
			expectError: true,
		},
		{
			name:          "Blank node",
			input:         `_:b1 <http://example.org/predicate> <http://example.org/object> .`,
			expectError:   false,
			expectedCount: 1,
		},
		{
			name: "Blank node with multiple triples",
			input: `_:b1 <http://example.org/predicate1> <http://example.org/object1> .
				_:b1 <http://example.org/predicate2> <http://example.org/object2> .`,
			expectError:   false,
			expectedCount: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewTurtleParser()
			graph, err := parser.ParseString(tt.input)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}

				if len(graph.Triples) != tt.expectedCount {
					t.Errorf("Expected %d triples, got %d", tt.expectedCount, len(graph.Triples))
				}

				// For blank node tests, verify the node ID is preserved
				if strings.Contains(tt.name, "Blank node") {
					for _, triple := range graph.Triples {
						if blank, ok := triple.Subject.(BNode); ok {
							if !strings.HasPrefix(blank.ID, "b") {
								t.Errorf("Expected blank node ID to start with 'b', got %s", blank.ID)
							}
						}
					}
				}
			}
		})
	}
}

func TestTurtleWriter(t *testing.T) {
	tests := []struct {
		name     string
		graph    *Graph
		prefixes map[string]string
		expected string
	}{
		{
			name:     "Empty graph",
			graph:    &Graph{},
			expected: "@prefix rdf: <http://www.w3.org/1999/02/22-rdf-syntax-ns#> .\n@prefix rdfs: <http://www.w3.org/2000/01/rdf-schema#> .\n@prefix xsd: <http://www.w3.org/2001/XMLSchema#> .\n@prefix foaf: <http://xmlns.com/foaf/0.1/> .\n\n",
		},
		{
			name: "Graph with one triple",
			graph: func() *Graph {
				g := &Graph{}
				g.AddTriple(
					IRIRef{IRI: "http://example.org/subject"},
					IRIRef{IRI: "http://example.org/predicate"},
					IRIRef{IRI: "http://example.org/object"},
				)
				return g
			}(),
			expected: "@prefix rdf: <http://www.w3.org/1999/02/22-rdf-syntax-ns#> .\n@prefix rdfs: <http://www.w3.org/2000/01/rdf-schema#> .\n@prefix xsd: <http://www.w3.org/2001/XMLSchema#> .\n@prefix foaf: <http://xmlns.com/foaf/0.1/> .\n\n<http://example.org/subject> <http://example.org/predicate> <http://example.org/object> .\n",
		},
		{
			name: "Graph with multiple triples",
			graph: func() *Graph {
				g := &Graph{}
				g.AddTriple(
					IRIRef{IRI: "http://example.org/subject1"},
					IRIRef{IRI: "http://example.org/predicate1"},
					IRIRef{IRI: "http://example.org/object1"},
				)
				g.AddTriple(
					IRIRef{IRI: "http://example.org/subject2"},
					IRIRef{IRI: "http://example.org/predicate2"},
					Literal{Value: "literal value"},
				)
				return g
			}(),
			expected: "@prefix rdf: <http://www.w3.org/1999/02/22-rdf-syntax-ns#> .\n@prefix rdfs: <http://www.w3.org/2000/01/rdf-schema#> .\n@prefix xsd: <http://www.w3.org/2001/XMLSchema#> .\n@prefix foaf: <http://xmlns.com/foaf/0.1/> .\n\n<http://example.org/subject1> <http://example.org/predicate1> <http://example.org/object1> .\n<http://example.org/subject2> <http://example.org/predicate2> \"literal value\" .\n",
		},
		{
			name: "Graph with custom prefix",
			graph: func() *Graph {
				g := &Graph{}
				g.AddTriple(
					IRIRef{IRI: "http://example.org/subject"},
					IRIRef{IRI: "http://example.org/predicate"},
					IRIRef{IRI: "http://example.org/object"},
				)
				return g
			}(),
			prefixes: map[string]string{"ex": "http://example.org/"},
			expected: "@prefix rdf: <http://www.w3.org/1999/02/22-rdf-syntax-ns#> .\n@prefix rdfs: <http://www.w3.org/2000/01/rdf-schema#> .\n@prefix xsd: <http://www.w3.org/2001/XMLSchema#> .\n@prefix foaf: <http://xmlns.com/foaf/0.1/> .\n@prefix ex: <http://example.org/> .\n\n<http://example.org/subject> <http://example.org/predicate> <http://example.org/object> .\n",
		},
		{
			name: "Graph with blank nodes",
			graph: func() *Graph {
				g := &Graph{}
				g.AddTriple(
					BNode{ID: "b1"},
					IRIRef{IRI: "http://example.org/predicate1"},
					IRIRef{IRI: "http://example.org/object1"},
				)
				g.AddTriple(
					BNode{ID: "b1"},
					IRIRef{IRI: "http://example.org/predicate2"},
					Literal{Value: "literal value"},
				)
				g.AddTriple(
					IRIRef{IRI: "http://example.org/subject"},
					IRIRef{IRI: "http://example.org/predicate3"},
					BNode{ID: "b2"},
				)
				return g
			}(),
			expected: "@prefix rdf: <http://www.w3.org/1999/02/22-rdf-syntax-ns#> .\n@prefix rdfs: <http://www.w3.org/2000/01/rdf-schema#> .\n@prefix xsd: <http://www.w3.org/2001/XMLSchema#> .\n@prefix foaf: <http://xmlns.com/foaf/0.1/> .\n\n_:b1 <http://example.org/predicate1> <http://example.org/object1> .\n_:b1 <http://example.org/predicate2> \"literal value\" .\n<http://example.org/subject> <http://example.org/predicate3> _:b2 .\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			writer := NewTurtleWriter()

			// Add custom prefixes
			if tt.prefixes != nil {
				for prefix, uri := range tt.prefixes {
					writer.AddPrefix(prefix, uri)
				}
			}

			err := writer.Write(tt.graph, &buf)
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			// Compare output, ignoring prefix order which may vary
			output := buf.String()
			outputParts := strings.Split(output, "\n\n")
			expectedParts := strings.Split(tt.expected, "\n\n")

			if len(outputParts) != 2 || len(expectedParts) != 2 {
				t.Errorf("Output format incorrect")
				return
			}

			// Check prefixes - order doesn't matter
			outputPrefixes := strings.Split(outputParts[0], "\n")
			expectedPrefixes := strings.Split(expectedParts[0], "\n")
			if len(outputPrefixes) != len(expectedPrefixes) {
				t.Errorf("Expected %d prefixes, got %d", len(expectedPrefixes), len(outputPrefixes))
			}

			// Check triples
			if outputParts[1] != expectedParts[1] {
				t.Errorf("Expected triples\n%s\nGot\n%s", expectedParts[1], outputParts[1])
			}
		})
	}
}
