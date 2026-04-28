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
		URL:      "https://example.com/",
		FinalURL: "https://example.com/",
		RawHTML:  raw,
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
