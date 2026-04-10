package headings

import (
	"fmt"
	"strings"

	"github.com/cars24/seo-automation/internal/models"
)

// PageChecks returns per-page heading checks.
func PageChecks() []models.PageCheck {
	return []models.PageCheck{
		&h1Missing{},
		&h1Multiple{},
		&h1EmptyOrShort{},
		&h1TooLong{},
		&h2Missing{},
		&headingHierarchySkipped{},
	}
}

// SiteChecks returns site-wide heading checks.
func SiteChecks() []models.SiteCheck {
	return []models.SiteCheck{
		&h1Duplicate{},
	}
}

type h1Missing struct{}

func (c *h1Missing) Run(p *models.PageData) []models.CheckResult {
	if len(p.H1s) == 0 {
		return []models.CheckResult{{
			ID:       "headings.h1.missing",
			Category: "Headings",
			Severity: models.SeverityError,
			Message:  "Page has no H1 heading",
			URL:      p.URL,
		}}
	}
	return nil
}

type h1Multiple struct{}

func (c *h1Multiple) Run(p *models.PageData) []models.CheckResult {
	if len(p.H1s) > 1 {
		return []models.CheckResult{{
			ID:       "headings.h1.multiple",
			Category: "Headings",
			Severity: models.SeverityWarning,
			Message:  fmt.Sprintf("Page has multiple H1 headings (%d)", len(p.H1s)),
			URL:      p.URL,
		}}
	}
	return nil
}

type h1EmptyOrShort struct{}

func (c *h1EmptyOrShort) Run(p *models.PageData) []models.CheckResult {
	var results []models.CheckResult
	for _, h := range p.H1s {
		h = strings.TrimSpace(h)
		if h == "" {
			results = append(results, models.CheckResult{
				ID:       "headings.h1.empty",
				Category: "Headings",
				Severity: models.SeverityError,
				Message:  "H1 heading is empty",
				URL:      p.URL,
			})
		} else if len(h) < 5 {
			results = append(results, models.CheckResult{
				ID:       "headings.h1.too_short",
				Category: "Headings",
				Severity: models.SeverityWarning,
				Message:  fmt.Sprintf("H1 too short (%d chars)", len(h)),
				URL:      p.URL,
				Details:  h,
			})
		}
	}
	return results
}

type h1TooLong struct{}

func (c *h1TooLong) Run(p *models.PageData) []models.CheckResult {
	var results []models.CheckResult
	for _, h := range p.H1s {
		if len(strings.TrimSpace(h)) > 70 {
			results = append(results, models.CheckResult{
				ID:       "headings.h1.too_long",
				Category: "Headings",
				Severity: models.SeverityWarning,
				Message:  fmt.Sprintf("H1 too long (%d chars, max 70)", len(strings.TrimSpace(h))),
				URL:      p.URL,
				Details:  h,
			})
		}
	}
	return results
}

type h2Missing struct{}

func (c *h2Missing) Run(p *models.PageData) []models.CheckResult {
	if p.WordCount > 300 && len(p.H2s) == 0 {
		return []models.CheckResult{{
			ID:       "headings.h2.missing",
			Category: "Headings",
			Severity: models.SeverityNotice,
			Message:  "Page has substantial content but no H2 headings",
			URL:      p.URL,
		}}
	}
	return nil
}

type headingHierarchySkipped struct{}

func (c *headingHierarchySkipped) Run(p *models.PageData) []models.CheckResult {
	// Check for H3 without H2, or H2 without H1
	hasH1 := len(p.H1s) > 0
	hasH2 := len(p.H2s) > 0
	hasH3 := len(p.H3s) > 0

	if hasH3 && !hasH2 {
		return []models.CheckResult{{
			ID:       "headings.hierarchy.skipped_level",
			Category: "Headings",
			Severity: models.SeverityWarning,
			Message:  "Heading hierarchy skips H2 (H3 without H2)",
			URL:      p.URL,
		}}
	}
	if hasH2 && !hasH1 {
		return []models.CheckResult{{
			ID:       "headings.hierarchy.skipped_level",
			Category: "Headings",
			Severity: models.SeverityWarning,
			Message:  "Heading hierarchy skips H1 (H2 without H1)",
			URL:      p.URL,
		}}
	}
	_ = hasH1
	return nil
}

type h1Duplicate struct{}

func (c *h1Duplicate) Run(pages []*models.PageData) []models.CheckResult {
	counts := map[string][]string{}
	for _, p := range pages {
		if len(p.H1s) == 0 {
			continue
		}
		h := strings.ToLower(strings.TrimSpace(p.H1s[0]))
		if h == "" {
			continue
		}
		counts[h] = append(counts[h], p.URL)
	}
	var results []models.CheckResult
	for h, urls := range counts {
		if len(urls) > 1 {
			for _, u := range urls {
				results = append(results, models.CheckResult{
					ID:       "headings.h1.duplicate",
					Category: "Headings",
					Severity: models.SeverityWarning,
					Message:  fmt.Sprintf("Duplicate H1 shared by %d pages", len(urls)),
					URL:      u,
					Details:  h,
				})
			}
		}
	}
	return results
}
