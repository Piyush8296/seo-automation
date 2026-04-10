package sitemapcheck

import (
	"fmt"
	"strings"

	"github.com/cars24/seo-automation/internal/models"
)

// PageChecks returns per-page sitemap checks.
func PageChecks() []models.PageCheck {
	return []models.PageCheck{
		&sitemapURL4xx{},
		&sitemapURLRedirected{},
		&sitemapURLNoindex{},
		&sitemapURLBlocked{},
	}
}

// SiteChecks returns site-wide sitemap checks.
func SiteChecks() []models.SiteCheck {
	return []models.SiteCheck{
		&sitemapMissing{},
		&sitemapTooLarge{},
		&sitemapCoverageLow{},
	}
}

type sitemapURL4xx struct{}

func (c *sitemapURL4xx) Run(p *models.PageData) []models.CheckResult {
	if p.InSitemap && p.StatusCode >= 400 && p.StatusCode < 500 {
		return []models.CheckResult{{
			ID:       "sitemap.url_4xx",
			Category: "Sitemap",
			Severity: models.SeverityError,
			Message:  fmt.Sprintf("Sitemap URL returns %d", p.StatusCode),
			URL:      p.URL,
		}}
	}
	return nil
}

type sitemapURLRedirected struct{}

func (c *sitemapURLRedirected) Run(p *models.PageData) []models.CheckResult {
	if p.InSitemap && p.URL != p.FinalURL && len(p.RedirectChain) > 0 {
		return []models.CheckResult{{
			ID:       "sitemap.url_redirected",
			Category: "Sitemap",
			Severity: models.SeverityWarning,
			Message:  "Sitemap URL is redirected (should use final URL in sitemap)",
			URL:      p.URL,
			Details:  "→ " + p.FinalURL,
		}}
	}
	return nil
}

type sitemapURLNoindex struct{}

func (c *sitemapURLNoindex) Run(p *models.PageData) []models.CheckResult {
	if p.InSitemap && strings.Contains(p.RobotsTag, "noindex") {
		return []models.CheckResult{{
			ID:       "sitemap.url_noindex",
			Category: "Sitemap",
			Severity: models.SeverityWarning,
			Message:  "Sitemap URL has noindex directive",
			URL:      p.URL,
		}}
	}
	return nil
}

type sitemapURLBlocked struct{}

func (c *sitemapURLBlocked) Run(p *models.PageData) []models.CheckResult {
	if p.InSitemap && strings.Contains(p.Error, "robots") {
		return []models.CheckResult{{
			ID:       "sitemap.url_blocked",
			Category: "Sitemap",
			Severity: models.SeverityWarning,
			Message:  "Sitemap URL is blocked by robots.txt",
			URL:      p.URL,
		}}
	}
	return nil
}

// Site-wide checks

type sitemapMissing struct{}

func (c *sitemapMissing) Run(pages []*models.PageData) []models.CheckResult {
	// If no pages are marked InSitemap, we may not have found a sitemap
	for _, p := range pages {
		if p.InSitemap {
			return nil
		}
	}
	if len(pages) == 0 {
		return nil
	}
	return []models.CheckResult{{
		ID:       "sitemap.missing",
		Category: "Sitemap",
		Severity: models.SeverityError,
		Message:  "No XML sitemap found or no sitemap URLs matched crawled pages",
		URL:      pages[0].URL,
	}}
}

type sitemapTooLarge struct{}

func (c *sitemapTooLarge) Run(pages []*models.PageData) []models.CheckResult {
	inSitemap := 0
	for _, p := range pages {
		if p.InSitemap {
			inSitemap++
		}
	}
	if inSitemap > 50000 {
		return []models.CheckResult{{
			ID:       "sitemap.too_large",
			Category: "Sitemap",
			Severity: models.SeverityWarning,
			Message:  fmt.Sprintf("Sitemap contains %d URLs (max 50,000 per sitemap)", inSitemap),
		}}
	}
	return nil
}

type sitemapCoverageLow struct{}

func (c *sitemapCoverageLow) Run(pages []*models.PageData) []models.CheckResult {
	total := 0
	inSitemap := 0
	for _, p := range pages {
		if p.StatusCode >= 200 && p.StatusCode < 300 {
			total++
			if p.InSitemap {
				inSitemap++
			}
		}
	}
	if total == 0 {
		return nil
	}
	coverage := float64(inSitemap) / float64(total)
	if coverage < 0.8 && total > 5 {
		return []models.CheckResult{{
			ID:       "sitemap.coverage_low",
			Category: "Sitemap",
			Severity: models.SeverityWarning,
			Message:  fmt.Sprintf("Low sitemap coverage: %d/%d crawled pages in sitemap (%.0f%%)", inSitemap, total, coverage*100),
		}}
	}
	return nil
}
