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

func TestUrlNonASCII(t *testing.T) {
	check := &urlNonASCII{}

	tests := []struct {
		name    string
		url     string
		wantHit bool
	}{
		{"ascii", "https://example.com/used-cars/bangalore", false},
		{"direct unicode", "https://example.com/café", true},
		{"encoded unicode", "https://example.com/caf%C3%A9", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results := check.Run(&models.PageData{URL: tt.url, FinalURL: tt.url})
			if tt.wantHit && len(results) == 0 {
				t.Fatalf("expected check to fire for %s", tt.url)
			}
			if !tt.wantHit && len(results) > 0 {
				t.Fatalf("expected no check for %s, got: %s", tt.url, results[0].Message)
			}
		})
	}
}

func TestUrlKeywordTopicMismatch(t *testing.T) {
	check := &urlKeywordTopicMismatch{}

	mismatch := &models.PageData{
		URL:      "https://example.com/used-cars/suv-bangalore",
		FinalURL: "https://example.com/used-cars/suv-bangalore",
		Title:    "Car finance calculator",
		H1s:      []string{"Loan eligibility"},
		BodyText: "Compare loan tenures and down payment options.",
	}
	if results := check.Run(mismatch); len(results) == 0 || results[0].ID != "url.keyword_topic_mismatch" {
		t.Fatalf("expected keyword topic mismatch, got %#v", results)
	}

	match := &models.PageData{
		URL:      "https://example.com/used-cars/suv-bangalore",
		FinalURL: "https://example.com/used-cars/suv-bangalore",
		Title:    "Used SUV cars in Bangalore",
		H1s:      []string{"SUV cars for sale in Bangalore"},
	}
	if results := check.Run(match); len(results) > 0 {
		t.Fatalf("expected matching URL keywords to pass, got %#v", results)
	}
}

func TestUrlPrintCanonicalConflict(t *testing.T) {
	check := &urlPrintCanonicalConflict{}

	missing := &models.PageData{
		URL:      "https://example.com/cars/honda-city?print=1",
		FinalURL: "https://example.com/cars/honda-city?print=1",
	}
	if results := check.Run(missing); len(results) == 0 || results[0].ID != "url.print_canonical_conflict" {
		t.Fatalf("expected missing print canonical conflict, got %#v", results)
	}

	selfCanonical := &models.PageData{
		URL:       "https://example.com/print/honda-city",
		FinalURL:  "https://example.com/print/honda-city",
		Canonical: "https://example.com/print/honda-city",
	}
	if results := check.Run(selfCanonical); len(results) == 0 {
		t.Fatalf("expected self-canonical print URL to fail")
	}

	mainCanonical := &models.PageData{
		URL:       "https://example.com/print/honda-city",
		FinalURL:  "https://example.com/print/honda-city",
		Canonical: "https://example.com/cars/honda-city",
	}
	if results := check.Run(mainCanonical); len(results) > 0 {
		t.Fatalf("expected print URL canonicalized to main page to pass, got %#v", results)
	}
}

func TestUrlBreadcrumbMismatch(t *testing.T) {
	check := &urlBreadcrumbMismatch{}

	mismatch := &models.PageData{
		URL:      "https://example.com/cars/maruti-swift",
		FinalURL: "https://example.com/cars/maruti-swift",
		Title:    "Maruti Swift used car",
		SchemaJSONRaw: []string{`{
			"@context":"https://schema.org",
			"@type":"BreadcrumbList",
			"itemListElement":[
				{"@type":"ListItem","position":1,"name":"Cars","item":"https://example.com/cars"},
				{"@type":"ListItem","position":2,"name":"Honda City","item":"https://example.com/cars/honda-city"}
			]
		}`},
	}
	if results := check.Run(mismatch); len(results) == 0 || results[0].ID != "url.breadcrumb_mismatch" {
		t.Fatalf("expected breadcrumb mismatch, got %#v", results)
	}

	match := &models.PageData{
		URL:      "https://example.com/cars/maruti-swift",
		FinalURL: "https://example.com/cars/maruti-swift",
		Title:    "Maruti Swift used car",
		SchemaJSONRaw: []string{`{
			"@context":"https://schema.org",
			"@type":"BreadcrumbList",
			"itemListElement":[
				{"@type":"ListItem","position":1,"name":"Cars","item":"https://example.com/cars"},
				{"@type":"ListItem","position":2,"name":"Maruti Swift","item":"https://example.com/cars/maruti-swift"}
			]
		}`},
	}
	if results := check.Run(match); len(results) > 0 {
		t.Fatalf("expected matching breadcrumb to pass, got %#v", results)
	}
}

func TestUrlConsistentStructure(t *testing.T) {
	check := &urlConsistentStructure{}

	pages := []*models.PageData{
		{URL: "https://example.com/cars/honda-city", FinalURL: "https://example.com/cars/honda-city", StatusCode: 200},
		{URL: "https://example.com/cars/maruti-swift", FinalURL: "https://example.com/cars/maruti-swift", StatusCode: 200},
		{URL: "https://example.com/cars/used/delhi/suv/cheap", FinalURL: "https://example.com/cars/used/delhi/suv/cheap", StatusCode: 200},
		{URL: "https://example.com/cars/used/mumbai/sedan/cheap", FinalURL: "https://example.com/cars/used/mumbai/sedan/cheap", StatusCode: 200},
	}
	if results := check.Run(pages); len(results) == 0 || results[0].ID != "url.consistent_structure" {
		t.Fatalf("expected inconsistent structure finding, got %#v", results)
	}
}

func TestUrlTrailingSlashInconsistent(t *testing.T) {
	check := &urlTrailingSlashInconsistent{}

	pages := []*models.PageData{
		{URL: "https://example.com/cars", FinalURL: "https://example.com/cars", StatusCode: 200},
		{URL: "https://example.com/cars/", FinalURL: "https://example.com/cars/", StatusCode: 200},
	}
	if results := check.Run(pages); len(results) == 0 || results[0].ID != "url.trailing_slash_inconsistent" {
		t.Fatalf("expected trailing slash inconsistency, got %#v", results)
	}
}
