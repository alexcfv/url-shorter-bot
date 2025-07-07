package handlers

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
	"url-shorter-bot/pkg/models"
)

func init() {
	models.Config.HostName = "localhost"
	models.Config.Port = "8000"
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
