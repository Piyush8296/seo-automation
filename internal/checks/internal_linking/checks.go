package internal_linking

import (
	"fmt"
	"strings"

	"github.com/cars24/seo-automation/internal/models"
)

var genericAnchors = map[string]bool{
	"click here": true, "here": true, "read more": true, "more": true,
	"learn more": true, "link": true, "this": true, "go": true,
	"page": true, "website": true, "click": true,
}

// PageChecks returns per-page internal linking checks.
func PageChecks() []models.PageCheck {
	return []models.PageCheck{
		&brokenInternalLinks{},
		&internalLinkNofollow{},
		&externalLinkMissingNoopener{},
		&emptyAnchor{},
		&genericAnchor{},
		&tooManyOutlinks{},
		&externalBroken4xx{},
		&externalBroken5xx{},
		&externalTimeout{},
		&externalRedirect{},
		&footerHeavy{},
		&noContentLinks{},
	}
}

// SiteChecks returns site-wide internal linking checks.
func SiteChecks() []models.SiteCheck {
	return []models.SiteCheck{
		&orphanPage{},
		&lowInlinks{},
		&navOrphan{},
	}
}

type brokenInternalLinks struct{}

func (c *brokenInternalLinks) Run(p *models.PageData) []models.CheckResult {
	var results []models.CheckResult
	for _, link := range p.Links {
		if !link.IsInternal || link.StatusCode == 0 {
			continue
		}
		if link.StatusCode >= 400 && link.StatusCode < 500 {
			results = append(results, models.CheckResult{
				ID:       "links.internal.broken_4xx",
				Category: "Internal Linking",
				Severity: models.SeverityError,
				Message:  fmt.Sprintf("Internal link returns %d", link.StatusCode),
				URL:      p.URL,
				Details:  link.URL,
			})
		} else if link.StatusCode >= 500 {
			results = append(results, models.CheckResult{
				ID:       "links.internal.broken_5xx",
				Category: "Internal Linking",
				Severity: models.SeverityError,
				Message:  fmt.Sprintf("Internal link returns %d", link.StatusCode),
				URL:      p.URL,
				Details:  link.URL,
			})
		} else if link.StatusCode == 301 || link.StatusCode == 302 {
			results = append(results, models.CheckResult{
				ID:       "links.internal.to_redirect",
				Category: "Internal Linking",
				Severity: models.SeverityWarning,
				Message:  fmt.Sprintf("Internal link redirects (%d)", link.StatusCode),
				URL:      p.URL,
				Details:  link.URL,
			})
		}
	}
	return results
}

type internalLinkNofollow struct{}

func (c *internalLinkNofollow) Run(p *models.PageData) []models.CheckResult {
	var results []models.CheckResult
	for _, link := range p.Links {
		if link.IsInternal && strings.Contains(link.Rel, "nofollow") {
			results = append(results, models.CheckResult{
				ID:       "links.internal.nofollow",
				Category: "Internal Linking",
				Severity: models.SeverityNotice,
				Message:  "Internal link has rel=nofollow (blocks PageRank)",
				URL:      p.URL,
				Details:  link.URL,
			})
		}
	}
	return results
}

type externalLinkMissingNoopener struct{}

func (c *externalLinkMissingNoopener) Run(p *models.PageData) []models.CheckResult {
	var results []models.CheckResult
	for _, link := range p.Links {
		if !link.IsInternal && !strings.Contains(link.Rel, "noopener") {
			results = append(results, models.CheckResult{
				ID:       "links.external.missing_noopener",
				Category: "Internal Linking",
				Severity: models.SeverityWarning,
				Message:  "External link missing rel=noopener (security risk)",
				URL:      p.URL,
				Details:  link.URL,
			})
		}
	}
	return results
}

type emptyAnchor struct{}

func (c *emptyAnchor) Run(p *models.PageData) []models.CheckResult {
	var results []models.CheckResult
	for _, link := range p.Links {
		if strings.TrimSpace(link.Text) == "" {
			results = append(results, models.CheckResult{
				ID:       "links.anchor.empty",
				Category: "Internal Linking",
				Severity: models.SeverityWarning,
				Message:  "Link has empty anchor text",
				URL:      p.URL,
				Details:  link.URL,
			})
		}
	}
	return results
}

type genericAnchor struct{}

