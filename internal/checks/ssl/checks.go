package ssl

import (
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/cars24/seo-automation/internal/models"
)

// PageChecks returns all SSL / TLS certificate checks.
func PageChecks() []models.PageCheck {
	return []models.PageCheck{
		&certExpired{},
		&certExpiringSoon{},
		&tlsVersionOld{},
		&chainIncomplete{},
		&certMismatch{},
		&hstsPreloadMissing{},
	}
}

// ── ssl.cert_expired ────────────────────────────────────────────────────────

type certExpired struct{}

func (c *certExpired) Run(p *models.PageData) []models.CheckResult {
	if p.TLSInfo == nil || p.TLSInfo.CertNotAfter.IsZero() {
		return nil
	}
	if time.Now().After(p.TLSInfo.CertNotAfter) {
		return []models.CheckResult{{
			ID:       "ssl.cert_expired",
			Category: "HTTPS & Security",
			Severity: models.SeverityError,
			Message:  "SSL certificate has expired",
			URL:      p.URL,
			Details:  fmt.Sprintf("expired on %s", p.TLSInfo.CertNotAfter.Format("2006-01-02")),
		}}
	}
	return nil
}

// ── ssl.cert_expiring_soon ──────────────────────────────────────────────────

type certExpiringSoon struct{}

func (c *certExpiringSoon) Run(p *models.PageData) []models.CheckResult {
	if p.TLSInfo == nil || p.TLSInfo.CertNotAfter.IsZero() {
		return nil
	}
	daysLeft := time.Until(p.TLSInfo.CertNotAfter).Hours() / 24
	if daysLeft > 0 && daysLeft <= 30 {
		return []models.CheckResult{{
			ID:       "ssl.cert_expiring_soon",
			Category: "HTTPS & Security",
			Severity: models.SeverityWarning,
			Message:  fmt.Sprintf("SSL certificate expires in %d days", int(daysLeft)),
			URL:      p.URL,
			Details:  fmt.Sprintf("expires on %s", p.TLSInfo.CertNotAfter.Format("2006-01-02")),
		}}
	}
	return nil
}

// ── ssl.tls_version_old ─────────────────────────────────────────────────────

type tlsVersionOld struct{}

func (c *tlsVersionOld) Run(p *models.PageData) []models.CheckResult {
	if p.TLSInfo == nil || p.TLSInfo.Version == "" {
		return nil
	}
	if p.TLSInfo.Version == "TLS 1.0" || p.TLSInfo.Version == "TLS 1.1" {
		return []models.CheckResult{{
			ID:       "ssl.tls_version_old",
			Category: "HTTPS & Security",
			Severity: models.SeverityError,
			Message:  fmt.Sprintf("Server uses deprecated %s (minimum TLS 1.2 recommended)", p.TLSInfo.Version),
			URL:      p.URL,
			Details:  p.TLSInfo.Version,
		}}
	}
	return nil
}

// ── ssl.chain_incomplete ────────────────────────────────────────────────────

type chainIncomplete struct{}

func (c *chainIncomplete) Run(p *models.PageData) []models.CheckResult {
	if p.TLSInfo == nil {
		return nil
	}
	// A valid chain needs at least 2 certs: leaf + at least one intermediate.
	// A chain of exactly 1 means no intermediates were sent (self-signed or missing).
	if p.TLSInfo.ChainLength == 1 {
		return []models.CheckResult{{
			ID:       "ssl.chain_incomplete",
			Category: "HTTPS & Security",
			Severity: models.SeverityWarning,
			Message:  "SSL certificate chain may be incomplete (no intermediate certificates)",
			URL:      p.URL,
			Details:  fmt.Sprintf("chain length: %d (expected ≥ 2)", p.TLSInfo.ChainLength),
		}}
	}
	return nil
}

// ── ssl.cert_mismatch ───────────────────────────────────────────────────────

type certMismatch struct{}

func (c *certMismatch) Run(p *models.PageData) []models.CheckResult {
	if p.TLSInfo == nil || p.TLSInfo.CertSubject == "" {
		return nil
	}
	parsed, err := url.Parse(p.FinalURL)
	if err != nil || parsed.Hostname() == "" {
		return nil
	}
	host := parsed.Hostname()

	// Check if host matches the CN or any SAN DNS name.
	if matchesDomain(host, p.TLSInfo.CertSubject) {
		return nil
	}
	for _, san := range p.TLSInfo.CertDNSNames {
		if matchesDomain(host, san) {
			return nil
		}
	}

	return []models.CheckResult{{
		ID:       "ssl.cert_mismatch",
		Category: "HTTPS & Security",
		Severity: models.SeverityError,
		Message:  fmt.Sprintf("SSL certificate does not match domain %q", host),
		URL:      p.URL,
		Details:  fmt.Sprintf("cert CN=%s, SANs=%v", p.TLSInfo.CertSubject, p.TLSInfo.CertDNSNames),
	}}
}

// matchesDomain checks if a hostname matches a certificate name (supports wildcard).
func matchesDomain(host, pattern string) bool {
	host = strings.ToLower(host)
	pattern = strings.ToLower(pattern)
	if host == pattern {
		return true
	}
	// Wildcard: *.example.com matches sub.example.com but not example.com
	if strings.HasPrefix(pattern, "*.") {
		suffix := pattern[1:] // ".example.com"
		// host must have exactly one more label: "sub.example.com"
		if strings.HasSuffix(host, suffix) && !strings.Contains(host[:len(host)-len(suffix)], ".") {
			return true
		}
	}
	return false
}

// ── ssl.hsts_preload_missing ────────────────────────────────────────────────

type hstsPreloadMissing struct{}

func (c *hstsPreloadMissing) Run(p *models.PageData) []models.CheckResult {
	if !strings.HasPrefix(p.FinalURL, "https://") {
		return nil
	}
	hsts := strings.ToLower(p.Headers["strict-transport-security"])
	if hsts == "" {
		return nil // hstsMissing check already covers this
	}

	hasIncludeSub := strings.Contains(hsts, "includesubdomains")
	hasPreload := strings.Contains(hsts, "preload")

	// Only flag if the site has HSTS with includeSubDomains but is missing preload,
	// meaning it's close to being preload-eligible but hasn't opted in.
	if hasIncludeSub && !hasPreload {
		return []models.CheckResult{{
			ID:       "ssl.hsts_preload_missing",
			Category: "HTTPS & Security",
			Severity: models.SeverityNotice,
			Message:  "HSTS header has includeSubDomains but missing preload directive",
			URL:      p.URL,
			Details:  "Add 'preload' to the HSTS header and submit to hstspreload.org",
		}}
	}
	return nil
}
