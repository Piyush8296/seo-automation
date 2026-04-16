package images

import (
	"fmt"
	"strings"

	"github.com/cars24/seo-automation/internal/models"
)

var imageExtensions = []string{".jpg", ".jpeg", ".png", ".gif", ".webp", ".svg", ".avif", ".bmp"}

// PageChecks returns per-page image checks.
func PageChecks() []models.PageCheck {
	return []models.PageCheck{
		&imageAltMissing{},
		&imageAltEmptyNonDecorative{},
		&imageAltTooLong{},
		&imageAltIsFilename{},
		&imageDimensionsMissing{},
		&imageLazyAboveFold{},
		&imageMissingSrcset{},
		&imageFormatNotModern{},
		&imageSizeTooLarge{},
		&imageBroken{},
		&imageNoWidthHeightCLS{},
	}
}

type imageAltMissing struct{}

func (c *imageAltMissing) Run(p *models.PageData) []models.CheckResult {
	var results []models.CheckResult
	for _, img := range p.Images {
		if !img.AltPresent {
			results = append(results, models.CheckResult{
				ID:       "images.alt.missing",
				Category: "Images",
				Severity: models.SeverityError,
				Message:  "Image is missing alt attribute",
				URL:      p.URL,
				Details:  img.Src,
			})
		}
	}
	return results
}

type imageAltEmptyNonDecorative struct{}

func (c *imageAltEmptyNonDecorative) Run(p *models.PageData) []models.CheckResult {
	var results []models.CheckResult
	for _, img := range p.Images {
		if img.AltPresent && img.Alt == "" {
			// Check if likely decorative (small icon, spacer, etc.)
			srcLower := strings.ToLower(img.Src)
			isLikelyDecorative := strings.Contains(srcLower, "icon") ||
				strings.Contains(srcLower, "spacer") ||
				strings.Contains(srcLower, "pixel") ||
				strings.Contains(srcLower, "tracking") ||
				(img.Width > 0 && img.Width <= 2) ||
				(img.Height > 0 && img.Height <= 2)
			if !isLikelyDecorative {
				results = append(results, models.CheckResult{
					ID:       "images.alt.empty_non_decorative",
					Category: "Images",
					Severity: models.SeverityWarning,
					Message:  "Non-decorative image has empty alt text",
					URL:      p.URL,
					Details:  img.Src,
				})
			}
		}
	}
	return results
}

type imageAltTooLong struct{}

func (c *imageAltTooLong) Run(p *models.PageData) []models.CheckResult {
	var results []models.CheckResult
	for _, img := range p.Images {
		if len(img.Alt) > 100 {
			results = append(results, models.CheckResult{
				ID:       "images.alt.too_long",
				Category: "Images",
				Severity: models.SeverityNotice,
				Message:  fmt.Sprintf("Image alt text too long (%d chars, max 100)", len(img.Alt)),
				URL:      p.URL,
				Details:  img.Src,
			})
		}
	}
	return results
}

type imageAltIsFilename struct{}

func (c *imageAltIsFilename) Run(p *models.PageData) []models.CheckResult {
	var results []models.CheckResult
	for _, img := range p.Images {
		if img.Alt == "" {
			continue
		}
		altLower := strings.ToLower(img.Alt)
		for _, ext := range imageExtensions {
			if strings.HasSuffix(altLower, ext) {
				results = append(results, models.CheckResult{
					ID:       "images.alt.is_filename",
					Category: "Images",
					Severity: models.SeverityWarning,
					Message:  "Image alt text looks like a filename",
					URL:      p.URL,
					Details:  fmt.Sprintf("alt=%q src=%s", img.Alt, img.Src),
				})
				break
			}
		}
	}
	return results
}

type imageDimensionsMissing struct{}

func (c *imageDimensionsMissing) Run(p *models.PageData) []models.CheckResult {
	count := 0
	for _, img := range p.Images {
		if img.Width == 0 || img.Height == 0 {
			count++
		}
	}
	if count > 0 {
		return []models.CheckResult{{
			ID:       "images.dimensions.missing",
			Category: "Images",
			Severity: models.SeverityWarning,
			Message:  fmt.Sprintf("%d image(s) missing width/height attributes", count),
			URL:      p.URL,
		}}
	}
	return nil
}

type imageLazyAboveFold struct{}

