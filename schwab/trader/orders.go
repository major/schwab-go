package trader

import (
	"context"
	"errors"
	"strconv"

	schwab "github.com/major/schwab-go/schwab"
)

// Session identifies the trading session for an order.
type Session string

const (
	SessionNormal   Session = "NORMAL"
	SessionAM       Session = "AM"
	SessionPM       Session = "PM"
	SessionSeamless Session = "SEAMLESS"
)

// Duration identifies how long an order remains active.
type Duration string

const (
	DurationDay               Duration = "DAY"
	DurationGoodTillCancel    Duration = "GOOD_TILL_CANCEL"
	DurationFillOrKill        Duration = "FILL_OR_KILL"
	DurationImmediateOrCancel Duration = "IMMEDIATE_OR_CANCEL"
	DurationEndOfWeek         Duration = "END_OF_WEEK"
	DurationEndOfMonth        Duration = "END_OF_MONTH"
	DurationNextEndOfMonth    Duration = "NEXT_END_OF_MONTH"
	DurationUnknown           Duration = "UNKNOWN"
)

// OrderType identifies the order pricing and execution type.
type OrderType string

const (
	OrderTypeMarket            OrderType = "MARKET"
	OrderTypeLimit             OrderType = "LIMIT"
	OrderTypeStop              OrderType = "STOP"
	OrderTypeStopLimit         OrderType = "STOP_LIMIT"
	OrderTypeTrailingStop      OrderType = "TRAILING_STOP"
	OrderTypeCabinet           OrderType = "CABINET"
	OrderTypeNonMarketable     OrderType = "NON_MARKETABLE"
	OrderTypeMarketOnClose     OrderType = "MARKET_ON_CLOSE"
	OrderTypeExercise          OrderType = "EXERCISE"
	OrderTypeTrailingStopLimit OrderType = "TRAILING_STOP_LIMIT"
	OrderTypeNetDebit          OrderType = "NET_DEBIT"
	OrderTypeNetCredit         OrderType = "NET_CREDIT"
	OrderTypeNetZero           OrderType = "NET_ZERO"
	OrderTypeLimitOnClose      OrderType = "LIMIT_ON_CLOSE"
	OrderTypeUnknown           OrderType = "UNKNOWN"
)

// OrderTypeRequest identifies order types accepted in request bodies.
type OrderTypeRequest string

const (
	OrderTypeRequestMarket            OrderTypeRequest = "MARKET"
	OrderTypeRequestLimit             OrderTypeRequest = "LIMIT"
	OrderTypeRequestStop              OrderTypeRequest = "STOP"
	OrderTypeRequestStopLimit         OrderTypeRequest = "STOP_LIMIT"
	OrderTypeRequestTrailingStop      OrderTypeRequest = "TRAILING_STOP"
	OrderTypeRequestCabinet           OrderTypeRequest = "CABINET"
	OrderTypeRequestNonMarketable     OrderTypeRequest = "NON_MARKETABLE"
	OrderTypeRequestMarketOnClose     OrderTypeRequest = "MARKET_ON_CLOSE"
	OrderTypeRequestExercise          OrderTypeRequest = "EXERCISE"
	OrderTypeRequestTrailingStopLimit OrderTypeRequest = "TRAILING_STOP_LIMIT"
	OrderTypeRequestNetDebit          OrderTypeRequest = "NET_DEBIT"
	OrderTypeRequestNetCredit         OrderTypeRequest = "NET_CREDIT"
	OrderTypeRequestNetZero           OrderTypeRequest = "NET_ZERO"
	OrderTypeRequestLimitOnClose      OrderTypeRequest = "LIMIT_ON_CLOSE"
)

// ComplexOrderStrategyType identifies a multi-leg order strategy.
type ComplexOrderStrategyType string

