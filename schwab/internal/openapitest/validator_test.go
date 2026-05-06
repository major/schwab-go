package openapitest

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestValidatorValidatesRequestAndJSONResponse(t *testing.T) {
	validator := NewValidator(t, "market_data.openapi.json")
	request := httptest.NewRequest(http.MethodGet, "/pricehistory?symbol=AAPL", nil)
	response := map[string]any{
		"candles": []map[string]any{
			{
				"open":     150.0,
				"high":     155.0,
				"low":      149.0,
				"close":    154.0,
				"volume":   1000000.0,
				"datetime": 1625097600000.0,
			},
		},
		"symbol": "AAPL",
		"empty":  false,
	}

	validator.ValidateRequest(t, request, "getPriceHistory")
	validator.ValidateJSONResponse(t, request, "getPriceHistory", http.StatusOK, response)
}
