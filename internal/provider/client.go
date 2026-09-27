package provider

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"sync"
	"time"
)

const (
	defaultBaseURL = "https://api.zone.eu/v2"

	// Rate limiting constants
	defaultRateLimit     = 60 // requests per minute
	rateLimitResetPeriod = time.Minute
	maxRetries           = 3
)

// DNS record types, as they appear in the /dns/{zone}/{type} API paths.
const (
	recordTypeA     = "a"
	recordTypeAAAA  = "aaaa"
	recordTypeCAA   = "caa"
	recordTypeCNAME = "cname"
	recordTypeMX    = "mx"
	recordTypeNS    = "ns"
	recordTypeSRV   = "srv"
	recordTypeSSHFP = "sshfp"
	recordTypeTLSA  = "tlsa"
	recordTypeTXT   = "txt"
	recordTypeURL   = "url"
)

// errNotFound is returned when the API answers 404, or answers a single-object
// GET with an empty array.
var errNotFound = errors.New("not found")

// APIError is a non-2xx response from the Zone.EU API.
//
// Validation errors (422) come back shaped like the resource being written,
// e.g. {"name":"invalid_host"}, not like the ErrorResponse schema the spec
// declares, so the body is kept verbatim rather than parsed.
type APIError struct {
	StatusCode    int
	Body          string
	StatusMessage string
}

func (e *APIError) Error() string {
	if e.StatusMessage != "" {
		return fmt.Sprintf("API error (status %d): %s (X-Status-Message: %s)", e.StatusCode, e.Body, e.StatusMessage)
	}
	return fmt.Sprintf("API error (status %d): %s", e.StatusCode, e.Body)
}

func (e *APIError) Unwrap() error {
	if e.StatusCode == http.StatusNotFound {
		return errNotFound
	}
	return nil
}

// isNotFound reports whether err means the requested object does not exist.
func isNotFound(err error) bool {
	return errors.Is(err, errNotFound)
}

// Client represents the Zone.EU API client
type Client struct {
	httpClient *http.Client
	baseURL    string
	username   string
	apiKey     string

	// Rate limiting
	mu                 sync.Mutex
	rateLimitLimit     int
	rateLimitRemaining int
	rateLimitResetAt   time.Time
}

// NewClient creates a new Zone.EU API client
func NewClient(username, apiKey string) *Client {
	return &Client{
		httpClient:         &http.Client{Timeout: 30 * time.Second},
		baseURL:            defaultBaseURL,
		username:           username,
		apiKey:             apiKey,
		rateLimitLimit:     defaultRateLimit,
		rateLimitRemaining: defaultRateLimit,
	}
}

// authHeader returns the Basic Auth header value
func (c *Client) authHeader() string {
	auth := c.username + ":" + c.apiKey
	return "Basic " + base64.StdEncoding.EncodeToString([]byte(auth))
}

// parseSingle parses a single-object response. The API wraps single objects
// in a one-element array; an empty array means the object does not exist.
func parseSingle[T any](resp []byte) (*T, error) {
	var items []T
	if err := json.Unmarshal(resp, &items); err != nil {
		return nil, fmt.Errorf("error parsing response: %w", err)
	}
	if len(items) == 0 {
		return nil, errNotFound
	}
	return &items[0], nil
}

// updateRateLimitInfo updates rate limit info from response headers
func (c *Client) updateRateLimitInfo(resp *http.Response) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if limit := resp.Header.Get("X-Ratelimit-Limit"); limit != "" {
		if val, err := strconv.Atoi(limit); err == nil {
			c.rateLimitLimit = val
		}
	}

	if remaining := resp.Header.Get("X-Ratelimit-Remaining"); remaining != "" {
		if val, err := strconv.Atoi(remaining); err == nil {
			c.rateLimitRemaining = val
		}
	}
}

// waitForRateLimit waits if we've hit the rate limit
func (c *Client) waitForRateLimit(ctx context.Context) error {
	c.mu.Lock()
	remaining, resetAt := c.rateLimitRemaining, c.rateLimitResetAt
	c.mu.Unlock()

	if remaining > 0 {
		return nil
	}
	return sleepCtx(ctx, time.Until(resetAt))
}

