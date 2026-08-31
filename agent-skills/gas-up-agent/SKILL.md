---
name: gas-up-agent
description: Guide AI agents through Gas Up native gas top-up workflows using the Gas Up MCP server or REST API. Use when users ask an agent to find supported gas target chains, compare payment methods, estimate gas, preview or lock a gas order quote, create a gas order, submit payment transaction hashes, track dispatch progress, or troubleshoot Gas Up API errors.
---

# Gas Up Agent

## Overview

Use this skill to operate Gas Up safely through MCP-compatible agents. Prefer the Gas Up MCP tools when available. Use the REST API only when MCP is unavailable or when debugging integration details.

The hosted MCP endpoint is:

```text
https://gasup-mcp.owlto.finance/mcp
```

MCP requests must include:

```text
Accept: application/json, text/event-stream
```

## Authentication

Use request-scoped credentials.

For normal user-facing integrations, prefer a wallet session:

```text
X-Wallet-Session-Id: ws_...
```

Wallet sessions are created by the Gas Up wallet sign-in flow. Never ask users to paste private keys or seed phrases.

Developer integrations may use a Gas Up API key:

```text
X-API-Key: gs_live_...
```

or:

```text
Authorization: Bearer gs_live_...
```

If an API key is not bound to a wallet, also pass `X-Wallet-Session-Id` for wallet-scoped write tools. Payment submission currently requires a wallet session.

On the official hosted MCP endpoint, normal wallet-session requests use Gas Up's server-side service credential. That credential is never exposed to users or agent clients.

Do not store API keys or wallet sessions in generated files unless the user explicitly asks for local configuration.

## Standard Flow

1. Call `list_gas_prices`.
2. Call `list_payment_methods`.
3. Choose only a target chain where `status` is `open` and `acceptingNewOrders` is `true`.
4. Call `estimate_gas_amounts` if the user wants to understand the amount of native gas received before pricing.
5. Call `preview_gas_order_quote` for a non-locking estimate.
6. Ask the user to confirm recipient addresses, gas times, payment chain, payment asset, payment amount, payment recipient, and quote expiration.
7. Call `create_gas_order_quote` only after confirmation.
8. Call `create_gas_order` with exactly the same quote fields.
9. If payment is required, instruct the user to send the exact `paymentAmountRaw` amount to `paymentRecipient` on the payment chain.
10. Call `submit_gas_order_payment` after the user provides the payment transaction hash.
11. Poll `get_gas_order_dispatch_summary` every 3 to 5 seconds until the order is completed, paused, failed, expired, or waiting for user action.

Read `references/workflows.md` before creating orders, submitting payments, or recovering from quote/order errors.

## Tool Rules

Use `estimate_gas_amounts` when the user wants to know what a single recipient will receive without pricing or reserving inventory.

Use `list_wallet_benefits` when the user asks whether free gas or campaign benefits apply.

Use `preview_gas_order_quote` before any paid order confirmation.

Use `create_gas_order_quote` only after the user confirms the quote details. This locks a quote and may reserve quote state.

Use `create_gas_order` only after a locked quote exists and the user has confirmed the order.

Use `submit_gas_order_payment` only after the user supplies a payment transaction hash.

## Confirmation Rules

Before quote locking or order creation, summarize:

- target chain
- recipient count and sample recipient addresses
- gas times and total gas units
- payment chain and asset
- payment display amount and raw amount
- payment recipient
- quote expiration
- delivery schedule
- execution gas price mode

Ask the user to confirm before calling a write tool.

## Amount Handling

Treat all amounts as strings. Never convert raw blockchain quantities such as `paymentAmountRaw`, `nativeGasAmountRaw`, or `amountRaw` through floating point numbers.

For user-facing summaries, show both display amount and raw amount when payment is required.

## Error Handling

MCP tool failures may be returned as HTTP 200 responses with `result.isError: true`. Always check both HTTP status and `result.isError`.

Recover according to `references/workflows.md`.