const (
	ComplexOrderStrategyTypeNone                   ComplexOrderStrategyType = "NONE"
	ComplexOrderStrategyTypeCovered                ComplexOrderStrategyType = "COVERED"
	ComplexOrderStrategyTypeVertical               ComplexOrderStrategyType = "VERTICAL"
	ComplexOrderStrategyTypeBackRatio              ComplexOrderStrategyType = "BACK_RATIO"
	ComplexOrderStrategyTypeCalendar               ComplexOrderStrategyType = "CALENDAR"
	ComplexOrderStrategyTypeDiagonal               ComplexOrderStrategyType = "DIAGONAL"
	ComplexOrderStrategyTypeStraddle               ComplexOrderStrategyType = "STRADDLE"
	ComplexOrderStrategyTypeStrangle               ComplexOrderStrategyType = "STRANGLE"
	ComplexOrderStrategyTypeCollarSynthetic        ComplexOrderStrategyType = "COLLAR_SYNTHETIC"
	ComplexOrderStrategyTypeButterfly              ComplexOrderStrategyType = "BUTTERFLY"
	ComplexOrderStrategyTypeCondor                 ComplexOrderStrategyType = "CONDOR"
	ComplexOrderStrategyTypeIronCondor             ComplexOrderStrategyType = "IRON_CONDOR"
	ComplexOrderStrategyTypeVerticalRoll           ComplexOrderStrategyType = "VERTICAL_ROLL"
	ComplexOrderStrategyTypeCollarWithStock        ComplexOrderStrategyType = "COLLAR_WITH_STOCK"
	ComplexOrderStrategyTypeDoubleDiagonal         ComplexOrderStrategyType = "DOUBLE_DIAGONAL"
	ComplexOrderStrategyTypeUnbalancedButterfly    ComplexOrderStrategyType = "UNBALANCED_BUTTERFLY"
	ComplexOrderStrategyTypeUnbalancedCondor       ComplexOrderStrategyType = "UNBALANCED_CONDOR"
	ComplexOrderStrategyTypeUnbalancedIronCondor   ComplexOrderStrategyType = "UNBALANCED_IRON_CONDOR"
	ComplexOrderStrategyTypeUnbalancedVerticalRoll ComplexOrderStrategyType = "UNBALANCED_VERTICAL_ROLL"
	ComplexOrderStrategyTypeMutualFundSwap         ComplexOrderStrategyType = "MUTUAL_FUND_SWAP"
	ComplexOrderStrategyTypeCustom                 ComplexOrderStrategyType = "CUSTOM"
)

// RequestedDestination identifies a requested order destination.
type RequestedDestination string

const (
	RequestedDestinationINET    RequestedDestination = "INET"
	RequestedDestinationECNARCA RequestedDestination = "ECN_ARCA"
	RequestedDestinationCBOE    RequestedDestination = "CBOE"
	RequestedDestinationAMEX    RequestedDestination = "AMEX"
	RequestedDestinationPHLX    RequestedDestination = "PHLX"
	RequestedDestinationISE     RequestedDestination = "ISE"
	RequestedDestinationBOX     RequestedDestination = "BOX"
	RequestedDestinationNYSE    RequestedDestination = "NYSE"
	RequestedDestinationNASDAQ  RequestedDestination = "NASDAQ"
	RequestedDestinationBATS    RequestedDestination = "BATS"
	RequestedDestinationC2      RequestedDestination = "C2"
	RequestedDestinationAuto    RequestedDestination = "AUTO"
)

// PriceLinkBasis identifies which price a linked price is based on.
type PriceLinkBasis string

const (
	PriceLinkBasisManual  PriceLinkBasis = "MANUAL"
	PriceLinkBasisBase    PriceLinkBasis = "BASE"
	PriceLinkBasisTrigger PriceLinkBasis = "TRIGGER"
	PriceLinkBasisLast    PriceLinkBasis = "LAST"
	PriceLinkBasisBid     PriceLinkBasis = "BID"
	PriceLinkBasisAsk     PriceLinkBasis = "ASK"
	PriceLinkBasisAskBid  PriceLinkBasis = "ASK_BID"
	PriceLinkBasisMark    PriceLinkBasis = "MARK"
	PriceLinkBasisAverage PriceLinkBasis = "AVERAGE"
)

// PriceLinkType identifies how a linked price offset is interpreted.
type PriceLinkType string

const (
	PriceLinkTypeValue   PriceLinkType = "VALUE"
	PriceLinkTypePercent PriceLinkType = "PERCENT"
	PriceLinkTypeTick    PriceLinkType = "TICK"
)

