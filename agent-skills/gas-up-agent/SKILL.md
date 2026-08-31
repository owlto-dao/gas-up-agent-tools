---
name: gas-up-agent
description: Guide AI agents through Gas Up native gas top-up workflows using the Gas Up MCP server or REST API. Use when users ask an agent to find supported gas target chains, compare payment methods, estimate gas, preview or lock a gas order quote, create a gas order, submit payment transaction hashes, track dispatch progress, or troubleshoot Gas Up API errors.
---

# Gas Up Agent

## Overview

Use this skill to operate Gas Up safely through MCP-compatible agents. Prefer the Gas Up MCP tools when available; use the REST API only when MCP is unavailable or when debugging integration details.

## Authentication

Use request-scoped credentials. Prefer a wallet-bound Gas Up API key passed to the MCP server as `X-API-Key: gs_live_...` or `Authorization: Bearer gs_live_...`.

Use `X-Wallet-Session-Id: ws_...` for wallet-session flows. Payment submission currently requires a wallet session.

Never ask the user to paste private keys or seed phrases. Do not store API keys in generated files unless the user explicitly asks for local configuration.

## Standard Flow

1. Call `list_gas_prices`.
2. Call `list_payment_methods`.
3. Choose only a target chain where `status` is `open` and `acceptingNewOrders` is `true`.
4. Use `preview_gas_order_quote` for a non-locking estimate.
5. Ask the user to confirm recipient addresses, gasTimes, payment chain, payment asset, payment amount, payment recipient, and quote expiration.
6. Call `create_gas_order_quote` only after confirmation.
7. Call `create_gas_order` with exactly the same quote fields.
8. If payment is required, instruct the user to send the exact `paymentAmountRaw` amount to `paymentRecipient`.
9. Call `submit_gas_order_payment` after the user provides the payment transaction hash.
10. Poll `get_gas_order_dispatch_summary` every 3 to 5 seconds until the order is completed, paused, or failed.

Read `references/workflows.md` before creating orders, submitting payments, or recovering from quote/order errors.

## Tool Rules

Use `estimate_gas_amounts` when the user wants to know what a single recipient will receive without pricing or reserving inventory.

Use `list_wallet_benefits` when the user asks whether free gas or campaign benefits apply.

Use `preview_gas_order_quote` before any paid order confirmation.

Use `create_gas_order_quote` to lock a quote. This may reserve quote state and should follow user confirmation.

Use `create_gas_order` only after a locked quote exists and the user has confirmed the order.

Use `submit_gas_order_payment` only after the user supplies a payment transaction hash.

## Amount Handling

Treat all amounts as strings. Never convert raw blockchain quantities such as `paymentAmountRaw`, `nativeGasAmountRaw`, or `amountRaw` through floating point numbers.

For user-facing summaries, show both display amount and raw amount when payment is required.
