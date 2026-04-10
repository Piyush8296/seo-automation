package parser

import (
	"net/http"
	"strings"
)

// ExtractHeaders converts http.Header to a lowercase-keyed map.
func ExtractHeaders(headers http.Header) map[string]string {
	m := make(map[string]string, len(headers))
	for k, v := range headers {
		if len(v) > 0 {
			m[strings.ToLower(k)] = v[0]
		}
	}
	return m
}

// GetHeader returns the value of a header by name (case-insensitive).
func GetHeader(headers map[string]string, name string) string {
	return headers[strings.ToLower(name)]
}
