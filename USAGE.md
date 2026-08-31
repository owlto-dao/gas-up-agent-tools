# Gas Up MCP Usage Guide

This guide explains how to connect an AI agent to Gas Up through MCP, quote native gas top-ups, create orders, submit payment transaction hashes, and track dispatch progress.

## Hosted Endpoint

Use the hosted MCP endpoint:

```text
https://gasup-mcp.owlto.finance/mcp
```

Production clients should use HTTPS.

## Required MCP Headers

Every MCP request must include:

```text
Content-Type: application/json
Accept: application/json, text/event-stream
```

The MCP server returns JSON responses, but the Streamable HTTP MCP transport still requires `text/event-stream` in the `Accept` header. Requests that only send `Accept: application/json` may return `400`.

Agent clients should also send a clear `User-Agent`, for example:

```text
User-Agent: your-agent-name/1.0
```

Some network protection systems may block generic HTTP library user agents.

## Authentication

Gas Up supports request-scoped authentication. The hosted MCP server keeps any service-level backend credential on the server side. Users and agent clients should only pass user-scoped credentials.

### Wallet Session

For normal user-facing integrations, pass the wallet session returned by the Gas Up wallet sign-in flow:

```text
X-Wallet-Session-Id: ws_...
```

This is the recommended authentication mode for consumer-facing agents.

### API Key

Developer integrations may pass a Gas Up API key:

```text
X-API-Key: gs_live_...
```

or:

```text
Authorization: Bearer gs_live_...
```

When a request supplies its own API key, write tools require `gas:write` or `gas_orders:write` scope. If the API key is not bound to a wallet, also pass `X-Wallet-Session-Id` so Gas Up can identify the wallet that owns the order.

On the official hosted MCP endpoint, normal wallet-session requests use Gas Up's server-side service credential. That credential is never exposed to users or agent clients.

Payment submission currently requires `X-Wallet-Session-Id`.

Never ask users to paste wallet private keys, seed phrases, or unrelated account credentials into an agent client.

## Creating a Wallet Session

Wallet sessions are created through the Gas Up wallet sign-in flow.

1. Request a challenge.
2. Ask the user's wallet to sign the returned message with `personal_sign`.
3. Verify the signature.
4. Send the returned `walletSessionId` to the MCP server as `X-Wallet-Session-Id`.

Challenge request:

```bash
curl -X POST https://gasup.owlto.finance/v1/wallet/challenge \
  -H 'Content-Type: application/json' \
  --data '{
    "walletAddress": "0xUserWallet...",
    "chainId": "8453",
    "domain": "gasup.owlto.finance"
  }'
```

Verify request:

```bash
curl -X POST https://gasup.owlto.finance/v1/wallet/verify \
  -H 'Content-Type: application/json' \
  --data '{
    "challengeId": "wc_...",
    "signature": "0xSignatureFromWallet..."
  }'
```

The verify response includes:

```json
{
  "walletSessionId": "ws_...",
  "walletAddress": "0xUserWallet...",
  "expiresAt": "..."
}
```

## Available Tools

| Tool | Purpose | Authentication |
| --- | --- | --- |
| `list_gas_prices` | List target chains, gas price guards, and inventory status. | None |
| `list_payment_methods` | List supported payment chains and assets. | None |
| `estimate_gas_amounts` | Estimate native gas received per recipient. | None |
| `get_platform_fee_discount_rules` | Read platform fee rules. | None |
| `list_wallet_benefits` | List wallet-specific free gas or campaign benefits. | Wallet session or wallet-bound API key |
| `preview_gas_order_quote` | Preview a non-locking quote. | Wallet session or wallet-bound API key |
| `create_gas_order_quote` | Lock a quote for a future order. | Wallet session or wallet-bound API key |
| `create_gas_order` | Create an order from a locked quote. | Wallet session or wallet-bound API key |
| `list_gas_orders` | List the authenticated wallet's orders. | Wallet session or wallet-bound API key |
| `get_gas_order` | Get one order and its gas unit details. | Wallet session or wallet-bound API key |
| `get_gas_order_dispatch_summary` | Get lightweight dispatch progress. | Wallet session or wallet-bound API key |
| `submit_gas_order_payment` | Submit a payment transaction hash. | Wallet session |

## Recommended Agent Flow

1. Call `list_gas_prices`.
2. Call `list_payment_methods`.
3. Select a target chain where `status` is `open` and `acceptingNewOrders` is `true`.
4. Call `estimate_gas_amounts` if the user wants to understand the native gas amount before pricing.
5. Call `preview_gas_order_quote` for a non-locking estimate.
6. Ask the user to confirm recipients, gas times, payment chain, payment asset, payment amount, payment recipient, and quote expiration.
7. Call `create_gas_order_quote` after confirmation.
8. Call `create_gas_order` using the exact same quote fields.
9. If payment is required, ask the user to send the exact `paymentAmountRaw` amount to `paymentRecipient` on the payment chain.
10. Call `submit_gas_order_payment` after the user provides the payment transaction hash.
11. Poll `get_gas_order_dispatch_summary` every 3 to 5 seconds until the order reaches a final or user-action state.

