package url_structure

import (
	"testing"

	"github.com/cars24/seo-automation/internal/models"
)

func TestUrlPathDepthTooDeep(t *testing.T) {
	check := &urlPathDepthTooDeep{}

	tests := []struct {
		name    string
		url     string
		wantHit bool
	}{
		{"root", "https://example.com/", false},
		{"1 segment", "https://example.com/cars", false},
		{"4 segments ok", "https://example.com/a/b/c/d", false},
		{"5 segments too deep", "https://example.com/a/b/c/d/e", true},
		{"6 segments too deep", "https://example.com/a/b/c/d/e/f", true},
		{"trailing slash 4", "https://example.com/a/b/c/d/", false},
		{"trailing slash 5", "https://example.com/a/b/c/d/e/", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &models.PageData{URL: tt.url}
			results := check.Run(p)
			if tt.wantHit && len(results) == 0 {
				t.Errorf("expected check to fire for %s", tt.url)
			}
			if !tt.wantHit && len(results) > 0 {
				t.Errorf("expected no check for %s, got: %s", tt.url, results[0].Message)
			}
			if tt.wantHit && len(results) > 0 && results[0].ID != "url.path_depth_too_deep" {
				t.Errorf("wrong ID: %s", results[0].ID)
			}
		})
	}
}

func TestUrlContainsStopWords(t *testing.T) {
	check := &urlContainsStopWords{}

	tests := []struct {
		name    string
		url     string
		wantHit bool
	}{
		{"clean url", "https://example.com/buy-used-cars", false},
		{"has 'the'", "https://example.com/the-best-cars", true},
		{"has 'and'", "https://example.com/cars-and-bikes", true},
		{"has 'for'", "https://example.com/tips-for-buying", true},
		{"root path", "https://example.com/", false},
		{"stop word in segment", "https://example.com/blog/a-guide", true},
		{"underscore separator", "https://example.com/how_to_buy", true},
		{"no stop words with hyphens", "https://example.com/used-car-prices", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &models.PageData{URL: tt.url}
			results := check.Run(p)
			if tt.wantHit && len(results) == 0 {
				t.Errorf("expected check to fire for %s", tt.url)
			}
			if !tt.wantHit && len(results) > 0 {
				t.Errorf("expected no check for %s, got: %s", tt.url, results[0].Message)
			}
			if tt.wantHit && len(results) > 0 && results[0].ID != "url.contains_stop_words" {
				t.Errorf("wrong ID: %s", results[0].ID)
			}
		})
	}
}