func (c *genericAnchor) Run(p *models.PageData) []models.CheckResult {
	var results []models.CheckResult
	for _, link := range p.Links {
		text := strings.ToLower(strings.TrimSpace(link.Text))
		if genericAnchors[text] {
			results = append(results, models.CheckResult{
				ID:       "links.anchor.generic",
				Category: "Internal Linking",
				Severity: models.SeverityWarning,
				Message:  fmt.Sprintf("Generic anchor text: %q", link.Text),
				URL:      p.URL,
				Details:  link.URL,
			})
		}
	}
	return results
}

type tooManyOutlinks struct{}

func (c *tooManyOutlinks) Run(p *models.PageData) []models.CheckResult {
	internal := 0
	for _, link := range p.Links {
		if link.IsInternal {
			internal++
		}
	}
	if internal > 100 {
		return []models.CheckResult{{
			ID:       "links.page.too_many_outlinks",
			Category: "Internal Linking",
			Severity: models.SeverityWarning,
			Message:  fmt.Sprintf("Page has too many internal outlinks (%d)", internal),
			URL:      p.URL,
		}}
	}
	return nil
}

// Site-wide checks

// buildInlinkMap counts how many internal links point to each URL and populates
// the InlinkCount field on every PageData.
func buildInlinkMap(pages []*models.PageData) map[string]int {
	inlinks := map[string]int{}
	for _, p := range pages {
		for _, link := range p.Links {
			if link.IsInternal {
				inlinks[link.URL]++
			}
		}
	}
	// Populate InlinkCount on each page for report output
	for _, p := range pages {
		count := inlinks[p.URL]
		if p.FinalURL != "" && p.FinalURL != p.URL {
			count += inlinks[p.FinalURL]
		}
		p.InlinkCount = count
	}
	return inlinks
}

type orphanPage struct{}

func (o *orphanPage) Run(pages []*models.PageData) []models.CheckResult {
	inlinks := buildInlinkMap(pages)
	var results []models.CheckResult
	for _, p := range pages {
		if p.Depth == 0 {
			continue
		}
		if inlinks[p.URL] == 0 && inlinks[p.FinalURL] == 0 {
			results = append(results, models.CheckResult{
				ID:       "links.page.orphan",
				Category: "Internal Linking",
				Severity: models.SeverityWarning,
				Message:  "Page has no internal inlinks (orphan)",
				URL:      p.URL,
			})
		}
	}
	return results
}

// External link checks — only produce results when --validate-external-links is enabled
// (i.e. when StatusCode > 0 or Timeout is set).

type externalBroken4xx struct{}

func (c *externalBroken4xx) Run(p *models.PageData) []models.CheckResult {
	var results []models.CheckResult
	for _, link := range p.Links {
		if link.IsInternal || (link.StatusCode == 0 && !link.Timeout) {
			continue
		}
		if link.StatusCode >= 400 && link.StatusCode < 500 {
			results = append(results, models.CheckResult{
				ID:       "links.external.broken_4xx",
				Category: "Internal Linking",
				Severity: models.SeverityError,
				Message:  fmt.Sprintf("External link returns %d", link.StatusCode),
				URL:      p.URL,
				Details:  link.URL,
			})
		}
	}
	return results
}

type externalBroken5xx struct{}

func (c *externalBroken5xx) Run(p *models.PageData) []models.CheckResult {
	var results []models.CheckResult
	for _, link := range p.Links {
		if link.IsInternal || (link.StatusCode == 0 && !link.Timeout) {
			continue
		}
		if link.StatusCode >= 500 {
			results = append(results, models.CheckResult{
				ID:       "links.external.broken_5xx",
				Category: "Internal Linking",
				Severity: models.SeverityError,
				Message:  fmt.Sprintf("External link returns %d", link.StatusCode),
				URL:      p.URL,
				Details:  link.URL,
			})
		}
	}
	return results
}

type externalTimeout struct{}

func (c *externalTimeout) Run(p *models.PageData) []models.CheckResult {
	var results []models.CheckResult
	for _, link := range p.Links {
		if link.IsInternal || !link.Timeout {
			continue
		}
		results = append(results, models.CheckResult{
			ID:       "links.external.timeout",
			Category: "Internal Linking",
			Severity: models.SeverityWarning,
			Message:  "External link timed out",
			URL:      p.URL,
			Details:  link.URL,
		})
	}
	return results
}

type externalRedirect struct{}

