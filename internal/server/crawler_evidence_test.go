package server

import (
	"testing"

	"github.com/cars24/seo-automation/internal/models"
)

func TestCrawlerEvidenceChecksCoverPackItems(t *testing.T) {
	checks := crawlerEvidenceChecks()
	if got, want := len(checks), 5; got != want {
		t.Fatalf("crawlerEvidenceChecks() len=%d, want %d", got, want)
	}

	seen := map[string]bool{}
	for _, check := range checks {
		seen[check.ID] = true
	}
	for _, id := range []string{"ROBOTS-010", "ROBOTS-020", "SITEMAP-022", "IMG-013", "INDEX-015"} {
		if !seen[id] {
			t.Fatalf("missing crawler evidence check %s", id)
		}
	}
}

func TestParameterCoverageFindsCoveredAndMissingParams(t *testing.T) {
	robots := "User-agent: *\nDisallow: /*?sort=\nDisallow: /*color=\n"
	covered, missing := parameterCoverage(robots, []string{"sort", "color", "city"})
	if len(covered) != 2 || covered[0] != "sort" || covered[1] != "color" {
		t.Fatalf("covered=%v, want sort/color", covered)
	}
	if len(missing) != 1 || missing[0] != "city" {
		t.Fatalf("missing=%v, want city", missing)
	}
}

func TestAnalyzeSitemapInventoryFailsMissingExpectedURLs(t *testing.T) {
	audit := &models.SiteAudit{
		SitemapURLs: []string{"https://example.com/cars/a", "https://example.com/cars/b"},
	}
	item := analyzeSitemapInventory(audit, []string{"https://example.com/cars/a", "https://example.com/cars/c"})
	if item.Status != "fail" {
		t.Fatalf("status=%s, want fail", item.Status)
	}
	if len(item.Evidence) != 1 || item.Evidence[0] != "https://example.com/cars/c" {
		t.Fatalf("evidence=%v, want missing c URL", item.Evidence)
	}
}

func TestAnalyzeImageCDNFlagsHostMismatch(t *testing.T) {
	audit := &models.SiteAudit{Pages: []*models.PageData{
		{
			Images: []models.Image{
				{Src: "https://cdn.example.com/a.webp", StatusCode: 200, ContentType: "image/webp"},
				{Src: "https://other.example.com/b.webp", StatusCode: 200, ContentType: "image/webp"},
			},
		},
	}}
	item := analyzeImageCDN(audit, []string{"cdn.example.com"})
	if item.Status != "fail" {
		t.Fatalf("status=%s, want fail", item.Status)
	}
}
