package trader

import (
	"net/url"
	"strconv"
	"strings"
)

func setOptionalString(q url.Values, key, value string) {
	if value != "" {
		q.Set(key, value)
	}
}

func setOptionalInt(q url.Values, key string, value int) {
	if value != 0 {
		q.Set(key, strconv.Itoa(value))
	}
}

func accountPath(accountHash string, segments ...string) string {
	parts := make([]string, 0, 2+len(segments))
	parts = append(parts, "accounts", url.PathEscape(accountHash))
	parts = append(parts, segments...)
	return "/" + strings.Join(parts, "/")
}
