package mobile

import (
	"regexp"
	"strings"

	"github.com/cars24/seo-automation/internal/models"
)

var smallFontRegex = regexp.MustCompile(`font-size\s*:\s*([0-9]+(?:\.[0-9]+)?)\s*px`)

// PageChecks returns per-page mobile checks.
func PageChecks() []models.PageCheck {
	return []models.PageCheck{
		&viewportMissing{},
		&viewportInvalid{},
		&fontSizeTooSmall{},
		&userScalableDisabled{},
	}
}

type viewportMissing struct{}

func (c *viewportMissing) Run(p *models.PageData) []models.CheckResult {
	if !p.HasViewport {
		return []models.CheckResult{{
			ID:       "mobile.viewport.missing",
			Category: "Mobile",
			Severity: models.SeverityError,
			Message:  "Missing viewport meta tag",
			URL:      p.URL,
			Platform: models.PlatformMobile,
		}}
	}
	return nil
}

type viewportInvalid struct{}

func (c *viewportInvalid) Run(p *models.PageData) []models.CheckResult {
	if !p.HasViewport {
		return nil
	}
	vp := strings.ToLower(p.ViewportContent)
	if !strings.Contains(vp, "width=") {
		return []models.CheckResult{{
			ID:       "mobile.viewport.invalid",
			Category: "Mobile",
			Severity: models.SeverityWarning,
			Message:  "Viewport meta tag missing width directive",
			URL:      p.URL,
			Details:  p.ViewportContent,
			Platform: models.PlatformMobile,
		}}
	}
	return nil
}

type fontSizeTooSmall struct{}

func (c *fontSizeTooSmall) Run(p *models.PageData) []models.CheckResult {
	// Scan inline styles for small font sizes
	matches := smallFontRegex.FindAllStringSubmatch(p.RawHTML, -1)
	for _, m := range matches {
		if len(m) < 2 {
			continue
		}
		// Parse font size
		var size float64
		for _, ch := range m[1] {
			if ch >= '0' && ch <= '9' {
				size = size*10 + float64(ch-'0')
			} else if ch == '.' {
				// simplified: ignore decimal for now
			}
		}
		if size > 0 && size < 12 {
			return []models.CheckResult{{
				ID:       "mobile.font_size.too_small",
				Category: "Mobile",
				Severity: models.SeverityNotice,
				Message:  "Page uses font sizes smaller than 12px (poor mobile readability)",
				URL:      p.URL,
				Platform: models.PlatformMobile,
			}}
		}
	}
	return nil
}

type userScalableDisabled struct{}

func (c *userScalableDisabled) Run(p *models.PageData) []models.CheckResult {
	if !p.HasViewport {
		return nil
	}
	vp := strings.ToLower(p.ViewportContent)
	if strings.Contains(vp, "user-scalable=no") || strings.Contains(vp, "user-scalable=0") {
		return []models.CheckResult{{
			ID:       "mobile.user_scalable.disabled",
			Category: "Mobile",
			Severity: models.SeverityWarning,
			Message:  "Viewport disables user zoom (user-scalable=no) — accessibility issue",
			URL:      p.URL,
			Platform: models.PlatformMobile,
		}}
	}
	return nil
}
