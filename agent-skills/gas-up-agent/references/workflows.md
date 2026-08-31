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
- recipient count and first few recipient addresses
- gasTimes and total gas units
- payment chain and asset
- payment display amount and raw amount
- payment recipient
- quote expiration
- delivery schedule
- execution gas price mode

Ask the user to confirm before calling `create_gas_order_quote` or `create_gas_order`.

## Payment

When the created order has `paymentStatus` equal to `unpaid` and `status` equal to `payment_required`, instruct the user to send the exact amount specified by `paymentAmountRaw` to `paymentRecipient` on `paymentChainId` using `paymentAsset`.

After the user sends payment, call `submit_gas_order_payment` with:

- `orderId`
- `payment.paymentChainId`
- `payment.paymentAsset`
- `payment.paymentTxHash`
- `payment.payerAddress`

If the transaction is pending or not yet observed, explain that verification can require confirmations and poll again later.

## Dispatch Tracking

Use `get_gas_order_dispatch_summary` for active orders. Stop polling when status is one of:

- `completed`
- `failed`
- `payment_required`
- `expired`

If `pausedGasUnits` is greater than zero, summarize `pauseReasons` and explain whether the user needs to wait or adjust a custom gas price setting.

## Error Recovery

For `target_chain_not_found` or `payment_method_not_found`, refresh `list_gas_prices` and `list_payment_methods`.

For `quote_expired`, create a fresh quote and ask the user to confirm the new payment details.

For `wallet_benefit_unavailable`, retry quote without assuming the benefit still applies.

For `risk_blocked`, do not retry with the same payload repeatedly. Explain that the backend risk policy blocked the request.

For `eoa_low_balance`, tell the user the target chain inventory is temporarily unavailable.

For `payment_payer_mismatch` or `payment_method_mismatch`, compare the order payment details with the submitted transaction details and ask the user for the correct transaction or payer address.
