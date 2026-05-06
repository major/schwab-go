// Package openapitest provides OpenAPI contract assertions for package tests.
package openapitest

import (
	"context"
	"encoding/json"
	"net/http"
	"path/filepath"
	"runtime"
	"slices"
	"sort"
	"strings"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/getkin/kin-openapi/openapi3gen"
	"github.com/getkin/kin-openapi/routers"
	"github.com/getkin/kin-openapi/routers/gorillamux"
	"github.com/stretchr/testify/require"
)

// Validator validates test HTTP traffic against one checked-in OpenAPI document.
type Validator struct {
	doc     *openapi3.T
	router  routers.Router
	options *openapi3filter.Options
}

// NewValidator loads and validates an OpenAPI specification from docs/.
func NewValidator(t testing.TB, specName string) *Validator {
	t.Helper()

	specPath := repoPath(t, "docs", specName)
	ctx := context.Background()
	loader := &openapi3.Loader{Context: ctx, IsExternalRefsAllowed: true}
	doc, err := loader.LoadFromFile(specPath)
	require.NoError(t, err)

	// Tests run against httptest servers, not Schwab's production host. A relative
	// server lets kin-openapi match the request path while still validating the
	// checked-in operation contracts.
	doc.Servers = openapi3.Servers{{URL: "/"}}
	require.NoError(t, doc.Validate(ctx, openapi3.DisableExamplesValidation()))

	router, err := gorillamux.NewRouter(doc)
	require.NoError(t, err)

	return &Validator{
		doc:    doc,
		router: router,
		options: &openapi3filter.Options{
			AuthenticationFunc:    openapi3filter.NoopAuthenticationFunc,
			IncludeResponseStatus: true,
		},
	}
}

// ValidateRequest validates an HTTP request against the operation it should match.
func (v *Validator) ValidateRequest(t testing.TB, r *http.Request, operationID string) {
	t.Helper()

	input := v.requestValidationInput(t, r, operationID)
	require.NoError(t, openapi3filter.ValidateRequest(context.Background(), input))
}

// ValidateJSONResponse validates a JSON response payload against the matched operation.
func (v *Validator) ValidateJSONResponse(t testing.TB, r *http.Request, operationID string, status int, value any) {
	t.Helper()

	body, err := json.Marshal(value)
	require.NoError(t, err)

	input := &openapi3filter.ResponseValidationInput{
		RequestValidationInput: v.requestValidationInput(t, r, operationID),
		Status:                 status,
		Header:                 http.Header{"Content-Type": []string{"application/json"}},
		Options:                v.options,
	}
	input.SetBodyBytes(body)

	require.NoError(t, openapi3filter.ValidateResponse(context.Background(), input))
}

// ValidateGoSchemaProperties validates that a Go value exposes the same JSON
// property names and basic wire shapes as a component schema.
func (v *Validator) ValidateGoSchemaProperties(t testing.TB, componentName string, value any) {
	t.Helper()

	specRef := v.doc.Components.Schemas[componentName]
	require.NotNil(t, specRef, "OpenAPI component %q must exist", componentName)
	require.NotNil(t, specRef.Value, "OpenAPI component %q must be resolved", componentName)

	generatedRef, err := openapi3gen.NewSchemaRefForValue(value, openapi3.Schemas{})
	require.NoError(t, err)
	require.NotNil(t, generatedRef.Value, "generated schema for %T must be resolved", value)

	want := summarizeSchemaProperties(specRef.Value.Properties)
	got := summarizeSchemaProperties(generatedRef.Value.Properties)
	require.Equal(t, want, got, "%T JSON schema properties must match OpenAPI component %q", value, componentName)
}

func (v *Validator) requestValidationInput(
	t testing.TB,
	r *http.Request,
	operationID string,
) *openapi3filter.RequestValidationInput {
	t.Helper()

	route, pathParams, err := v.router.FindRoute(r)
	require.NoError(t, err)
	require.Equal(t, operationID, route.Operation.OperationID)

	return &openapi3filter.RequestValidationInput{
		Request:    r,
		PathParams: pathParams,
		Route:      route,
		Options:    v.options,
	}
}

func repoPath(t testing.TB, elem ...string) string {
	t.Helper()

	_, file, _, ok := runtime.Caller(0)
	require.True(t, ok)
	parts := append([]string{filepath.Dir(file), "..", "..", ".."}, elem...)
	return filepath.Clean(filepath.Join(parts...))
}

type propertySummary struct {
	Types    []string `json:"types,omitempty"`
	Format   string   `json:"format,omitempty"`
	Ref      string   `json:"ref,omitempty"`
	Nullable bool     `json:"nullable,omitempty"`
	Items    *propertySummary
}

func summarizeSchemaProperties(properties openapi3.Schemas) map[string]propertySummary {
	summary := make(map[string]propertySummary, len(properties))
	for name, property := range properties {
		summary[name] = summarizeSchemaRef(property)
	}
	return summary
}

func summarizeSchemaRef(ref *openapi3.SchemaRef) propertySummary {
	if ref == nil {
		return propertySummary{}
	}

	summary := propertySummary{}
	if ref.Value == nil {
		summary.Ref = componentName(ref.Ref)
		return summary
	}

	summary.Types = schemaTypes(ref.Value.Type)
	if !schemaHasType(summary.Types, "string") {
		summary.Format = ref.Value.Format
	}
	summary.Nullable = ref.Value.Nullable
	if ref.Value.Items != nil {
		items := summarizeSchemaRef(ref.Value.Items)
		summary.Items = &items
	}
	return summary
}

func schemaTypes(types *openapi3.Types) []string {
	if types == nil {
		return nil
	}

	values := append([]string(nil), types.Slice()...)
	sort.Strings(values)
	return values
}

func schemaHasType(types []string, target string) bool {
	return slices.Contains(types, target)
}

func componentName(ref string) string {
	if ref == "" {
		return ""
	}
	if index := strings.LastIndex(ref, "/"); index >= 0 {
		return ref[index+1:]
	}
	return ref
}
