package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"url-shorter-bot/pkg/middleware"
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
	bad  bool
}

func (m *mockSupabase) Get(table string, target map[string]string) ([]byte, error) {
	hash := target["Hash"]
	if m.bad {
		return []byte(`{malformed_json}`), nil
	}
	val, ok := m.data[hash]
	if !ok {
		return nil, fmt.Errorf("not found")
	}
	return json.Marshal(models.Url{
		Hash: hash,
		Url:  val,
	})
}

func (m *mockSupabase) Insert(table string, data interface{}) ([]byte, error) {
	u, ok := data.(models.Url)
	if !ok {
		return nil, fmt.Errorf("invalid insert format")
	}
	m.data[u.Hash] = u.Url
	return json.Marshal(u)
}

func (m *mockSupabase) Delete(table, filter string) ([]byte, error) {
	return []byte(`{}`), nil
}

type mockLogger struct {
	db *mockSupabase
}

func (l *mockLogger) LogAction(telegramID int64, action string)      {}
func (l *mockLogger) LogError(telegramID int64, errMsg, code string) {}

func TestHandlerHashUrl(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		url            string
		setupCache     func(m *mockCache)
		setupDB        func(db *mockSupabase)
		expectedStatus int
	}{
		{
			name:   "redirect from cache",
			method: http.MethodGet,
			url:    "/abc123",
			setupCache: func(m *mockCache) {
				m.Set("abc123", "https://cached.com", 10*time.Minute)
			},
			setupDB:        func(db *mockSupabase) {},
			expectedStatus: http.StatusFound,
		},
		{
			name:       "redirect from Supabase",
			method:     http.MethodGet,
			url:        "/xyz456",
			setupCache: func(m *mockCache) {},
			setupDB: func(db *mockSupabase) {
				db.data["xyz456"] = "https://from-db.com"
			},
			expectedStatus: http.StatusFound,
		},
		{
			name:           "missing method",
			method:         http.MethodPost,
			url:            "/abc123",
			setupCache:     func(m *mockCache) {},
			setupDB:        func(db *mockSupabase) {},
			expectedStatus: http.StatusMethodNotAllowed,
		},
		{
			name:           "hash not found",
			method:         http.MethodGet,
			url:            "/notfound",
			setupCache:     func(m *mockCache) {},
			setupDB:        func(db *mockSupabase) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:       "invalid JSON from Supabase",
			method:     http.MethodGet,
			url:        "/badjson",
			setupCache: func(m *mockCache) {},
			setupDB: func(db *mockSupabase) {
				db.bad = true
				db.data["badjson"] = "https://broken.com"
			},
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:           "missing hash path param",
			method:         http.MethodGet,
			url:            "/",
			setupCache:     func(m *mockCache) {},
			setupDB:        func(db *mockSupabase) {},
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cache := &mockCache{data: make(map[string]string)}
			db := &mockSupabase{data: make(map[string]string)}
			log := &mockLogger{db: db}
			tt.setupCache(cache)
			tt.setupDB(db)

			handler := NewHashedUrlHandler(cache, db, log)
			r := mux.NewRouter()
			r.HandleFunc("/{url}", handler.HandlerHashUrl)

			req := httptest.NewRequest(tt.method, tt.url, nil)
			w := httptest.NewRecorder()

			r.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}

func TestHandlerUrlShort(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		requestBody    string
		contentType    string
		expectedStatus int
		expectedText   string
	}{
		{
			name:           "valid short URL",
			method:         http.MethodPost,
			requestBody:    `{"Url":"https://valid.com"}`,
			contentType:    "application/json",
			expectedStatus: http.StatusOK,
			expectedText:   "http://localhost/",
		},
		{
			name:           "invalid JSON",
			method:         http.MethodPost,
			requestBody:    `{notjson}`,
			contentType:    "application/json",
			expectedStatus: http.StatusUnsupportedMediaType,
			expectedText:   "invalid JSON body",
		},
		{
			name:           "invalid URL format",
			method:         http.MethodPost,
			requestBody:    `{"Url":"htp:/bad"}`,
			contentType:    "application/json",
			expectedStatus: http.StatusUnsupportedMediaType,
			expectedText:   "invalid URL",
		},
		{
			name:           "missing Content-Type",
			method:         http.MethodPost,
			requestBody:    `{"Url":"https://valid.com"}`,
			contentType:    "",
			expectedStatus: http.StatusUnsupportedMediaType,
		},
		{
			name:           "empty body",
			method:         http.MethodPost,
			requestBody:    ``,
			contentType:    "application/json",
			expectedStatus: http.StatusUnsupportedMediaType,
		},
		{
			name:           "invalid method GET",
			method:         http.MethodGet,
			requestBody:    ``,
			contentType:    "application/json",
			expectedStatus: http.StatusMethodNotAllowed,
			expectedText:   "must be only POST",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, "/short", bytes.NewBufferString(tt.requestBody))
			if tt.contentType != "" {
				req.Header.Set("Content-Type", tt.contentType)
			}
			req.Header.Set("X-Telegram-ID", "123456")

			w := httptest.NewRecorder()
			db := &mockSupabase{data: map[string]string{}}
			log := &mockLogger{db: db}
			handler := NewShortdUrlHandler(db, log)

			r := mux.NewRouter()
			r.HandleFunc("/short", handler.HandlerUrlShort)
			r.Use(middleware.TelegramIDMiddleware)

			r.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.expectedText != "" && !bytes.Contains(w.Body.Bytes(), []byte(tt.expectedText)) {
				t.Errorf("expected body to contain %q, got %q", tt.expectedText, w.Body.String())
			}
		})
	}
}
