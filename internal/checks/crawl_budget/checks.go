// Package crawl_budget implements crawl budget analysis checks.
// Crawl budget — the number of pages Googlebot crawls and indexes per day —
// is critical for large sites. Wasting it on low-value URLs means important
// pages get crawled less frequently. This is a top-3 technical SEO issue for
// sites with 10,000+ pages.
package crawl_budget

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"

	"github.com/cars24/seo-automation/internal/models"
)

// Common parameter names that generate duplicate content / waste crawl budget.
var duplicateContentParams = []string{
	"ref", "source", "utm_source", "utm_medium", "utm_campaign", "utm_content",
	"utm_term", "gclid", "fbclid", "msclkid", "affiliate", "partner",
	"sort", "order", "orderby", "sortby", "filter", "color", "size",
	"currency", "lang", "locale", "from", "origin", "tracking",
	"_ga", "mc_cid", "mc_eid",
}

// Patterns suggesting infinite crawl spaces or faceted navigation.
var infiniteCrawlPatterns = []*regexp.Regexp{
	regexp.MustCompile(`[?&](color|size|brand|price|rating|category|tag|type|style)=[^&]+`),
	regexp.MustCompile(`/tag/[^/]+/`),
	regexp.MustCompile(`/category/[^/]+/page/\d+`),
	regexp.MustCompile(`/search\?`),
	regexp.MustCompile(`/results\?`),
	regexp.MustCompile(`/filter\?`),
}

// PageChecks returns per-page crawl budget checks.
func PageChecks() []models.PageCheck {
	return []models.PageCheck{
		&trackingParamsInURL{},
		&facetedNavigationURL{},
		&lowValuePage{},
		&sitemapButNoIndex{},
		&internalSearchPage{},
	}
}

// SiteChecks returns site-wide crawl budget checks.
func SiteChecks() []models.SiteCheck {
	return []models.SiteCheck{
		&crawlBudgetWasteEstimate{},
	}
}

// trackingParamsInURL: UTM and tracking parameters in crawled URLs waste
// crawl budget and create duplicate content.
type trackingParamsInURL struct{}

func (c *trackingParamsInURL) Run(p *models.PageData) []models.CheckResult {
	parsed, err := url.Parse(p.URL)
	if err != nil {
		return nil
	}
	params := parsed.Query()
	for _, bad := range duplicateContentParams {
		if params.Get(bad) != "" {
			return []models.CheckResult{{
				ID:       "crawl_budget.tracking_params",
				Category: "Crawl Budget",
				Severity: models.SeverityWarning,
				Message:  fmt.Sprintf("URL contains tracking/duplicate-content parameter: %s", bad),
				URL:      p.URL,
				Details:  "Add to robots.txt Disallow or use canonical to consolidate. Block via Google Search Console parameter handling.",
			}}
		}
	}
	return nil
}

// facetedNavigationURL: faceted navigation (filter/sort combinations) creates
// thousands of near-duplicate URLs. Critical crawl budget issue for e-commerce.
type facetedNavigationURL struct{}

func (c *facetedNavigationURL) Run(p *models.PageData) []models.CheckResult {
	for _, re := range infiniteCrawlPatterns {
		if re.MatchString(p.URL) {
			return []models.CheckResult{{
				ID:       "crawl_budget.faceted_navigation",
				Category: "Crawl Budget",
				Severity: models.SeverityWarning,
				Message:  "URL appears to be faceted navigation (filter/sort combination)",
				URL:      p.URL,
				Details:  "Consider: noindex+follow, canonical to base category, or robots.txt Disallow for filter combinations",
			}}
		}
	}
	return nil
}

// lowValuePage: pages that are unlikely to rank and waste crawl budget.
// Definition: very thin content, noindex, or error pages that are still linked.
type lowValuePage struct{}

