package content_meta_desc

import (
	"fmt"
	"strings"

	"github.com/cars24/seo-automation/internal/models"
)

// PageChecks returns per-page meta description checks.
func PageChecks() []models.PageCheck {
	return []models.PageCheck{
		&metaDescMissingOrEmpty{},
		&metaDescTooShort{},
		&metaDescTooLong{},
	}
}

// SiteChecks returns site-wide meta description checks.
func SiteChecks() []models.SiteCheck {
	return []models.SiteCheck{
		&metaDescDuplicate{},
	}
}

type metaDescMissingOrEmpty struct{}

func (c *metaDescMissingOrEmpty) Run(p *models.PageData) []models.CheckResult {
	if strings.TrimSpace(p.MetaDesc) == "" {
		return []models.CheckResult{{
			ID:       "meta_desc.missing",
			Category: "Meta Descriptions",
			Severity: models.SeverityWarning,
			Message:  "Meta description is missing or empty",
			URL:      p.URL,
		}}
	}
	return nil
}

type metaDescTooShort struct{}

func (c *metaDescTooShort) Run(p *models.PageData) []models.CheckResult {
	d := strings.TrimSpace(p.MetaDesc)
	if len(d) > 0 && len(d) < 50 {
		return []models.CheckResult{{
			ID:       "meta_desc.too_short",
			Category: "Meta Descriptions",
			Severity: models.SeverityNotice,
			Message:  fmt.Sprintf("Meta description too short (%d chars, min 50)", len(d)),
			URL:      p.URL,
			Details:  d,
		}}
	}
	return nil
}

type metaDescTooLong struct{}

func (c *metaDescTooLong) Run(p *models.PageData) []models.CheckResult {
	d := strings.TrimSpace(p.MetaDesc)
	if len(d) > 160 {
		return []models.CheckResult{{
			ID:       "meta_desc.too_long",
			Category: "Meta Descriptions",
			Severity: models.SeverityWarning,
			Message:  fmt.Sprintf("Meta description too long (%d chars, max 160)", len(d)),
			URL:      p.URL,
		}}
	}
	return nil
}

type metaDescDuplicate struct{}

func (c *metaDescDuplicate) Run(pages []*models.PageData) []models.CheckResult {
	counts := map[string][]string{}
	for _, p := range pages {
		d := strings.ToLower(strings.TrimSpace(p.MetaDesc))
		if d == "" {
			continue
		}
		counts[d] = append(counts[d], p.URL)
	}
	var results []models.CheckResult
	for desc, urls := range counts {
		if len(urls) > 1 {
			for _, u := range urls {
				results = append(results, models.CheckResult{
					ID:       "meta_desc.duplicate",
					Category: "Meta Descriptions",
					Severity: models.SeverityWarning,
					Message:  fmt.Sprintf("Duplicate meta description shared by %d pages", len(urls)),
					URL:      u,
					Details:  desc,
				})
			}
		}
	}
	return results
}
