package n3

// Store represents an N3 store that can contain and query RDF triples
type Store interface {
	// CountQuads counts the number of quads matching the given pattern
	CountQuads(subject, predicate, object, graph interface{}) int
	// GetObjects returns all objects matching the given subject and predicate
	GetObjects(subject, predicate, graph interface{}) []Term
	// AddQuad adds a quad to the store
	AddQuad(quad Quad)
}

// Term represents an RDF term (subject, predicate, object, or graph)
type Term interface {
	// Value returns the string value of the term
	Value() string
}

// BasicTerm is a simple implementation of the Term interface
type BasicTerm struct {
	value string
}

// Value implements Term.Value
func (t *BasicTerm) Value() string {
	return t.value
}

// BasicStore is a simple implementation of the Store interface
type BasicStore struct {
	quads []Quad
}

// Quad represents an RDF quad (subject, predicate, object, graph)
type Quad struct {
	Subject   Term
	Predicate Term
	Object    Term
	Graph     Term
}

// NewBasicStore creates a new BasicStore
func NewBasicStore() *BasicStore {
	return &BasicStore{
		quads: make([]Quad, 0),
	}
}

// AddQuad adds a quad to the store
func (s *BasicStore) AddQuad(quad Quad) {
	s.quads = append(s.quads, quad)
}

// CountQuads implements Store.CountQuads
func (s *BasicStore) CountQuads(subject, predicate, object, graph interface{}) int {
	count := 0
	for _, quad := range s.quads {
		if matches(quad.Subject, subject) &&
			matches(quad.Predicate, predicate) &&
			matches(quad.Object, object) &&
			matches(quad.Graph, graph) {
			count++
		}
	}
	return count
}

// GetObjects implements Store.GetObjects
func (s *BasicStore) GetObjects(subject, predicate, graph interface{}) []Term {
	var objects []Term
	for _, quad := range s.quads {
		if matches(quad.Subject, subject) &&
			matches(quad.Predicate, predicate) &&
			matches(quad.Graph, graph) {
			objects = append(objects, quad.Object)
		}
	}
	return objects
}

// matches checks if a term matches a pattern
func matches(term Term, pattern interface{}) bool {
	if pattern == nil {
		return true
	}
	if p, ok := pattern.(Term); ok {
		return term.Value() == p.Value()
	}
	return false
}
