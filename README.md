# Gas Up Agent Tools

MCP server and portable Agent Skill for Gas Up native gas top-up workflows.

The MCP server is a thin adapter over the Gas Up backend API. It does not contain order, pricing, treasury, or payment verification business logic.

## Run Locally

```bash
cp .env.example .env
set -a
source ./.env
set +a
go run ./cmd/mcp
```

The default MCP endpoint is:

```text
http://localhost:4010/mcp
```

Health check:

```bash
curl http://localhost:4010/health
```

## Configuration

- `MCP_HOST`: listen host, default `0.0.0.0`
- `MCP_PORT`: listen port, default `4010`
- `GAS_UP_API_BASE_URL`: Gas Up backend API base URL, default `http://localhost:4000`
- `GAS_UP_API_KEY`: optional fallback API key for internal deployments only

For user-facing deployments, prefer request-scoped credentials:

```text
X-API-Key: gs_live_...
```

or:

```text
Authorization: Bearer gs_live_...
```

Wallet-session flows can also pass:

```text
X-Wallet-Session-Id: ws_...
```

## MCP Tools

- `list_gas_prices`
- `list_payment_methods`
- `estimate_gas_amounts`
- `get_platform_fee_discount_rules`
- `list_wallet_benefits`
- `preview_gas_order_quote`
- `create_gas_order_quote`
- `create_gas_order`
- `list_gas_orders`
- `get_gas_order`
- `get_gas_order_dispatch_summary`
- `submit_gas_order_payment`

The intended Agent flow is:

1. List gas prices and payment methods.
2. Preview a quote.
3. Ask the user to confirm payment details.
4. Create a locked quote.
5. Create an order using the exact same quote fields.
6. Submit payment and poll the dispatch summary.

## Agent Skill

The portable Skill package is in:

```text
agent-skills/gas-up-agent
```

Package that folder with the MCP endpoint configuration when distributing the workflow to Skill-compatible agents.

## Deployment

Build:

```bash
go build -o bin/gas-up-mcp ./cmd/mcp
```

Run with PM2:

```bash
set -a
source ./.env
set +a
pm2 start ./bin/gas-up-mcp --name gas-up-mcp --update-env
pm2 save
```

Recommended production URL:

```text
https://mcp.gasup.owlto.finance/mcp
```
