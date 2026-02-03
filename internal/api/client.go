package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

// Client represents a Joplin API client
type Client struct {
	baseURL    string
	token      string
	httpClient *http.Client
}

// NewClient creates a new Joplin API client
func NewClient(host string, port int, token string, timeout time.Duration) *Client {
	return &Client{
		baseURL: fmt.Sprintf("http://%s:%d", host, port),
		token:   token,
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}
}

// Ping checks if the Joplin API is reachable
func (c *Client) Ping(ctx context.Context) error {
	endpoint := fmt.Sprintf("%s/ping?token=%s", c.baseURL, url.QueryEscape(c.token))

	req, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("cannot connect to Joplin. Is the desktop app running with Web Clipper enabled? %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		return fmt.Errorf("authentication failed. Check your API token")
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	// Joplin returns "JoplinClipperServer" as plain text
	if string(body) != "JoplinClipperServer" {
		return fmt.Errorf("unexpected response from ping endpoint: %s", string(body))
	}

	return nil
}

// doRequestWithRetry performs an HTTP request with exponential backoff retry
func (c *Client) doRequestWithRetry(ctx context.Context, method, endpoint string, body io.Reader) (*http.Response, error) {
	maxRetries := 3
	baseDelay := 500 * time.Millisecond

	var lastErr error
	for attempt := 0; attempt < maxRetries; attempt++ {
		if attempt > 0 {
			delay := baseDelay * time.Duration(1<<uint(attempt-1))
			select {
			case <-time.After(delay):
			case <-ctx.Done():
				return nil, ctx.Err()
			}
		}

		req, err := http.NewRequestWithContext(ctx, method, endpoint, body)
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}

		if body != nil {
			req.Header.Set("Content-Type", "application/json")
		}

		resp, err := c.httpClient.Do(req)
		if err != nil {
			lastErr = err
			continue
		}

		if resp.StatusCode >= 500 {
			resp.Body.Close()
			lastErr = fmt.Errorf("server error: %d", resp.StatusCode)
			continue
		}

		return resp, nil
	}

	return nil, fmt.Errorf("request failed after %d attempts: %w", maxRetries, lastErr)
}

// get performs a GET request
func (c *Client) get(ctx context.Context, path string, params url.Values) (*http.Response, error) {
	if params == nil {
		params = url.Values{}
	}
	params.Set("token", c.token)

	endpoint := fmt.Sprintf("%s%s?%s", c.baseURL, path, params.Encode())
	return c.doRequestWithRetry(ctx, "GET", endpoint, nil)
}

// parseJSON reads and parses JSON response
func parseJSON(resp *http.Response, v interface{}) error {
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		return fmt.Errorf("authentication failed. Check your API token")
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("unexpected status code %d: %s", resp.StatusCode, string(body))
	}

	if err := json.NewDecoder(resp.Body).Decode(v); err != nil {
		return fmt.Errorf("failed to decode JSON: %w", err)
	}

	return nil
}