func (c *externalRedirect) Run(p *models.PageData) []models.CheckResult {
	var results []models.CheckResult
	for _, link := range p.Links {
		if link.IsInternal || (link.StatusCode == 0 && !link.Timeout) {
			continue
		}
		if link.StatusCode >= 300 && link.StatusCode < 400 {
			results = append(results, models.CheckResult{
				ID:       "links.external.redirect",
				Category: "Internal Linking",
				Severity: models.SeverityNotice,
				Message:  fmt.Sprintf("External link redirects (%d)", link.StatusCode),
				URL:      p.URL,
				Details:  link.URL,
			})
		}
	}
	return results
}

type footerHeavy struct{}

func (c *footerHeavy) Run(p *models.PageData) []models.CheckResult {
	total := 0
	footer := 0
	for _, l := range p.Links {
		if !l.IsInternal {
			continue
		}
		total++
		if l.Position == models.PositionFooter {
			footer++
		}
	}
	if total < 10 {
		return nil
	}
	pct := float64(footer) / float64(total)
	if pct > 0.70 {
		return []models.CheckResult{{
			ID:       "links.footer_heavy",
			Category: "Internal Linking",
			Severity: models.SeverityWarning,
			Message:  fmt.Sprintf("%.0f%% of internal links are in the footer (%d of %d)", pct*100, footer, total),
			URL:      p.URL,
		}}
	}
	return nil
}

type noContentLinks struct{}

func (c *noContentLinks) Run(p *models.PageData) []models.CheckResult {
	total := 0
	content := 0
	for _, l := range p.Links {
		if !l.IsInternal {
			continue
		}
		total++
		if l.Position == models.PositionContent {
			content++
		}
	}
	if total < 5 {
		return nil
	}
	if content == 0 {
		return []models.CheckResult{{
			ID:       "links.no_content_links",
			Category: "Internal Linking",
			Severity: models.SeverityWarning,
			Message:  fmt.Sprintf("Page has no internal links in main content (%d links, all in nav/header/footer/sidebar)", total),
			URL:      p.URL,
		}}
	}
	return nil
}

type navOrphan struct{}

func (c *navOrphan) Run(pages []*models.PageData) []models.CheckResult {
	// Skip entirely if the site uses no nav-positioned links (no <nav>/role=navigation anywhere).
	hasAnyNav := false
	for _, p := range pages {
		for _, l := range p.Links {
			if l.IsInternal && l.Position == models.PositionNav {
				hasAnyNav = true
				break
			}
		}
		if hasAnyNav {
			break
		}
	}
	if !hasAnyNav {
		return nil
	}

	navInlinks := map[string]int{}
	anyInlinks := map[string]int{}
	for _, p := range pages {
		for _, l := range p.Links {
			if !l.IsInternal {
				continue
			}
			anyInlinks[l.URL]++
			if l.Position == models.PositionNav {
				navInlinks[l.URL]++
			}
		}
	}

	var results []models.CheckResult
	for _, p := range pages {
		if p.Depth == 0 {
			continue
		}
		total := anyInlinks[p.URL] + anyInlinks[p.FinalURL]
		if total < 3 {
			continue // too few inlinks to be meaningful; orphan/low_inlinks cover these
		}
		if navInlinks[p.URL]+navInlinks[p.FinalURL] > 0 {
			continue
		}
		results = append(results, models.CheckResult{
			ID:       "links.nav_orphan",
			Category: "Internal Linking",
			Severity: models.SeverityNotice,
			Message:  fmt.Sprintf("Page has %d internal inlinks but none from site navigation", total),
			URL:      p.URL,
		})
	}
	return results
}

type lowInlinks struct{}

func (l *lowInlinks) Run(pages []*models.PageData) []models.CheckResult {
	inlinks := buildInlinkMap(pages)
	var results []models.CheckResult
	for _, p := range pages {
		if p.Depth == 0 {
			continue // skip seed URL
		}
		count := inlinks[p.URL]
		if p.FinalURL != "" && p.FinalURL != p.URL {
			count += inlinks[p.FinalURL]
		}
		if count > 0 && count < 3 {
			results = append(results, models.CheckResult{
				ID:       "links.page.low_inlinks",
				Category: "Internal Linking",
				Severity: models.SeverityWarning,
				Message:  fmt.Sprintf("Page has very few internal inlinks (%d, recommend 3+)", count),
				URL:      p.URL,
			})
		}
	}
	return results
}
