package ssl

import (
	"testing"
	"time"

	"github.com/cars24/seo-automation/internal/models"
)

// ── helpers ─────────────────────────────────────────────────────────────────

func basePage() *models.PageData {
	return &models.PageData{
		URL:      "https://example.com",
		FinalURL: "https://example.com",
		Headers:  map[string]string{},
	}
}

func withTLS(p *models.PageData, t *models.TLSInfo) *models.PageData {
	p.TLSInfo = t
	return p
}

func expectNil(t *testing.T, results []models.CheckResult, checkID string) {
	t.Helper()
	for _, r := range results {
		if r.ID == checkID {
			t.Errorf("expected no result for %s, got: %s", checkID, r.Message)
		}
	}
}

func expectOne(t *testing.T, results []models.CheckResult, checkID string, sev models.Severity) {
	t.Helper()
	for _, r := range results {
		if r.ID == checkID {
			if r.Severity != sev {
				t.Errorf("%s: expected severity %s, got %s", checkID, sev, r.Severity)
			}
			return
		}
	}
	t.Errorf("expected result for %s, got none", checkID)
}

// ── ssl.cert_expired ────────────────────────────────────────────────────────

func TestCertExpired_Expired(t *testing.T) {
	p := withTLS(basePage(), &models.TLSInfo{
		CertNotAfter: time.Now().Add(-24 * time.Hour),
	})
	results := (&certExpired{}).Run(p)
	expectOne(t, results, "ssl.cert_expired", models.SeverityError)
}

func TestCertExpired_Valid(t *testing.T) {
	p := withTLS(basePage(), &models.TLSInfo{
		CertNotAfter: time.Now().Add(90 * 24 * time.Hour),
	})
	results := (&certExpired{}).Run(p)
	expectNil(t, results, "ssl.cert_expired")
}

func TestCertExpired_NoTLS(t *testing.T) {
	results := (&certExpired{}).Run(basePage())
	expectNil(t, results, "ssl.cert_expired")
}

// ── ssl.cert_expiring_soon ──────────────────────────────────────────────────

func TestCertExpiringSoon_15Days(t *testing.T) {
	p := withTLS(basePage(), &models.TLSInfo{
		CertNotAfter: time.Now().Add(15 * 24 * time.Hour),
	})
	results := (&certExpiringSoon{}).Run(p)
	expectOne(t, results, "ssl.cert_expiring_soon", models.SeverityWarning)
}

func TestCertExpiringSoon_1Day(t *testing.T) {
	p := withTLS(basePage(), &models.TLSInfo{
		CertNotAfter: time.Now().Add(1 * 24 * time.Hour),
	})
	results := (&certExpiringSoon{}).Run(p)
	expectOne(t, results, "ssl.cert_expiring_soon", models.SeverityWarning)
}

func TestCertExpiringSoon_60Days(t *testing.T) {
	p := withTLS(basePage(), &models.TLSInfo{
		CertNotAfter: time.Now().Add(60 * 24 * time.Hour),
	})
	results := (&certExpiringSoon{}).Run(p)
	expectNil(t, results, "ssl.cert_expiring_soon")
}

func TestCertExpiringSoon_AlreadyExpired(t *testing.T) {
	p := withTLS(basePage(), &models.TLSInfo{
		CertNotAfter: time.Now().Add(-1 * 24 * time.Hour),
	})
	results := (&certExpiringSoon{}).Run(p)
	// Already expired — handled by cert_expired, not cert_expiring_soon
	expectNil(t, results, "ssl.cert_expiring_soon")
}

// ── ssl.tls_version_old ─────────────────────────────────────────────────────

func TestTLSVersionOld_TLS10(t *testing.T) {
	p := withTLS(basePage(), &models.TLSInfo{Version: "TLS 1.0"})
	results := (&tlsVersionOld{}).Run(p)
	expectOne(t, results, "ssl.tls_version_old", models.SeverityError)
}