// StopType identifies which market price triggers a stop.
type StopType string

const (
	StopTypeStandard StopType = "STANDARD"
	StopTypeBid      StopType = "BID"
	StopTypeAsk      StopType = "ASK"
	StopTypeLast     StopType = "LAST"
	StopTypeMark     StopType = "MARK"
)

// TaxLotMethod identifies the tax lot selection method for an order.
type TaxLotMethod string

const (
	TaxLotMethodFIFO          TaxLotMethod = "FIFO"
	TaxLotMethodLIFO          TaxLotMethod = "LIFO"
	TaxLotMethodHighCost      TaxLotMethod = "HIGH_COST"
	TaxLotMethodLowCost       TaxLotMethod = "LOW_COST"
	TaxLotMethodAverageCost   TaxLotMethod = "AVERAGE_COST"
	TaxLotMethodSpecificLot   TaxLotMethod = "SPECIFIC_LOT"
	TaxLotMethodLossHarvester TaxLotMethod = "LOSS_HARVESTER"
)

// SpecialInstruction identifies extra handling instructions for an order.
type SpecialInstruction string

const (
	SpecialInstructionAllOrNone            SpecialInstruction = "ALL_OR_NONE"
	SpecialInstructionDoNotReduce          SpecialInstruction = "DO_NOT_REDUCE"
	SpecialInstructionAllOrNoneDoNotReduce SpecialInstruction = "ALL_OR_NONE_DO_NOT_REDUCE"
)

// OrderStrategyType identifies how an order relates to other orders.
type OrderStrategyType string

const (
	OrderStrategyTypeSingle     OrderStrategyType = "SINGLE"
	OrderStrategyTypeCancel     OrderStrategyType = "CANCEL"
	OrderStrategyTypeRecall     OrderStrategyType = "RECALL"
	OrderStrategyTypePair       OrderStrategyType = "PAIR"
	OrderStrategyTypeFlatten    OrderStrategyType = "FLATTEN"
	OrderStrategyTypeTwoDaySwap OrderStrategyType = "TWO_DAY_SWAP"
	OrderStrategyTypeBlastAll   OrderStrategyType = "BLAST_ALL"
	OrderStrategyTypeOCO        OrderStrategyType = "OCO"
	OrderStrategyTypeTrigger    OrderStrategyType = "TRIGGER"
)

// OrderStatus identifies the current status of an order.
type OrderStatus string

const (
	OrderStatusAwaitingParentOrder   OrderStatus = "AWAITING_PARENT_ORDER"
	OrderStatusAwaitingCondition     OrderStatus = "AWAITING_CONDITION"
	OrderStatusAwaitingStopCondition OrderStatus = "AWAITING_STOP_CONDITION"
	OrderStatusAwaitingManualReview  OrderStatus = "AWAITING_MANUAL_REVIEW"
	OrderStatusAccepted              OrderStatus = "ACCEPTED"
	OrderStatusAwaitingUROut         OrderStatus = "AWAITING_UR_OUT"
	OrderStatusPendingActivation     OrderStatus = "PENDING_ACTIVATION"
	OrderStatusQueued                OrderStatus = "QUEUED"
	OrderStatusWorking               OrderStatus = "WORKING"
	OrderStatusRejected              OrderStatus = "REJECTED"
	OrderStatusPendingCancel         OrderStatus = "PENDING_CANCEL"
	OrderStatusCanceled              OrderStatus = "CANCELED"
	OrderStatusPendingReplace        OrderStatus = "PENDING_REPLACE"
	OrderStatusReplaced              OrderStatus = "REPLACED"
	OrderStatusFilled                OrderStatus = "FILLED"
	OrderStatusExpired               OrderStatus = "EXPIRED"
	OrderStatusNew                   OrderStatus = "NEW"
	OrderStatusAwaitingReleaseTime   OrderStatus = "AWAITING_RELEASE_TIME"
	OrderStatusPendingAcknowledgment OrderStatus = "PENDING_ACKNOWLEDGEMENT"
	OrderStatusPendingRecall         OrderStatus = "PENDING_RECALL"
	OrderStatusUnknown               OrderStatus = "UNKNOWN"
)

