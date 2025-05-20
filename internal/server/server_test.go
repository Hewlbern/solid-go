package server

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

type mockStorage struct {
	data map[string][]byte
	err  error
}

func (m *mockStorage) Get(ctx context.Context, path string) ([]byte, error) {
	if m.err != nil {
		return nil, m.err
	}
	if data, exists := m.data[path]; exists {
		return data, nil
	}
	return nil, nil
}

func (m *mockStorage) Put(ctx context.Context, path string, data []byte) error {
	if m.err != nil {
		return m.err
	}
	m.data[path] = data
	return nil
}

func (m *mockStorage) Delete(ctx context.Context, path string) error {
	if m.err != nil {
		return m.err
	}
	delete(m.data, path)
	return nil
}

func (m *mockStorage) List(ctx context.Context, path string) ([]string, error) {
	if m.err != nil {
		return nil, m.err
	}
	var paths []string
	for p := range m.data {
		if p != path {
			paths = append(paths, p)
		}
	}
	return paths, nil
}

func (m *mockStorage) Exists(path string) bool {
	_, exists := m.data[path]
	return exists
}

func TestServer(t *testing.T) {
	tests := []struct {
		name           string
		storage        *mockStorage
		request        *http.Request
		expectedStatus int
		expectedBody   interface{}
	}{
		{
			name: "Data Directory Query",
			storage: &mockStorage{
				data: map[string][]byte{
					"/resource1":     []byte("data1"),
					"/container/":    []byte(""),
					"/resource2.acl": []byte("acl"),
				},
			},
			request:        httptest.NewRequest(http.MethodGet, "/data", nil),
			expectedStatus: http.StatusOK,
			expectedBody: DataDirectoryInfo{
				Path:       "/",
				Resources:  []string{"/resource1"},
				Containers: []string{"/container/"},
				ACLs:       []string{"/resource2.acl"},
			},
		},
		{
			name: "WebID Request",
			storage: &mockStorage{
				data: map[string][]byte{
					"/profile/user": []byte("profile data"),
				},
			},
			request:        httptest.NewRequest(http.MethodGet, "/profile/user", nil),
			expectedStatus: http.StatusOK,
			expectedBody:   nil, // Profile data will be in Turtle format
		},
		{
			name: "ACL Request",
			storage: &mockStorage{
				data: map[string][]byte{
					"/resource.acl": []byte("acl data"),
				},
			},
			request:        httptest.NewRequest(http.MethodGet, "/resource.acl", nil),
			expectedStatus: http.StatusOK,
			expectedBody:   nil, // ACL data will be in Turtle format
		},
		{
			name: "Container Request",
			storage: &mockStorage{
				data: map[string][]byte{
					"/container/":          []byte(""),
					"/container/resource1": []byte("data1"),
					"/container/resource2": []byte("data2"),
				},
			},
			request:        httptest.NewRequest(http.MethodGet, "/container/", nil),
			expectedStatus: http.StatusOK,
			expectedBody:   nil, // Container listing will be in specific format
		},
		{
			name: "Resource Request",
			storage: &mockStorage{
				data: map[string][]byte{
					"/resource": []byte("resource data"),
				},
			},
			request:        httptest.NewRequest(http.MethodGet, "/resource", nil),
			expectedStatus: http.StatusOK,
			expectedBody:   []byte("resource data"),
		},
		{
			name: "Method Not Allowed",
			storage: &mockStorage{
				data: map[string][]byte{},
			},
			request:        httptest.NewRequest(http.MethodPatch, "/resource", nil),
			expectedStatus: http.StatusMethodNotAllowed,
			expectedBody:   nil,
		},
		{
			name: "Resource Not Found",
			storage: &mockStorage{
				data: map[string][]byte{},
			},
			request:        httptest.NewRequest(http.MethodGet, "/nonexistent", nil),
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := NewServer(tt.storage)
			recorder := httptest.NewRecorder()

			server.ServeHTTP(recorder, tt.request)

			if recorder.Code != tt.expectedStatus {
				t.Errorf("ServeHTTP() status = %v, want %v",
					recorder.Code, tt.expectedStatus)
			}

			if tt.expectedBody != nil {
				switch body := tt.expectedBody.(type) {
				case []byte:
					if !bytes.Equal(recorder.Body.Bytes(), body) {
						t.Errorf("ServeHTTP() body = %v, want %v",
							recorder.Body.String(), string(body))
					}
				case DataDirectoryInfo:
					var info DataDirectoryInfo
					if err := json.NewDecoder(recorder.Body).Decode(&info); err != nil {
						t.Errorf("ServeHTTP() failed to decode response: %v", err)
						return
					}
					if !compareDataDirectoryInfo(info, body) {
						t.Errorf("ServeHTTP() body = %+v, want %+v",
							info, body)
					}
				}
			}
		})
	}
}

func compareDataDirectoryInfo(a, b DataDirectoryInfo) bool {
	if a.Path != b.Path {
		return false
	}
	if len(a.Resources) != len(b.Resources) {
		return false
	}
	if len(a.Containers) != len(b.Containers) {
		return false
	}
	if len(a.ACLs) != len(b.ACLs) {
		return false
	}
	// TODO: Implement proper slice comparison
	return true
}
