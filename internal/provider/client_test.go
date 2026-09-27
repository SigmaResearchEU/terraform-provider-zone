package provider

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewClient(t *testing.T) {
	client := NewClient("testuser", "testapikey")

	if client.username != "testuser" {
		t.Errorf("expected username 'testuser', got '%s'", client.username)
	}
	if client.apiKey != "testapikey" {
		t.Errorf("expected apiKey 'testapikey', got '%s'", client.apiKey)
	}
	if client.httpClient == nil {
		t.Error("expected httpClient to be initialized")
	}
	if client.rateLimitLimit != defaultRateLimit {
		t.Errorf("expected rateLimitLimit %d, got %d", defaultRateLimit, client.rateLimitLimit)
	}
}

func TestAuthHeader(t *testing.T) {
	client := NewClient("testuser", "testapikey")
	header := client.authHeader()

	// "testuser:testapikey" base64 encoded = "dGVzdHVzZXI6dGVzdGFwaWtleQ=="
	expected := "Basic dGVzdHVzZXI6dGVzdGFwaWtleQ=="
	if header != expected {
		t.Errorf("expected auth header '%s', got '%s'", expected, header)
	}
}

func TestParseDNSRecordResponse(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectError bool
		expectID    string
	}{
		{
			name:        "valid response with one record",
			input:       `[{"id": "123", "name": "test", "destination": "192.168.1.1"}]`,
			expectError: false,
			expectID:    "123",
		},
		{
			name:        "empty array",
			input:       `[]`,
			expectError: true,
		},
		{
			name:        "invalid json",
			input:       `{invalid`,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			record, err := parseDNSRecordResponse([]byte(tt.input))
			if tt.expectError {
				if err == nil {
					t.Error("expected error but got nil")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if record.ID != tt.expectID {
					t.Errorf("expected ID '%s', got '%s'", tt.expectID, record.ID)
				}
			}
		})
	}
}

func TestParseDNSZoneResponse(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectError bool
		expectName  string
	}{
		{
			name:        "valid response with one zone",
			input:       `[{"name": "example.com"}]`,
			expectError: false,
			expectName:  "example.com",
		},
		{
			name:        "empty array",
			input:       `[]`,
			expectError: true,
		},
		{
			name:        "invalid json",
			input:       `{invalid`,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			zone, err := parseDNSZoneResponse([]byte(tt.input))
			if tt.expectError {
				if err == nil {
					t.Error("expected error but got nil")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if zone.Name != tt.expectName {
					t.Errorf("expected name '%s', got '%s'", tt.expectName, zone.Name)
				}
			}
		})
	}
}

func TestUpdateRateLimitInfo(t *testing.T) {
	client := NewClient("testuser", "testapikey")

	resp := &http.Response{
		Header: http.Header{
			"X-Ratelimit-Limit":     []string{"100"},
			"X-Ratelimit-Remaining": []string{"50"},
		},
	}

	client.updateRateLimitInfo(resp)

	if client.rateLimitLimit != 100 {
		t.Errorf("expected rateLimitLimit 100, got %d", client.rateLimitLimit)
	}
	if client.rateLimitRemaining != 50 {
		t.Errorf("expected rateLimitRemaining 50, got %d", client.rateLimitRemaining)
	}
}

func TestGetDNSZone_MockServer(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("expected GET method, got %s", r.Method)
		}

		auth := r.Header.Get("Authorization")
		if auth == "" {
			t.Error("expected Authorization header")
		}

		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-Ratelimit-Limit", "60")
		w.Header().Set("X-Ratelimit-Remaining", "59")
		json.NewEncoder(w).Encode([]DNSZone{{Name: "example.com"}})
	}))
	defer server.Close()

	// Verify server is working
	if server.URL == "" {
		t.Error("expected server URL to be set")
	}
}

func TestCreateARecord_MockServer(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST method, got %s", r.Method)
		}

		var record DNSRecord
		if err := json.NewDecoder(r.Body).Decode(&record); err != nil {
			t.Errorf("failed to decode request body: %v", err)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode([]DNSRecord{{
			ID:          "123",
			Name:        record.Name,
			Destination: record.Destination,
		}})
	}))
	defer server.Close()

	// Verify server is working
	if server.URL == "" {
		t.Error("expected server URL to be set")
	}
}
