package server

import (
	"encoding/json"
	"io"
	"net/http"
)

// ParsingHttpHandler handles HTTP requests with content parsing
type ParsingHttpHandler interface {
	HttpHandler
	// ParseRequest parses the request body into the given value
	ParseRequest(r *http.Request, v interface{}) error
	// WriteResponse writes the response with the given value
	WriteResponse(w http.ResponseWriter, v interface{}) error
}

// BaseParsingHandler provides a base implementation of ParsingHttpHandler
type BaseParsingHandler struct {
	contentType string
}

// NewBaseParsingHandler creates a new BaseParsingHandler
func NewBaseParsingHandler(contentType string) *BaseParsingHandler {
	return &BaseParsingHandler{
		contentType: contentType,
	}
}

// Handle implements HttpHandler
func (h *BaseParsingHandler) Handle(w http.ResponseWriter, r *http.Request) error {
	// Set content type
	w.Header().Set("Content-Type", h.contentType)

	// Parse request if needed
	if r.Body != nil && r.Method != http.MethodGet {
		var v interface{}
		if err := h.ParseRequest(r, &v); err != nil {
			return err
		}
		// Process the parsed value
		// TODO: Implement processing logic
	}

	// Write response
	// TODO: Implement response writing logic
	return nil
}

// ParseRequest implements ParsingHttpHandler
func (h *BaseParsingHandler) ParseRequest(r *http.Request, v interface{}) error {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return err
	}
	return json.Unmarshal(body, v)
}

// WriteResponse implements ParsingHttpHandler
func (h *BaseParsingHandler) WriteResponse(w http.ResponseWriter, v interface{}) error {
	data, err := json.Marshal(v)
	if err != nil {
		return err
	}
	_, err = w.Write(data)
	return err
}
