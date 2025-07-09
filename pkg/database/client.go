package database

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type client struct {
	baseURL string
	apiKey  string
	client  *http.Client
}

func NewClient(baseURL, apiKey string) SupabaseClient {
	return &client{
		baseURL: baseURL,
		apiKey:  apiKey,
		client:  &http.Client{},
	}
}

func (c *client) Get(table string) ([]byte, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/rest/v1/%s", c.baseURL, table), nil)
	if err != nil {
		return nil, err
	}
	c.setHeaders(req)
	resp, err := c.client.Do(req)
	return handleResponse(resp, err)
}

func (c *client) Insert(table string, data interface{}) ([]byte, error) {
	body, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/rest/v1/%s", c.baseURL, table), bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	c.setHeaders(req)
	req.Header.Set("Prefer", "return=representation")
	resp, err := c.client.Do(req)
	return handleResponse(resp, err)
}

func (c *client) Delete(table string, filter string) ([]byte, error) {
	req, err := http.NewRequest("DELETE", fmt.Sprintf("%s/rest/v1/%s?%s", c.baseURL, table, filter), nil)
	if err != nil {
		return nil, err
	}
	c.setHeaders(req)
	resp, err := c.client.Do(req)
	return handleResponse(resp, err)
}

func (c *client) setHeaders(req *http.Request) {
	req.Header.Set("apikey", c.apiKey)
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")
}

func handleResponse(resp *http.Response, err error) ([]byte, error) {
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
	}
	return body, err
}
