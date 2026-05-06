package marketdata

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	schwab "github.com/major/schwab-go/schwab"
	"github.com/major/schwab-go/schwab/internal/openapitest"
)

func writeJSON(t *testing.T, w http.ResponseWriter, value any) {
	t.Helper()
	require.NoError(t, json.NewEncoder(w).Encode(value))
}

func newTestClient(t *testing.T, handler http.HandlerFunc) *Client {
	t.Helper()
	ts := httptest.NewServer(http.StripPrefix(apiPathPrefix, handler))
	t.Cleanup(ts.Close)
	return NewClient(
		schwab.WithToken("test-token"),
		schwab.WithHTTPClient(ts.Client()),
		schwab.WithBaseURL(ts.URL),
	)
}

func newOpenAPIValidator(t *testing.T) *openapitest.Validator {
	t.Helper()
	return openapitest.NewValidator(t, "market_data.openapi.json")
}
