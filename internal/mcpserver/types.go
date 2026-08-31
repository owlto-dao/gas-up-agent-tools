package mcpserver

type emptyInput struct{}

type estimateGasInput struct {
	TargetChainID  string `json:"targetChainId" jsonschema:"Target chain ID that receives native gas, for example 1, 8453, 42161, 4663, or 83797601."`
	CustomGasTimes int    `json:"customGasTimes,omitempty" jsonschema:"Optional integer from 1 to 100 for estimating a custom number of gas units."`
}

type paginationInput struct {
	Limit  int `json:"limit,omitempty" jsonschema:"Optional page size. Defaults to the backend endpoint default."`
	Offset int `json:"offset,omitempty" jsonschema:"Optional pagination offset. Defaults to 0."`
}

type orderIDInput struct {
	OrderID string `json:"orderId" jsonschema:"Gas order ID, usually prefixed with go_."`
}

type getGasOrderInput struct {
	OrderID    string          `json:"orderId" jsonschema:"Gas order ID, usually prefixed with go_."`
	Pagination paginationInput `json:"pagination,omitempty" jsonschema:"Optional pagination for gas unit details."`
}

type paymentMethodInput struct {
	PaymentChainID string `json:"paymentChainId" jsonschema:"Payment chain ID returned by list_payment_methods."`
	Asset          string `json:"asset" jsonschema:"Payment asset symbol returned by list_payment_methods, for example ETH or BNB."`
}

type deliveryScheduleInput struct {
	Mode               string `json:"mode" jsonschema:"Delivery mode: immediate or random_interval."`
	MinIntervalSeconds int    `json:"minIntervalSeconds,omitempty" jsonschema:"Required when mode is random_interval."`
	MaxIntervalSeconds int    `json:"maxIntervalSeconds,omitempty" jsonschema:"Required when mode is random_interval."`
}

type executionGasPriceInput struct {
	Mode        string `json:"mode" jsonschema:"Gas price execution mode: realtime or custom_max."`
	MaxGasPrice string `json:"maxGasPrice,omitempty" jsonschema:"Required when mode is custom_max. EVM unit is gwei; Solana unit is lamports."`
	Unit        string `json:"unit" jsonschema:"Target-chain gas price unit, usually gwei for EVM or lamports for Solana."`
}

type gasQuoteInput struct {
	WalletSessionID    string                 `json:"walletSessionId,omitempty" jsonschema:"Wallet session ID. Optional when the MCP request includes a wallet-bound API key."`
	TargetChainID      string                 `json:"targetChainId" jsonschema:"Target chain that receives native gas."`
	ReceivingAddresses []string               `json:"receivingAddresses" jsonschema:"Recipient addresses. Provide 1 to 1000 addresses."`
	ReceivingCSV       string                 `json:"receivingCsv,omitempty" jsonschema:"Optional CSV address input. May be used together with receivingAddresses."`
	GasTimes           int                    `json:"gasTimes" jsonschema:"Number of gas units per recipient, from 1 to 100."`
	PaymentMethod      paymentMethodInput     `json:"paymentMethod" jsonschema:"Payment method chosen from list_payment_methods."`
	DeliverySchedule   deliveryScheduleInput  `json:"deliverySchedule" jsonschema:"Delivery schedule. Use mode immediate unless the user asks for random spacing."`
	ExecutionGasPrice  executionGasPriceInput `json:"executionGasPrice" jsonschema:"Target-chain execution gas price preference."`
}

type createGasOrderInput struct {
	WalletSessionID    string                 `json:"walletSessionId,omitempty" jsonschema:"Wallet session ID. Optional when the MCP request includes a wallet-bound API key."`
	QuoteID            string                 `json:"quoteId" jsonschema:"Locked quote ID returned by create_gas_order_quote."`
	OrderName          string                 `json:"orderName,omitempty" jsonschema:"Optional display name for this gas order. Does not affect quote matching."`
	TargetChainID      string                 `json:"targetChainId" jsonschema:"Must exactly match the locked quote."`
	ReceivingAddresses []string               `json:"receivingAddresses" jsonschema:"Must exactly match the locked quote after backend normalization."`
	ReceivingCSV       string                 `json:"receivingCsv,omitempty" jsonschema:"Must exactly match the locked quote after backend normalization."`
	GasTimes           int                    `json:"gasTimes" jsonschema:"Must exactly match the locked quote."`
	PaymentMethod      paymentMethodInput     `json:"paymentMethod" jsonschema:"Must exactly match the locked quote."`
	DeliverySchedule   deliveryScheduleInput  `json:"deliverySchedule" jsonschema:"Must exactly match the locked quote."`
	ExecutionGasPrice  executionGasPriceInput `json:"executionGasPrice" jsonschema:"Must exactly match the locked quote."`
	AppliedBenefitIDs  []string               `json:"appliedBenefitIds,omitempty" jsonschema:"Optional applied wallet benefit IDs returned by the locked quote."`
}

type submitPaymentInput struct {
	OrderID string              `json:"orderId" jsonschema:"Gas order ID to mark as paid or payment-verifying."`
	Payment confirmPaymentInput `json:"payment" jsonschema:"Payment transaction details."`
}

type confirmPaymentInput struct {
	PaymentChainID string `json:"paymentChainId" jsonschema:"Payment chain ID from the order quote."`
	PaymentAsset   string `json:"paymentAsset" jsonschema:"Payment asset from the order quote, for example ETH or BNB."`
	PaymentTxHash  string `json:"paymentTxHash" jsonschema:"Payment transaction hash. For EVM production, use 0x followed by 64 hex characters."`
	PayerAddress   string `json:"payerAddress" jsonschema:"Wallet address that sent the payment."`
}
