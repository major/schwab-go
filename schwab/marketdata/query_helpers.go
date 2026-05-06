package marketdata

import (
	"net/url"
	"strconv"
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

func setOptionalInt64(q url.Values, key string, value int64) {
	if value != 0 {
		q.Set(key, strconv.FormatInt(value, 10))
	}
}

func setOptionalFloat64(q url.Values, key string, value float64) {
	if value != 0 {
		q.Set(key, strconv.FormatFloat(value, 'f', -1, 64))
	}
}

func setOptionalBool(q url.Values, key string, value *bool) {
	if value != nil {
		q.Set(key, strconv.FormatBool(*value))
	}
}
