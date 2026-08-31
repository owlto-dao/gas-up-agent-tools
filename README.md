# Gas Up Agent Tools

MCP server and portable Agent Skill for Gas Up native gas top-up workflows.

This repository contains a thin adapter over the Gas Up backend API. It does not contain pricing, treasury, payment verification, or order dispatch business logic.

## Hosted MCP

Most users do not need to run this repository. Use the hosted MCP endpoint:

```text
https://gasup-mcp.owlto.finance/mcp
```

MCP requests must include both response types in the `Accept` header:

```text
Accept: application/json, text/event-stream
```

Hosted MCP supports request-scoped user credentials:

```text
X-Wallet-Session-Id: ws_...
```

or:

```text
X-API-Key: gs_live_...
```

or:

```text
Authorization: Bearer gs_live_...
```

For normal user-facing integrations, use wallet sessions created by the Gas Up wallet sign-in flow. The hosted MCP service keeps its backend service credential on the server side; users should never receive that credential.

## Security

Never ask users to paste private keys, seed phrases, or unrelated account credentials into an agent client.

Do not configure `GAS_UP_API_KEY` for public user-facing deployments unless the deployment is controlled by your own trusted backend. That variable is only a server-side fallback credential. Public deployments should use request-scoped user credentials so permissions, rate limits, and audit logs stay tied to the correct user.

Agents should ask for explicit user confirmation before quote locking, order creation, and payment submission.

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

Recommended flow:

1. Call `list_gas_prices` and `list_payment_methods`.
2. Call `preview_gas_order_quote`.
3. Ask the user to confirm recipients, gas units, payment chain, payment asset, payment amount, payment recipient, and quote expiration.
4. Call `create_gas_order_quote`.
5. Call `create_gas_order` with the exact same quote fields.
6. Ask the user to send payment when required.
7. Call `submit_gas_order_payment` after the user provides the payment transaction hash.
8. Poll `get_gas_order_dispatch_summary`.

See `USAGE.md` for a user-facing integration guide.

## Agent Skill

The portable Agent Skill is in:

```text
agent-skills/gas-up-agent
```

Package that folder with the hosted MCP endpoint when distributing the workflow to Skill-compatible agents.

## Run Locally

```bash
cp .env.example .env
set -a
source ./.env
set +a
go run ./cmd/mcp
```

The default local MCP endpoint is:

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
- `GAS_UP_API_KEY`: optional server-side fallback API key for trusted private deployments

For self-hosted public deployments, keep `GAS_UP_API_KEY` empty unless your own backend controls all traffic to the MCP server. If `GAS_UP_API_KEY` is empty, callers must provide their own API key for write tools that require backend API-key authorization.

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

Recommended production endpoint:

```text
https://gasup-mcp.owlto.finance/mcp
```

## License

MIT
