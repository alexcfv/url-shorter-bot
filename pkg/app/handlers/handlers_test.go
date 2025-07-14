package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
	"url-shorter-bot/pkg/models"

	"github.com/gorilla/mux"
)

func init() {
	models.Config.HostName = "localhost"
	models.Config.Port = "80"
	models.Protocol = "http"
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

type mockSupabase struct {
	data map[string]string
}

type mockLogger struct {
	db *mockSupabase
}

func (l *mockLogger) LogAction(telegramID int64, action string) {
	payload := models.LogAction{
		Telegram_id: telegramID,
		Action:      action,
	}
	if _, err := l.db.Insert("log_action", payload); err != nil {
		log.Printf("log action failed: %v", err)
	}
}

func (l *mockLogger) LogError(telegramID int64, errMsg, code string) {
	payload := models.LogError{
		Telegram_id: telegramID,
		Error:       errMsg,
		Error_code:  code,
	}
	if _, err := l.db.Insert("log_error", payload); err != nil {
		log.Printf("log error failed: %v", err)
	}
}

func (m *mockSupabase) Get(table string, target map[string]string) ([]byte, error) {
	key, ok := target["Hash"]
	if !ok {
		return nil, fmt.Errorf("missing filter key 'hash'")
	}

	val, exists := m.data[key]
	if !exists {
		return nil, fmt.Errorf("no record found for hash: %s", key)
	}

	result := map[string]string{
		"Hash":         key,
		"original_url": val,
	}

	return json.Marshal(result)
}

func (m *mockSupabase) Insert(table string, data interface{}) ([]byte, error) {
	if m.data == nil {
		m.data = map[string]string{}
	}
	u, ok := data.(models.Url)
	if !ok {
		return nil, fmt.Errorf("invalid format")
	}
	m.data[u.Hash] = u.Url
	return json.Marshal(u)
}

func (m *mockSupabase) Delete(table string, filter string) ([]byte, error) {
	delete(m.data, "12345")
	return []byte(`{}`), nil
}

func TestHandlerHashUrl(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		url            string
		setupCache     func(m *mockCache)
		setupDB        func(db *mockSupabase)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:   "valid GET — found in cache",
			method: http.MethodGet,
			url:    "/12345",
			setupCache: func(m *mockCache) {
				m.Set("12345", "https://cached-url.com", 10*time.Minute)
			},
			setupDB:        func(db *mockSupabase) {},
			expectedStatus: http.StatusOK,
			expectedBody:   "From cache: https://cached-url.com",
		},
		{
			name:   "valid GET — not in cache, found in Supabase",
			method: http.MethodGet,
			url:    "/67890",
			setupCache: func(m *mockCache) {
			},
			setupDB: func(db *mockSupabase) {
				db.data["67890"] = "https://supabase-url.com"
			},
			expectedStatus: http.StatusOK,
			expectedBody:   "Fetched and cached: https://supabase-url.com",
		},
		{
			name:           "invalid method POST",
			method:         http.MethodPost,
			url:            "/12345",
			setupCache:     func(m *mockCache) {},
			setupDB:        func(db *mockSupabase) {},
			expectedStatus: http.StatusMethodNotAllowed,
			expectedBody:   "must be only GET\n",
		},
		{
			name:           "non-numeric URL — not matched",
			method:         http.MethodGet,
			url:            "/abc",
			setupCache:     func(m *mockCache) {},
			setupDB:        func(db *mockSupabase) {},
			expectedStatus: http.StatusNotFound,
			expectedBody:   "404 page not found\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockCache{data: make(map[string]string)}
			db := &mockSupabase{data: make(map[string]string)}
			log := &mockLogger{db: db}

			tt.setupCache(mock)
			tt.setupDB(db)

			handler := NewHashedUrlHandler(mock, db, log)

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
			requestBody:    `{"Url":"https://example.com"}`,
			expectedStatus: http.StatusOK,
			expectedInBody: "http://localhost/",
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
			requestBody:    `{"Url":"invalid-url"}`,
			expectedStatus: http.StatusUnsupportedMediaType,
			expectedInBody: "invalid JSON body",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, "/short", bytes.NewBufferString(tt.requestBody))
			w := httptest.NewRecorder()

			db := &mockSupabase{data: make(map[string]string)}
			log := &mockLogger{db: db}

			handler := NewShortdUrlHandler(db, log)

			r := mux.NewRouter()
			r.HandleFunc("/short", handler.HandlerUrlShort)

			r.ServeHTTP(w, req)

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
