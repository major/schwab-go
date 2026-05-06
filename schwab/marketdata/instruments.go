package marketdata

import (
	"context"
	"fmt"
	"net/url"

	schwab "github.com/major/schwab-go/schwab"
)

// InstrumentProjection defines the type of instrument search.
type InstrumentProjection string

const (
	ProjectionSymbolSearch InstrumentProjection = "symbol-search"
	ProjectionSymbolRegex  InstrumentProjection = "symbol-regex"
	ProjectionDescSearch   InstrumentProjection = "desc-search"
	ProjectionDescRegex    InstrumentProjection = "desc-regex"
	ProjectionSearch       InstrumentProjection = "search"
	ProjectionFundamental  InstrumentProjection = "fundamental"
)

// InstrumentResponse is the response from GET /instruments.
type InstrumentResponse struct {
	Instruments []Instrument `json:"instruments"`
}

// Instrument represents a financial instrument.
type Instrument struct {
	Cusip              string           `json:"cusip"`
	Symbol             string           `json:"symbol"`
	Description        string           `json:"description"`
	Exchange           string           `json:"exchange"`
	AssetType          schwab.AssetType `json:"assetType"`
	BondFactor         string           `json:"bondFactor,omitempty"`
	BondMultiplier     string           `json:"bondMultiplier,omitempty"`
	BondPrice          float64          `json:"bondPrice,omitempty"`
	BondInstrumentInfo *BondInfo        `json:"bondInstrumentInfo,omitempty"`
	InstrumentInfo     *InstrumentInfo  `json:"instrumentInfo,omitempty"`
	Type               string           `json:"type,omitempty"`
	Fundamental        *FundamentalData `json:"fundamental,omitempty"`
}

// BondInfo contains bond-specific instrument details.
type BondInfo struct {
	AssetType      schwab.AssetType `json:"assetType"`
	BondFactor     string           `json:"bondFactor"`
	BondMultiplier string           `json:"bondMultiplier"`
	BondPrice      float64          `json:"bondPrice"`
	Cusip          string           `json:"cusip"`
	Description    string           `json:"description"`
	Exchange       string           `json:"exchange"`
	Symbol         string           `json:"symbol"`
	Type           string           `json:"type"`
}

// InstrumentInfo contains basic instrument details for nested references.
type InstrumentInfo struct {
	AssetType   schwab.AssetType `json:"assetType"`
	Cusip       string           `json:"cusip"`
	Description string           `json:"description"`
	Exchange    string           `json:"exchange"`
	Symbol      string           `json:"symbol"`
	Type        string           `json:"type"`
}

// FundamentalData contains financial metrics returned when projection=fundamental.
type FundamentalData struct {
	Symbol                  string  `json:"symbol"`
	High52                  float64 `json:"high52"`
	Low52                   float64 `json:"low52"`
	DividendAmount          float64 `json:"dividendAmount"`
	DividendYield           float64 `json:"dividendYield"`
	DividendDate            string  `json:"dividendDate"`
	PeRatio                 float64 `json:"peRatio"`
	PegRatio                float64 `json:"pegRatio"`
	PbRatio                 float64 `json:"pbRatio"`
	PrRatio                 float64 `json:"prRatio"`
	PcfRatio                float64 `json:"pcfRatio"`
	GrossMarginTTM          float64 `json:"grossMarginTTM"`
	GrossMarginMRQ          float64 `json:"grossMarginMRQ"`
	NetProfitMarginTTM      float64 `json:"netProfitMarginTTM"`
	NetProfitMarginMRQ      float64 `json:"netProfitMarginMRQ"`
	OperatingMarginTTM      float64 `json:"operatingMarginTTM"`
	OperatingMarginMRQ      float64 `json:"operatingMarginMRQ"`
	ReturnOnEquity          float64 `json:"returnOnEquity"`
	ReturnOnAssets          float64 `json:"returnOnAssets"`
	ReturnOnInvestment      float64 `json:"returnOnInvestment"`
	QuickRatio              float64 `json:"quickRatio"`
	CurrentRatio            float64 `json:"currentRatio"`
	InterestCoverage        float64 `json:"interestCoverage"`
	TotalDebtToCapital      float64 `json:"totalDebtToCapital"`
	LtDebtToEquity          float64 `json:"ltDebtToEquity"`
	TotalDebtToEquity       float64 `json:"totalDebtToEquity"`
	EpsTTM                  float64 `json:"epsTTM"`
	EpsChangePercentTTM     float64 `json:"epsChangePercentTTM"`
	EpsChangeYear           float64 `json:"epsChangeYear"`
	EpsChange               float64 `json:"epsChange"`
	RevChangeYear           float64 `json:"revChangeYear"`
	RevChangeTTM            float64 `json:"revChangeTTM"`
	RevChangeIn             float64 `json:"revChangeIn"`
	SharesOutstanding       float64 `json:"sharesOutstanding"`
	MarketCapFloat          float64 `json:"marketCapFloat"`
	MarketCap               float64 `json:"marketCap"`
	BookValuePerShare       float64 `json:"bookValuePerShare"`
	ShortIntToFloat         float64 `json:"shortIntToFloat"`
	ShortIntDayToCover      float64 `json:"shortIntDayToCover"`
	DivGrowthRate3Year      float64 `json:"divGrowthRate3Year"`
	DividendPayAmount       float64 `json:"dividendPayAmount"`
	DividendPayDate         string  `json:"dividendPayDate"`
	Beta                    float64 `json:"beta"`
	Vol1DayAvg              float64 `json:"vol1DayAvg"`
	Vol10DayAvg             float64 `json:"vol10DayAvg"`
	Vol3MonthAvg            float64 `json:"vol3MonthAvg"`
	Avg1DayVolume           int64   `json:"avg1DayVolume"`
	Avg10DaysVolume         int64   `json:"avg10DaysVolume"`
	Avg3MonthVolume         int64   `json:"avg3MonthVolume"`
	Avg1YearVolume          int64   `json:"avg1YearVolume"`
	DeclarationDate         string  `json:"declarationDate"`
	DividendFreq            int     `json:"dividendFreq"`
	Eps                     float64 `json:"eps"`
	DtnVolume               int64   `json:"dtnVolume"`
	NextDividendPayDate     string  `json:"nextDividendPayDate"`
	NextDividendDate        string  `json:"nextDividendDate"`
	FundLeverageFactor      float64 `json:"fundLeverageFactor"`
	FundStrategy            string  `json:"fundStrategy"`
	CorpactionDate          string  `json:"corpactionDate"`
}

// SearchInstruments searches for instruments by symbol with the given projection.
// symbol: the symbol to search for (required)
// projection: the type of search to perform (required)
func (c *Client) SearchInstruments(ctx context.Context, symbol string, projection InstrumentProjection) (*InstrumentResponse, error) {
	req, err := c.newRequest(ctx, "GET", "/instruments", nil)
	if err != nil {
		return nil, err
	}

	q := req.URL.Query()
	q.Set("symbol", symbol)
	q.Set("projection", string(projection))
	req.URL.RawQuery = q.Encode()

	var result InstrumentResponse
	if err := c.do(req, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// GetInstrumentByCUSIP retrieves an instrument by its CUSIP ID.
// cusipID: the CUSIP identifier
func (c *Client) GetInstrumentByCUSIP(ctx context.Context, cusipID string) (*Instrument, error) {
	path := fmt.Sprintf("/instruments/%s", url.PathEscape(cusipID))
	req, err := c.newRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}

	var result Instrument
	if err := c.do(req, &result); err != nil {
		return nil, err
	}
	return &result, nil
}
