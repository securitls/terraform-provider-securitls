package securitls

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"
)

type Client struct {
	BaseURL    string
	APIKey     string
	HTTPClient *http.Client

	mu        sync.Mutex
	jwt       string
	jwtExpiry time.Time
}

func NewClient(baseURL, apiKey string) *Client {
	return &Client{
		BaseURL: strings.TrimRight(baseURL, "/"),
		APIKey:  apiKey,
		HTTPClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

func (c *Client) authenticate(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.jwt != "" && time.Now().Before(c.jwtExpiry.Add(-2*time.Minute)) {
		return nil
	}

	payload, _ := json.Marshal(map[string]string{"apikey": c.APIKey})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.BaseURL+"/authenticate", bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("authenticate request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read authenticate response: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("authentication failed: HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var result struct {
		JWT string `json:"jwt"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return fmt.Errorf("decode authenticate response: %w", err)
	}
	if result.JWT == "" {
		return fmt.Errorf("authentication response did not contain jwt")
	}

	c.jwt = result.JWT
	// SecuriTLS currently issues API JWTs for one hour.
	c.jwtExpiry = time.Now().Add(time.Hour)
	return nil
}

func (c *Client) Do(ctx context.Context, method, path string, payload any, out any) (int, error) {
	if err := c.authenticate(ctx); err != nil {
		return 0, err
	}
	return c.doAuthenticated(ctx, method, path, payload, out, true)
}

func (c *Client) doAuthenticated(ctx context.Context, method, path string, payload any, out any, retry401 bool) (int, error) {
	var body io.Reader
	if payload != nil {
		b, err := json.Marshal(payload)
		if err != nil {
			return 0, fmt.Errorf("encode request body: %w", err)
		}
		body = bytes.NewReader(b)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.BaseURL+path, body)
	if err != nil {
		return 0, err
	}
	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	c.mu.Lock()
	token := c.jwt
	c.mu.Unlock()
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return 0, fmt.Errorf("SecuriTLS request failed: %w", err)
	}
	defer resp.Body.Close()

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return resp.StatusCode, fmt.Errorf("read SecuriTLS response: %w", err)
	}

	if resp.StatusCode == http.StatusUnauthorized && retry401 {
		c.mu.Lock()
		c.jwt = ""
		c.jwtExpiry = time.Time{}
		c.mu.Unlock()
		if err := c.authenticate(ctx); err != nil {
			return resp.StatusCode, err
		}
		return c.doAuthenticated(ctx, method, path, payload, out, false)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return resp.StatusCode, fmt.Errorf("SecuriTLS API HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(b)))
	}

	if out != nil && len(bytes.TrimSpace(b)) > 0 {
		if err := json.Unmarshal(b, out); err != nil {
			return resp.StatusCode, fmt.Errorf("decode SecuriTLS response: %w; body=%s", err, string(b))
		}
	}
	return resp.StatusCode, nil
}

func stringFromMap(m map[string]any, keys ...string) string {
	for _, k := range keys {
		if v, ok := m[k]; ok && v != nil {
			return fmt.Sprint(v)
		}
	}
	return ""
}

func boolFromMap(m map[string]any, key string) bool {
	v, ok := m[key]
	if !ok || v == nil {
		return false
	}
	b, _ := v.(bool)
	return b
}

func floatFromMap(m map[string]any, key string) float64 {
	v, ok := m[key]
	if !ok || v == nil {
		return 0
	}
	f, _ := v.(float64)
	return f
}

func mapFromMap(m map[string]any, key string) map[string]any {
	v, ok := m[key]
	if !ok || v == nil {
		return nil
	}
	ret, _ := v.(map[string]any)
	return ret
}
