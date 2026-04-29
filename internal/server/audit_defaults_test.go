package server

import "testing"

func TestNormalizeStartAuditRequestAppliesBackendDefaults(t *testing.T) {
	req, err := NormalizeStartAuditRequest(StartAuditRequest{URL: " https://example.com "})
	if err != nil {
		t.Fatalf("NormalizeStartAuditRequest returned error: %v", err)
	}

	if req.URL != "https://example.com" {
		t.Fatalf("expected trimmed URL, got %q", req.URL)
	}
	if req.MaxDepth != -1 {
		t.Fatalf("expected unlimited max depth, got %d", req.MaxDepth)
	}
	if req.MaxPages != 0 {
		t.Fatalf("expected unlimited max pages, got %d", req.MaxPages)
	}
	if req.Concurrency != defaultConcur {
		t.Fatalf("expected default concurrency %d, got %d", defaultConcur, req.Concurrency)
	}
	if req.SitemapMode != string(defaultSitemapMode) {
		t.Fatalf("expected default sitemap mode %q, got %q", defaultSitemapMode, req.SitemapMode)
	}
	if req.ValidateExternalLinks == nil || !*req.ValidateExternalLinks {
		t.Fatal("expected external link validation default to be enabled")
	}
	if req.DiscoverResources == nil || !*req.DiscoverResources {
		t.Fatal("expected resource discovery default to be enabled")
	}
	if req.EnableCrawlerEvidence == nil || !*req.EnableCrawlerEvidence {
		t.Fatal("expected crawler evidence default to be enabled")
	}
	if req.EnableRenderedSEO == nil || !*req.EnableRenderedSEO {
		t.Fatal("expected rendered SEO default to be enabled")
	}
	if req.RenderedSampleLimit != defaultRenderedSampleLimit {
		t.Fatalf("expected rendered sample limit %d, got %d", defaultRenderedSampleLimit, req.RenderedSampleLimit)
	}
}

func TestNormalizeStartAuditRequestAllowsExplicitDisabledBooleans(t *testing.T) {
	no := false
	req, err := NormalizeStartAuditRequest(StartAuditRequest{
		URL:                   "https://example.com",
		ValidateExternalLinks: &no,
		DiscoverResources:     &no,
		EnableCrawlerEvidence: &no,
		EnableRenderedSEO:     &no,
	})
	if err != nil {
		t.Fatalf("NormalizeStartAuditRequest returned error: %v", err)
	}
	if *req.ValidateExternalLinks || *req.DiscoverResources || *req.EnableCrawlerEvidence || *req.EnableRenderedSEO {
		t.Fatalf("expected explicit false booleans to be preserved: %+v", req)
	}
}

func TestNormalizeStartAuditRequestRejectsInvalidLimits(t *testing.T) {
	tests := []StartAuditRequest{
		{URL: "https://example.com", MaxPages: -1},
		{URL: "https://example.com", MaxDepth: -2},
		{URL: "https://example.com", Concurrency: maxAuditConcurrency + 1},
		{URL: "https://example.com", Platform: "tablet"},
		{URL: "https://example.com", SitemapMode: "everything"},
	}

	for _, tt := range tests {
		if _, err := NormalizeStartAuditRequest(tt); err == nil {
			t.Fatalf("expected error for request %+v", tt)
		}
	}
}
