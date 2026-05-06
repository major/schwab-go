package schwab

// AssetType represents the type of a financial asset.
type AssetType string

const (
	AssetTypeBond         AssetType = "BOND"
	AssetTypeEquity       AssetType = "EQUITY"
	AssetTypeETF          AssetType = "ETF"
	AssetTypeForex        AssetType = "FOREX"
	AssetTypeFuture       AssetType = "FUTURE"
	AssetTypeFutureOption AssetType = "FUTURE_OPTION"
	AssetTypeIndex        AssetType = "INDEX"
	AssetTypeMutualFund   AssetType = "MUTUAL_FUND"
	AssetTypeOption       AssetType = "OPTION"
)
