package url_structure

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"

	"github.com/cars24/seo-automation/internal/models"
)

var sessionParams = []string{"jsessionid", "sessionid", "phpsessid", "aspsessionid", "sid"}

var nonDescriptivePattern = regexp.MustCompile(`^/(p|page|post|node|item|product|article|id|detail)s?/\d+/?$`)

// PageChecks returns per-page URL structure checks.
func PageChecks() []models.PageCheck {
	return []models.PageCheck{
		&urlTooLong{},
		&urlHasUnderscores{},
		&urlHasUppercase{},
		&urlHasSpaces{},
		&urlHasSessionParams{},
		&urlTooManyParams{},
		&urlDoubleSlash{},
		&urlNonDescriptive{},
	}
}

type urlTooLong struct{}

func (c *urlTooLong) Run(p *models.PageData) []models.CheckResult {
	if len(p.URL) > 100 {
		return []models.CheckResult{{
			ID:       "url.too_long",
			Category: "URL Structure",
			Severity: models.SeverityWarning,
			Message:  fmt.Sprintf("URL too long (%d chars, max 100)", len(p.URL)),
			URL:      p.URL,
		}}
	}
	return nil
}

type urlHasUnderscores struct{}

func (c *urlHasUnderscores) Run(p *models.PageData) []models.CheckResult {
	parsed, err := url.Parse(p.URL)
	if err != nil {
		return nil
	}
	if strings.Contains(parsed.Path, "_") {
		return []models.CheckResult{{
			ID:       "url.has_underscores",
			Category: "URL Structure",
			Severity: models.SeverityNotice,
			Message:  "URL path contains underscores (prefer hyphens)",
			URL:      p.URL,
		}}
	}
	return nil
}

type urlHasUppercase struct{}

func (c *urlHasUppercase) Run(p *models.PageData) []models.CheckResult {
	parsed, err := url.Parse(p.URL)
	if err != nil {
		return nil
	}
	if strings.ToLower(parsed.Path) != parsed.Path {
		return []models.CheckResult{{
			ID:       "url.has_uppercase",
			Category: "URL Structure",
			Severity: models.SeverityWarning,
			Message:  "URL path contains uppercase characters",
			URL:      p.URL,
		}}
	}
	return nil
}

type urlHasSpaces struct{}

func (c *urlHasSpaces) Run(p *models.PageData) []models.CheckResult {
	if strings.Contains(p.URL, "%20") || strings.Contains(p.URL, "+") {
		return []models.CheckResult{{
			ID:       "url.has_spaces",
			Category: "URL Structure",
			Severity: models.SeverityError,
			Message:  "URL contains spaces (encoded)",
			URL:      p.URL,
		}}
	}
	return nil
}

type urlHasSessionParams struct{}

func (c *urlHasSessionParams) Run(p *models.PageData) []models.CheckResult {
	lowerURL := strings.ToLower(p.URL)
	for _, param := range sessionParams {
		if strings.Contains(lowerURL, param) {
			return []models.CheckResult{{
				ID:       "url.has_session_params",
				Category: "URL Structure",
				Severity: models.SeverityError,
				Message:  fmt.Sprintf("URL contains session parameter: %s", param),
				URL:      p.URL,
			}}
		}
	}
	return nil
}

type urlTooManyParams struct{}

func (c *urlTooManyParams) Run(p *models.PageData) []models.CheckResult {
	parsed, err := url.Parse(p.URL)
	if err != nil {
		return nil
	}
	params := parsed.Query()
	if len(params) > 3 {
		return []models.CheckResult{{
			ID:       "url.too_many_params",
			Category: "URL Structure",
			Severity: models.SeverityWarning,
			Message:  fmt.Sprintf("URL has too many query parameters (%d, max 3)", len(params)),
			URL:      p.URL,
		}}
	}
	return nil
}

type urlDoubleSlash struct{}

func (c *urlDoubleSlash) Run(p *models.PageData) []models.CheckResult {
	parsed, err := url.Parse(p.URL)
	if err != nil {
		return nil
	}
	// Skip protocol's double slash
	path := parsed.Path
	if strings.Contains(path, "//") {
		return []models.CheckResult{{
			ID:       "url.double_slash",
			Category: "URL Structure",
			Severity: models.SeverityWarning,
			Message:  "URL path contains double slashes",
			URL:      p.URL,
		}}
	}
	return nil
}

type urlNonDescriptive struct{}

func (c *urlNonDescriptive) Run(p *models.PageData) []models.CheckResult {
	parsed, err := url.Parse(p.URL)
	if err != nil {
		return nil
	}
	path := parsed.Path
	if path == "/" || path == "" {
		return nil
	}
	if nonDescriptivePattern.MatchString(path) {
		return []models.CheckResult{{
			ID:       "url.non_descriptive",
			Category: "URL Structure",
			Severity: models.SeverityNotice,
			Message:  "URL path is non-descriptive (contains only numeric ID)",
			URL:      p.URL,
		}}
	}
	return nil
}
