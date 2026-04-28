package server

import (
	"testing"

	"github.com/cars24/seo-automation/internal/models"
	"github.com/temoto/robotstxt"
)

func TestCrawlerEvidenceChecksCoverPackItems(t *testing.T) {
	checks := crawlerEvidenceChecks()
	if got, want := len(checks), 23; got != want {
		t.Fatalf("crawlerEvidenceChecks() len=%d, want %d", got, want)
	}

	seen := map[string]bool{}
	for _, check := range checks {
		seen[check.ID] = true
	}
	for _, id := range []string{
		"ROBOTS-001", "ROBOTS-002", "ROBOTS-003", "ROBOTS-004", "ROBOTS-005",
		"ROBOTS-006", "ROBOTS-007", "ROBOTS-008", "ROBOTS-009", "ROBOTS-010",
		"ROBOTS-011", "ROBOTS-012", "ROBOTS-013", "ROBOTS-014", "ROBOTS-015",
		"ROBOTS-016", "ROBOTS-017", "ROBOTS-018", "ROBOTS-019", "ROBOTS-020",
		"SITEMAP-022", "IMG-013", "INDEX-015",
	} {
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

func TestRobotsRuleConflictsDetectIdenticalAllowDisallow(t *testing.T) {
	raw := "User-agent: *\nAllow: /cars/\nDisallow: /cars/\n"
	conflicts := robotsRuleConflicts(raw)
	if len(conflicts) != 1 {
		t.Fatalf("conflicts=%v, want one conflict", conflicts)
	}
}

func TestAnalyzeRobotsResourceAccessFlagsBlockedAssets(t *testing.T) {
	robots := robotsEvidence{
		Exists:  true,
		ParseOK: true,
		Raw:     "User-agent: *\nDisallow: /assets/\n",
	}
	data, err := robotstxt.FromString(robots.Raw)
	if err != nil {
		t.Fatal(err)
	}
	robots.Data = data
	audit := &models.SiteAudit{Pages: []*models.PageData{
		{
			Resources: []models.Resource{
				{URL: "https://example.com/assets/app.js", Type: models.ResourceScript, IsInternal: true},
			},
		},
	}}
	item := analyzeRobotsResourceAccess(audit, robots)
	if item.Status != "fail" {
		t.Fatalf("status=%s, want fail", item.Status)
	}
}

func TestAnalyzeCrawlDelayWarnsWhenHigh(t *testing.T) {
	robots := robotsEvidence{
		Exists:  true,
		ParseOK: true,
		Raw:     "User-agent: *\nCrawl-delay: 15\n",
	}
	data, err := robotstxt.FromString(robots.Raw)
	if err != nil {
		t.Fatal(err)
	}
	robots.Data = data
	item := analyzeCrawlDelay(robots)
	if item.Status != "warning" {
		t.Fatalf("status=%s, want warning", item.Status)
	}
}