// sleepCtx sleeps for d or until ctx is cancelled.
func sleepCtx(ctx context.Context, d time.Duration) error {
	if d <= 0 {
		return nil
	}
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-t.C:
		return nil
	}
}

// doRequest performs an HTTP request with authentication, rate limiting, and context support
func (c *Client) doRequest(ctx context.Context, method, path string, body interface{}) ([]byte, error) {
	var lastErr error

	for attempt := 0; attempt < maxRetries; attempt++ {
		if err := c.waitForRateLimit(ctx); err != nil {
			return nil, err
		}

		result, err := c.doRequestOnce(ctx, method, path, body)
		if err == nil {
			return result, nil
		}

		var rateLimitErr *RateLimitError
		if !errors.As(err, &rateLimitErr) {
			return nil, err
		}

		c.mu.Lock()
		c.rateLimitRemaining = 0
		c.rateLimitResetAt = time.Now().Add(rateLimitErr.RetryAfter)
		c.mu.Unlock()
		lastErr = err
	}

	return nil, fmt.Errorf("max retries exceeded: %w", lastErr)
}

// RateLimitError represents a rate limit error from the API
type RateLimitError struct {
	RetryAfter time.Duration
	Message    string
}

func (e *RateLimitError) Error() string {
	return fmt.Sprintf("rate limit exceeded, retry after %v: %s", e.RetryAfter, e.Message)
}

// doRequestOnce performs a single HTTP request
func (c *Client) doRequestOnce(ctx context.Context, method, path string, body interface{}) ([]byte, error) {
	var bodyReader io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("error marshaling request body: %w", err)
		}
		bodyReader = bytes.NewBuffer(jsonBody)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Authorization", c.authHeader())
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error performing request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	c.updateRateLimitInfo(resp)

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %w", err)
	}

	if resp.StatusCode == http.StatusTooManyRequests {
		retryAfter := rateLimitResetPeriod
		if retryHeader := resp.Header.Get("Retry-After"); retryHeader != "" {
			if seconds, err := strconv.Atoi(retryHeader); err == nil {
				retryAfter = time.Duration(seconds) * time.Second
			}
		}
		return nil, &RateLimitError{
			RetryAfter: retryAfter,
			Message:    resp.Header.Get("X-Status-Message"),
		}
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, &APIError{
			StatusCode:    resp.StatusCode,
			Body:          string(respBody),
			StatusMessage: resp.Header.Get("X-Status-Message"),
		}
	}

	return respBody, nil
}

// DNSRecord represents a generic DNS record.
//
// Name is always a fully-qualified name: the API returns FQDNs (the apex as the
// bare zone name) and rejects short names on write with 422 invalid_host.
// There is no TTL field; Zone.EU fixes TTL server-side.
type DNSRecord struct {
	ID          string `json:"id,omitempty"`
	Name        string `json:"name"`
	Destination string `json:"destination"`
	// For MX records
	Priority int `json:"priority,omitempty"`
	// For SRV records
	Weight int `json:"weight,omitempty"`
	Port   int `json:"port,omitempty"`
	// For CAA records
	Flag int    `json:"flag,omitempty"`
	Tag  string `json:"tag,omitempty"`
	// For TLSA records
	CertificateUsage int `json:"certificate_usage,omitempty"`
	Selector         int `json:"selector,omitempty"`
	MatchingType     int `json:"matching_type,omitempty"`
	// For SSHFP records (algorithm, type) and URL records (type=redirect code)
	Algorithm int `json:"algorithm,omitempty"`
	Type      int `json:"type,omitempty"`
}

// DNSZone represents a DNS zone
type DNSZone struct {
	Name   string `json:"name"`
	Active bool   `json:"active"`
	IPv6   bool   `json:"ipv6"`
}

func recordsPath(zone, recordType string) string {
	return fmt.Sprintf("/dns/%s/%s", url.PathEscape(zone), recordType)
}

func recordPath(zone, recordType, id string) string {
	return recordsPath(zone, recordType) + "/" + url.PathEscape(id)
}

// ==================== DNS Records ====================

