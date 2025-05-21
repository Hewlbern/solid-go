package server

// Representation represents a resource representation
type Representation struct {
	Data     []byte
	Metadata map[string]string
}

// Operation represents an HTTP operation
type Operation struct {
	Method      string
	Target      string
	ContentType string
	Headers     map[string][]string
	Body        *Representation
}

// NewOperation creates a new Operation
func NewOperation(method, target, contentType string, headers map[string][]string) Operation {
	return Operation{
		Method:      method,
		Target:      target,
		ContentType: contentType,
		Headers:     headers,
	}
}