func TestTLSVersionOld_TLS11(t *testing.T) {
	p := withTLS(basePage(), &models.TLSInfo{Version: "TLS 1.1"})
	results := (&tlsVersionOld{}).Run(p)
	expectOne(t, results, "ssl.tls_version_old", models.SeverityError)
}

func TestTLSVersionOld_TLS12(t *testing.T) {
	p := withTLS(basePage(), &models.TLSInfo{Version: "TLS 1.2"})
	results := (&tlsVersionOld{}).Run(p)
	expectNil(t, results, "ssl.tls_version_old")
}

func TestTLSVersionOld_TLS13(t *testing.T) {
	p := withTLS(basePage(), &models.TLSInfo{Version: "TLS 1.3"})
	results := (&tlsVersionOld{}).Run(p)
	expectNil(t, results, "ssl.tls_version_old")
}

func TestTLSVersionOld_NoTLS(t *testing.T) {
	results := (&tlsVersionOld{}).Run(basePage())
	expectNil(t, results, "ssl.tls_version_old")
}

// ── ssl.chain_incomplete ────────────────────────────────────────────────────

func TestChainIncomplete_OnlyleafCert(t *testing.T) {
	p := withTLS(basePage(), &models.TLSInfo{ChainLength: 1})
	results := (&chainIncomplete{}).Run(p)
	expectOne(t, results, "ssl.chain_incomplete", models.SeverityWarning)
}

func TestChainIncomplete_FullChain(t *testing.T) {
	p := withTLS(basePage(), &models.TLSInfo{ChainLength: 3})
	results := (&chainIncomplete{}).Run(p)
	expectNil(t, results, "ssl.chain_incomplete")
}

func TestChainIncomplete_TwoCerts(t *testing.T) {
	p := withTLS(basePage(), &models.TLSInfo{ChainLength: 2})
	results := (&chainIncomplete{}).Run(p)
	expectNil(t, results, "ssl.chain_incomplete")
}

func TestChainIncomplete_NoTLS(t *testing.T) {
	results := (&chainIncomplete{}).Run(basePage())
	expectNil(t, results, "ssl.chain_incomplete")
}

// ── ssl.cert_mismatch ───────────────────────────────────────────────────────

func TestCertMismatch_CNMatches(t *testing.T) {
	p := withTLS(basePage(), &models.TLSInfo{
		CertSubject: "example.com",
	})
	results := (&certMismatch{}).Run(p)
	expectNil(t, results, "ssl.cert_mismatch")
}

func TestCertMismatch_SANMatches(t *testing.T) {
	p := withTLS(basePage(), &models.TLSInfo{
		CertSubject:  "other.com",
		CertDNSNames: []string{"other.com", "example.com"},
	})
	results := (&certMismatch{}).Run(p)
	expectNil(t, results, "ssl.cert_mismatch")
}

func TestCertMismatch_WildcardMatches(t *testing.T) {
	p := basePage()
	p.FinalURL = "https://www.example.com"
	withTLS(p, &models.TLSInfo{
		CertSubject:  "*.example.com",
		CertDNSNames: []string{"*.example.com"},
	})
	results := (&certMismatch{}).Run(p)
	expectNil(t, results, "ssl.cert_mismatch")
}

func TestCertMismatch_WildcardDoesNotMatchApex(t *testing.T) {
	p := withTLS(basePage(), &models.TLSInfo{
		CertSubject:  "*.example.com",
		CertDNSNames: []string{"*.example.com"},
	})
	// FinalURL is example.com (apex), wildcard *.example.com should NOT match
	results := (&certMismatch{}).Run(p)
	expectOne(t, results, "ssl.cert_mismatch", models.SeverityError)
}

func TestCertMismatch_WildcardDoesNotMatchDeepSub(t *testing.T) {
	p := basePage()
	p.FinalURL = "https://a.b.example.com"
	withTLS(p, &models.TLSInfo{
		CertSubject:  "*.example.com",
		CertDNSNames: []string{"*.example.com"},
	})
	// *.example.com should not match a.b.example.com (multi-level)
	results := (&certMismatch{}).Run(p)
	expectOne(t, results, "ssl.cert_mismatch", models.SeverityError)
}