// Instruction identifies the action for an order leg.
type Instruction string

const (
	InstructionBuy             Instruction = "BUY"
	InstructionSell            Instruction = "SELL"
	InstructionBuyToCover      Instruction = "BUY_TO_COVER"
	InstructionSellShort       Instruction = "SELL_SHORT"
	InstructionBuyToOpen       Instruction = "BUY_TO_OPEN"
	InstructionBuyToClose      Instruction = "BUY_TO_CLOSE"
	InstructionSellToOpen      Instruction = "SELL_TO_OPEN"
	InstructionSellToClose     Instruction = "SELL_TO_CLOSE"
	InstructionExchange        Instruction = "EXCHANGE"
	InstructionSellShortExempt Instruction = "SELL_SHORT_EXEMPT"
)

// OrderLegType identifies the instrument class for an order leg.
type OrderLegType string

const (
	OrderLegTypeEquity               OrderLegType = "EQUITY"
	OrderLegTypeOption               OrderLegType = "OPTION"
	OrderLegTypeIndex                OrderLegType = "INDEX"
	OrderLegTypeMutualFund           OrderLegType = "MUTUAL_FUND"
	OrderLegTypeCashEquivalent       OrderLegType = "CASH_EQUIVALENT"
	OrderLegTypeFixedIncome          OrderLegType = "FIXED_INCOME"
	OrderLegTypeCurrency             OrderLegType = "CURRENCY"
	OrderLegTypeCollectiveInvestment OrderLegType = "COLLECTIVE_INVESTMENT"
)

// PositionEffect identifies whether an order opens or closes a position.
type PositionEffect string

const (
	PositionEffectOpening   PositionEffect = "OPENING"
	PositionEffectClosing   PositionEffect = "CLOSING"
	PositionEffectAutomatic PositionEffect = "AUTOMATIC"
)

// QuantityType identifies how an order leg quantity is represented.
type QuantityType string

const (
	QuantityTypeAllShares QuantityType = "ALL_SHARES"
	QuantityTypeDollars   QuantityType = "DOLLARS"
	QuantityTypeShares    QuantityType = "SHARES"
)

// DivCapGains identifies dividend and capital gains handling for mutual fund orders.
type DivCapGains string

const (
	DivCapGainsReinvest DivCapGains = "REINVEST"
	DivCapGainsPayout   DivCapGains = "PAYOUT"
)

// AmountIndicator identifies how an order strategy amount is expressed.
type AmountIndicator string

const (
	AmountIndicatorDollars    AmountIndicator = "DOLLARS"
	AmountIndicatorShares     AmountIndicator = "SHARES"
	AmountIndicatorAllShares  AmountIndicator = "ALL_SHARES"
	AmountIndicatorPercentage AmountIndicator = "PERCENTAGE"
	AmountIndicatorUnknown    AmountIndicator = "UNKNOWN"
)

// SettlementInstruction identifies settlement timing for an order strategy.
type SettlementInstruction string

const (
	SettlementInstructionRegular SettlementInstruction = "REGULAR"
	SettlementInstructionCash    SettlementInstruction = "CASH"
	SettlementInstructionNextDay SettlementInstruction = "NEXT_DAY"
	SettlementInstructionUnknown SettlementInstruction = "UNKNOWN"
)

// AdvancedOrderType identifies preview response advanced order grouping.
type AdvancedOrderType string

const (
	AdvancedOrderTypeNone     AdvancedOrderType = "NONE"
	AdvancedOrderTypeOTO      AdvancedOrderType = "OTO"
	AdvancedOrderTypeOCO      AdvancedOrderType = "OCO"
	AdvancedOrderTypeOTOCO    AdvancedOrderType = "OTOCO"
	AdvancedOrderTypeOT2OCO   AdvancedOrderType = "OT2OCO"
	AdvancedOrderTypeOT3OCO   AdvancedOrderType = "OT3OCO"
	AdvancedOrderTypeBlastAll AdvancedOrderType = "BLAST_ALL"
	AdvancedOrderTypeOTA      AdvancedOrderType = "OTA"
	AdvancedOrderTypePair     AdvancedOrderType = "PAIR"
)

