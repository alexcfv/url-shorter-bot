package database

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestClient_Get_tableQuery(t *testing.T) {
	tests := []struct {
		name          string
		table         string
		query         map[string]string
		expectedQuery string
		mockResponse  string
		statusCode    int
		expectErr     bool
	}{
		{
			name:          "Get with single filter",
			table:         "users",
			query:         map[string]string{"id": "123"},
			expectedQuery: "id=eq.123",
			mockResponse:  `{"id": "123", "name": "Alice"}`,
			statusCode:    http.StatusOK,
			expectErr:     false,
		},
		{
			name:          "Get with multiple filters",
			table:         "orders",
			query:         map[string]string{"status": "paid", "user_id": "42"},
			expectedQuery: "", // dynamic
			mockResponse:  `[{"id":1},{"id":2}]`,
			statusCode:    http.StatusOK,
			expectErr:     false,
		},
		{
			name:         "Get 404 error",
			table:        "products",
			query:        map[string]string{"id": "999"},
			mockResponse: `Not found`,
			statusCode:   http.StatusNotFound,
			expectErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if !strings.Contains(r.URL.Path, tt.table) {
					t.Errorf("expected table %q in path, got %q", tt.table, r.URL.Path)
				}
				if tt.expectedQuery != "" && r.URL.RawQuery != tt.expectedQuery {
					t.Errorf("expected query %q, got %q", tt.expectedQuery, r.URL.RawQuery)
				}
				w.WriteHeader(tt.statusCode)
				w.Write([]byte(tt.mockResponse))
			}))
			defer ts.Close()

			client := NewClient(ts.URL, "test-api-key")
			resp, err := client.Get(tt.table, tt.query)

			if tt.expectErr && err == nil {
				t.Errorf("expected error, got nil")
			}
			if !tt.expectErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if !tt.expectErr && string(resp) != tt.mockResponse {
				t.Errorf("expected %q, got %q", tt.mockResponse, string(resp))
			}
		})
	}
}

func TestClient_Insert_tableQuery(t *testing.T) {
	tests := []struct {
		name         string
		table        string
		data         map[string]interface{}
		mockResponse string
		statusCode   int
		expectErr    bool
	}{
		{
			name:         "Insert success",
			table:        "users",
			data:         map[string]interface{}{"name": "Bob"},
			mockResponse: `[{"id": 1, "name": "Bob"}]`,
			statusCode:   http.StatusCreated,
			expectErr:    false,
		},
		{
			name:         "Insert failure",
			table:        "users",
			data:         map[string]interface{}{},
			mockResponse: `{"error":"bad request"}`,
			statusCode:   http.StatusBadRequest,
			expectErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if !strings.Contains(r.URL.Path, tt.table) {
					t.Errorf("expected table %q in path, got %q", tt.table, r.URL.Path)
				}
				body, _ := io.ReadAll(r.Body)
				if tt.table == "users" && !bytes.Contains(body, []byte("Bob")) && len(tt.data) > 0 {
					t.Errorf("expected request body to contain 'Bob', got: %s", string(body))
				}
				w.WriteHeader(tt.statusCode)
				w.Write([]byte(tt.mockResponse))
			}))
			defer ts.Close()

			client := NewClient(ts.URL, "test-api-key")
			resp, err := client.Insert(tt.table, tt.data)

			if tt.expectErr && err == nil {
				t.Errorf("expected error, got nil")
			}
			if !tt.expectErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if !tt.expectErr && string(resp) != tt.mockResponse {
				t.Errorf("expected %q, got %q", tt.mockResponse, string(resp))
			}
		})
	}
}

func TestClient_Delete_tableQuery(t *testing.T) {
	tests := []struct {
		name         string
		table        string
		query        string
		mockResponse string
		statusCode   int
		expectErr    bool
	}{
		{
			name:         "Delete success",
			table:        "sessions",
			query:        "id=eq.10",
			mockResponse: ``,
			statusCode:   http.StatusNoContent,
			expectErr:    false,
		},
		{
			name:         "Delete not found",
			table:        "sessions",
			query:        "id=eq.999",
			mockResponse: `Not found`,
			statusCode:   http.StatusNotFound,
			expectErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if !strings.Contains(r.URL.Path, tt.table) {
					t.Errorf("expected table %q in path, got %q", tt.table, r.URL.Path)
				}
				if r.URL.RawQuery != tt.query {
					t.Errorf("expected query %q, got %q", tt.query, r.URL.RawQuery)
				}
				w.WriteHeader(tt.statusCode)
				w.Write([]byte(tt.mockResponse))
			}))
			defer ts.Close()

			client := NewClient(ts.URL, "test-api-key")
			resp, err := client.Delete(tt.table, tt.query)

			if tt.expectErr && err == nil {
				t.Errorf("expected error, got nil")
			}
			if !tt.expectErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if !tt.expectErr && string(resp) != tt.mockResponse {
				t.Errorf("expected %q, got %q", tt.mockResponse, string(resp))
			}
		})
	}
}