func (c *imageLazyAboveFold) Run(p *models.PageData) []models.CheckResult {
	for _, img := range p.Images {
		if img.IsAboveFold && img.Loading == "lazy" {
			return []models.CheckResult{{
				ID:       "images.lazy.above_fold",
				Category: "Images",
				Severity: models.SeverityWarning,
				Message:  "Above-fold image has lazy loading (may delay LCP)",
				URL:      p.URL,
				Details:  img.Src,
			}}
		}
	}
	return nil
}

type imageMissingSrcset struct{}

func (c *imageMissingSrcset) Run(p *models.PageData) []models.CheckResult {
	count := 0
	for _, img := range p.Images {
		srcLower := strings.ToLower(img.Src)
		isLikelyContent := !strings.Contains(srcLower, "icon") &&
			!strings.Contains(srcLower, "logo") &&
			!strings.Contains(srcLower, "sprite") &&
			!strings.Contains(srcLower, "pixel") &&
			!img.HasSrcset
		if isLikelyContent {
			count++
		}
	}
	if count > 0 {
		return []models.CheckResult{{
			ID:       "images.missing_srcset",
			Category: "Images",
			Severity: models.SeverityNotice,
			Message:  fmt.Sprintf("%d content image(s) missing srcset (responsive images)", count),
			URL:      p.URL,
		}}
	}
	return nil
}

// modernFormats are image formats considered modern/optimal for web.
var modernFormats = map[string]bool{
	"webp": true,
	"avif": true,
	"svg":  true,
}

// ── images.format.not_modern ──────────────────────────────────────────────────

type imageFormatNotModern struct{}

func (c *imageFormatNotModern) Run(p *models.PageData) []models.CheckResult {
	var results []models.CheckResult
	for _, img := range p.Images {
		if img.Format == "" || modernFormats[img.Format] {
			continue
		}
		results = append(results, models.CheckResult{
			ID:       "images.format.not_modern",
			Category: "Images",
			Severity: models.SeverityNotice,
			Message:  fmt.Sprintf("Image not in modern format (current: %s, recommend WebP/AVIF)", img.Format),
			URL:      p.URL,
			Details:  img.Src,
		})
	}
	return results
}

// ── images.size.too_large ─────────────────────────────────────────────────────

const maxImageBytes = 200 * 1024 // 200 KB

type imageSizeTooLarge struct{}

func (c *imageSizeTooLarge) Run(p *models.PageData) []models.CheckResult {
	var results []models.CheckResult
	for _, img := range p.Images {
		if img.FileSize <= 0 {
			continue
		}
		if img.FileSize > maxImageBytes {
			sizeKB := img.FileSize / 1024
			results = append(results, models.CheckResult{
				ID:       "images.size.too_large",
				Category: "Images",
				Severity: models.SeverityWarning,
				Message:  fmt.Sprintf("Image file too large (%d KB, max 200 KB)", sizeKB),
				URL:      p.URL,
				Details:  img.Src,
			})
		}
	}
	return results
}

// ── images.broken ─────────────────────────────────────────────────────────────

type imageBroken struct{}

func (c *imageBroken) Run(p *models.PageData) []models.CheckResult {
	var results []models.CheckResult
	for _, img := range p.Images {
		if img.StatusCode == 0 {
			// No response or network error — skip (could be data URI, relative, etc.)
			continue
		}
		if img.StatusCode >= 400 {
			results = append(results, models.CheckResult{
				ID:       "images.broken",
				Category: "Images",
				Severity: models.SeverityError,
				Message:  fmt.Sprintf("Broken image (HTTP %d)", img.StatusCode),
				URL:      p.URL,
				Details:  img.Src,
			})
		}
	}
	return results
}

// ── images.no_width_height_cls ────────────────────────────────────────────────

type imageNoWidthHeightCLS struct{}

func (c *imageNoWidthHeightCLS) Run(p *models.PageData) []models.CheckResult {
	var results []models.CheckResult
	for _, img := range p.Images {
		if !img.IsAboveFold {
			continue
		}
		if img.Width == 0 || img.Height == 0 {
			results = append(results, models.CheckResult{
				ID:       "images.no_width_height_cls",
				Category: "Images",
				Severity: models.SeverityWarning,
				Message:  "Above-fold image missing explicit width/height (CLS risk)",
				URL:      p.URL,
				Details:  img.Src,
			})
		}
	}
	return results
}
