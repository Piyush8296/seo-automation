package crawlability

import (
	"fmt"
	"strings"

	"github.com/cars24/seo-automation/internal/models"
)

// PageChecks returns per-page crawlability checks.
func PageChecks() []models.PageCheck {
	return []models.PageCheck{
		&response4xx{},
		&response5xx{},
		&responseTimeout{},
		&redirectChain{},
		&redirectLoop{},
		&redirect302Permanent{},
		&noindexHasInlinks{},
		&pageDepthTooDeep{},
	}
}

// SiteChecks returns site-wide crawlability checks.
func SiteChecks() []models.SiteCheck {
	return []models.SiteCheck{
		&robotsTxtMissing{},
		&robotsBlocksAll{},
		&robotsMissingSitemapDirective{},
		&noindexInSitemapSite{},
		&orphanExternalOnly{},
		&robotsPageBlockedButLinked{},
	}
}

// --- Per-page checks ---

type response4xx struct{}

func (r *response4xx) Run(p *models.PageData) []models.CheckResult {
	if p.StatusCode >= 400 && p.StatusCode < 500 {
		return []models.CheckResult{{
			ID:       "crawl.response.4xx",
			Category: "Crawlability",
			Severity: models.SeverityError,
			Message:  fmt.Sprintf("4xx response code (%d)", p.StatusCode),
			URL:      p.URL,
		}}
	}
	return nil
}

type response5xx struct{}

func (r *response5xx) Run(p *models.PageData) []models.CheckResult {
	if p.StatusCode >= 500 {
		return []models.CheckResult{{
			ID:       "crawl.response.5xx",
			Category: "Crawlability",
			Severity: models.SeverityError,
			Message:  fmt.Sprintf("5xx response code (%d)", p.StatusCode),
			URL:      p.URL,
		}}
	}
	return nil
}

type responseTimeout struct{}

func (r *responseTimeout) Run(p *models.PageData) []models.CheckResult {
	if strings.Contains(strings.ToLower(p.Error), "timeout") ||
		strings.Contains(strings.ToLower(p.Error), "deadline exceeded") {
		return []models.CheckResult{{
			ID:       "crawl.response.timeout",
			Category: "Crawlability",
			Severity: models.SeverityError,
			Message:  "Request timed out",
			URL:      p.URL,
			Details:  p.Error,
		}}
	}
	return nil
}

type redirectChain struct{}

func (r *redirectChain) Run(p *models.PageData) []models.CheckResult {
	if len(p.RedirectChain) > 3 {
		return []models.CheckResult{{
			ID:       "crawl.redirect.chain",
			Category: "Crawlability",
			Severity: models.SeverityWarning,
			Message:  fmt.Sprintf("Redirect chain too long (%d hops)", len(p.RedirectChain)),
			URL:      p.URL,
		}}
	}
	return nil
}

type redirectLoop struct{}

func (r *redirectLoop) Run(p *models.PageData) []models.CheckResult {
	if strings.Contains(strings.ToLower(p.Error), "redirect loop") ||
		strings.Contains(strings.ToLower(p.Error), "stopped after") {
		return []models.CheckResult{{
			ID:       "crawl.redirect.loop",
			Category: "Crawlability",
			Severity: models.SeverityError,
			Message:  "Redirect loop detected",
			URL:      p.URL,
		}}
	}
	// Check for repeated URL in chain
	seen := map[string]bool{}
	for _, hop := range p.RedirectChain {
		if seen[hop.URL] {
			return []models.CheckResult{{
				ID:       "crawl.redirect.loop",
				Category: "Crawlability",
				Severity: models.SeverityError,
				Message:  "Redirect loop detected",
				URL:      p.URL,
			}}
		}
		seen[hop.URL] = true
	}
	return nil
}

type redirect302Permanent struct{}

func (r *redirect302Permanent) Run(p *models.PageData) []models.CheckResult {
	// If there is a redirect and the page URL differs from final URL, and status was 302
	if p.StatusCode == 302 && p.URL != p.FinalURL {
		return []models.CheckResult{{
			ID:       "crawl.redirect.302_permanent",
			Category: "Crawlability",
			Severity: models.SeverityWarning,
			Message:  "Using 302 (temporary) redirect — consider 301 if permanent",
			URL:      p.URL,
			Details:  fmt.Sprintf("Redirects to: %s", p.FinalURL),
		}}
	}
	return nil
}

