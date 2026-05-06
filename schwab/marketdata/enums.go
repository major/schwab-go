package marketdata

// QuoteType identifies the quote feed type in a quote envelope.
type QuoteType string

const (
	QuoteTypeNBBO QuoteType = "NBBO"
)

// OptionChainContractType filters or identifies option chain contracts.
type OptionChainContractType string

const (
	OptionChainContractTypeCall OptionChainContractType = "CALL"
	OptionChainContractTypePut  OptionChainContractType = "PUT"
	OptionChainContractTypeAll  OptionChainContractType = "ALL"
)

// OptionContractType identifies option contract side in quote reference payloads.
type OptionContractType string

const (
	OptionContractTypePut  OptionContractType = "P"
	OptionContractTypeCall OptionContractType = "C"
)

// OptionChainStrategy identifies option chain strategy filters.
type OptionChainStrategy string

const (
	OptionChainStrategySingle     OptionChainStrategy = "SINGLE"
	OptionChainStrategyAnalytical OptionChainStrategy = "ANALYTICAL"
	OptionChainStrategyCovered    OptionChainStrategy = "COVERED"
	OptionChainStrategyVertical   OptionChainStrategy = "VERTICAL"
	OptionChainStrategyCalendar   OptionChainStrategy = "CALENDAR"
	OptionChainStrategyStrangle   OptionChainStrategy = "STRANGLE"
	OptionChainStrategyStraddle   OptionChainStrategy = "STRADDLE"
	OptionChainStrategyButterfly  OptionChainStrategy = "BUTTERFLY"
	OptionChainStrategyCondor     OptionChainStrategy = "CONDOR"
	OptionChainStrategyDiagonal   OptionChainStrategy = "DIAGONAL"
	OptionChainStrategyCollar     OptionChainStrategy = "COLLAR"
	OptionChainStrategyRoll       OptionChainStrategy = "ROLL"
)

// OptionChainRange identifies option chain strike range filters.
type OptionChainRange string

const (
	OptionChainRangeInTheMoney         OptionChainRange = "ITM"
	OptionChainRangeNearTheMoney       OptionChainRange = "NTM"
	OptionChainRangeOutOfTheMoney      OptionChainRange = "OTM"
	OptionChainRangeStrikesAboveMarket OptionChainRange = "SAK"
	OptionChainRangeStrikesBelowMarket OptionChainRange = "SBK"
	OptionChainRangeStrikesNearMarket  OptionChainRange = "SNK"
	OptionChainRangeAll                OptionChainRange = "ALL"
)

// ExpirationMonth identifies option expiration month filters.
type ExpirationMonth string

const (
	ExpirationMonthJanuary   ExpirationMonth = "JAN"
	ExpirationMonthFebruary  ExpirationMonth = "FEB"
	ExpirationMonthMarch     ExpirationMonth = "MAR"
	ExpirationMonthApril     ExpirationMonth = "APR"
	ExpirationMonthMay       ExpirationMonth = "MAY"
	ExpirationMonthJune      ExpirationMonth = "JUN"
	ExpirationMonthJuly      ExpirationMonth = "JUL"
	ExpirationMonthAugust    ExpirationMonth = "AUG"
	ExpirationMonthSeptember ExpirationMonth = "SEP"
	ExpirationMonthOctober   ExpirationMonth = "OCT"
	ExpirationMonthNovember  ExpirationMonth = "NOV"
	ExpirationMonthDecember  ExpirationMonth = "DEC"
	ExpirationMonthAll       ExpirationMonth = "ALL"
)

// OptionChainType identifies standard versus non-standard option chain filters.
type OptionChainType string

const (
	OptionChainTypeStandard    OptionChainType = "S"
	OptionChainTypeNonStandard OptionChainType = "NS"
	OptionChainTypeAll         OptionChainType = "ALL"
)

// OptionEntitlement identifies entitlement filters for option chains.
type OptionEntitlement string

const (
	OptionEntitlementPayingNonProfessional OptionEntitlement = "PN"
	OptionEntitlementNonProfessional       OptionEntitlement = "NP"
	OptionEntitlementPayingProfessional    OptionEntitlement = "PP"
)

// OptionExerciseType identifies American or European exercise styles.
type OptionExerciseType string

const (
	OptionExerciseTypeAmerican OptionExerciseType = "A"
	OptionExerciseTypeEuropean OptionExerciseType = "E"
)

// OptionExpirationType identifies the option expiration cycle.
type OptionExpirationType string

const (
	OptionExpirationTypeMonthly   OptionExpirationType = "M"
	OptionExpirationTypeQuarterly OptionExpirationType = "Q"
	OptionExpirationTypeStandard  OptionExpirationType = "S"
	OptionExpirationTypeWeekly    OptionExpirationType = "W"
)

// OptionSettlementType identifies physical or cash settlement.
type OptionSettlementType string

const (
	OptionSettlementTypeAM OptionSettlementType = "A"
	OptionSettlementTypePM OptionSettlementType = "P"
)

// MarketID identifies supported market hours resources.
type MarketID string

const (
	MarketIDEquity MarketID = "equity"
	MarketIDOption MarketID = "option"
	MarketIDBond   MarketID = "bond"
	MarketIDFuture MarketID = "future"
	MarketIDForex  MarketID = "forex"
)
