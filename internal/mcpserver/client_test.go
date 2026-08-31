package mcpserver

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAPIClientForwardsAuthAndIdempotency(t *testing.T) {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/gas-orders/quote" {
			t.Fatalf("path = %s, want /v1/gas-orders/quote", r.URL.Path)
		}
		if got := r.Header.Get("X-API-Key"); got != "gs_test_123" {
			t.Fatalf("X-API-Key = %q", got)
		}
		if got := r.Header.Get("X-Wallet-Session-Id"); got != "ws_123" {
			t.Fatalf("X-Wallet-Session-Id = %q", got)
		}
		if got := r.Header.Get("Idempotency-Key"); got == "" {
			t.Fatal("Idempotency-Key is empty")
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"quoteId": "gq_123"})
	}))
	defer backend.Close()

	client := NewAPIClient(backend.URL, "gs_test_123", "ws_123")
	result, err := client.Post(context.Background(), "/v1/gas-orders/quote", map[string]any{"gasTimes": 1}, true)
	if err != nil {
		t.Fatal(err)
	}
	if got := result["quoteId"]; got != "gq_123" {
		t.Fatalf("quoteId = %v", got)
	}
}

func TestAPIClientReturnsAPIError(t *testing.T) {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"error": map[string]any{
				"code":    "invalid_gas_times",
				"message": "gasTimes must be between 1 and 100",
			},
		})
	}))
	defer backend.Close()

	client := NewAPIClient(backend.URL, "", "")
	_, err := client.Post(context.Background(), "/v1/gas-orders/quote", map[string]any{"gasTimes": 0}, true)
	if err == nil {
		t.Fatal("expected error")
	}
	if got := err.Error(); got != "gas up API error status=400 code=invalid_gas_times message=gasTimes must be between 1 and 100" {
		t.Fatalf("error = %q", got)
	}
}
