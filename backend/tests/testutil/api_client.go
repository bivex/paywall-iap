package testutil

import (
	"bytes"
	"encoding/json"
	"net/http"
)

// NewTestRequest creates a new HTTP request for testing
func NewTestRequest(method, url string, body interface{}, token string) (*http.Request, error) {
	var req *http.Request
	var err error

	if body != nil {
		bodyBytes, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		req, err = http.NewRequest(method, url, bytes.NewReader(bodyBytes))
		if err != nil {
			return nil, err
		}
		req.Header.Set("Content-Type", "application/json")
	} else {
		req, err = http.NewRequest(method, url, nil)
		if err != nil {
			return nil, err
		}
	}

	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	return req, nil
}

// DoRequest executes a request and returns the response
func DoRequest(client *http.Client, req *http.Request) (*http.Response, []byte, error) {
	if client == nil {
		client = &http.Client{}
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()

	buf := new(bytes.Buffer)
	if _, err := buf.ReadFrom(resp.Body); err != nil {
		return nil, nil, err
	}
	body := buf.Bytes()

	return resp, body, nil
}

// DecodeResponse decodes a JSON response into the provided struct
func DecodeResponse(body []byte, v interface{}) error {
	return json.Unmarshal(body, v)
}

// APIResponse represents a standard API response
type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
	Message string      `json:"message,omitempty"`
}

// AuthResponse represents authentication response
type AuthResponse struct {
	Data AuthData `json:"data"`
}

type AuthData struct {
	UserID       string `json:"user_id"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
}

// SubscriptionResponse represents subscription response
type SubscriptionResponse struct {
	Data SubscriptionData `json:"data"`
}

type SubscriptionData struct {
	ID        string `json:"id"`
	Status    string `json:"status"`
	Source    string `json:"source"`
	PlanType  string `json:"plan_type"`
	ExpiresAt string `json:"expires_at"`
	AutoRenew bool   `json:"auto_renew"`
}

// AccessCheckResponse represents access check response
type AccessCheckResponse struct {
	Data AccessData `json:"data"`
}

type AccessData struct {
	HasAccess bool   `json:"has_access"`
	ExpiresAt string `json:"expires_at,omitempty"`
	PlanType  string `json:"plan_type,omitempty"`
}