func (c *lowValuePage) Run(p *models.PageData) []models.CheckResult {
	// Print/PDF versions of pages
	urlLower := strings.ToLower(p.URL)
	if strings.Contains(urlLower, "/print/") ||
		strings.Contains(urlLower, "?print=") ||
		strings.Contains(urlLower, "format=pdf") ||
		strings.Contains(urlLower, "/pdf/") {
		return []models.CheckResult{{
			ID:       "crawl_budget.low_value_page",
			Category: "Crawl Budget",
			Severity: models.SeverityNotice,
			Message:  "Low-value page variant (print/PDF) consuming crawl budget",
			URL:      p.URL,
			Details:  "Add noindex or block via robots.txt",
		}}
	}
	// Tag/archive pages with no unique content beyond listing
	if strings.Contains(urlLower, "/tag/") || strings.Contains(urlLower, "/author/") ||
		strings.Contains(urlLower, "/date/") || strings.Contains(urlLower, "/archive/") {
		if p.WordCount < 200 {
			return []models.CheckResult{{
				ID:       "crawl_budget.low_value_archive",
				Category: "Crawl Budget",
				Severity: models.SeverityNotice,
				Message:  "Low-value archive/tag page with thin content",
				URL:      p.URL,
				Details:  "Consider noindex or add unique content to archive pages",
			}}
		}
	}
	return nil
}

// sitemapButNoIndex: pages in sitemap with noindex waste crawl budget and
// send conflicting signals to Googlebot.
type sitemapButNoIndex struct{}

func (c *sitemapButNoIndex) Run(p *models.PageData) []models.CheckResult {
	if p.InSitemap && strings.Contains(p.RobotsTag, "noindex") {
		return []models.CheckResult{{
			ID:       "crawl_budget.sitemap_noindex_conflict",
			Category: "Crawl Budget",
			Severity: models.SeverityError,
			Message:  "Page in sitemap has noindex directive — conflicting signals waste crawl budget",
			URL:      p.URL,
			Details:  "Remove from sitemap OR remove noindex. Googlebot must crawl to see noindex, wasting budget.",
		}}
	}
	return nil
}

// internalSearchPage: internal site search result pages should never be indexed.
type internalSearchPage struct{}

func (c *internalSearchPage) Run(p *models.PageData) []models.CheckResult {
	parsed, err := url.Parse(p.URL)
	if err != nil {
		return nil
	}
	path := strings.ToLower(parsed.Path)
	query := strings.ToLower(parsed.RawQuery)

	isSearchPage := strings.Contains(path, "/search") ||
		strings.Contains(path, "/results") ||
		strings.Contains(query, "q=") ||
		strings.Contains(query, "query=") ||
		strings.Contains(query, "search=") ||
		strings.Contains(query, "s=")

	if isSearchPage && !strings.Contains(p.RobotsTag, "noindex") {
		return []models.CheckResult{{
			ID:       "crawl_budget.search_page_indexable",
			Category: "Crawl Budget",
			Severity: models.SeverityWarning,
			Message:  "Internal search results page is indexable (should be noindex)",
			URL:      p.URL,
			Details:  "Search pages are low-value, near-duplicate content. Add noindex+noarchive.",
		}}
	}
	return nil
}

// Site-wide crawl budget waste estimate.
type crawlBudgetWasteEstimate struct{}

func (c *crawlBudgetWasteEstimate) Run(pages []*models.PageData) []models.CheckResult {
	waste := 0
	for _, p := range pages {
		if strings.Contains(p.RobotsTag, "noindex") {
			waste++
		}
		if p.StatusCode >= 400 {
			waste++
		}
	}
	total := len(pages)
	if total == 0 {
		return nil
	}
	wastePercent := float64(waste) / float64(total) * 100
	if wastePercent > 20 {
		return []models.CheckResult{{
			ID:       "crawl_budget.high_waste_ratio",
			Category: "Crawl Budget",
			Severity: models.SeverityError,
			Message: fmt.Sprintf("High crawl budget waste: %.0f%% of pages are noindex or error (%d/%d)",
				wastePercent, waste, total),
			Details: "Review and remove or noindex low-value pages. Use robots.txt to block entire sections.",
		}}
	} else if wastePercent > 10 {
		return []models.CheckResult{{
			ID:       "crawl_budget.moderate_waste_ratio",
			Category: "Crawl Budget",
			Severity: models.SeverityWarning,
			Message: fmt.Sprintf("Moderate crawl budget waste: %.0f%% of pages are noindex or error (%d/%d)",
				wastePercent, waste, total),
		}}
	}
	return nil
}