func TestCertMismatch_Mismatch(t *testing.T) {
	p := withTLS(basePage(), &models.TLSInfo{
		CertSubject:  "other.com",
		CertDNSNames: []string{"other.com", "www.other.com"},
	})
	results := (&certMismatch{}).Run(p)
	expectOne(t, results, "ssl.cert_mismatch", models.SeverityError)
}

func TestCertMismatch_CaseInsensitive(t *testing.T) {
	p := withTLS(basePage(), &models.TLSInfo{
		CertSubject: "Example.COM",
	})
	results := (&certMismatch{}).Run(p)
	expectNil(t, results, "ssl.cert_mismatch")
}

func TestCertMismatch_NoTLS(t *testing.T) {
	results := (&certMismatch{}).Run(basePage())
	expectNil(t, results, "ssl.cert_mismatch")
}

// ── ssl.hsts_preload_missing ────────────────────────────────────────────────

func TestHSTSPreloadMissing_HasIncludeSubNoPreload(t *testing.T) {
	p := basePage()
	p.Headers["strict-transport-security"] = "max-age=31536000; includeSubDomains"
	results := (&hstsPreloadMissing{}).Run(p)
	expectOne(t, results, "ssl.hsts_preload_missing", models.SeverityNotice)
}

func TestHSTSPreloadMissing_HasBoth(t *testing.T) {
	p := basePage()
	p.Headers["strict-transport-security"] = "max-age=31536000; includeSubDomains; preload"
	results := (&hstsPreloadMissing{}).Run(p)
	expectNil(t, results, "ssl.hsts_preload_missing")
}

func TestHSTSPreloadMissing_NoIncludeSubDomains(t *testing.T) {
	p := basePage()
	p.Headers["strict-transport-security"] = "max-age=31536000"
	results := (&hstsPreloadMissing{}).Run(p)
	// No includeSubDomains — not close to preload-eligible, don't flag
	expectNil(t, results, "ssl.hsts_preload_missing")
}

func TestHSTSPreloadMissing_NoHSTS(t *testing.T) {
	p := basePage()
	results := (&hstsPreloadMissing{}).Run(p)
	expectNil(t, results, "ssl.hsts_preload_missing")
}

func TestHSTSPreloadMissing_HTTPPage(t *testing.T) {
	p := basePage()
	p.FinalURL = "http://example.com"
	p.Headers["strict-transport-security"] = "max-age=31536000; includeSubDomains"
	results := (&hstsPreloadMissing{}).Run(p)
	expectNil(t, results, "ssl.hsts_preload_missing")
}

// ── matchesDomain unit tests ────────────────────────────────────────────────

func TestMatchesDomain(t *testing.T) {
	tests := []struct {
		host    string
		pattern string
		want    bool
	}{
		{"example.com", "example.com", true},
		{"Example.COM", "example.com", true},
		{"www.example.com", "*.example.com", true},
		{"sub.example.com", "*.example.com", true},
		{"example.com", "*.example.com", false},         // apex doesn't match wildcard
		{"a.b.example.com", "*.example.com", false},     // multi-level doesn't match
		{"example.com", "other.com", false},
		{"example.com", "*.other.com", false},
	}
	for _, tt := range tests {
		t.Run(tt.host+"_vs_"+tt.pattern, func(t *testing.T) {
			got := matchesDomain(tt.host, tt.pattern)
			if got != tt.want {
				t.Errorf("matchesDomain(%q, %q) = %v, want %v", tt.host, tt.pattern, got, tt.want)
			}
		})
	}
}

// ── PageChecks registration ─────────────────────────────────────────────────

func TestPageChecks_Returns6(t *testing.T) {
	checks := PageChecks()
	if len(checks) != 6 {
		t.Errorf("expected 6 SSL checks, got %d", len(checks))
	}
}
