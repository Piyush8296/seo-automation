package international

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"

	"github.com/cars24/seo-automation/internal/models"
)

var bcp47Pattern = regexp.MustCompile(`^[a-zA-Z]{2,3}(-[a-zA-Z]{2,4})?(-[a-zA-Z0-9]{1,8})*$`)

// PageChecks returns per-page hreflang checks.
func PageChecks() []models.PageCheck {
	return []models.PageCheck{
		&hreflangURLNotAbsolute{},
		&hreflangInvalidLangCode{},
		&hreflangMissingXDefault{},
		&hreflangMissingSelfRef{},
		&hreflangNonCanonicalURL{},
		&hreflangPointsToRedirect{},
	}
}

// SiteChecks returns site-wide hreflang checks.
func SiteChecks() []models.SiteCheck {
	return []models.SiteCheck{
		&hreflangMissingReturnLink{},
	}
}

type hreflangURLNotAbsolute struct{}

func (c *hreflangURLNotAbsolute) Run(p *models.PageData) []models.CheckResult {
	if len(p.HreflangTags) == 0 {
		return nil
	}
	var results []models.CheckResult
	for _, h := range p.HreflangTags {
		parsed, err := url.Parse(h.URL)
		if err != nil || !parsed.IsAbs() {
			results = append(results, models.CheckResult{
				ID:       "i18n.hreflang.url_not_absolute",
				Category: "International",
				Severity: models.SeverityError,
				Message:  fmt.Sprintf("Hreflang URL is not absolute: %s", h.URL),
				URL:      p.URL,
				Details:  fmt.Sprintf("lang=%s url=%s", h.Lang, h.URL),
			})
		}
	}
	return results
}

type hreflangInvalidLangCode struct{}

func (c *hreflangInvalidLangCode) Run(p *models.PageData) []models.CheckResult {
	if len(p.HreflangTags) == 0 {
		return nil
	}
	var results []models.CheckResult
	for _, h := range p.HreflangTags {
		if h.Lang == "x-default" {
			continue
		}
		if !bcp47Pattern.MatchString(h.Lang) {
			results = append(results, models.CheckResult{
				ID:       "i18n.hreflang.invalid_lang_code",
				Category: "International",
				Severity: models.SeverityWarning,
				Message:  fmt.Sprintf("Invalid hreflang language code: %s", h.Lang),
				URL:      p.URL,
			})
		}
	}
	return results
}

type hreflangMissingXDefault struct{}

func (c *hreflangMissingXDefault) Run(p *models.PageData) []models.CheckResult {
	if len(p.HreflangTags) == 0 {
		return nil
	}
	for _, h := range p.HreflangTags {
		if h.Lang == "x-default" {
			return nil
		}
	}
	return []models.CheckResult{{
		ID:       "i18n.hreflang.missing_x_default",
		Category: "International",
		Severity: models.SeverityWarning,
		Message:  "Hreflang tags present but no x-default entry",
		URL:      p.URL,
	}}
}

type hreflangMissingSelfRef struct{}

func (c *hreflangMissingSelfRef) Run(p *models.PageData) []models.CheckResult {
	if len(p.HreflangTags) == 0 {
		return nil
	}
	pageURL := strings.TrimRight(p.FinalURL, "/")
	for _, h := range p.HreflangTags {
		if strings.TrimRight(h.URL, "/") == pageURL {
			return nil
		}
	}
	return []models.CheckResult{{
		ID:       "i18n.hreflang.missing_self_ref",
		Category: "International",
		Severity: models.SeverityWarning,
		Message:  "Page is missing a self-referencing hreflang tag",
		URL:      p.URL,
	}}
}

type hreflangNonCanonicalURL struct{}

func (c *hreflangNonCanonicalURL) Run(p *models.PageData) []models.CheckResult {
	if len(p.HreflangTags) == 0 || p.Canonical == "" {
		return nil
	}
	canonical := strings.TrimRight(p.Canonical, "/")
	pageURL := strings.TrimRight(p.FinalURL, "/")
	if strings.EqualFold(canonical, pageURL) {
		return nil
	}
	// Self-referencing hreflang should use canonical
	for _, h := range p.HreflangTags {
		hURL := strings.TrimRight(h.URL, "/")
		if strings.EqualFold(hURL, pageURL) && !strings.EqualFold(hURL, canonical) {
			return []models.CheckResult{{
				ID:       "i18n.hreflang.non_canonical_url",
				Category: "International",
				Severity: models.SeverityWarning,
				Message:  "Hreflang self-reference uses non-canonical URL",
				URL:      p.URL,
			}}
		}
	}
	return nil
}

type hreflangPointsToRedirect struct{}

func (c *hreflangPointsToRedirect) Run(p *models.PageData) []models.CheckResult {
	// Can't verify without fetching; handled at crawler level
	return nil
}

// Site-wide: check that for every hreflang A→B, B also has hreflang B→A.
type hreflangMissingReturnLink struct{}

func (c *hreflangMissingReturnLink) Run(pages []*models.PageData) []models.CheckResult {
	// Build map: url → hreflang tags
	pageHreflang := map[string][]models.Hreflang{}
	for _, p := range pages {
		if len(p.HreflangTags) > 0 {
			pageHreflang[strings.TrimRight(p.FinalURL, "/")] = p.HreflangTags
		}
	}

	var results []models.CheckResult
	for pageURL, tags := range pageHreflang {
		for _, h := range tags {
			if h.Lang == "x-default" {
				continue
			}
			targetURL := strings.TrimRight(h.URL, "/")
			targetTags, exists := pageHreflang[targetURL]
			if !exists {
				continue
			}
			// Check if target has a return link to this page
			hasReturn := false
			for _, t := range targetTags {
				if strings.TrimRight(t.URL, "/") == pageURL {
					hasReturn = true
					break
				}
			}
			if !hasReturn {
				results = append(results, models.CheckResult{
					ID:       "i18n.hreflang.missing_return_link",
					Category: "International",
					Severity: models.SeverityError,
					Message:  fmt.Sprintf("Hreflang target missing return link (lang=%s)", h.Lang),
					URL:      pageURL,
					Details:  targetURL,
				})
				break // one result per page
			}
		}
	}
	return results
}
