package ldp

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"path"
	"strings"
	"time"

	"github.com/yourusername/solid-go/internal/events"
	"github.com/yourusername/solid-go/internal/storage"
)

// ResourceType represents the type of a LDP resource.
type ResourceType string

// LDP resource types
const (
	TypeResource  ResourceType = "Resource"
	TypeContainer ResourceType = "Container"
	TypeRDFSource ResourceType = "RDFSource"
)

// Handler handles LDP requests.
type Handler struct {
	// storageFactory is the factory for creating storage instances
	storageFactory storage.Factory
	// eventDispatcher is the dispatcher for LDP events
	eventDispatcher events.Dispatcher
	// defaultStorage is the default storage instance
	defaultStorage storage.Storage
}

// NewHandler creates a new LDP handler.
func NewHandler(storageFactory storage.Factory, eventDispatcher events.Dispatcher) (*Handler, error) {
	// Create the default storage (file-based)
	defaultStorage, err := storageFactory.CreateStorage("file")
	if err != nil {
		return nil, fmt.Errorf("failed to create storage for LDP: %w", err)
	}

	return &Handler{
		storageFactory:  storageFactory,
		eventDispatcher: eventDispatcher,
		defaultStorage:  defaultStorage,
	}, nil
}

// ServeHTTP implements http.Handler.
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Parse the path
	resourcePath := r.URL.Path
	// Normalize the path to ensure it starts with a slash
	if !strings.HasPrefix(resourcePath, "/") {
		resourcePath = "/" + resourcePath
	}

	// Extract the resource path from the URL
	if resourcePath == "/" {
		// Root path - handle specially
		h.handleRoot(w, r)
		return
	}

	// Handle based on the HTTP method
	switch r.Method {
	case http.MethodGet:
		h.handleGet(w, r, resourcePath)
	case http.MethodHead:
		h.handleHead(w, r, resourcePath)
	case http.MethodOptions:
		h.handleOptions(w, r, resourcePath)
	case http.MethodPost:
		h.handlePost(w, r, resourcePath)
	case http.MethodPut:
		h.handlePut(w, r, resourcePath)
	case http.MethodPatch:
		h.handlePatch(w, r, resourcePath)
	case http.MethodDelete:
		h.handleDelete(w, r, resourcePath)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleRoot handles requests to the root path.
func (h *Handler) handleRoot(w http.ResponseWriter, r *http.Request) {
	// For now, just return basic information about the server
	w.Header().Set("Content-Type", "text/turtle")
	w.Write([]byte(`@prefix solid: <http://www.w3.org/ns/solid/terms#> .
@prefix ldp: <http://www.w3.org/ns/ldp#> .

<> a solid:Container, ldp:BasicContainer ;
   solid:status "Welcome to SolidGo!" .
`))
}

// handleGet handles GET requests.
func (h *Handler) handleGet(w http.ResponseWriter, r *http.Request, resourcePath string) {
	// Check if the resource exists
	exists, err := h.defaultStorage.ResourceExists(resourcePath)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error checking resource: %v", err), http.StatusInternalServerError)
		return
	}

	if !exists {
		http.Error(w, "Resource not found", http.StatusNotFound)
		return
	}

	// Check if it's a container
	metadata, err := h.defaultStorage.GetResourceMetadata(resourcePath)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error retrieving metadata: %v", err), http.StatusInternalServerError)
		return
	}

	if metadata.IsContainer {
		h.handleGetContainer(w, r, resourcePath)
		return
	}

	// It's a regular resource
	h.handleGetResource(w, r, resourcePath)
}

// handleGetResource retrieves a regular resource.
func (h *Handler) handleGetResource(w http.ResponseWriter, r *http.Request, resourcePath string) {
	// Get the resource data
	data, contentType, err := h.defaultStorage.GetResource(resourcePath)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error retrieving resource: %v", err), http.StatusInternalServerError)
		return
	}

	// Set content type header
	w.Header().Set("Content-Type", contentType)
	
	// Add LDP headers
	h.addLDPResourceHeaders(w, TypeResource)

	// Return the data
	w.Write(data)
}

