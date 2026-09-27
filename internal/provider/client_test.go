package provider

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func newTestClient(t *testing.T, handler http.HandlerFunc) *Client {
	t.Helper()
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)
	c := NewClient("testuser", "testapikey")
	c.baseURL = server.URL
	return c
}

func TestNewClient(t *testing.T) {
	client := NewClient("testuser", "testapikey")
	if client.baseURL != defaultBaseURL {
		t.Errorf("expected baseURL %q, got %q", defaultBaseURL, client.baseURL)
	}
	if client.rateLimitLimit != defaultRateLimit {
		t.Errorf("expected rateLimitLimit %d, got %d", defaultRateLimit, client.rateLimitLimit)
	}
}

func TestAuthHeader(t *testing.T) {
	client := NewClient("testuser", "testapikey")
	// "testuser:testapikey" base64 encoded
	expected := "Basic dGVzdHVzZXI6dGVzdGFwaWtleQ=="
	if got := client.authHeader(); got != expected {
		t.Errorf("expected auth header %q, got %q", expected, got)
	}
}

func TestParseSingle(t *testing.T) {
	record, err := parseSingle[DNSRecord]([]byte(`[{"id":"950615","name":"node.example.com","destination":"192.0.2.1"}]`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if record.ID != "950615" {
		t.Errorf("expected ID 950615, got %q", record.ID)
	}

	if _, err := parseSingle[DNSRecord]([]byte(`[]`)); !isNotFound(err) {
		t.Errorf("empty array: expected not-found, got %v", err)
	}
	if _, err := parseSingle[DNSRecord]([]byte(`{invalid`)); err == nil || isNotFound(err) {
		t.Errorf("invalid json: expected parse error, got %v", err)
	}
}

func TestUpdateRateLimitInfo(t *testing.T) {
	client := NewClient("testuser", "testapikey")
	resp := &http.Response{Header: http.Header{}}
	resp.Header.Set("X-Ratelimit-Limit", "100")
	resp.Header.Set("X-Ratelimit-Remaining", "50")

	client.updateRateLimitInfo(resp)

	if client.rateLimitLimit != 100 {
		t.Errorf("expected rateLimitLimit 100, got %d", client.rateLimitLimit)
	}
	if client.rateLimitRemaining != 50 {
		t.Errorf("expected rateLimitRemaining 50, got %d", client.rateLimitRemaining)
	}
}

func TestRecordCRUD(t *testing.T) {
	type call struct{ method, path, body string }
	var calls []call

	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		_ = json.NewDecoder(r.Body).Decode(&body)
		b, _ := json.Marshal(body)
		calls = append(calls, call{r.Method, r.URL.Path, string(b)})

		if got := r.Header.Get("Authorization"); got != "Basic dGVzdHVzZXI6dGVzdGFwaWtleQ==" {
			t.Errorf("unexpected Authorization header %q", got)
		}

		switch r.Method {
		case http.MethodDelete:
			w.WriteHeader(http.StatusNoContent)
		case http.MethodPost:
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(`[{"id":"950615","name":"_tftest.example.com","destination":"\"v=spf1 -all\""}]`))
		default:
			_, _ = w.Write([]byte(`[{"id":"950615","name":"_tftest.example.com","destination":"\"v=spf1 -all\""}]`))
		}
	})
	ctx := context.Background()

	created, err := client.CreateRecord(ctx, recordTypeTXT, "example.com", &DNSRecord{
		Name:        "_tftest.example.com",
		Destination: `"v=spf1 -all"`,
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if created.ID != "950615" {
		t.Errorf("create: expected string ID 950615, got %q", created.ID)
	}
	if created.Destination != `"v=spf1 -all"` {
		t.Errorf("create: TXT destination must pass through unmodified, got %q", created.Destination)
	}

	if _, err := client.GetRecord(ctx, recordTypeTXT, "example.com", "950615"); err != nil {
		t.Fatalf("get: %v", err)
	}
	if _, err := client.UpdateRecord(ctx, recordTypeTXT, "example.com", "950615", &DNSRecord{Name: "_tftest.example.com", Destination: "x"}); err != nil {
		t.Fatalf("update: %v", err)
	}
	if err := client.DeleteRecord(ctx, recordTypeTXT, "example.com", "950615"); err != nil {
		t.Fatalf("delete: %v", err)
	}

	want := []call{
		{http.MethodPost, "/dns/example.com/txt", `{"destination":"\"v=spf1 -all\"","name":"_tftest.example.com"}`},
		{http.MethodGet, "/dns/example.com/txt/950615", "null"},
		{http.MethodPut, "/dns/example.com/txt/950615", `{"destination":"x","name":"_tftest.example.com"}`},
		{http.MethodDelete, "/dns/example.com/txt/950615", "null"},
	}
	if len(calls) != len(want) {
		t.Fatalf("expected %d calls, got %d: %+v", len(want), len(calls), calls)
	}
	for i := range want {
		if calls[i] != want[i] {
			t.Errorf("call %d: expected %+v, got %+v", i, want[i], calls[i])
		}
	}
}

func TestGetRecordNotFound(t *testing.T) {
	tests := map[string]http.HandlerFunc{
		"404": func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		},
		"empty array": func(w http.ResponseWriter, r *http.Request) {
			_, _ = w.Write([]byte(`[]`))
		},
	}
	for name, handler := range tests {
		t.Run(name, func(t *testing.T) {
			client := newTestClient(t, handler)
			_, err := client.GetRecord(context.Background(), recordTypeA, "example.com", "1")
			if !isNotFound(err) {
				t.Errorf("expected not-found, got %v", err)
			}
		})
	}
}

func TestValidationErrorKeepsBody(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Status-Message", "Puudulik sisend")
		w.WriteHeader(http.StatusUnprocessableEntity)
		_, _ = w.Write([]byte(`{"name":"invalid_host"}`))
	})

	_, err := client.CreateRecord(context.Background(), recordTypeA, "example.com", &DNSRecord{Name: "short", Destination: "192.0.2.1"})
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected *APIError, got %T: %v", err, err)
	}
	if apiErr.StatusCode != http.StatusUnprocessableEntity || apiErr.Body != `{"name":"invalid_host"}` {
		t.Errorf("unexpected error contents: %+v", apiErr)
	}
	if isNotFound(err) {
		t.Error("422 must not be treated as not-found")
	}
}

func TestRateLimitRetry(t *testing.T) {
	attempts := 0
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts == 1 {
			w.Header().Set("Retry-After", "0")
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		_, _ = w.Write([]byte(`[{"id":"1","name":"example.com","destination":"192.0.2.1"}]`))
	})

	if _, err := client.GetRecord(context.Background(), recordTypeA, "example.com", "1"); err != nil {
		t.Fatalf("expected retry to succeed, got %v", err)
	}
	if attempts != 2 {
		t.Errorf("expected 2 attempts, got %d", attempts)
	}
}

func TestGetDNSZone(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/dns/example.com" {
			t.Errorf("unexpected path %q", r.URL.Path)
		}
		// The API wraps the zone in a one-element array.
		_, _ = w.Write([]byte(`[{"name":"example.com","active":true,"ipv6":true,"dnssec":false}]`))
	})

	zone, err := client.GetDNSZone(context.Background(), "example.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if zone.Name != "example.com" || !zone.Active {
		t.Errorf("unexpected zone: %+v", zone)
	}
}
