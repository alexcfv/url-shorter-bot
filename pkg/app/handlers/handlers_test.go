package handlers

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
	"url-shorter-bot/pkg/models"

	"github.com/gorilla/mux"
)

func init() {
	models.Config.HostName = "localhost"
	models.Config.Port = "8000"
}

type mockCache struct {
	data map[string]string
}

func (m *mockCache) Set(key string, value interface{}, duration time.Duration) {
	m.data[key] = value.(string)
}

func (m *mockCache) Get(key string) (interface{}, bool) {
	val, ok := m.data[key]
	return val, ok
}

func (m *mockCache) Delete(key string) {
	delete(m.data, key)
}

func TestHandlerHashUrl(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		url            string
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "valid GET numeric URL (not in cache)",
			method:         http.MethodGet,
			url:            "/12345",
			expectedStatus: http.StatusOK,
			expectedBody:   "Fetched and cached: 12345",
		},
		{
			name:           "invalid method POST",
			method:         http.MethodPost,
			url:            "/12345",
			expectedStatus: http.StatusMethodNotAllowed,
			expectedBody:   "must be only GET\n",
		},
		{
			name:           "non-numeric URL - not matched by router",
			method:         http.MethodGet,
			url:            "/abc",
			expectedStatus: http.StatusNotFound,
			expectedBody:   "404 page not found\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockCache{data: make(map[string]string)}
			handler := NewHashedUrlHandler(mock)

			if strings.Contains(tt.name, "valid GET") {
				mock.Set("12345", "https://yourdomain.com/12345", 10*time.Minute)
				tt.expectedBody = "From cache: https://yourdomain.com/12345"
			}

			r := mux.NewRouter()
			r.HandleFunc("/{url:[0-9]+}", handler.HandlerHashUrl)

			req := httptest.NewRequest(tt.method, tt.url, nil)
			w := httptest.NewRecorder()

			r.ServeHTTP(w, req)

			resp := w.Result()
			defer resp.Body.Close()

			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, resp.StatusCode)
			}

			body := w.Body.String()
			if body != tt.expectedBody {
				t.Errorf("expected body %q, got %q", tt.expectedBody, body)
			}
		})
	}
}

func TestHandlerUrlShort(t *testing.T) {

	tests := []struct {
		name           string
		method         string
		requestBody    string
		expectedStatus int
		expectedInBody string
	}{
		{
			name:           "valid POST with correct URL",
			method:         http.MethodPost,
			requestBody:    `{"url":"https://example.com"}`,
			expectedStatus: http.StatusOK,
			expectedInBody: "http://localhost:8000/",
		},
		{
			name:           "invalid method GET",
			method:         http.MethodGet,
			requestBody:    ``,
			expectedStatus: http.StatusMethodNotAllowed,
			expectedInBody: "must be only POST",
		},
		{
			name:           "invalid JSON",
			method:         http.MethodPost,
			requestBody:    `not a json`,
			expectedStatus: http.StatusUnsupportedMediaType,
			expectedInBody: "invalid JSON body",
		},
		{
			name:           "invalid URL inside JSON",
			method:         http.MethodPost,
			requestBody:    `{"url":"invalid-url"}`,
			expectedStatus: http.StatusUnsupportedMediaType,
			expectedInBody: "invalid JSON body",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, "/short", bytes.NewBufferString(tt.requestBody))
			w := httptest.NewRecorder()

			HandlerUrlShort(w, req)

			resp := w.Result()
			defer resp.Body.Close()

			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, resp.StatusCode)
			}

			buf := new(bytes.Buffer)
			buf.ReadFrom(resp.Body)
			body := buf.String()

			if tt.expectedInBody != "" && !bytes.Contains([]byte(body), []byte(tt.expectedInBody)) {
				t.Errorf("expected body to contain %q, got %q", tt.expectedInBody, body)
			}
		})
	}
}
