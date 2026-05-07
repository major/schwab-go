package marketdata_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"strings"

	schwab "github.com/major/schwab-go/schwab"
	"github.com/major/schwab-go/schwab/marketdata"
)

func newMarketDataExampleClient(handler http.HandlerFunc) (*marketdata.Client, func()) {
	wrapped := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handler(w, r)
	})
	server := httptest.NewServer(http.StripPrefix("/marketdata/v1", wrapped))

	client := marketdata.NewClient(
		schwab.WithToken("example-token"),
		schwab.WithHTTPClient(server.Client()),
		schwab.WithBaseURL(server.URL),
	)
	return client, server.Close
}

func writeExampleJSON(w http.ResponseWriter, value any) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(value); err != nil {
		panic(err)
	}
}

func mustRawExample(value any) json.RawMessage {
	raw, err := json.Marshal(value)
	if err != nil {
		panic(err)
	}
	return raw
}

func printExampleLines(lines ...string) {
	var output strings.Builder
	for _, line := range lines {
		output.WriteString(line)
		output.WriteByte('\n')
	}
	if _, err := os.Stdout.WriteString(output.String()); err != nil {
		panic(err)
	}
}

func formatExampleFloat(value float64) string {
	return strconv.FormatFloat(value, 'f', 2, 64)
}