// handleGetContainer retrieves a container.
func (h *Handler) handleGetContainer(w http.ResponseWriter, r *http.Request, containerPath string) {
	// Set content type to text/turtle by default for containers
	w.Header().Set("Content-Type", "text/turtle")

	// Add LDP headers
	h.addLDPResourceHeaders(w, TypeContainer)

	// List the container
	resources, err := h.defaultStorage.ListContainer(containerPath)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error listing container: %v", err), http.StatusInternalServerError)
		return
	}

	// Build Turtle response
	var builder strings.Builder
	builder.WriteString("@prefix ldp: <http://www.w3.org/ns/ldp#> .\n")
	builder.WriteString("@prefix dc: <http://purl.org/dc/terms/> .\n")
	builder.WriteString("\n")
	
	// Container information
	builder.WriteString("<> a ldp:BasicContainer, ldp:Container ;\n")
	
	// Add contains triples for each resource
	if len(resources) > 0 {
		for i, resource := range resources {
			resourceName := path.Base(resource.Path)
			
			if i < len(resources)-1 {
				builder.WriteString(fmt.Sprintf("    ldp:contains <%s> ;\n", resourceName))
			} else {
				builder.WriteString(fmt.Sprintf("    ldp:contains <%s> .\n", resourceName))
			}
		}
	} else {
		builder.WriteString("    dc:title \"Empty Container\" .\n")
	}
	
	// Add information about each resource
	for _, resource := range resources {
		resourceName := path.Base(resource.Path)
		
		builder.WriteString(fmt.Sprintf("\n<%s> a ", resourceName))
		if resource.IsContainer {
			builder.WriteString("ldp:BasicContainer, ldp:Container ;\n")
		} else {
			builder.WriteString("ldp:Resource ;\n")
		}
		
		builder.WriteString(fmt.Sprintf("    dc:title \"%s\" ;\n", resourceName))
		builder.WriteString(fmt.Sprintf("    dc:modified \"%s\" ;\n", resource.LastModified))
		builder.WriteString(fmt.Sprintf("    dc:contentType \"%s\" .\n", resource.ContentType))
	}
	
	w.Write([]byte(builder.String()))
}

// handleHead handles HEAD requests.
func (h *Handler) handleHead(w http.ResponseWriter, r *http.Request, resourcePath string) {
	// Check if the resource exists
	exists, err := h.defaultStorage.ResourceExists(resourcePath)
	if err != nil {
		http.Error(w, "", http.StatusInternalServerError)
		return
	}

	if !exists {
		http.Error(w, "", http.StatusNotFound)
		return
	}

	// Get metadata
	metadata, err := h.defaultStorage.GetResourceMetadata(resourcePath)
	if err != nil {
		http.Error(w, "", http.StatusInternalServerError)
		return
	}

	// Set content type header
	w.Header().Set("Content-Type", metadata.ContentType)
	
	// Add LDP headers
	if metadata.IsContainer {
		h.addLDPResourceHeaders(w, TypeContainer)
	} else {
		h.addLDPResourceHeaders(w, TypeResource)
	}
}

// handleOptions handles OPTIONS requests.
func (h *Handler) handleOptions(w http.ResponseWriter, r *http.Request, resourcePath string) {
	// Check if the resource exists
	exists, err := h.defaultStorage.ResourceExists(resourcePath)
	if err != nil {
		http.Error(w, "", http.StatusInternalServerError)
		return
	}

	// Get the resource type
	resourceType := TypeResource
	if exists {
		metadata, err := h.defaultStorage.GetResourceMetadata(resourcePath)
		if err != nil {
			http.Error(w, "", http.StatusInternalServerError)
			return
		}
		
		if metadata.IsContainer {
			resourceType = TypeContainer
		}
	}

	// Set CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, HEAD, OPTIONS, POST, PUT, PATCH, DELETE")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Link, Accept, Accept-Encoding, Accept-Language, Authorization")
	
	// Add LDP headers
	h.addLDPResourceHeaders(w, resourceType)
	
	// Add allowed methods
	if resourceType == TypeContainer {
		w.Header().Set("Allow", "GET, HEAD, OPTIONS, POST, PUT, PATCH, DELETE")
	} else {
		w.Header().Set("Allow", "GET, HEAD, OPTIONS, PUT, PATCH, DELETE")
	}
}

