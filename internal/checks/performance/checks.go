package performance

import (
	"fmt"
	"strings"

	"github.com/cars24/seo-automation/internal/models"
)

// PageChecks returns all performance checks.
func PageChecks() []models.PageCheck {
	return []models.PageCheck{
		&responseTimeSlow{},
		&responseTimeCritical{},
		&htmlSizeLarge{},
		&compressionMissing{},
		&cacheControlMissing{},
		&renderBlockingScripts{},
		&renderBlockingCSSCount{},
		&externalScriptsCount{},
		&lcpCandidateLazyLoaded{},
		&clsRiskImagesNoDimensions{},
		&missingPreconnect{},
		&inlineCSSLarge{},
	}
}

type responseTimeSlow struct{}

func (c *responseTimeSlow) Run(p *models.PageData) []models.CheckResult {
	if p.ResponseTimeMs > 4000 {
		return nil // let critical handle it
	}
	if p.ResponseTimeMs > 2000 {
		return []models.CheckResult{{
			ID:       "perf.response_time.slow",
			Category: "Performance",
			Severity: models.SeverityWarning,
			Message:  fmt.Sprintf("Slow response time (%dms)", p.ResponseTimeMs),
			URL:      p.URL,
		}}
	}
	return nil
}

type responseTimeCritical struct{}

func (c *responseTimeCritical) Run(p *models.PageData) []models.CheckResult {
	if p.ResponseTimeMs > 4000 {
		return []models.CheckResult{{
			ID:       "perf.response_time.critical",
			Category: "Performance",
			Severity: models.SeverityError,
			Message:  fmt.Sprintf("Critical response time (%dms)", p.ResponseTimeMs),
			URL:      p.URL,
		}}
	}
	return nil
}

type htmlSizeLarge struct{}

func (c *htmlSizeLarge) Run(p *models.PageData) []models.CheckResult {
	if p.HTMLSizeBytes > 100*1024 {
		return []models.CheckResult{{
			ID:       "perf.html_size.large",
			Category: "Performance",
			Severity: models.SeverityWarning,
			Message:  fmt.Sprintf("Large HTML page size (%dKB)", p.HTMLSizeBytes/1024),
			URL:      p.URL,
		}}
	}
	return nil
}

type compressionMissing struct{}

func (c *compressionMissing) Run(p *models.PageData) []models.CheckResult {
	enc := strings.ToLower(p.Headers["content-encoding"])
	if enc == "" && p.HTMLSizeBytes > 1024 {
		return []models.CheckResult{{
			ID:       "perf.compression.missing",
			Category: "Performance",
			Severity: models.SeverityWarning,
			Message:  "Response is not compressed (no gzip/brotli)",
			URL:      p.URL,
		}}
	}
	return nil
}

type cacheControlMissing struct{}

func (c *cacheControlMissing) Run(p *models.PageData) []models.CheckResult {
	if p.Headers["cache-control"] == "" {
		return []models.CheckResult{{
			ID:       "perf.cache_control.missing",
			Category: "Performance",
			Severity: models.SeverityNotice,
			Message:  "Missing Cache-Control header",
			URL:      p.URL,
		}}
	}
	return nil
}

type renderBlockingScripts struct{}

func (c *renderBlockingScripts) Run(p *models.PageData) []models.CheckResult {
	if p.RenderBlockingScripts > 0 {
		return []models.CheckResult{{
			ID:       "perf.render_blocking.scripts",
			Category: "Performance",
			Severity: models.SeverityWarning,
			Message:  fmt.Sprintf("%d render-blocking script(s) in <head>", p.RenderBlockingScripts),
			URL:      p.URL,
		}}
	}
	return nil
}

type renderBlockingCSSCount struct{}

func (c *renderBlockingCSSCount) Run(p *models.PageData) []models.CheckResult {
	if p.RenderBlockingCSS > 3 {
		return []models.CheckResult{{
			ID:       "perf.render_blocking.css_count",
			Category: "Performance",
			Severity: models.SeverityWarning,
			Message:  fmt.Sprintf("Too many render-blocking stylesheets (%d)", p.RenderBlockingCSS),
			URL:      p.URL,
		}}
	}
	return nil
}

type externalScriptsCount struct{}

func (c *externalScriptsCount) Run(p *models.PageData) []models.CheckResult {
	if p.ExternalScriptCount > 5 {
		return []models.CheckResult{{
			ID:       "perf.external_scripts.count",
			Category: "Performance",
			Severity: models.SeverityWarning,
			Message:  fmt.Sprintf("Too many external scripts (%d)", p.ExternalScriptCount),
			URL:      p.URL,
		}}
	}
	return nil
}

type lcpCandidateLazyLoaded struct{}

func (c *lcpCandidateLazyLoaded) Run(p *models.PageData) []models.CheckResult {
	for _, img := range p.Images {
		if img.IsAboveFold && img.Loading == "lazy" {
			return []models.CheckResult{{
				ID:       "perf.lcp_candidate.lazy_loaded",
				Category: "Performance",
				Severity: models.SeverityWarning,
				Message:  "Potential LCP image has lazy loading (above-fold image)",
				URL:      p.URL,
				Details:  img.Src,
			}}
		}
	}
	return nil
}

type clsRiskImagesNoDimensions struct{}

func (c *clsRiskImagesNoDimensions) Run(p *models.PageData) []models.CheckResult {
	count := 0
	for _, img := range p.Images {
		if img.Width == 0 || img.Height == 0 {
			count++
		}
	}
	if count > 0 {
		return []models.CheckResult{{
			ID:       "perf.cls_risk.images_no_dimensions",
			Category: "Performance",
			Severity: models.SeverityWarning,
			Message:  fmt.Sprintf("%d image(s) missing width/height (CLS risk)", count),
			URL:      p.URL,
		}}
	}
	return nil
}

type missingPreconnect struct{}

func (c *missingPreconnect) Run(p *models.PageData) []models.CheckResult {
	if p.ExternalScriptCount > 0 && !strings.Contains(p.RawHTML, "rel=\"preconnect\"") &&
		!strings.Contains(p.RawHTML, "rel='preconnect'") {
		return []models.CheckResult{{
			ID:       "perf.missing_preconnect",
			Category: "Performance",
			Severity: models.SeverityNotice,
			Message:  "No preconnect hints for external origins",
			URL:      p.URL,
		}}
	}
	return nil
}

type inlineCSSLarge struct{}

func (c *inlineCSSLarge) Run(p *models.PageData) []models.CheckResult {
	if p.InlineCSSBytes > 50*1024 {
		return []models.CheckResult{{
			ID:       "perf.inline_css.large",
			Category: "Performance",
			Severity: models.SeverityNotice,
			Message:  fmt.Sprintf("Large inline CSS (%dKB)", p.InlineCSSBytes/1024),
			URL:      p.URL,
		}}
	}
	return nil
}
