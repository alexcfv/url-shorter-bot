package migration_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"url-shorter-bot/pkg/models"
)

type testMigrator struct {
	ProjectUrl string
	ApiKey     string
	Client     *http.Client
}

func newTestMigrator(projectUrl, apiKey string, client *http.Client) *testMigrator {
	return &testMigrator{
		ProjectUrl: projectUrl,
		ApiKey:     apiKey,
		Client:     client,
	}
}

func (m *testMigrator) TablesExists(tableName string) (bool, error) {
	url := fmt.Sprintf("%s/rest/v1/rpc/table_exists", m.ProjectUrl)

	bodyData := map[string]string{"tbl": tableName}
	bodyBytes, _ := json.Marshal(bodyData)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return false, err
	}
	req.Header.Set("apikey", m.ApiKey)
	req.Header.Set("Authorization", "Bearer "+m.ApiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResp map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&errResp)
		return false, fmt.Errorf("supabase RPC error: %v (status %d)", errResp["message"], resp.StatusCode)
	}

	var exists bool
	if err := json.NewDecoder(resp.Body).Decode(&exists); err != nil {
		return false, err
	}
	return exists, nil
}

func (m *testMigrator) CreateTable(table, request string) error {
	payload := map[string]string{"sql": request}
	body, _ := json.Marshal(payload)

	url := fmt.Sprintf("%s/rest/v1/rpc/execute_sql", m.ProjectUrl)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	req.Header.Set("apikey", m.ApiKey)
	req.Header.Set("Authorization", "Bearer "+m.ApiKey)
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return fmt.Errorf("failed to create table %s: %s", table, resp.Status)
	}
	return nil
}

func TestTableExists(t *testing.T) {
	tests := []struct {
		name           string
		tableName      string
		mockResponse   interface{}
		statusCode     int
		expectedResult bool
		expectError    bool
	}{
		{
			name:           "table exists",
			tableName:      "urls",
			mockResponse:   true,
			statusCode:     200,
			expectedResult: true,
			expectError:    false,
		},
		{
			name:           "table does not exist",
			tableName:      "anything",
			mockResponse:   false,
			statusCode:     200,
			expectedResult: false,
			expectError:    false,
		},
		{
			name:         "supabase error",
			tableName:    "urls",
			mockResponse: map[string]string{"message": "access denied"},
			statusCode:   403,
			expectError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
				if tt.mockResponse != nil {
					json.NewEncoder(w).Encode(tt.mockResponse)
				}
			}))
			defer server.Close()

			m := newTestMigrator(server.URL, "fake-key", server.Client())

			result, err := m.TablesExists(tt.tableName)

			if (err != nil) != tt.expectError {
				t.Errorf("expected error = %v, got %v", tt.expectError, err)
			}
			if result != tt.expectedResult {
				t.Errorf("expected result = %v, got %v", tt.expectedResult, result)
			}
		})
	}
}

func TestCreateTable(t *testing.T) {
	tests := []struct {
		table       string
		query       string
		name        string
		statusCode  int
		expectError bool
	}{
		{
			table:       "name",
			query:       models.SqlRequests["urls"],
			name:        "success",
			statusCode:  200,
			expectError: false,
		},
		{
			table:       "name",
			query:       models.SqlRequests["urls"],
			name:        "supabase returns error",
			statusCode:  500,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
				io.Copy(io.Discard, r.Body)
			}))
			defer server.Close()

			m := newTestMigrator(server.URL, "fake-key", server.Client())

			err := m.CreateTable(tt.table, tt.query)
			if (err != nil) != tt.expectError {
				t.Errorf("expected error = %v, got %v", tt.expectError, err)
			}
		})
	}
}