// handlePost handles POST requests (creating resources in containers).
func (h *Handler) handlePost(w http.ResponseWriter, r *http.Request, containerPath string) {
	// Check if the container exists
	exists, err := h.defaultStorage.ResourceExists(containerPath)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error checking container: %v", err), http.StatusInternalServerError)
		return
	}

	if !exists {
		// Try to create the container
		err = h.defaultStorage.CreateContainer(containerPath)
		if err != nil {
			http.Error(w, fmt.Sprintf("Container not found and could not be created: %v", err), http.StatusNotFound)
			return
		}
	}

	// Check if it's a container
	metadata, err := h.defaultStorage.GetResourceMetadata(containerPath)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error retrieving metadata: %v", err), http.StatusInternalServerError)
		return
	}

	if !metadata.IsContainer {
		http.Error(w, "Cannot POST to a non-container resource", http.StatusMethodNotAllowed)
		return
	}

	// Get the Slug header to determine the resource name
	slug := r.Header.Get("Slug")
	if slug == "" {
		// Generate a unique name
		slug = fmt.Sprintf("resource-%d", time.Now().UnixNano())
	}

	// Ensure the slug doesn't contain invalid characters
	slug = sanitizeSlug(slug)

	// Build the resource path
	resourcePath := path.Join(containerPath, slug)

	// Check if a resource with this name already exists
	exists, err = h.defaultStorage.ResourceExists(resourcePath)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error checking resource: %v", err), http.StatusInternalServerError)
		return
	}

	if exists {
		http.Error(w, "Resource already exists", http.StatusConflict)
		return
	}

	// Get the content type
	contentType := r.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "text/plain"
	}

	// Read the request body
	data, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error reading request body: %v", err), http.StatusInternalServerError)
		return
	}

	// Store the resource
	err = h.defaultStorage.StoreResource(resourcePath, data, contentType)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error storing resource: %v", err), http.StatusInternalServerError)
		return
	}

	// Publish an event
	if h.eventDispatcher != nil {
		event := events.NewEvent(events.TypeResourceCreated, resourcePath, getAgentID(r))
		h.eventDispatcher.Dispatch(event)
	}

	// Set the Location header to the new resource
	w.Header().Set("Location", resourcePath)
	
	// Set status code to 201 Created
	w.WriteHeader(http.StatusCreated)
}

// handlePut handles PUT requests (creating or replacing resources).
func (h *Handler) handlePut(w http.ResponseWriter, r *http.Request, resourcePath string) {
	// Check if the resource exists
	exists, err := h.defaultStorage.ResourceExists(resourcePath)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error checking resource: %v", err), http.StatusInternalServerError)
		return
	}

	// If the resource is a container, the PUT operation is not allowed
	if exists {
		metadata, err := h.defaultStorage.GetResourceMetadata(resourcePath)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error retrieving metadata: %v", err), http.StatusInternalServerError)
			return
		}

		if metadata.IsContainer {
			http.Error(w, "Cannot PUT to a container", http.StatusMethodNotAllowed)
			return
		}
	}

	// Get the content type
	contentType := r.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "text/plain"
	}

	// Read the request body
	data, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error reading request body: %v", err), http.StatusInternalServerError)
		return
	}

	// Store the resource
	err = h.defaultStorage.StoreResource(resourcePath, data, contentType)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error storing resource: %v", err), http.StatusInternalServerError)
		return
	}

	// Publish an event
	if h.eventDispatcher != nil {
		if exists {
			event := events.NewEvent(events.TypeResourceUpdated, resourcePath, getAgentID(r))
			h.eventDispatcher.Dispatch(event)
		} else {
			event := events.NewEvent(events.TypeResourceCreated, resourcePath, getAgentID(r))
			h.eventDispatcher.Dispatch(event)
		}
	}

	// Set status code based on whether the resource existed
	if exists {
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusCreated)
	}
}

// handlePatch handles PATCH requests (partial updates to resources).
func (h *Handler) handlePatch(w http.ResponseWriter, r *http.Request, resourcePath string) {
	// Check if the resource exists
	exists, err := h.defaultStorage.ResourceExists(resourcePath)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error checking resource: %v", err), http.StatusInternalServerError)
		return
	}

	if !exists {
		http.Error(w, "Resource not found", http.StatusNotFound)
		return
	}

	// Check if it's a container
	metadata, err := h.defaultStorage.GetResourceMetadata(resourcePath)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error retrieving metadata: %v", err), http.StatusInternalServerError)
		return
	}

	if metadata.IsContainer {
		http.Error(w, "Cannot PATCH a container", http.StatusMethodNotAllowed)
		return
	}

	// Get the content type of the patch
	contentType := r.Header.Get("Content-Type")
	
	// Currently only support text/n3 (N3 Patch)
	if contentType != "text/n3" {
		http.Error(w, "Unsupported patch content type, only text/n3 is supported", http.StatusUnsupportedMediaType)
		return
	}

	// Read the patch data
	patchData, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error reading patch data: %v", err), http.StatusInternalServerError)
		return
	}

	// For now, we'll just implement a simple patch that replaces the resource
	// In a real implementation, you'd parse the N3 patch and apply it to the resource
	
	// Get the current resource
	data, originalContentType, err := h.defaultStorage.GetResource(resourcePath)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error retrieving resource: %v", err), http.StatusInternalServerError)
		return
	}

	// Apply the patch (simplified implementation)
	// In a real implementation, you'd parse the N3 patch and apply it to the resource
	// For now, we'll just append the patch data
	newData := append(data, patchData...)

	// Store the updated resource
	err = h.defaultStorage.StoreResource(resourcePath, newData, originalContentType)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error storing resource: %v", err), http.StatusInternalServerError)
		return
	}

	// Publish an event
	if h.eventDispatcher != nil {
		event := events.NewEvent(events.TypeResourceUpdated, resourcePath, getAgentID(r))
		h.eventDispatcher.Dispatch(event)
	}

	w.WriteHeader(http.StatusOK)
}