// OrderListParams contains parameters for listing orders.
// FromEnteredTime and ToEnteredTime are required by the API in yyyy-MM-dd format.
type OrderListParams struct {
	MaxResults      int
	FromEnteredTime string
	ToEnteredTime   string
	Status          OrderStatus
}

// OrderRequest contains the writable payload for creating, replacing, or previewing orders.
type OrderRequest struct {
	Session                  Session                  `json:"session,omitempty"`
	Duration                 Duration                 `json:"duration,omitempty"`
	OrderType                OrderTypeRequest         `json:"orderType,omitempty"`
	ComplexOrderStrategyType ComplexOrderStrategyType `json:"complexOrderStrategyType,omitempty"`
	Quantity                 float64                  `json:"quantity,omitempty"`
	FilledQuantity           float64                  `json:"filledQuantity,omitempty"`
	RemainingQuantity        float64                  `json:"remainingQuantity,omitempty"`
	DestinationLinkName      string                   `json:"destinationLinkName,omitempty"`
	Price                    float64                  `json:"price,omitempty"`
	StopPrice                float64                  `json:"stopPrice,omitempty"`
	StopPriceLinkBasis       PriceLinkBasis           `json:"stopPriceLinkBasis,omitempty"`
	StopPriceLinkType        PriceLinkType            `json:"stopPriceLinkType,omitempty"`
	StopPriceOffset          float64                  `json:"stopPriceOffset,omitempty"`
	StopType                 StopType                 `json:"stopType,omitempty"`
	PriceLinkBasis           PriceLinkBasis           `json:"priceLinkBasis,omitempty"`
	PriceLinkType            PriceLinkType            `json:"priceLinkType,omitempty"`
	TaxLotMethod             TaxLotMethod             `json:"taxLotMethod,omitempty"`
	ActivationPrice          float64                  `json:"activationPrice,omitempty"`
	SpecialInstruction       SpecialInstruction       `json:"specialInstruction,omitempty"`
	OrderStrategyType        OrderStrategyType        `json:"orderStrategyType,omitempty"`
	CancelTime               string                   `json:"cancelTime,omitempty"`
	ReleaseTime              string                   `json:"releaseTime,omitempty"`
	OrderLegCollection       []OrderLegRequest        `json:"orderLegCollection,omitempty"`
	OrderActivityCollection  []OrderActivity          `json:"orderActivityCollection,omitempty"`
	ChildOrderStrategies     []*OrderRequest          `json:"childOrderStrategies,omitempty"`
	ReplacingOrderCollection []*OrderRequest          `json:"replacingOrderCollection,omitempty"`
	OrderID                  int64                    `json:"orderId,omitempty"`
	Cancelable               bool                     `json:"cancelable,omitempty"`
	Editable                 bool                     `json:"editable,omitempty"`
	Status                   OrderStatus              `json:"status,omitempty"`
	EnteredTime              string                   `json:"enteredTime,omitempty"`
	CloseTime                string                   `json:"closeTime,omitempty"`
	AccountNumber            int64                    `json:"accountNumber,omitempty"`
	StatusDescription        string                   `json:"statusDescription,omitempty"`
}

// OrderLegRequest represents one leg in an order request body.
type OrderLegRequest struct {
	OrderLegType   OrderLegType    `json:"orderLegType,omitempty"`
	LegID          int64           `json:"legId,omitempty"`
	Instrument     OrderInstrument `json:"instrument"`
	Instruction    Instruction     `json:"instruction,omitempty"`
	PositionEffect PositionEffect  `json:"positionEffect,omitempty"`
	Quantity       float64         `json:"quantity,omitempty"`
	QuantityType   QuantityType    `json:"quantityType,omitempty"`
	DivCapGains    DivCapGains     `json:"divCapGains,omitempty"`
	ToSymbol       string          `json:"toSymbol,omitempty"`
}

