package rdf

import (
	"testing"
)

func TestTripleString(t *testing.T) {
	tests := []struct {
		name     string
		triple   Triple
		expected string
	}{
		{
			name: "IRIRef Subject, Predicate, and Object",
			triple: Triple{
				Subject:   IRIRef{IRI: "http://example.org/subject"},
				Predicate: IRIRef{IRI: "http://example.org/predicate"},
				Object:    IRIRef{IRI: "http://example.org/object"},
			},
			expected: "<http://example.org/subject> <http://example.org/predicate> <http://example.org/object> .",
		},
		{
			name: "BNode Subject",
			triple: Triple{
				Subject:   BNode{ID: "b1"},
				Predicate: IRIRef{IRI: "http://example.org/predicate"},
				Object:    IRIRef{IRI: "http://example.org/object"},
			},
			expected: "_:b1 <http://example.org/predicate> <http://example.org/object> .",
		},
		{
			name: "Literal Object",
			triple: Triple{
				Subject:   IRIRef{IRI: "http://example.org/subject"},
				Predicate: IRIRef{IRI: "http://example.org/predicate"},
				Object:    Literal{Value: "value"},
			},
			expected: `<http://example.org/subject> <http://example.org/predicate> "value" .`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.triple.String()
			if result != tt.expected {
				t.Errorf("Triple.String() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestIRIRefString(t *testing.T) {
	tests := []struct {
		name     string
		iri      IRIRef
		expected string
	}{
		{
			name:     "Simple IRI",
			iri:      IRIRef{IRI: "http://example.org/resource"},
			expected: "<http://example.org/resource>",
		},
		{
			name:     "Empty IRI",
			iri:      IRIRef{IRI: ""},
			expected: "<>",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.iri.String()
			if result != tt.expected {
				t.Errorf("IRIRef.String() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestBNodeString(t *testing.T) {
	tests := []struct {
		name     string
		bnode    BNode
		expected string
	}{
		{
			name:     "Simple BNode",
			bnode:    BNode{ID: "b1"},
			expected: "_:b1",
		},
		{
			name:     "Empty BNode ID",
			bnode:    BNode{ID: ""},
			expected: "_:",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.bnode.String()
			if result != tt.expected {
				t.Errorf("BNode.String() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestLiteralString(t *testing.T) {
	tests := []struct {
		name     string
		literal  Literal
		expected string
	}{
		{
			name:     "Simple Literal",
			literal:  Literal{Value: "test"},
			expected: `"test"`,
		},
		{
			name:     "Literal with Language",
			literal:  Literal{Value: "test", Language: "en"},
			expected: `"test"@en`,
		},
		{
			name:     "Literal with Datatype",
			literal:  Literal{Value: "42", Datatype: IRIRef{IRI: "http://www.w3.org/2001/XMLSchema#integer"}},
			expected: `"42"^^<http://www.w3.org/2001/XMLSchema#integer>`,
		},
		{
			name:     "Literal with Escapes",
			literal:  Literal{Value: "line1\nline2\t\"quoted\""},
			expected: `"line1\nline2\t\"quoted\""`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.literal.String()
			if result != tt.expected {
				t.Errorf("Literal.String() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestGraphAddTriple(t *testing.T) {
	g := Graph{}
	subject := IRIRef{IRI: "http://example.org/subject"}
	predicate := IRIRef{IRI: "http://example.org/predicate"}
	object := IRIRef{IRI: "http://example.org/object"}

	g.AddTriple(subject, predicate, object)

	if len(g.Triples) != 1 {
		t.Errorf("Graph.AddTriple() did not add a triple to the graph, got %d triples", len(g.Triples))
	}

	if g.Triples[0].Subject != subject || g.Triples[0].Predicate != predicate || g.Triples[0].Object != object {
		t.Errorf("Graph.AddTriple() added incorrect triple")
	}
}

func TestGraphString(t *testing.T) {
	g := Graph{}
	g.AddTriple(
		IRIRef{IRI: "http://example.org/subject1"},
		IRIRef{IRI: "http://example.org/predicate1"},
		IRIRef{IRI: "http://example.org/object1"},
	)
	g.AddTriple(
		IRIRef{IRI: "http://example.org/subject2"},
		IRIRef{IRI: "http://example.org/predicate2"},
		Literal{Value: "value2"},
	)

	expected := "<http://example.org/subject1> <http://example.org/predicate1> <http://example.org/object1> .\n" +
		"<http://example.org/subject2> <http://example.org/predicate2> \"value2\" .\n"

	result := g.String()
	if result != expected {
		t.Errorf("Graph.String() = %v, want %v", result, expected)
	}
} 