// ListRecords retrieves all records of a type in a zone. DNS record lists are
// not paginated.
func (c *Client) ListRecords(ctx context.Context, recordType, zone string) ([]DNSRecord, error) {
	resp, err := c.doRequest(ctx, http.MethodGet, recordsPath(zone, recordType), nil)
	if err != nil {
		return nil, err
	}
	var records []DNSRecord
	if err := json.Unmarshal(resp, &records); err != nil {
		return nil, fmt.Errorf("error parsing response: %w", err)
	}
	return records, nil
}

func (c *Client) GetRecord(ctx context.Context, recordType, zone, id string) (*DNSRecord, error) {
	resp, err := c.doRequest(ctx, http.MethodGet, recordPath(zone, recordType, id), nil)
	if err != nil {
		return nil, err
	}
	return parseSingle[DNSRecord](resp)
}

func (c *Client) CreateRecord(ctx context.Context, recordType, zone string, record *DNSRecord) (*DNSRecord, error) {
	resp, err := c.doRequest(ctx, http.MethodPost, recordsPath(zone, recordType), record)
	if err != nil {
		return nil, err
	}
	return parseSingle[DNSRecord](resp)
}

// UpdateRecord updates a record. PUT merges: fields omitted from the body keep
// their current value.
func (c *Client) UpdateRecord(ctx context.Context, recordType, zone, id string, record *DNSRecord) (*DNSRecord, error) {
	resp, err := c.doRequest(ctx, http.MethodPut, recordPath(zone, recordType, id), record)
	if err != nil {
		return nil, err
	}
	return parseSingle[DNSRecord](resp)
}

func (c *Client) DeleteRecord(ctx context.Context, recordType, zone, id string) error {
	_, err := c.doRequest(ctx, http.MethodDelete, recordPath(zone, recordType, id), nil)
	return err
}

// ==================== DNS Zone ====================

func (c *Client) GetDNSZone(ctx context.Context, zone string) (*DNSZone, error) {
	resp, err := c.doRequest(ctx, http.MethodGet, "/dns/"+url.PathEscape(zone), nil)
	if err != nil {
		return nil, err
	}
	return parseSingle[DNSZone](resp)
}

// ==================== Domain ====================

// Domain represents a domain in Zone.EU
type Domain struct {
	ResourceURL          string `json:"resource_url,omitempty"`
	Name                 string `json:"name"`
	Delegated            string `json:"delegated,omitempty"`
	Expires              string `json:"expires,omitempty"`
	DNSSEC               bool   `json:"dnssec"`
	Autorenew            bool   `json:"autorenew"`
	RenewOrder           string `json:"renew_order,omitempty"`
	RenewalNotifications bool   `json:"renewal_notifications"`
	HasPendingTrade      *int   `json:"has_pending_trade,omitempty"`
	HasPendingDNSSEC     bool   `json:"has_pending_dnssec,omitempty"`
	Reactivate           bool   `json:"reactivate,omitempty"`
	AuthKeyEnabled       bool   `json:"auth_key_enabled,omitempty"`
	SigningRequired      bool   `json:"signing_required,omitempty"`
	NameserversCustom    bool   `json:"nameservers_custom"`
}

// DomainPreferences represents domain preferences
type DomainPreferences struct {
	ResourceURL          string `json:"resource_url,omitempty"`
	RenewalNotifications bool   `json:"renewal_notifications"`
}

// GetDomain retrieves a specific domain
func (c *Client) GetDomain(ctx context.Context, name string) (*Domain, error) {
	resp, err := c.doRequest(ctx, http.MethodGet, "/domain/"+url.PathEscape(name), nil)
	if err != nil {
		return nil, err
	}
	return parseSingle[Domain](resp)
}

// GetDomainPreferences retrieves domain preferences
func (c *Client) GetDomainPreferences(ctx context.Context, name string) (*DomainPreferences, error) {
	resp, err := c.doRequest(ctx, http.MethodGet, "/domain/"+url.PathEscape(name)+"/preferences", nil)
	if err != nil {
		return nil, err
	}
	return parseSingle[DomainPreferences](resp)
}
