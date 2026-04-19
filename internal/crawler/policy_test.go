package crawler

import (
	"testing"

	"github.com/cars24/seo-automation/internal/models"
)

func TestEffectiveScopePrefix_DefaultsToSeedPath(t *testing.T) {
	cfg := &models.CrawlConfig{
		SeedURL: "https://www.example.com/buy-used-cars/city/",
	}
	if got, want := effectiveScopePrefix(cfg), "/buy-used-cars/city"; got != want {
		t.Fatalf("effectiveScopePrefix() = %q, want %q", got, want)
	}
}

func TestPathWithinScope(t *testing.T) {
	tests := []struct {
		name   string
		rawURL string
		prefix string
		want   bool
	}{
		{"same path", "https://example.com/buy-used-cars", "/buy-used-cars", true},
		{"child path", "https://example.com/buy-used-cars/delhi", "/buy-used-cars", true},
		{"sibling path", "https://example.com/sell-used-cars", "/buy-used-cars", false},
		{"prefix lookalike", "https://example.com/buy-used-cars-plus", "/buy-used-cars", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := pathWithinScope(tt.rawURL, tt.prefix); got != tt.want {
				t.Fatalf("pathWithinScope(%q, %q) = %t, want %t", tt.rawURL, tt.prefix, got, tt.want)
			}
		})
	}
}

func TestShouldCrawlURL_AppliesScopeAndURLFilters(t *testing.T) {
	cfg := &models.CrawlConfig{
		SeedURL:        "https://www.example.com/buy-used-cars/",
		Scope:          models.CrawlScopeSubfolder,
		MaxURLLength:   60,
		MaxQueryParams: 2,
	}

	if !shouldCrawlURL("https://www.example.com/buy-used-cars/delhi", cfg) {
		t.Fatalf("expected in-scope URL to be crawlable")
	}
	if shouldCrawlURL("https://www.example.com/sell-used-cars", cfg) {
		t.Fatalf("expected out-of-scope URL to be rejected")
	}
	if shouldCrawlURL("https://blog.example.com/buy-used-cars/delhi", cfg) {
		t.Fatalf("expected different host to be rejected")
	}
	if shouldCrawlURL("https://www.example.com/buy-used-cars/delhi?one=1&two=2&three=3", cfg) {
		t.Fatalf("expected URL with too many query params to be rejected")
	}
	if shouldCrawlURL("https://www.example.com/buy-used-cars/this-is-a-very-long-city-slug", cfg) {
		t.Fatalf("expected URL longer than max length to be rejected")
	}
}

func TestCanExpandFromPage_RespectsNoindexAndCanonicalSettings(t *testing.T) {
	page := &models.PageData{
		URL:              "https://example.com/a",
		FinalURL:         "https://example.com/a",
		Canonical:        "https://example.com/b",
		RobotsDirectives: []string{"noindex"},
	}

	if canExpandFromPage(page, &models.CrawlConfig{
		ExpandNoindexPages:       false,
		ExpandCanonicalizedPages: true,
	}) {
		t.Fatalf("expected noindex page expansion to be blocked")
	}

	if canExpandFromPage(page, &models.CrawlConfig{
		ExpandNoindexPages:       true,
		ExpandCanonicalizedPages: false,
	}) {
		t.Fatalf("expected canonicalized page expansion to be blocked")
	}

	if !canExpandFromPage(page, &models.CrawlConfig{
		ExpandNoindexPages:       true,
		ExpandCanonicalizedPages: true,
	}) {
		t.Fatalf("expected page expansion to be allowed when both toggles are on")
	}
}