type noindexHasInlinks struct{}

func (n *noindexHasInlinks) Run(p *models.PageData) []models.CheckResult {
	// Per-page: flag if noindex is set (site-wide check will verify inlinks)
	if strings.Contains(p.RobotsTag, "noindex") && p.InSitemap {
		return []models.CheckResult{{
			ID:       "crawl.noindex.in_sitemap",
			Category: "Crawlability",
			Severity: models.SeverityWarning,
			Message:  "Noindex page is listed in sitemap",
			URL:      p.URL,
		}}
	}
	return nil
}

type pageDepthTooDeep struct{}

func (pd *pageDepthTooDeep) Run(p *models.PageData) []models.CheckResult {
	if p.Depth > 3 {
		return []models.CheckResult{{
			ID:       "crawl.page_depth.too_deep",
			Category: "Crawlability",
			Severity: models.SeverityWarning,
			Message:  fmt.Sprintf("Page is too deep in site structure (depth: %d)", p.Depth),
			URL:      p.URL,
		}}
	}
	return nil
}

// --- Site-wide checks ---

type robotsTxtMissing struct{}

func (r *robotsTxtMissing) Run(pages []*models.PageData) []models.CheckResult {
	if len(pages) == 0 {
		return nil
	}
	// Check the audit-level flag passed via first page's parent audit
	// We detect this via absence of robots data — use a heuristic: if the site
	// served pages fine but we have no robots info, we can't tell from pages alone.
	// This will be set by the crawler in audit.RobotsTxtMissing; for now skip.
	return nil
}

type robotsBlocksAll struct{}

func (r *robotsBlocksAll) Run(pages []*models.PageData) []models.CheckResult {
	return nil
}

type robotsMissingSitemapDirective struct{}

func (r *robotsMissingSitemapDirective) Run(pages []*models.PageData) []models.CheckResult {
	return nil
}

type noindexInSitemapSite struct{}

func (n *noindexInSitemapSite) Run(pages []*models.PageData) []models.CheckResult {
	var results []models.CheckResult
	for _, p := range pages {
		if strings.Contains(p.RobotsTag, "noindex") && p.InSitemap {
			results = append(results, models.CheckResult{
				ID:       "crawl.noindex.in_sitemap",
				Category: "Crawlability",
				Severity: models.SeverityWarning,
				Message:  "Noindex page listed in XML sitemap",
				URL:      p.URL,
			})
		}
	}
	return results
}

type orphanExternalOnly struct{}

func (o *orphanExternalOnly) Run(pages []*models.PageData) []models.CheckResult {
	// Build inlink map
	inlinks := map[string]int{}
	for _, p := range pages {
		for _, link := range p.Links {
			if link.IsInternal {
				inlinks[link.URL]++
			}
		}
	}
	var results []models.CheckResult
	for _, p := range pages {
		if p.Depth == 0 {
			continue // seed page is never orphan
		}
		if inlinks[p.URL] == 0 && inlinks[p.FinalURL] == 0 {
			results = append(results, models.CheckResult{
				ID:       "crawl.page_depth.orphan_external_only",
				Category: "Crawlability",
				Severity: models.SeverityWarning,
				Message:  "Page has no internal inlinks (orphan page)",
				URL:      p.URL,
			})
		}
	}
	return results
}

type robotsPageBlockedButLinked struct{}

func (r *robotsPageBlockedButLinked) Run(pages []*models.PageData) []models.CheckResult {
	blocked := map[string]bool{}
	for _, p := range pages {
		if strings.Contains(p.Error, "robots") {
			blocked[p.URL] = true
		}
	}
	if len(blocked) == 0 {
		return nil
	}
	linked := map[string]bool{}
	for _, p := range pages {
		for _, link := range p.Links {
			if link.IsInternal {
				linked[link.URL] = true
			}
		}
	}
	var results []models.CheckResult
	for url := range blocked {
		if linked[url] {
			results = append(results, models.CheckResult{
				ID:       "crawl.robots.page_blocked_but_linked",
				Category: "Crawlability",
				Severity: models.SeverityWarning,
				Message:  "Page is blocked by robots.txt but has internal inlinks",
				URL:      url,
			})
		}
	}
	return results
}