## Quote Matching Rule

The `create_gas_order` request must match the locked quote exactly for these fields:

```text
targetChainId
receivingAddresses
receivingCsv
gasTimes
paymentMethod.paymentChainId
paymentMethod.asset
deliverySchedule
executionGasPrice
```

`orderName` is display-only and may be different.

If the backend returns `quote_request_mismatch`, create a new locked quote with the intended final fields. Do not guess which field changed.

## Example MCP Requests

### Initialize

```bash
curl -X POST https://gasup-mcp.owlto.finance/mcp \
  -H 'Content-Type: application/json' \
  -H 'Accept: application/json, text/event-stream' \
  -H 'User-Agent: example-agent/1.0' \
  --data '{
    "jsonrpc": "2.0",
    "id": 1,
    "method": "initialize",
    "params": {
      "protocolVersion": "2025-11-25",
      "capabilities": {},
      "clientInfo": {
        "name": "example-agent",
        "version": "1.0"
      }
    }
  }'
```

### List Tools

```bash
curl -X POST https://gasup-mcp.owlto.finance/mcp \
  -H 'Content-Type: application/json' \
  -H 'Accept: application/json, text/event-stream' \
  -H 'User-Agent: example-agent/1.0' \
  --data '{
    "jsonrpc": "2.0",
    "id": 2,
    "method": "tools/list",
    "params": {}
  }'
```

### Preview a Quote

This request does not lock a quote or create an order.

```bash
curl -X POST https://gasup-mcp.owlto.finance/mcp \
  -H 'Content-Type: application/json' \
  -H 'Accept: application/json, text/event-stream' \
  -H 'User-Agent: example-agent/1.0' \
  -H 'X-Wallet-Session-Id: ws_...' \
  --data '{
    "jsonrpc": "2.0",
    "id": 3,
    "method": "tools/call",
    "params": {
      "name": "preview_gas_order_quote",
      "arguments": {
        "targetChainId": "56",
        "receivingAddresses": [
          "0xRecipientAddress..."
        ],
        "gasTimes": 1,
        "paymentMethod": {
          "paymentChainId": "8453",
          "asset": "ETH"
        },
        "deliverySchedule": {
          "mode": "immediate"
        },
        "executionGasPrice": {
          "mode": "realtime",
          "unit": "gwei"
        }
      }
    }
  }'
```

### Create a Locked Quote

Call this only after the user confirms the quote details.

```bash
curl -X POST https://gasup-mcp.owlto.finance/mcp \
  -H 'Content-Type: application/json' \
  -H 'Accept: application/json, text/event-stream' \
  -H 'User-Agent: example-agent/1.0' \
  -H 'X-Wallet-Session-Id: ws_...' \
  --data '{
    "jsonrpc": "2.0",
    "id": 4,
    "method": "tools/call",
    "params": {
      "name": "create_gas_order_quote",
      "arguments": {
        "targetChainId": "56",
        "receivingAddresses": [
          "0xRecipientAddress..."
        ],
        "gasTimes": 1,
        "paymentMethod": {
          "paymentChainId": "8453",
          "asset": "ETH"
        },
        "deliverySchedule": {
          "mode": "immediate"
        },
        "executionGasPrice": {
          "mode": "realtime",
          "unit": "gwei"
        }
      }
    }
  }'
```

Save the returned `quoteId` and reuse the exact same request fields when creating the order.

### Create an Order

```bash
curl -X POST https://gasup-mcp.owlto.finance/mcp \
  -H 'Content-Type: application/json' \
  -H 'Accept: application/json, text/event-stream' \
  -H 'User-Agent: example-agent/1.0' \
  -H 'X-Wallet-Session-Id: ws_...' \
  --data '{
    "jsonrpc": "2.0",
    "id": 5,
    "method": "tools/call",
    "params": {
      "name": "create_gas_order",
      "arguments": {
        "quoteId": "gq_...",
        "orderName": "Gas Up order",
        "targetChainId": "56",
        "receivingAddresses": [
          "0xRecipientAddress..."
        ],
        "gasTimes": 1,
        "paymentMethod": {
          "paymentChainId": "8453",
          "asset": "ETH"
        },
        "deliverySchedule": {
          "mode": "immediate"
        },
        "executionGasPrice": {
          "mode": "realtime",
          "unit": "gwei"
        }
      }
    }
  }'
```

If the order returns `paymentStatus: unpaid` and `status: payment_required`, the user must send the exact payment amount to the returned payment recipient.

## Payment Submission

Use the raw amount for payment execution:

```text
paymentAmountRaw
```

Use display amounts only for UI:

```text
paymentAmount
payAmountUsd
```

After the user sends the payment transaction, call `submit_gas_order_payment`:

