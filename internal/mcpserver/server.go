package mcpserver

import (
	"context"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

const defaultBackendBaseURL = "http://localhost:4000"

type Options struct {
	BackendBaseURL string
	DefaultAPIKey  string
}

func OptionsFromEnv() Options {
	return Options{
		BackendBaseURL: env("GAS_UP_API_BASE_URL", defaultBackendBaseURL),
		DefaultAPIKey:  os.Getenv("GAS_UP_API_KEY"),
	}
}

func NewHTTPHandler(opts Options) http.Handler {
	return mcp.NewStreamableHTTPHandler(func(r *http.Request) *mcp.Server {
		return NewServer(clientFromRequest(r, opts))
	}, &mcp.StreamableHTTPOptions{
		Stateless:                    true,
		JSONResponse:                 true,
		PropagateRequestCancellation: true,
	})
}

func NewServer(api *APIClient) *mcp.Server {
	server := mcp.NewServer(&mcp.Implementation{
		Name:    "gas-up-mcp",
		Version: "0.1.0",
	}, &mcp.ServerOptions{
		Instructions: instructions,
		Capabilities: &mcp.ServerCapabilities{},
	})
	registerTools(server, api)
	return server
}

func clientFromRequest(r *http.Request, opts Options) *APIClient {
	apiKey := strings.TrimSpace(r.Header.Get("X-API-Key"))
	if apiKey == "" {
		apiKey = apiKeyFromAuthorization(r.Header.Get("Authorization"))
	}
	if apiKey == "" {
		apiKey = opts.DefaultAPIKey
	}
	return NewAPIClient(opts.BackendBaseURL, apiKey, r.Header.Get("X-Wallet-Session-Id"))
}

func apiKeyFromAuthorization(header string) string {
	header = strings.TrimSpace(header)
	if header == "" {
		return ""
	}
	parts := strings.Fields(header)
	if len(parts) == 2 && strings.EqualFold(parts[0], "Bearer") && strings.HasPrefix(parts[1], "gs_") {
		return parts[1]
	}
	return ""
}

func registerTools(server *mcp.Server, api *APIClient) {
	readOnly := true
	notReadOnly := false
	idempotent := true
	nonDestructive := false

	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_gas_prices",
		Title:       "List Gas Prices",
		Description: "List target chains that can receive gas top-ups, including gas price guard and treasury inventory status. Use this before creating a quote.",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: readOnly, IdempotentHint: true},
	}, func(ctx context.Context, _ *mcp.CallToolRequest, _ emptyInput) (*mcp.CallToolResult, any, error) {
		return toolGet(ctx, api, "/v1/gas-prices", nil)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_payment_methods",
		Title:       "List Payment Methods",
		Description: "List enabled payment chains and assets. Use paymentChainId and asset from this tool when quoting or creating gas orders.",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: readOnly, IdempotentHint: true},
	}, func(ctx context.Context, _ *mcp.CallToolRequest, _ emptyInput) (*mcp.CallToolResult, any, error) {
		return toolGet(ctx, api, "/v1/payment-methods", nil)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "estimate_gas_amounts",
		Title:       "Estimate Gas Amounts",
		Description: "Estimate native gas amount per recipient for a target chain without creating a quote or reserving inventory.",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: readOnly, IdempotentHint: true},
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input estimateGasInput) (*mcp.CallToolResult, any, error) {
		query := url.Values{"targetChainId": []string{input.TargetChainID}}
		if input.CustomGasTimes > 0 {
			query.Set("customGasTimes", strconv.Itoa(input.CustomGasTimes))
		}
		return toolGet(ctx, api, "/v1/gas-estimates", query)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_platform_fee_discount_rules",
		Title:       "Get Platform Fee Rules",
		Description: "Return the platform fee tiers used to calculate gas top-up quotes.",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: readOnly, IdempotentHint: true},
	}, func(ctx context.Context, _ *mcp.CallToolRequest, _ emptyInput) (*mcp.CallToolResult, any, error) {
		return toolGet(ctx, api, "/v1/platform-fee/discount-rules", nil)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_wallet_benefits",
		Title:       "List Wallet Benefits",
		Description: "List available and reserved free gas benefits for the authenticated wallet. Requires X-Wallet-Session-Id or a wallet-bound API key.",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: readOnly, IdempotentHint: true},
	}, func(ctx context.Context, _ *mcp.CallToolRequest, _ emptyInput) (*mcp.CallToolResult, any, error) {
		return toolGet(ctx, api, "/v1/wallet-benefits", nil)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "preview_gas_order_quote",
		Title:       "Preview Gas Order Quote",
		Description: "Return a live gas order quote preview without locking a quote. Use this before asking the user to confirm a paid gas top-up.",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: readOnly, IdempotentHint: true},
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input gasQuoteInput) (*mcp.CallToolResult, any, error) {
		return toolPost(ctx, api, "/v1/gas-orders/check", input, false)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "create_gas_order_quote",
		Title:       "Create Gas Order Quote",
		Description: "Create a locked quote for a gas top-up order. The next create_gas_order call must reuse the same targetChainId, recipients, gasTimes, paymentMethod, deliverySchedule, and executionGasPrice.",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: notReadOnly, IdempotentHint: true, DestructiveHint: &nonDestructive},
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input gasQuoteInput) (*mcp.CallToolResult, any, error) {
		return toolPost(ctx, api, "/v1/gas-orders/quote", input, idempotent)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "create_gas_order",
		Title:       "Create Gas Order",
		Description: "Create a gas top-up order from a locked quote. Ask the user to confirm before using this tool. The request must exactly match the locked quote except orderName.",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: notReadOnly, IdempotentHint: true, DestructiveHint: &nonDestructive},
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input createGasOrderInput) (*mcp.CallToolResult, any, error) {
		return toolPost(ctx, api, "/v1/gas-orders", input, idempotent)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_gas_orders",
		Title:       "List Gas Orders",
		Description: "List gas top-up orders for the authenticated wallet. Requires X-Wallet-Session-Id or a wallet-bound API key.",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: readOnly, IdempotentHint: true},
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input paginationInput) (*mcp.CallToolResult, any, error) {
		query := paginationQuery(input, 50, 100)
		return toolGet(ctx, api, "/v1/gas-orders", query)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_gas_order",
		Title:       "Get Gas Order",
		Description: "Get a gas top-up order and paginated gas unit details for the authenticated wallet.",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: readOnly, IdempotentHint: true},
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input getGasOrderInput) (*mcp.CallToolResult, any, error) {
		query := paginationQuery(input.Pagination, 200, 1000)
		return toolGet(ctx, api, "/v1/gas-orders/"+url.PathEscape(input.OrderID), query)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_gas_order_dispatch_summary",
		Title:       "Get Dispatch Summary",
		Description: "Get lightweight gas order dispatch progress. Poll every 3 to 5 seconds while an order is active.",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: readOnly, IdempotentHint: true},
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input orderIDInput) (*mcp.CallToolResult, any, error) {
		return toolGet(ctx, api, "/v1/gas-orders/"+url.PathEscape(input.OrderID)+"/dispatch-summary", nil)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "submit_gas_order_payment",
		Title:       "Submit Gas Order Payment",
		Description: "Submit a payment transaction hash for verification. Requires X-Wallet-Session-Id because the backend currently requires wallet-session authentication for payment submission.",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: notReadOnly, IdempotentHint: true, DestructiveHint: &nonDestructive},
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input submitPaymentInput) (*mcp.CallToolResult, any, error) {
		return toolPost(ctx, api, "/v1/gas-orders/"+url.PathEscape(input.OrderID)+"/pay", input.Payment, idempotent)
	})
}

