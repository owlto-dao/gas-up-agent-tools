package mcpserver

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHTTPHandlerListsTools(t *testing.T) {
	backend := httptest.NewServer(http.NotFoundHandler())
	defer backend.Close()

	handler := NewHTTPHandler(Options{BackendBaseURL: backend.URL})

	response := postMCP(t, handler, map[string]any{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "initialize",
		"params": map[string]any{
			"protocolVersion": "2026-07-28",
			"capabilities":    map[string]any{},
			"clientInfo":      map[string]any{"name": "test", "version": "1.0"},
		},
	})
	result := response["result"].(map[string]any)
	if result["serverInfo"].(map[string]any)["name"] != "gas-up-mcp" {
		t.Fatalf("serverInfo.name = %v", result["serverInfo"])
	}

	response = postMCP(t, handler, map[string]any{
		"jsonrpc": "2.0",
		"id":      2,
		"method":  "tools/list",
		"params":  map[string]any{},
	})
	result = response["result"].(map[string]any)
	tools := result["tools"].([]any)
	if len(tools) < 10 {
		t.Fatalf("got %d tools, want at least 10", len(tools))
	}
	found := false
	for _, item := range tools {
		tool := item.(map[string]any)
		if tool["name"] == "create_gas_order_quote" {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("create_gas_order_quote tool was not listed")
	}
}

func TestHTTPHandlerForwardsBearerAPIKey(t *testing.T) {
	var gotAPIKey string
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAPIKey = r.Header.Get("X-API-Key")
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"items": []any{}})
	}))
	defer backend.Close()

	handler := NewHTTPHandler(Options{BackendBaseURL: backend.URL})
	reqBody, _ := json.Marshal(map[string]any{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "tools/call",
		"params": map[string]any{
			"name":      "list_gas_prices",
			"arguments": map[string]any{},
		},
	})
	req := httptest.NewRequest(http.MethodPost, "/mcp", bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json, text/event-stream")
	req.Header.Set("Authorization", "Bearer gs_test_forwarded")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
	if gotAPIKey != "gs_test_forwarded" {
		t.Fatalf("forwarded API key = %q", gotAPIKey)
	}
}

func TestHTTPHandlerCopiesWalletSessionHeaderIntoQuoteBody(t *testing.T) {
	var gotSessionHeader string
	var gotWalletSessionID string
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotSessionHeader = r.Header.Get("X-Wallet-Session-Id")
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatal(err)
		}
		gotWalletSessionID, _ = body["walletSessionId"].(string)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"ok": true})
	}))
	defer backend.Close()

	handler := NewHTTPHandler(Options{BackendBaseURL: backend.URL})
	reqBody, _ := json.Marshal(map[string]any{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "tools/call",
		"params": map[string]any{
			"name": "preview_gas_order_quote",
			"arguments": map[string]any{
				"targetChainId":      "56",
				"receivingAddresses": []string{"0xDbe00f4f94B8931236713Be25E2798e24db0d648"},
				"gasTimes":           1,
				"paymentMethod":      map[string]any{"paymentChainId": "8453", "asset": "ETH"},
				"deliverySchedule":   map[string]any{"mode": "immediate"},
				"executionGasPrice":  map[string]any{"mode": "realtime", "unit": "gwei"},
			},
		},
	})
	req := httptest.NewRequest(http.MethodPost, "/mcp", bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json, text/event-stream")
	req.Header.Set("X-Wallet-Session-Id", "ws_test_session")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
	if gotSessionHeader != "ws_test_session" {
		t.Fatalf("forwarded wallet session header = %q", gotSessionHeader)
	}
	if gotWalletSessionID != "ws_test_session" {
		t.Fatalf("walletSessionId body = %q", gotWalletSessionID)
	}
}

func postMCP(t *testing.T, handler http.Handler, payload map[string]any) map[string]any {
	t.Helper()
	reqBody, err := json.Marshal(payload)
	if err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest(http.MethodPost, "/mcp", bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json, text/event-stream")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)
	body, _ := io.ReadAll(rec.Result().Body)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", rec.Code, string(body))
	}
	var response map[string]any
	if err := json.Unmarshal(body, &response); err != nil {
		t.Fatalf("unmarshal response: %v body=%s", err, string(body))
	}
	if response["error"] != nil {
		t.Fatalf("mcp error: %v", response["error"])
	}
	return response
}