// Order represents a Schwab order, including recursive child and replacement orders.
type Order struct {
	Session                  Session                  `json:"session"`
	Duration                 Duration                 `json:"duration"`
	OrderType                OrderType                `json:"orderType"`
	ComplexOrderStrategyType ComplexOrderStrategyType `json:"complexOrderStrategyType"`
	Quantity                 float64                  `json:"quantity"`
	FilledQuantity           float64                  `json:"filledQuantity"`
	RemainingQuantity        float64                  `json:"remainingQuantity"`
	RequestedDestination     RequestedDestination     `json:"requestedDestination"`
	DestinationLinkName      string                   `json:"destinationLinkName"`
	Price                    float64                  `json:"price"`
	StopPrice                float64                  `json:"stopPrice"`
	StopPriceLinkBasis       PriceLinkBasis           `json:"stopPriceLinkBasis"`
	StopPriceLinkType        PriceLinkType            `json:"stopPriceLinkType"`
	StopPriceOffset          float64                  `json:"stopPriceOffset"`
	StopType                 StopType                 `json:"stopType"`
	PriceLinkBasis           PriceLinkBasis           `json:"priceLinkBasis"`
	PriceLinkType            PriceLinkType            `json:"priceLinkType"`
	TaxLotMethod             TaxLotMethod             `json:"taxLotMethod"`
	ActivationPrice          float64                  `json:"activationPrice"`
	SpecialInstruction       SpecialInstruction       `json:"specialInstruction"`
	OrderStrategyType        OrderStrategyType        `json:"orderStrategyType"`
	CancelTime               string                   `json:"cancelTime"`
	ReleaseTime              string                   `json:"releaseTime"`
	OrderLegCollection       []OrderLeg               `json:"orderLegCollection"`
	OrderActivityCollection  []OrderActivity          `json:"orderActivityCollection"`
	ChildOrderStrategies     []*Order                 `json:"childOrderStrategies"`
	ReplacingOrderCollection []*Order                 `json:"replacingOrderCollection"`
	OrderID                  int64                    `json:"orderId"`
	Cancelable               bool                     `json:"cancelable"`
	Editable                 bool                     `json:"editable"`
	Status                   OrderStatus              `json:"status"`
	EnteredTime              string                   `json:"enteredTime"`
	CloseTime                string                   `json:"closeTime"`
	Tag                      string                   `json:"tag"`
	AccountNumber            int64                    `json:"accountNumber"`
	StatusDescription        string                   `json:"statusDescription"`
}

// OrderLeg represents one leg within an order.
type OrderLeg struct {
	OrderLegType   OrderLegType    `json:"orderLegType"`
	LegID          int64           `json:"legId"`
	Instrument     OrderInstrument `json:"instrument"`
	Instruction    Instruction     `json:"instruction"`
	PositionEffect PositionEffect  `json:"positionEffect"`
	Quantity       float64         `json:"quantity"`
	QuantityType   QuantityType    `json:"quantityType"`
	DivCapGains    DivCapGains     `json:"divCapGains"`
	ToSymbol       string          `json:"toSymbol"`
}

// OrderInstrument represents an instrument within an order leg.
type OrderInstrument struct {
	AssetType    schwab.AssetType `json:"assetType"`
	CUSIP        string           `json:"cusip"`
	Symbol       string           `json:"symbol"`
	Description  string           `json:"description"`
	InstrumentID int64            `json:"instrumentId"`
	Type         string           `json:"type"`
}

// OrderActivity represents activity recorded against an order.
type OrderActivity struct {
	ActivityType           string         `json:"activityType"`
	ExecutionType          string         `json:"executionType"`
	Quantity               float64        `json:"quantity"`
	OrderRemainingQuantity float64        `json:"orderRemainingQuantity"`
	ExecutionLegs          []ExecutionLeg `json:"executionLegs"`
}

// ExecutionLeg represents an execution for one order leg.
type ExecutionLeg struct {
	LegID             int64   `json:"legId"`
	Price             float64 `json:"price"`
	Quantity          float64 `json:"quantity"`
	MismarkedQuantity float64 `json:"mismarkedQuantity"`
	InstrumentID      int64   `json:"instrumentId"`
	Time              string  `json:"time"`
}

// PreviewOrder is the response returned by the order preview endpoint.
type PreviewOrder struct {
	OrderID               int64                  `json:"orderId"`
	OrderStrategy         *OrderStrategy         `json:"orderStrategy"`
	OrderValidationResult *OrderValidationResult `json:"orderValidationResult"`
	CommissionAndFee      *CommissionAndFee      `json:"commissionAndFee"`
}

