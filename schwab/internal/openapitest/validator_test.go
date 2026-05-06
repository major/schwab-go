package openapitest

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/stretchr/testify/require"
)

type schemaFixture struct {
	ID     string   `json:"id"`
	Amount float64  `json:"amount"`
	Tags   []string `json:"tags"`
}

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

func TestValidateGoSchemaProperties(t *testing.T) {
	validator := &Validator{
		doc: &openapi3.T{
			Components: &openapi3.Components{
				Schemas: openapi3.Schemas{
					"SchemaFixture": openapi3.NewSchemaRef("", openapi3.NewObjectSchema().
						WithProperty("id", openapi3.NewStringSchema()).
						WithProperty("amount", openapi3.NewFloat64Schema().WithFormat("double")).
						WithProperty("tags", openapi3.NewArraySchema().WithItems(openapi3.NewStringSchema()))),
				},
			},
		},
	}

	validator.ValidateGoSchemaProperties(t, "SchemaFixture", schemaFixture{})
}

func TestSummarizeSchemaRef(t *testing.T) {
	tests := []struct {
		name string
		ref  *openapi3.SchemaRef
		want propertySummary
	}{
		{
			name: "nil ref",
			want: propertySummary{},
		},
		{
			name: "unresolved component ref",
			ref:  openapi3.NewSchemaRef("#/components/schemas/Candle", nil),
			want: propertySummary{Ref: "Candle"},
		},
		{
			name: "string format ignored",
			ref:  openapi3.NewSchemaRef("", openapi3.NewStringSchema().WithFormat("yyyy-MM-dd")),
			want: propertySummary{Types: []string{"string"}},
		},
		{
			name: "non-string format retained",
			ref:  openapi3.NewSchemaRef("", openapi3.NewFloat64Schema().WithFormat("double")),
			want: propertySummary{Types: []string{"number"}, Format: "double"},
		},
		{
			name: "nullable array with summarized items",
			ref: openapi3.NewSchemaRef("", openapi3.NewArraySchema().
				WithItems(openapi3.NewStringSchema()).
				WithNullable()),
			want: propertySummary{
				Types:    []string{"array"},
				Nullable: true,
				Items:    &propertySummary{Types: []string{"string"}},
			},
		},
		{
			name: "types sorted before comparison",
			ref:  openapi3.NewSchemaRef("", &openapi3.Schema{Type: &openapi3.Types{"string", "number"}}),
			want: propertySummary{Types: []string{"number", "string"}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := summarizeSchemaRef(tt.ref)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestComponentName(t *testing.T) {
	tests := []struct {
		name string
		ref  string
		want string
	}{
		{name: "empty", want: ""},
		{name: "plain name", ref: "Candle", want: "Candle"},
		{name: "component path", ref: "#/components/schemas/Candle", want: "Candle"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := componentName(tt.ref)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestSchemaTypesReturnsNilForMissingTypes(t *testing.T) {
	got := schemaTypes(nil)
	require.Nil(t, got)
}
