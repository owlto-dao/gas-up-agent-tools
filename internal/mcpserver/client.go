package mcpserver

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const defaultTimeout = 30 * time.Second

type APIClient struct {
	baseURL         string
	apiKey          string
	walletSessionID string
	httpClient      *http.Client
}

func NewAPIClient(baseURL, apiKey, walletSessionID string) *APIClient {
	return &APIClient{
		baseURL:         strings.TrimRight(baseURL, "/"),
		apiKey:          strings.TrimSpace(apiKey),
		walletSessionID: strings.TrimSpace(walletSessionID),
		httpClient:      &http.Client{Timeout: defaultTimeout},
	}
}

func (c *APIClient) Get(ctx context.Context, path string, query url.Values) (map[string]any, error) {
	return c.do(ctx, http.MethodGet, path, query, nil, "")
}

func (c *APIClient) Post(ctx context.Context, path string, body any, idempotent bool) (map[string]any, error) {
	idempotencyKey := ""
	if idempotent {
		idempotencyKey = newID("mcp")
	}
	return c.do(ctx, http.MethodPost, path, nil, body, idempotencyKey)
}

func (c *APIClient) do(ctx context.Context, method, path string, query url.Values, body any, idempotencyKey string) (map[string]any, error) {
	if c.baseURL == "" {
		return nil, fmt.Errorf("GAS_UP_API_BASE_URL is required")
	}
	endpoint := c.baseURL + path
	if len(query) > 0 {
		endpoint += "?" + query.Encode()
	}

	var reader io.Reader
	if body != nil {
		encoded, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		reader = bytes.NewReader(encoded)
	}

	req, err := http.NewRequestWithContext(ctx, method, endpoint, reader)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "gas-up-mcp/0.1")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if c.apiKey != "" {
		req.Header.Set("X-API-Key", c.apiKey)
	}
	if c.walletSessionID != "" {
		req.Header.Set("X-Wallet-Session-Id", c.walletSessionID)
	}
	if idempotencyKey != "" {
		req.Header.Set("Idempotency-Key", idempotencyKey)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var payload map[string]any
	if len(raw) > 0 {
		if err := json.Unmarshal(raw, &payload); err != nil {
			return nil, fmt.Errorf("gas up API returned non-JSON response status=%d body=%q", resp.StatusCode, string(raw))
		}
	} else {
		payload = map[string]any{}
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, apiError(resp.StatusCode, payload)
	}
	return payload, nil
}

func apiError(status int, payload map[string]any) error {
	if raw, ok := payload["error"].(map[string]any); ok {
		code, _ := raw["code"].(string)
		message, _ := raw["message"].(string)
		if code != "" || message != "" {
			return fmt.Errorf("gas up API error status=%d code=%s message=%s", status, code, message)
		}
	}
	encoded, _ := json.Marshal(payload)
	return fmt.Errorf("gas up API error status=%d body=%s", status, string(encoded))
}

func newID(prefix string) string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return fmt.Sprintf("%s-%d", prefix, time.Now().UnixNano())
	}
	return prefix + "-" + hex.EncodeToString(b[:])
}
