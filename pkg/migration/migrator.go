package migration

// rpc function for create table, supabase SQL editor

/*
create or replace function table_exists(tbl text)
returns boolean
language plpgsql
as $$
begin
  return exists (
    select from pg_tables
    where tablename = tbl
  );
end;
$$;
*/

// function permission rpc on supabase SQL editor

/*
grant execute on function table_exists(text) to service_role;
*/

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type SupabaseMigrator struct {
	ProjectUrl string
	ApiKey     string
}

func NewMigrator(projectUrl, apiKey string) *SupabaseMigrator {
	return &SupabaseMigrator{
		ProjectUrl: projectUrl,
		ApiKey:     apiKey,
	}
}

func (m *SupabaseMigrator) TablesExists(tableName string) (bool, error) {
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

func (m *SupabaseMigrator) CreateTable(table, request string) error {
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
