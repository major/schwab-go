package marketdata

import "testing"

func TestMarketDataModelsMatchOpenAPISchemaProperties(t *testing.T) {
	openapi := newOpenAPIValidator(t)
	tests := []struct {
		name      string
		component string
		value     any
	}{
		{
			name:      "candle",
			component: "Candle",
			value:     Candle{},
		},
		{
			name:      "candle_list",
			component: "CandleList",
			value:     CandleList{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			openapi.ValidateGoSchemaProperties(t, tt.component, tt.value)
		})
	}
}
