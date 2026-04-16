package resources

import (
	"fmt"

	"github.com/cars24/seo-automation/internal/models"
)

const (
	totalSizeBudgetBytes = 3 * 1024 * 1024 // 3MB
	totalRequestBudget   = 50
)

// PageChecks returns all per-page sub-resource checks.
func PageChecks() []models.PageCheck {
	return []models.PageCheck{
		&scriptBroken{},
		&cssBroken{},
		&fontBroken{},
		&totalSizeTooLarge{},
		&tooManyRequests{},
		&fontNoDisplaySwap{},
	}
}

func isBroken(status int) bool {
	return status >= 400 && status < 600
}

type scriptBroken struct{}

func (c *scriptBroken) Run(p *models.PageData) []models.CheckResult {
	var results []models.CheckResult
	for _, r := range p.Resources {
		if r.Type != models.ResourceScript {
			continue
		}
		if !isBroken(r.StatusCode) {
			continue
		}
		results = append(results, models.CheckResult{
			ID:       "resources.script.broken",
			Category: "Resources",
			Severity: models.SeverityError,
			Message:  fmt.Sprintf("Script returned %d", r.StatusCode),
			URL:      p.URL,
			Details:  r.URL,
		})
	}
	return results
}

type cssBroken struct{}

func (c *cssBroken) Run(p *models.PageData) []models.CheckResult {
	var results []models.CheckResult
	for _, r := range p.Resources {
		if r.Type != models.ResourceCSS {
			continue
		}
		if !isBroken(r.StatusCode) {
			continue
		}
		results = append(results, models.CheckResult{
			ID:       "resources.css.broken",
			Category: "Resources",
			Severity: models.SeverityError,
			Message:  fmt.Sprintf("Stylesheet returned %d", r.StatusCode),
			URL:      p.URL,
			Details:  r.URL,
		})
	}
	return results
}

type fontBroken struct{}

func (c *fontBroken) Run(p *models.PageData) []models.CheckResult {
	var results []models.CheckResult
	for _, r := range p.Resources {
		if r.Type != models.ResourceFont {
			continue
		}
		if !isBroken(r.StatusCode) {
			continue
		}
		results = append(results, models.CheckResult{
			ID:       "resources.font.broken",
			Category: "Resources",
			Severity: models.SeverityError,
			Message:  fmt.Sprintf("Font returned %d", r.StatusCode),
			URL:      p.URL,
			Details:  r.URL,
		})
	}
	return results
}

type totalSizeTooLarge struct{}

func (c *totalSizeTooLarge) Run(p *models.PageData) []models.CheckResult {
	// Only run when sub-resource discovery actually ran (Resources populated with sizes).
	if len(p.Resources) == 0 {
		return nil
	}
	var total int64 = int64(p.HTMLSizeBytes)
	for _, img := range p.Images {
		total += img.FileSize
	}
	for _, r := range p.Resources {
		total += r.FileSize
	}
	if total <= totalSizeBudgetBytes {
		return nil
	}
	return []models.CheckResult{{
		ID:       "resources.total_size_too_large",
		Category: "Resources",
		Severity: models.SeverityWarning,
		Message:  fmt.Sprintf("Total page weight %.1fMB exceeds 3MB budget", float64(total)/1024/1024),
		URL:      p.URL,
	}}
}

type tooManyRequests struct{}

func (c *tooManyRequests) Run(p *models.PageData) []models.CheckResult {
	if len(p.Resources) == 0 {
		return nil
	}
	count := 1 + len(p.Images) + len(p.Resources) // +1 for the HTML document
	if count <= totalRequestBudget {
		return nil
	}
	return []models.CheckResult{{
		ID:       "resources.too_many_requests",
		Category: "Resources",
		Severity: models.SeverityWarning,
		Message:  fmt.Sprintf("Page requires %d HTTP requests (>%d)", count, totalRequestBudget),
		URL:      p.URL,
	}}
}

type fontNoDisplaySwap struct{}

func (c *fontNoDisplaySwap) Run(p *models.PageData) []models.CheckResult {
	if p.FontFaceNoDisplay == 0 {
		return nil
	}
	return []models.CheckResult{{
		ID:       "resources.font.no_display_swap",
		Category: "Resources",
		Severity: models.SeverityNotice,
		Message:  fmt.Sprintf("%d @font-face block(s) missing font-display", p.FontFaceNoDisplay),
		URL:      p.URL,
	}}
}
