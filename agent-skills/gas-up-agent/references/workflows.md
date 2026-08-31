# Gas Up Workflows

## Quote And Order Safety

The quote and create-order requests must match exactly for these fields:

- `targetChainId`
- `receivingAddresses` and `receivingCsv` after backend normalization
- `gasTimes`
- `paymentMethod.paymentChainId`
- `paymentMethod.asset`
- `deliverySchedule`
- `executionGasPrice`

`orderName` is display-only and can differ.

If `quote_request_mismatch` occurs, create a new locked quote using the intended final fields. Do not guess which field changed.

## Confirmation Text

Before locking a quote or creating an order, summarize:

- target chain
- recipient count and sample recipient addresses
- gas times and total gas units
- payment chain and asset
- payment display amount and raw amount
- payment recipient
- quote expiration
- delivery schedule
- execution gas price mode

Ask the user to confirm before calling `create_gas_order_quote` or `create_gas_order`.

## Payment

When the created order has `paymentStatus` equal to `unpaid` and `status` equal to `payment_required`, instruct the user to send the exact amount specified by `paymentAmountRaw` to `paymentRecipient` on `paymentChainId` using `paymentAsset`.

Use raw integer amount fields for transaction construction. Use display amount fields only for human-readable summaries.

After the user sends payment, ask for the transaction hash and call `submit_gas_order_payment` with:

- `orderId`
- `payment.paymentChainId`
- `payment.paymentAsset`
- `payment.paymentTxHash`
- `payment.payerAddress`

If the transaction is pending or not yet observed, explain that verification can require confirmations and poll again later.

## Dispatch Tracking

Use `get_gas_order_dispatch_summary` for active orders. Poll every 3 to 5 seconds while the order is still moving.

Common active or user-action states include:

- `payment_required`
- `payment_verifying`
- `paid`
- `dispatching`
- `paused`

Common terminal states include:

- `completed`
- `failed`
- `expired`
- `cancelled`

Stop polling when the order reaches a terminal state. If the order is paused, summarize `pauseReasons` and explain whether the user should wait, change a custom gas price setting, or contact support.

## Error Recovery

For `unauthorized`, ask the user to reconnect the wallet session or provide a valid API key.

For `forbidden`, ask for an API key with the required scope.

For `target_chain_not_found` or `payment_method_not_found`, refresh `list_gas_prices` and `list_payment_methods`.

For `quote_expired`, create a fresh quote and ask the user to confirm the new payment details.

For `quote_request_mismatch`, create a fresh quote with the final intended fields.

For `wallet_benefit_unavailable`, retry the quote without assuming the benefit still applies.

For `risk_blocked`, do not retry repeatedly with the same payload. Explain that the backend risk policy blocked the request.

For `eoa_low_balance`, tell the user the target-chain inventory is temporarily unavailable.

For `payment_payer_mismatch` or `payment_method_mismatch`, compare the order payment details with the submitted transaction details and ask the user for the correct transaction or payer address.

## Client Compatibility

MCP requests must include `Accept: application/json, text/event-stream`.

Agents should send a clear `User-Agent`. If a client receives `403 Error 1010`, a network protection rule may be blocking generic HTTP clients.

MCP tool failures may be returned as HTTP 200 responses with `result.isError: true`. Always inspect the tool result.
