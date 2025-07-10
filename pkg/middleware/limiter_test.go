package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestGetVisitor_NewAndExisting(t *testing.T) {
	ip := "203.0.113.1"
	lim1 := GetVisitor(ip)
	if lim1 == nil {
		t.Fatal("Expected limiter, got nil")
	}

	lim2 := GetVisitor(ip)
	if lim2 != lim1 {
		t.Error("Expected the same limiter instance for same IP")
	}
}

func TestCleanupVisitors(t *testing.T) {
	ip := "198.51.100.1"
	GetVisitor(ip)

	mu.Lock()
	visitors[ip].lastSeen = time.Now().Add(-3 * time.Minute)
	mu.Unlock()

	CleanupVisitors("testing")

	mu.Lock()
	_, exists := visitors[ip]
	mu.Unlock()

	if exists {
		t.Error("Visitor should have been cleaned up")
	}
}

func TestGetVisitor(t *testing.T) {
	tests := []struct {
		name string
		ip   string
	}{
		{"Normal IP", "192.168.0.1"},
		{"Another IP", "10.0.0.1"},
		{"Empty IP", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lim := GetVisitor(tt.ip)
			if lim == nil {
				t.Errorf("Expected nil limiter for IP %q", tt.ip)
			}
		})
	}
}

func TestRateLimitMiddleware(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	tests := []struct {
		name           string
		ip             string
		requests       int
		expectedStatus int
	}{
		{
			name:           "Under limit first",
			ip:             "192.0.2.1",
			requests:       1,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Under limit second",
			ip:             "192.0.2.1",
			requests:       1,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Over limit",
			ip:             "192.0.2.2",
			requests:       3,
			expectedStatus: http.StatusTooManyRequests,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mw := RateLimitMiddleware(handler)
			var finalCode int

			for i := 0; i < tt.requests; i++ {
				rec := httptest.NewRecorder()
				req := httptest.NewRequest("GET", "/", nil)
				req.RemoteAddr = tt.ip + ":12345"
				mw.ServeHTTP(rec, req)
				finalCode = rec.Code
			}

			if finalCode != tt.expectedStatus {
				t.Errorf("Final response code = %d; want %d", finalCode, tt.expectedStatus)
			}
		})
	}
}
