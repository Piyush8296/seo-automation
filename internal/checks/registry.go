package checks

import (
	"github.com/cars24/seo-automation/internal/checks/amp"
	"github.com/cars24/seo-automation/internal/checks/canonical"
	"github.com/cars24/seo-automation/internal/checks/content_body"
	"github.com/cars24/seo-automation/internal/checks/content_meta_desc"
	"github.com/cars24/seo-automation/internal/checks/content_title"
	"github.com/cars24/seo-automation/internal/checks/core_web_vitals"
	"github.com/cars24/seo-automation/internal/checks/crawl_budget"
	"github.com/cars24/seo-automation/internal/checks/crawlability"
	"github.com/cars24/seo-automation/internal/checks/eeat"
	"github.com/cars24/seo-automation/internal/checks/headings"
	"github.com/cars24/seo-automation/internal/checks/html_hygiene"
	"github.com/cars24/seo-automation/internal/checks/https_security"
	"github.com/cars24/seo-automation/internal/checks/images"
	"github.com/cars24/seo-automation/internal/checks/internal_linking"
	"github.com/cars24/seo-automation/internal/checks/international"
	"github.com/cars24/seo-automation/internal/checks/mobile"
	"github.com/cars24/seo-automation/internal/checks/mobile_desktop"
	"github.com/cars24/seo-automation/internal/checks/pagination"
	"github.com/cars24/seo-automation/internal/checks/performance"
	"github.com/cars24/seo-automation/internal/checks/resources"
	"github.com/cars24/seo-automation/internal/checks/sitemapcheck"
	"github.com/cars24/seo-automation/internal/checks/social"
	"github.com/cars24/seo-automation/internal/checks/ssl"
	"github.com/cars24/seo-automation/internal/checks/structured_data"
	"github.com/cars24/seo-automation/internal/checks/url_structure"
	"github.com/cars24/seo-automation/internal/models"
)

var (
	pageChecks []models.PageCheck
	siteChecks []models.SiteCheck
)

func init() {
	// ── Per-page checks (original 146) ──────────────────────────────────────
	for _, c := range crawlability.PageChecks() {
		pageChecks = append(pageChecks, c)
	}
	for _, c := range https_security.PageChecks() {
		pageChecks = append(pageChecks, c)
	}
	for _, c := range performance.PageChecks() {
		pageChecks = append(pageChecks, c)
	}
	for _, c := range internal_linking.PageChecks() {
		pageChecks = append(pageChecks, c)
	}
	for _, c := range content_title.PageChecks() {
		pageChecks = append(pageChecks, c)
	}
	for _, c := range content_meta_desc.PageChecks() {
		pageChecks = append(pageChecks, c)
	}
	for _, c := range content_body.PageChecks() {
		pageChecks = append(pageChecks, c)
	}
	for _, c := range headings.PageChecks() {
		pageChecks = append(pageChecks, c)
	}
	for _, c := range html_hygiene.PageChecks() {
		pageChecks = append(pageChecks, c)
	}
	for _, c := range canonical.PageChecks() {
		pageChecks = append(pageChecks, c)
	}
	for _, c := range images.PageChecks() {
		pageChecks = append(pageChecks, c)
	}
	for _, c := range structured_data.PageChecks() {
		pageChecks = append(pageChecks, c)
	}
	for _, c := range social.PageChecks() {
		pageChecks = append(pageChecks, c)
	}
	for _, c := range url_structure.PageChecks() {
		pageChecks = append(pageChecks, c)
	}
	for _, c := range mobile.PageChecks() {
		pageChecks = append(pageChecks, c)
	}
	for _, c := range mobile_desktop.PageChecks() {
		pageChecks = append(pageChecks, c)
	}
	for _, c := range international.PageChecks() {
		pageChecks = append(pageChecks, c)
	}
	for _, c := range sitemapcheck.PageChecks() {
		pageChecks = append(pageChecks, c)
	}

	// ── Per-page checks (expert additions) ──────────────────────────────────
	for _, c := range pagination.PageChecks() {
		pageChecks = append(pageChecks, c)
	}
	for _, c := range amp.PageChecks() {
		pageChecks = append(pageChecks, c)
	}
	for _, c := range crawl_budget.PageChecks() {
		pageChecks = append(pageChecks, c)
	}
	for _, c := range core_web_vitals.PageChecks() {
		pageChecks = append(pageChecks, c)
	}
	for _, c := range eeat.PageChecks() {
		pageChecks = append(pageChecks, c)
	}
	for _, c := range ssl.PageChecks() {
		pageChecks = append(pageChecks, c)
	}
	for _, c := range resources.PageChecks() {
		pageChecks = append(pageChecks, c)
	}

	// ── Site-wide checks (original) ──────────────────────────────────────────
	for _, c := range crawlability.SiteChecks() {
		siteChecks = append(siteChecks, c)
	}
	for _, c := range internal_linking.SiteChecks() {
		siteChecks = append(siteChecks, c)
	}
	for _, c := range content_title.SiteChecks() {
		siteChecks = append(siteChecks, c)
	}
	for _, c := range content_meta_desc.SiteChecks() {
		siteChecks = append(siteChecks, c)
	}
	for _, c := range content_body.SiteChecks() {
		siteChecks = append(siteChecks, c)
	}
	for _, c := range headings.SiteChecks() {
		siteChecks = append(siteChecks, c)
	}
	for _, c := range structured_data.SiteChecks() {
		siteChecks = append(siteChecks, c)
	}
	for _, c := range international.SiteChecks() {
		siteChecks = append(siteChecks, c)
	}
	for _, c := range sitemapcheck.SiteChecks() {
		siteChecks = append(siteChecks, c)
	}
	for _, c := range canonical.SiteChecks() {
		siteChecks = append(siteChecks, c)
	}

	// ── Site-wide checks (expert additions) ─────────────────────────────────
	for _, c := range pagination.SiteChecks() {
		siteChecks = append(siteChecks, c)
	}
	for _, c := range amp.SiteChecks() {
		siteChecks = append(siteChecks, c)
	}
	for _, c := range crawl_budget.SiteChecks() {
		siteChecks = append(siteChecks, c)
	}
	for _, c := range eeat.SiteChecks() {
		siteChecks = append(siteChecks, c)
	}
}

// Catalog describes the registered check surface for introspection via API.
type Catalog struct {
	Total      int `json:"total"`
	PageChecks int `json:"page_checks"`
	SiteChecks int `json:"site_checks"`
	CheckIDs   int `json:"check_ids"`
}

// GetCatalog returns the count of registered check runners.
func GetCatalog() Catalog {
	return Catalog{
		Total:      len(pageChecks) + len(siteChecks),
		PageChecks: len(pageChecks),
		SiteChecks: len(siteChecks),
		CheckIDs:   len(checkDescriptors),
	}
}

// RunPageChecks runs all per-page checks and returns combined results.
func RunPageChecks(page *models.PageData) []models.CheckResult {
	var results []models.CheckResult
	for _, c := range pageChecks {
		results = append(results, c.Run(page)...)
	}
	return results
}

// RunSiteWideChecks runs all site-wide checks across all pages.
func RunSiteWideChecks(pages []*models.PageData) []models.CheckResult {
	var results []models.CheckResult
	for _, c := range siteChecks {
		results = append(results, c.Run(pages)...)
	}
	return results
}