// PreviewOrderRequest contains the request payload for the preview endpoint.
type PreviewOrderRequest struct {
	OrderID               int64                  `json:"orderId,omitempty"`
	OrderStrategy         *OrderStrategy         `json:"orderStrategy,omitempty"`
	OrderValidationResult *OrderValidationResult `json:"orderValidationResult,omitempty"`
	CommissionAndFee      *CommissionAndFee      `json:"commissionAndFee,omitempty"`
}

// OrderStrategy contains the previewed order strategy details.
type OrderStrategy struct {
	AccountNumber          string                   `json:"accountNumber"`
	AdvancedOrderType      AdvancedOrderType        `json:"advancedOrderType"`
	CloseTime              string                   `json:"closeTime"`
	EnteredTime            string                   `json:"enteredTime"`
	OrderBalance           *OrderBalance            `json:"orderBalance"`
	OrderStrategyType      OrderStrategyType        `json:"orderStrategyType"`
	OrderVersion           float64                  `json:"orderVersion"`
	Session                Session                  `json:"session"`
	Status                 OrderStatus              `json:"status"`
	AllOrNone              bool                     `json:"allOrNone"`
	Discretionary          bool                     `json:"discretionary"`
	Duration               Duration                 `json:"duration"`
	FilledQuantity         float64                  `json:"filledQuantity"`
	OrderType              OrderType                `json:"orderType"`
	OrderValue             float64                  `json:"orderValue"`
	Price                  float64                  `json:"price"`
	Quantity               float64                  `json:"quantity"`
	RemainingQuantity      float64                  `json:"remainingQuantity"`
	SellNonMarginableFirst bool                     `json:"sellNonMarginableFirst"`
	SettlementInstruction  SettlementInstruction    `json:"settlementInstruction"`
	AmountIndicator        AmountIndicator          `json:"amountIndicator"`
	Strategy               ComplexOrderStrategyType `json:"strategy"`
	OrderLegs              []PreviewOrderLeg        `json:"orderLegs"`
}

// OrderBalance contains estimated balance details for a previewed order.
type OrderBalance struct {
	OrderValue             float64 `json:"orderValue"`
	ProjectedAvailableFund float64 `json:"projectedAvailableFund"`
	ProjectedBuyingPower   float64 `json:"projectedBuyingPower"`
	ProjectedCommission    float64 `json:"projectedCommission"`
}

// PreviewOrderLeg represents one leg in an order preview response.
type PreviewOrderLeg struct {
	AskPrice            float64          `json:"askPrice"`
	AssetType           schwab.AssetType `json:"assetType"`
	BidPrice            float64          `json:"bidPrice"`
	FinalSymbol         string           `json:"finalSymbol"`
	Instruction         Instruction      `json:"instruction"`
	LastPrice           float64          `json:"lastPrice"`
	LegID               float64          `json:"legId"`
	MarkPrice           float64          `json:"markPrice"`
	ProjectedCommission float64          `json:"projectedCommission"`
	Quantity            float64          `json:"quantity"`
}

// OrderValidationResult contains validation details for a previewed order.
type OrderValidationResult struct {
	Alerts  []OrderValidationDetail `json:"alerts"`
	Accepts []OrderValidationDetail `json:"accepts"`
	Rejects []OrderValidationDetail `json:"rejects"`
	Reviews []OrderValidationDetail `json:"reviews"`
	Warns   []OrderValidationDetail `json:"warns"`
}

// OrderValidationDetail represents a single validation message.
type OrderValidationDetail struct {
	ValidationRuleName string `json:"validationRuleName"`
	Message            string `json:"message"`
	ActivityMessage    string `json:"activityMessage"`
	OriginalSeverity   string `json:"originalSeverity"`
	OverrideName       string `json:"overrideName"`
	OverrideSeverity   string `json:"overrideSeverity"`
}

// CommissionAndFee contains the commission and fee breakdown for a previewed order.
type CommissionAndFee struct {
	Commission     *Commission `json:"commission"`
	Fee            *Fees       `json:"fee"`
	TrueCommission *Commission `json:"trueCommission"`
}

