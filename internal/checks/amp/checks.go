// Package amp checks for AMP (Accelerated Mobile Pages) SEO issues.
// AMP pages must have correct canonical relationships with their regular counterparts.
// Common issues: regular page missing amphtml link, AMP missing canonical back-link,
// AMP and regular page content mismatch.
package amp

import (
	"strings"

	"github.com/cars24/seo-automation/internal/models"
)

// isAMPPage returns true if the page is an AMP variant.
func isAMPPage(p *models.PageData) bool {
	// AMP pages typically have ?amp=1, /amp/ path, or amp in URL
	urlLower := strings.ToLower(p.URL)
	return strings.Contains(urlLower, "/amp/") ||
		strings.Contains(urlLower, "?amp=1") ||
		strings.Contains(urlLower, "&amp=1") ||
		strings.Contains(p.RawHTML, `<html amp`) ||
		strings.Contains(p.RawHTML, `<html ⚡`)
}

// hasAMPLink returns true if the page has <link rel="amphtml">.
func hasAMPLink(p *models.PageData) bool {
	return strings.Contains(strings.ToLower(p.RawHTML), `rel="amphtml"`) ||
		strings.Contains(strings.ToLower(p.RawHTML), `rel='amphtml'`)
}

// PageChecks returns per-page AMP checks.
func PageChecks() []models.PageCheck {
	return []models.PageCheck{
		&ampMissingCanonical{},
		&ampCanonicalPointsToSelf{},
		&regularPageMissingAMPLink{},
	}
}

// SiteChecks returns site-wide AMP checks.
func SiteChecks() []models.SiteCheck {
	return []models.SiteCheck{
		&ampCanonicalOrphan{},
	}
}

// ampMissingCanonical: AMP pages MUST have a canonical link back to the regular page.
type ampMissingCanonical struct{}

func (c *ampMissingCanonical) Run(p *models.PageData) []models.CheckResult {
	if !isAMPPage(p) {
		return nil
	}
	if p.Canonical == "" {
		return []models.CheckResult{{
			ID:       "amp.canonical.missing",
			Category: "AMP",
			Severity: models.SeverityError,
			Message:  "AMP page is missing a canonical link to the regular page",
			URL:      p.URL,
			Details:  "AMP pages must have <link rel=\"canonical\"> pointing to the regular HTML page",
		}}
	}
	return nil
}

// ampCanonicalPointsToSelf: AMP page canonical pointing to itself is a
// self-canonical AMP — acceptable only if no regular version exists, but
// usually indicates misconfiguration.
type ampCanonicalPointsToSelf struct{}

func (c *ampCanonicalPointsToSelf) Run(p *models.PageData) []models.CheckResult {
	if !isAMPPage(p) {
		return nil
	}
	if p.Canonical == "" {
		return nil
	}
	// Canonical pointing to itself (same AMP URL)
	if strings.TrimRight(p.Canonical, "/") == strings.TrimRight(p.FinalURL, "/") {
		// Self-canonical AMP — check if there's a regular version
		// Flag as notice: might be intentional (AMP-first) but often a mistake
		return []models.CheckResult{{
			ID:       "amp.canonical.self_reference",
			Category: "AMP",
			Severity: models.SeverityNotice,
			Message:  "AMP page canonical points to itself (self-canonical AMP)",
			URL:      p.URL,
			Details:  "If a regular HTML version exists, canonical should point to it",
		}}
	}
	// Check if canonical still points to an AMP URL
	canonicalLower := strings.ToLower(p.Canonical)
	if strings.Contains(canonicalLower, "/amp/") || strings.Contains(canonicalLower, "?amp=1") {
		return []models.CheckResult{{
			ID:       "amp.canonical.points_to_amp",
			Category: "AMP",
			Severity: models.SeverityError,
			Message:  "AMP page canonical points to another AMP URL (should point to regular page)",
			URL:      p.URL,
			Details:  p.Canonical,
		}}
	}
	return nil
}

// regularPageMissingAMPLink: regular pages with a corresponding AMP version
// should have <link rel="amphtml"> to help Googlebot discover the AMP page.
type regularPageMissingAMPLink struct{}

func (c *regularPageMissingAMPLink) Run(p *models.PageData) []models.CheckResult {
	if isAMPPage(p) {
		return nil
	}
	// Only flag if there are other AMP pages on the site (can't know per-page)
	// Instead: flag pages that look like article/product pages without amphtml
	// Heuristic: if RawHTML has article tag and no amphtml link
	// Too noisy without knowing if AMP versions exist — skip per-page, handle site-wide
	return nil
}

// ampCanonicalOrphan: site-wide check — AMP page's canonical target doesn't
// have a corresponding amphtml link back.
type ampCanonicalOrphan struct{}

func (c *ampCanonicalOrphan) Run(pages []*models.PageData) []models.CheckResult {
	// Build map of regular pages that have amphtml links
	amphlmlTargets := map[string]bool{}
	regularPages := map[string]*models.PageData{}

	for _, p := range pages {
		if isAMPPage(p) {
			continue
		}
		regularPages[strings.TrimRight(p.FinalURL, "/")] = p
		if hasAMPLink(p) {
			amphlmlTargets[strings.TrimRight(p.FinalURL, "/")] = true
		}
	}

	var results []models.CheckResult
	for _, p := range pages {
		if !isAMPPage(p) || p.Canonical == "" {
			continue
		}
		canonicalKey := strings.TrimRight(p.Canonical, "/")
		regularPage, exists := regularPages[canonicalKey]
		if !exists {
			continue
		}
		if !amphlmlTargets[canonicalKey] {
			results = append(results, models.CheckResult{
				ID:       "amp.regular.missing_amphtml",
				Category: "AMP",
				Severity: models.SeverityWarning,
				Message:  "Regular page is missing <link rel=\"amphtml\"> link to its AMP version",
				URL:      regularPage.URL,
				Details:  "AMP page: " + p.URL,
			})
		}
	}
	return results
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
