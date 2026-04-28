package crawlability

import (
	"strings"
	"testing"

	"github.com/cars24/seo-automation/internal/models"
)

func TestRedirect302PermanentUsesRedirectHopStatus(t *testing.T) {
	page := redirectPage("https://example.com/old", "https://example.com/new", 302)

	results := (&redirect302Permanent{}).Run(page)

	assertCheck(t, results, "crawl.redirect.302_permanent")
}

func TestSoft404DetectedFromTitleAndH1(t *testing.T) {
	page := htmlPage(`<html><head><title>Page not found</title></head><body><h1>404 Not Found</h1></body></html>`)
	page.StatusCode = 200
	page.ContentType = "text/html"
	page.Title = "Page not found"
	page.H1s = []string{"404 Not Found"}
	page.BodyText = "Page not found"
	page.WordCount = 3

	results := (&responseSoft404{}).Run(page)

	assertCheck(t, results, "crawl.response.soft_404")
}

func TestSoft404IgnoresRegularSearchNoResultsPage(t *testing.T) {
	page := htmlPage(`<html><head><title>Search results</title></head><body><h1>Search results</h1><p>No results matched your filter.</p></body></html>`)
	page.StatusCode = 200
	page.ContentType = "text/html"
	page.Title = "Search results"
	page.H1s = []string{"Search results"}
	page.BodyText = "No results matched your filter."
	page.WordCount = 6

	if results := (&responseSoft404{}).Run(page); len(results) != 0 {
		t.Fatalf("expected no soft 404 issue, got %#v", results)
	}
}

func TestImageNon200Detected(t *testing.T) {
	page := htmlPage(`<html><body><img src="https://example.com/missing.jpg"></body></html>`)
	page.Images = []models.Image{{
		Src:         "https://example.com/missing.jpg",
		StatusCode:  404,
		ContentType: "text/html",
	}}

	results := (&responseImageNon200{}).Run(page)

	assertCheck(t, results, "crawl.response.image_non_200")
}

func TestContentTypeMismatchDetected(t *testing.T) {
	page := &models.PageData{
		URL:         "https://example.com/photo.jpg",
		FinalURL:    "https://example.com/photo.jpg",
		StatusCode:  200,
		ContentType: "text/html; charset=utf-8",
	}

	results := (&responseContentTypeMismatch{}).Run(page)

	assertCheck(t, results, "crawl.response.content_type_mismatch")
}

func TestVaryHeaderInvalidForCompressedCacheableResponse(t *testing.T) {
	page := &models.PageData{
		URL:        "https://example.com/app.js",
		FinalURL:   "https://example.com/app.js",
		StatusCode: 200,
		Headers: map[string]string{
			"cache-control":    "public, max-age=3600",
			"content-encoding": "br",
			"vary":             "Origin",
		},
	}

	results := (&responseVaryHeaderInvalid{}).Run(page)

	assertCheck(t, results, "crawl.response.vary_header_invalid")
}

func TestVaryHeaderValidForCompressedCacheableResponse(t *testing.T) {
	page := &models.PageData{
		URL:        "https://example.com/app.js",
		FinalURL:   "https://example.com/app.js",
		StatusCode: 200,
		Headers: map[string]string{
			"cache-control":    "public, max-age=3600",
			"content-encoding": "gzip",
			"vary":             "Accept-Encoding",
		},
	}

	if results := (&responseVaryHeaderInvalid{}).Run(page); len(results) != 0 {
		t.Fatalf("expected no Vary issue, got %#v", results)
	}
}

func TestHTTPToHTTPSRedirectRequiresPermanentStatus(t *testing.T) {
	page := redirectPage("http://example.com/", "https://example.com/", 302)

	results := (&redirectHTTPToHTTPSPermanent{}).Run(page)

	assertCheck(t, results, "crawl.redirect.http_to_https_not_301")
}

func TestHTTPToHTTPSPermanentRedirectPasses(t *testing.T) {
	page := redirectPage("http://example.com/", "https://example.com/", 301)

	if results := (&redirectHTTPToHTTPSPermanent{}).Run(page); len(results) != 0 {
		t.Fatalf("expected no HTTP to HTTPS issue, got %#v", results)
	}
}

func TestWWWVariantRedirectRequiresPermanentStatus(t *testing.T) {
	page := redirectPage("https://www.example.com/cars", "https://example.com/cars", 302)

	results := (&redirectWWWVariantPermanent{}).Run(page)

	assertCheck(t, results, "crawl.redirect.www_variant_not_301")
}

func TestTrailingSlashRedirectRequiresPermanentStatus(t *testing.T) {
	page := redirectPage("https://example.com/cars", "https://example.com/cars/", 302)

	results := (&redirectTrailingSlashPermanent{}).Run(page)

	assertCheck(t, results, "crawl.redirect.trailing_slash_inconsistent")
}

func TestTrailingSlashVariantsReturning200AreFlagged(t *testing.T) {
	pages := []*models.PageData{
		{URL: "https://example.com/cars", FinalURL: "https://example.com/cars", StatusCode: 200},
		{URL: "https://example.com/cars/", FinalURL: "https://example.com/cars/", StatusCode: 200},
	}

	results := (&redirectTrailingSlashSiteConsistency{}).Run(pages)

	assertCheck(t, results, "crawl.redirect.trailing_slash_inconsistent")
}

func TestJavascriptRedirectDetected(t *testing.T) {
	page := htmlPage(`<html><head><script>window.location.replace('/new')</script></head><body></body></html>`)

	results := (&redirectJavascript{}).Run(page)

	assertCheck(t, results, "crawl.redirect.javascript")
}

func TestMetaRefreshRedirectDetected(t *testing.T) {
	page := htmlPage(`<html><head><meta http-equiv="refresh" content="0; url=/new"></head><body></body></html>`)

	results := (&redirectMetaRefresh{}).Run(page)

	assertCheck(t, results, "crawl.redirect.meta_refresh")
}

func TestRedirectDestinationNoindexDetected(t *testing.T) {
	page := redirectPage("https://example.com/old", "https://example.com/new", 301)
	page.StatusCode = 200
	page.RobotsTag = "noindex, follow"
	page.RobotsDirectives = []string{"noindex", "follow"}

	results := (&redirectDestinationIndexable{}).Run(page)

	assertCheck(t, results, "crawl.redirect.destination_not_indexable")
	if !strings.Contains(results[0].Details, "noindex") {
		t.Fatalf("expected noindex details, got %q", results[0].Details)
	}
}

func redirectPage(from string, to string, status int) *models.PageData {
	return &models.PageData{
		URL:        from,
		FinalURL:   to,
		StatusCode: 200,
		RedirectChain: []models.RedirectHop{{
			URL:        from,
			StatusCode: status,
		}},
	}
}

func htmlPage(raw string) *models.PageData {
	return &models.PageData{
		URL:         "https://example.com/",
		FinalURL:    "https://example.com/",
		ContentType: "text/html",
		RawHTML:     raw,
	}
}

func assertCheck(t *testing.T, results []models.CheckResult, id string) {
	t.Helper()
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d: %#v", len(results), results)
	}
	if results[0].ID != id {
		t.Fatalf("expected %q, got %q", id, results[0].ID)
	}
}
