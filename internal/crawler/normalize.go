package crawler

import (
	"net/url"
	"strings"
)

// NormalizeURL resolves rawURL relative to base, strips fragment, lowercases scheme+host.
func NormalizeURL(rawURL, base string) (string, error) {
	baseU, err := url.Parse(base)
	if err != nil {
		return "", err
	}
	ref, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}
	resolved := baseU.ResolveReference(ref)
	// Strip fragment
	resolved.Fragment = ""
	// Lowercase scheme and host
	resolved.Scheme = strings.ToLower(resolved.Scheme)
	resolved.Host = strings.ToLower(resolved.Host)
	return resolved.String(), nil
}

// DedupeKey returns a canonical key for deduplication (no trailing slash, lowercase).
func DedupeKey(u string) string {
	parsed, err := url.Parse(u)
	if err != nil {
		return strings.ToLower(u)
	}
	parsed.Fragment = ""
	parsed.Scheme = strings.ToLower(parsed.Scheme)
	parsed.Host = strings.ToLower(parsed.Host)
	// Remove trailing slash from path (except root)
	if len(parsed.Path) > 1 {
		parsed.Path = strings.TrimRight(parsed.Path, "/")
	}
	return parsed.String()
}

// SameHost returns true if a and b share the same host.
func SameHost(a, b string) bool {
	au, err := url.Parse(a)
	if err != nil {
		return false
	}
	bu, err := url.Parse(b)
	if err != nil {
		return false
	}
	return strings.ToLower(au.Hostname()) == strings.ToLower(bu.Hostname())
}

// IsHTTPScheme returns true if the URL uses http or https.
func IsHTTPScheme(u string) bool {
	parsed, err := url.Parse(u)
	if err != nil {
		return false
	}
	s := strings.ToLower(parsed.Scheme)
	return s == "http" || s == "https"
}

// OriginOf returns "scheme://host" for a URL.
func OriginOf(u string) string {
	parsed, err := url.Parse(u)
	if err != nil {
		return ""
	}
	return strings.ToLower(parsed.Scheme) + "://" + strings.ToLower(parsed.Host)
}