```bash
curl -X POST https://gasup-mcp.owlto.finance/mcp \
  -H 'Content-Type: application/json' \
  -H 'Accept: application/json, text/event-stream' \
  -H 'User-Agent: example-agent/1.0' \
  -H 'X-Wallet-Session-Id: ws_...' \
  --data '{
    "jsonrpc": "2.0",
    "id": 6,
    "method": "tools/call",
    "params": {
      "name": "submit_gas_order_payment",
      "arguments": {
        "orderId": "go_...",
        "payment": {
          "paymentChainId": "8453",
          "paymentAsset": "ETH",
          "paymentTxHash": "0xPaymentTransactionHash...",
          "payerAddress": "0xUserWallet..."
        }
      }
    }
  }'
```

Agents should confirm the order ID, payment chain, asset, payer address, and transaction hash with the user before submitting payment details.

## Tracking Dispatch

Poll `get_gas_order_dispatch_summary` while the order is active:

```bash
curl -X POST https://gasup-mcp.owlto.finance/mcp \
  -H 'Content-Type: application/json' \
  -H 'Accept: application/json, text/event-stream' \
  -H 'User-Agent: example-agent/1.0' \
  -H 'X-Wallet-Session-Id: ws_...' \
  --data '{
    "jsonrpc": "2.0",
    "id": 7,
    "method": "tools/call",
    "params": {
      "name": "get_gas_order_dispatch_summary",
      "arguments": {
        "orderId": "go_..."
      }
    }
  }'
```

Common active or user-action states include:

```text
payment_required
payment_verifying
paid
dispatching
paused
```

Common terminal states include:

```text
completed
failed
expired
cancelled
```

If `pausedGasUnits` is greater than zero, inspect `pauseReasons` and explain whether the user should wait, change a custom gas price setting, or contact support.

## Tool Error Handling

MCP tool failures are often returned as successful HTTP responses with `isError: true`:

```json
{
  "jsonrpc": "2.0",
  "id": 8,
  "result": {
    "isError": true,
    "content": [
      {
        "type": "text",
        "text": "gas up API error status=401 code=unauthorized message=..."
      }
    ]
  }
}
```

Agent clients must check both the HTTP status and `result.isError`.

Common recovery behavior:

- `unauthorized`: ask the user to reconnect the wallet session or provide a valid API key.
- `forbidden`: ask for an API key with the required scope.
- `target_chain_not_found`: refresh `list_gas_prices`.
- `payment_method_not_found`: refresh `list_payment_methods`.
- `quote_expired`: create a new quote and ask for confirmation again.
- `quote_request_mismatch`: create a new quote using the final intended fields.
- `wallet_benefit_unavailable`: retry without assuming the benefit still applies.
- `risk_blocked`: do not retry repeatedly with the same payload.
- `eoa_low_balance`: explain that target-chain inventory is temporarily unavailable.
- `payment_method_mismatch`: compare the order payment details with the submitted transaction.
- `payment_payer_mismatch`: ask for the correct payer address or transaction hash.

## Amount Safety

All token and blockchain quantities are returned as strings.

Do not convert raw amount fields through floating point numbers:

```text
paymentAmountRaw
nativeGasAmountRaw
amountRaw
```

Use integer or arbitrary-precision decimal types when implementing wallet payment flows.

## Agent Skill

The portable Agent Skill is located at:

```text
agent-skills/gas-up-agent
```

Skill-compatible agent platforms can load this folder to learn the recommended Gas Up workflow, safety rules, confirmation behavior, and error recovery behavior.

Distribute the Skill together with the hosted MCP endpoint:

```text
https://gasup-mcp.owlto.finance/mcp
```

## Self-Hosting

Most users do not need to self-host the MCP server.

If you self-host it, build and run:

```bash
go build -o bin/gas-up-mcp ./cmd/mcp
```

Create an environment file:

```bash
cp .env.example .env
```

Recommended values when the MCP server runs on the same private network as the Gas Up backend:

```text
MCP_HOST=0.0.0.0
MCP_PORT=4010
GAS_UP_API_BASE_URL=http://127.0.0.1:4000
GAS_UP_API_KEY=
```

For public self-hosted deployments, keep `GAS_UP_API_KEY` empty unless your own backend controls all traffic to the MCP server. If `GAS_UP_API_KEY` is empty, callers must provide their own API key for write tools that require backend API-key authorization.

For trusted private deployments, `GAS_UP_API_KEY` may be used as a server-side fallback credential.

Run with PM2:

```bash
set -a
source ./.env
set +a

pm2 start ./bin/gas-up-mcp --name gas-up-mcp --update-env
pm2 save
```

Health check:

```bash
curl http://127.0.0.1:4010/health
```

## Compatibility Notes

The hosted MCP endpoint is intended for agent clients and server-side integrations.

`OPTIONS /mcp` may return `405`. Browser-based clients that call MCP directly may need a backend proxy or explicit CORS preflight support.

If a client receives `403 Error 1010`, set a clear `User-Agent` and check whether a network protection rule is blocking generic HTTP clients.