func toolGet(ctx context.Context, api *APIClient, path string, query url.Values) (*mcp.CallToolResult, any, error) {
	output, err := api.Get(ctx, path, query)
	return nil, output, err
}

func toolPost(ctx context.Context, api *APIClient, path string, body any, idempotent bool) (*mcp.CallToolResult, any, error) {
	output, err := api.Post(ctx, path, body, idempotent)
	return nil, output, err
}

func paginationQuery(input paginationInput, defaultLimit, maxLimit int) url.Values {
	limit := input.Limit
	if limit <= 0 {
		limit = defaultLimit
	}
	if limit > maxLimit {
		limit = maxLimit
	}
	if input.Offset < 0 {
		input.Offset = 0
	}
	return url.Values{
		"limit":  []string{strconv.Itoa(limit)},
		"offset": []string{strconv.Itoa(input.Offset)},
	}
}

func env(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}

const instructions = `Gas Up MCP exposes tools for cross-chain native gas top-ups.

Use list_gas_prices and list_payment_methods before quoting.
Only create orders for target chains where status is "open" and acceptingNewOrders is true.
Use preview_gas_order_quote for estimates, then create_gas_order_quote only after the user is ready to lock a quote.
Before create_gas_order, ask the user to confirm recipient addresses, gasTimes, payment amount, payment asset, payment recipient, and quote expiration.
The create_gas_order payload must exactly match the locked quote for targetChainId, receivingAddresses/receivingCsv, gasTimes, paymentMethod, deliverySchedule, and executionGasPrice.
Amounts and *Raw fields are strings and must not be converted through floating point numbers.
submit_gas_order_payment requires X-Wallet-Session-Id because payment verification is wallet-session scoped.`
