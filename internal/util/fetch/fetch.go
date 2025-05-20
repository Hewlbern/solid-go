package fetch

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
)

// Representation represents a fetched dataset
type Representation struct {
	Data io.Reader
}

// FetchError represents an error that occurred during fetching
type FetchError struct {
	StatusCode int
	Message    string
}

func (e *FetchError) Error() string {
	return fmt.Sprintf("fetch error: %s (status code: %d)", e.Message, e.StatusCode)
}

// FetchDataset fetches a dataset from the given URL
func FetchDataset(url string) (*Representation, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, &FetchError{
			StatusCode: 0,
			Message:    fmt.Sprintf("failed to fetch dataset: %v", err),
		}
	}
	defer resp.Body.Close()

	// Check for successful status code
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, &FetchError{
			StatusCode: resp.StatusCode,
			Message:    fmt.Sprintf("unexpected status code: %s", resp.Status),
		}
	}

	// Check content type
	contentType := resp.Header.Get("Content-Type")
	if contentType == "" {
		return nil, &FetchError{
			StatusCode: resp.StatusCode,
			Message:    "missing Content-Type header",
		}
	}

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, &FetchError{
			StatusCode: resp.StatusCode,
			Message:    fmt.Sprintf("failed to read response body: %v", err),
		}
	}

	return &Representation{
		Data: bytes.NewReader(body),
	}, nil
}
