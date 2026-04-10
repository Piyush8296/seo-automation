package content_title

import (
	"fmt"
	"strings"

	"github.com/cars24/seo-automation/internal/models"
)

// PageChecks returns per-page title checks.
func PageChecks() []models.PageCheck {
	return []models.PageCheck{
		&titleMissingOrEmpty{},
		&titleTooShort{},
		&titleTooLong{},
	}
}

// SiteChecks returns site-wide title checks.
func SiteChecks() []models.SiteCheck {
	return []models.SiteCheck{
		&titleDuplicate{},
	}
}

type titleMissingOrEmpty struct{}

func (c *titleMissingOrEmpty) Run(p *models.PageData) []models.CheckResult {
	if strings.TrimSpace(p.Title) == "" {
		return []models.CheckResult{{
			ID:       "title.missing",
			Category: "Titles",
			Severity: models.SeverityError,
			Message:  "Page title is missing or empty",
			URL:      p.URL,
		}}
	}
	return nil
}

type titleTooShort struct{}

func (c *titleTooShort) Run(p *models.PageData) []models.CheckResult {
	if t := strings.TrimSpace(p.Title); len(t) > 0 && len(t) < 10 {
		return []models.CheckResult{{
			ID:       "title.too_short",
			Category: "Titles",
			Severity: models.SeverityWarning,
			Message:  fmt.Sprintf("Title too short (%d chars, min 10)", len(t)),
			URL:      p.URL,
			Details:  t,
		}}
	}
	return nil
}

type titleTooLong struct{}

func (c *titleTooLong) Run(p *models.PageData) []models.CheckResult {
	if t := strings.TrimSpace(p.Title); len(t) > 60 {
		return []models.CheckResult{{
			ID:       "title.too_long",
			Category: "Titles",
			Severity: models.SeverityWarning,
			Message:  fmt.Sprintf("Title too long (%d chars, max 60)", len(t)),
			URL:      p.URL,
			Details:  t,
		}}
	}
	return nil
}

type titleDuplicate struct{}

func (c *titleDuplicate) Run(pages []*models.PageData) []models.CheckResult {
	counts := map[string][]string{}
	for _, p := range pages {
		t := strings.ToLower(strings.TrimSpace(p.Title))
		if t == "" {
			continue
		}
		counts[t] = append(counts[t], p.URL)
	}
	var results []models.CheckResult
	for title, urls := range counts {
		if len(urls) > 1 {
			for _, u := range urls {
				results = append(results, models.CheckResult{
					ID:       "title.duplicate",
					Category: "Titles",
					Severity: models.SeverityWarning,
					Message:  fmt.Sprintf("Duplicate title shared by %d pages", len(urls)),
					URL:      u,
					Details:  title,
				})
			}
		}
	}
	return results
}
