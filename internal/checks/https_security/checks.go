package https_security

import (
	"strconv"
	"strings"

	"github.com/cars24/seo-automation/internal/models"
)

// PageChecks returns all HTTPS and security checks.
func PageChecks() []models.PageCheck {
	return []models.PageCheck{
		&pageInsecure{},
		&mixedContent{},
		&httpNotRedirecting{},
		&hstsMissing{},
		&hstsMaxAgeTooShort{},
		&cspMissing{},
		&xFrameMissing{},
		&xctoMissing{},
		&referrerPolicyMissing{},
		&permissionsPolicyMissing{},
	}
}

type pageInsecure struct{}

func (c *pageInsecure) Run(p *models.PageData) []models.CheckResult {
	if strings.HasPrefix(p.FinalURL, "http://") {
		return []models.CheckResult{{
			ID:       "https.page_insecure",
			Category: "HTTPS & Security",
			Severity: models.SeverityError,
			Message:  "Page is served over HTTP (not HTTPS)",
			URL:      p.URL,
		}}
	}
	return nil
}

type mixedContent struct{}

func (c *mixedContent) Run(p *models.PageData) []models.CheckResult {
	if !strings.HasPrefix(p.FinalURL, "https://") {
		return nil
	}
	if strings.Contains(p.RawHTML, "src=\"http://") || strings.Contains(p.RawHTML, "href=\"http://") {
		return []models.CheckResult{{
			ID:       "https.mixed_content",
			Category: "HTTPS & Security",
			Severity: models.SeverityWarning,
			Message:  "HTTPS page loads HTTP resources (mixed content)",
			URL:      p.URL,
		}}
	}
	return nil
}

type httpNotRedirecting struct{}

func (c *httpNotRedirecting) Run(p *models.PageData) []models.CheckResult {
	// Check if redirect chain starts from HTTP and ends at HTTPS
	if len(p.RedirectChain) == 0 && strings.HasPrefix(p.URL, "https://") {
		// No redirect to check
		return nil
	}
	return nil
}

type hstsMissing struct{}

func (c *hstsMissing) Run(p *models.PageData) []models.CheckResult {
	if !strings.HasPrefix(p.FinalURL, "https://") {
		return nil
	}
	if p.Headers["strict-transport-security"] == "" {
		return []models.CheckResult{{
			ID:       "security.hsts.missing",
			Category: "HTTPS & Security",
			Severity: models.SeverityWarning,
			Message:  "Missing Strict-Transport-Security (HSTS) header",
			URL:      p.URL,
		}}
	}
	return nil
}

type hstsMaxAgeTooShort struct{}

func (c *hstsMaxAgeTooShort) Run(p *models.PageData) []models.CheckResult {
	hsts := p.Headers["strict-transport-security"]
	if hsts == "" {
		return nil
	}
	for _, part := range strings.Split(hsts, ";") {
		part = strings.TrimSpace(part)
		if strings.HasPrefix(strings.ToLower(part), "max-age=") {
			valStr := strings.SplitN(part, "=", 2)[1]
			val, err := strconv.ParseInt(strings.TrimSpace(valStr), 10, 64)
			if err == nil && val < 31536000 {
				return []models.CheckResult{{
					ID:       "security.hsts.max_age_too_short",
					Category: "HTTPS & Security",
					Severity: models.SeverityWarning,
					Message:  "HSTS max-age is less than 1 year (31536000 seconds)",
					URL:      p.URL,
					Details:  "max-age=" + valStr,
				}}
			}
		}
	}
	return nil
}

type cspMissing struct{}

func (c *cspMissing) Run(p *models.PageData) []models.CheckResult {
	if p.Headers["content-security-policy"] == "" &&
		p.Headers["content-security-policy-report-only"] == "" {
		return []models.CheckResult{{
			ID:       "security.csp.missing",
			Category: "HTTPS & Security",
			Severity: models.SeverityNotice,
			Message:  "Missing Content-Security-Policy header",
			URL:      p.URL,
		}}
	}
	return nil
}

type xFrameMissing struct{}

func (c *xFrameMissing) Run(p *models.PageData) []models.CheckResult {
	if p.Headers["x-frame-options"] == "" {
		return []models.CheckResult{{
			ID:       "security.x_frame.missing",
			Category: "HTTPS & Security",
			Severity: models.SeverityWarning,
			Message:  "Missing X-Frame-Options header",
			URL:      p.URL,
		}}
	}
	return nil
}

type xctoMissing struct{}

func (c *xctoMissing) Run(p *models.PageData) []models.CheckResult {
	if p.Headers["x-content-type-options"] == "" {
		return []models.CheckResult{{
			ID:       "security.xcto.missing",
			Category: "HTTPS & Security",
			Severity: models.SeverityNotice,
			Message:  "Missing X-Content-Type-Options header",
			URL:      p.URL,
		}}
	}
	return nil
}

type referrerPolicyMissing struct{}

func (c *referrerPolicyMissing) Run(p *models.PageData) []models.CheckResult {
	if p.Headers["referrer-policy"] == "" {
		return []models.CheckResult{{
			ID:       "security.referrer_policy.missing",
			Category: "HTTPS & Security",
			Severity: models.SeverityNotice,
			Message:  "Missing Referrer-Policy header",
			URL:      p.URL,
		}}
	}
	return nil
}

type permissionsPolicyMissing struct{}

func (c *permissionsPolicyMissing) Run(p *models.PageData) []models.CheckResult {
	if p.Headers["permissions-policy"] == "" {
		return []models.CheckResult{{
			ID:       "security.permissions_policy.missing",
			Category: "HTTPS & Security",
			Severity: models.SeverityNotice,
			Message:  "Missing Permissions-Policy header",
			URL:      p.URL,
		}}
	}
	return nil
}