// Commission contains commission leg details.
type Commission struct {
	CommissionLegs []CommissionLeg `json:"commissionLegs"`
}

// CommissionLeg represents a single commission leg.
type CommissionLeg struct {
	CommissionValues []CommissionValue `json:"commissionValues"`
}

// CommissionValue represents a single commission value entry.
type CommissionValue struct {
	Value float64 `json:"value"`
	Type  string  `json:"type"`
}

// Fees contains fee leg details.
type Fees struct {
	FeeLegs []FeeLeg `json:"feeLegs"`
}

// FeeLeg represents a single fee leg.
type FeeLeg struct {
	FeeValues []FeeValue `json:"feeValues"`
}

// FeeValue represents a single fee value entry.
type FeeValue struct {
	Value float64 `json:"value"`
	Type  string  `json:"type"`
}

// GetOrders retrieves orders for a single account.
func (c *Client) GetOrders(ctx context.Context, accountHash string, params *OrderListParams) ([]Order, error) {
	return c.getOrders(ctx, accountPath(accountHash, "orders"), params)
}

// CreateOrder creates an order for a single account.
func (c *Client) CreateOrder(ctx context.Context, accountHash string, order *OrderRequest) error {
	req, err := c.newRequest(ctx, "POST", accountPath(accountHash, "orders"), order)
	if err != nil {
		return err
	}
	return c.do(req, nil)
}

// GetOrder retrieves a single order for a single account.
func (c *Client) GetOrder(ctx context.Context, accountHash string, orderID int64) (*Order, error) {
	req, err := c.newRequest(ctx, "GET", accountPath(accountHash, "orders", strconv.FormatInt(orderID, 10)), nil)
	if err != nil {
		return nil, err
	}

	var result Order
	if doErr := c.do(req, &result); doErr != nil {
		return nil, doErr
	}
	return &result, nil
}

// ReplaceOrder replaces an existing order for a single account.
func (c *Client) ReplaceOrder(ctx context.Context, accountHash string, orderID int64, order *OrderRequest) error {
	req, err := c.newRequest(ctx, "PUT", accountPath(accountHash, "orders", strconv.FormatInt(orderID, 10)), order)
	if err != nil {
		return err
	}
	return c.do(req, nil)
}

// CancelOrder cancels an existing order for a single account.
func (c *Client) CancelOrder(ctx context.Context, accountHash string, orderID int64) error {
	req, err := c.newRequest(ctx, "DELETE", accountPath(accountHash, "orders", strconv.FormatInt(orderID, 10)), nil)
	if err != nil {
		return err
	}
	return c.do(req, nil)
}

// PreviewOrder previews commissions and fees for an order without placing it.
func (c *Client) PreviewOrder(
	ctx context.Context, accountHash string, order *PreviewOrderRequest,
) (*PreviewOrder, error) {
	req, err := c.newRequest(ctx, "POST", accountPath(accountHash, "previewOrder"), order)
	if err != nil {
		return nil, err
	}

	var result PreviewOrder
	if doErr := c.do(req, &result); doErr != nil {
		return nil, doErr
	}
	return &result, nil
}

// GetAllOrders retrieves orders across all accounts.
func (c *Client) GetAllOrders(ctx context.Context, params *OrderListParams) ([]Order, error) {
	return c.getOrders(ctx, "/orders", params)
}

func (c *Client) getOrders(ctx context.Context, path string, params *OrderListParams) ([]Order, error) {
	if params == nil {
		return nil, errors.New("order list params are required")
	}
	if params.FromEnteredTime == "" {
		return nil, errors.New("fromEnteredTime is required")
	}
	if params.ToEnteredTime == "" {
		return nil, errors.New("toEnteredTime is required")
	}

	req, err := c.newRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}

	q := req.URL.Query()
	q.Set("fromEnteredTime", params.FromEnteredTime)
	q.Set("toEnteredTime", params.ToEnteredTime)
	setOptionalInt(q, "maxResults", params.MaxResults)
	setOptionalString(q, "status", string(params.Status))
	req.URL.RawQuery = q.Encode()

	var result []Order
	if doErr := c.do(req, &result); doErr != nil {
		return nil, doErr
	}
	return result, nil
}