// handleDelete handles DELETE requests.
func (h *Handler) handleDelete(w http.ResponseWriter, r *http.Request, resourcePath string) {
	// Check if the resource exists
	exists, err := h.defaultStorage.ResourceExists(resourcePath)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error checking resource: %v", err), http.StatusInternalServerError)
		return
	}

	if !exists {
		http.Error(w, "Resource not found", http.StatusNotFound)
		return
	}

	// Check if it's a container
	metadata, err := h.defaultStorage.GetResourceMetadata(resourcePath)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error retrieving metadata: %v", err), http.StatusInternalServerError)
		return
	}

	// If it's a container, make sure it's empty
	if metadata.IsContainer {
		resources, err := h.defaultStorage.ListContainer(resourcePath)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error listing container: %v", err), http.StatusInternalServerError)
			return
		}

		if len(resources) > 0 {
			http.Error(w, "Cannot delete non-empty container", http.StatusConflict)
			return
		}
	}

	// Delete the resource
	err = h.defaultStorage.DeleteResource(resourcePath)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error deleting resource: %v", err), http.StatusInternalServerError)
		return
	}

	// Publish an event
	if h.eventDispatcher != nil {
		eventType := events.TypeResourceDeleted
		if metadata.IsContainer {
			eventType = events.TypeContainerDeleted
		}
		
		event := events.NewEvent(eventType, resourcePath, getAgentID(r))
		h.eventDispatcher.Dispatch(event)
	}

	w.WriteHeader(http.StatusNoContent)
}

// addLDPResourceHeaders adds LDP-specific headers to the response.
func (h *Handler) addLDPResourceHeaders(w http.ResponseWriter, resourceType ResourceType) {
	// Add Link header with LDP resource type
	switch resourceType {
	case TypeContainer:
		w.Header().Add("Link", "<http://www.w3.org/ns/ldp#Container>; rel=\"type\"")
		w.Header().Add("Link", "<http://www.w3.org/ns/ldp#BasicContainer>; rel=\"type\"")
	case TypeResource:
		w.Header().Add("Link", "<http://www.w3.org/ns/ldp#Resource>; rel=\"type\"")
	case TypeRDFSource:
		w.Header().Add("Link", "<http://www.w3.org/ns/ldp#RDFSource>; rel=\"type\"")
	}
	
	// Add other common LDP headers
	w.Header().Add("Link", "<http://www.w3.org/ns/ldp#Resource>; rel=\"type\"")
	w.Header().Set("Accept-Post", "text/turtle, application/ld+json")
}

// sanitizeSlug sanitizes a slug to ensure it's a valid filename.
func sanitizeSlug(slug string) string {
	// Replace slashes with underscores
	slug = strings.ReplaceAll(slug, "/", "_")
	// Replace any other problematic characters
	slug = strings.ReplaceAll(slug, ":", "_")
	slug = strings.ReplaceAll(slug, "*", "_")
	slug = strings.ReplaceAll(slug, "?", "_")
	slug = strings.ReplaceAll(slug, "\"", "_")
	slug = strings.ReplaceAll(slug, "<", "_")
	slug = strings.ReplaceAll(slug, ">", "_")
	slug = strings.ReplaceAll(slug, "|", "_")
	
	return slug
}

// getAgentID extracts the agent ID from the request.
func getAgentID(r *http.Request) string {
	// In a real implementation, you'd get this from the authenticated agent
	// For now, just return a placeholder
	return "anonymous"
}

// errors
var (
	ErrResourceNotFound = errors.New("resource not found")
	ErrMethodNotAllowed = errors.New("method not allowed")
) 