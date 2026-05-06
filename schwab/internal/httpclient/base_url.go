package httpclient

import (
	"net/url"
	"strings"
)

// WithPathPrefix returns a copy of baseURL with pathPrefix appended once.
func WithPathPrefix(baseURL *url.URL, pathPrefix string) *url.URL {
	if baseURL == nil {
		baseURL = &url.URL{}
	}
	if hasPathSuffix(baseURL.Path, pathPrefix) {
		clone := *baseURL
		return &clone
	}
	return baseURL.JoinPath(pathPrefix)
}

func hasPathSuffix(basePath, suffix string) bool {
	return strings.HasSuffix(strings.TrimRight(basePath, "/"), strings.TrimRight(suffix, "/"))
}
