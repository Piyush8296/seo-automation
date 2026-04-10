package canonical

import (
	"net/url"
	"strings"

	"github.com/cars24/seo-automation/internal/models"
)

// PageChecks returns per-page canonical checks.
func PageChecks() []models.PageCheck {
	return []models.PageCheck{
		&canonicalMissing{},
		&canonicalNotAbsolute{},
		&canonicalInsecure{},
		&canonicalPointsElsewhere{},
		&canonicalConflictOGURL{},
	}
}

type canonicalMissing struct{}

func (c *canonicalMissing) Run(p *models.PageData) []models.CheckResult {
	if strings.TrimSpace(p.Canonical) == "" {
		return []models.CheckResult{{
			ID:       "canonical.missing",
			Category: "Canonical",
			Severity: models.SeverityWarning,
			Message:  "Page has no canonical URL tag",
			URL:      p.URL,
		}}
	}
	return nil
}

type canonicalNotAbsolute struct{}

func (c *canonicalNotAbsolute) Run(p *models.PageData) []models.CheckResult {
	can := strings.TrimSpace(p.Canonical)
	if can == "" {
		return nil
	}
	parsed, err := url.Parse(can)
	if err != nil || !parsed.IsAbs() {
		return []models.CheckResult{{
			ID:       "canonical.not_absolute",
			Category: "Canonical",
			Severity: models.SeverityError,
			Message:  "Canonical URL is not an absolute URL",
			URL:      p.URL,
			Details:  can,
		}}
	}
	return nil
}

type canonicalInsecure struct{}

func (c *canonicalInsecure) Run(p *models.PageData) []models.CheckResult {
	can := strings.TrimSpace(p.Canonical)
	if strings.HasPrefix(can, "http://") {
		return []models.CheckResult{{
			ID:       "canonical.insecure",
			Category: "Canonical",
			Severity: models.SeverityWarning,
			Message:  "Canonical URL uses HTTP instead of HTTPS",
			URL:      p.URL,
			Details:  can,
		}}
	}
	return nil
}

type canonicalPointsElsewhere struct{}

func (c *canonicalPointsElsewhere) Run(p *models.PageData) []models.CheckResult {
	can := strings.TrimSpace(p.Canonical)
	if can == "" {
		return nil
	}
	finalURL := strings.TrimRight(p.FinalURL, "/")
	canClean := strings.TrimRight(can, "/")
	if !strings.EqualFold(finalURL, canClean) {
		return []models.CheckResult{{
			ID:       "canonical.points_elsewhere",
			Category: "Canonical",
			Severity: models.SeverityNotice,
			Message:  "Canonical URL points to a different URL than the current page",
			URL:      p.URL,
			Details:  can,
		}}
	}
	return nil
}

type canonicalConflictOGURL struct{}

func (c *canonicalConflictOGURL) Run(p *models.PageData) []models.CheckResult {
	can := strings.TrimSpace(p.Canonical)
	ogURL := strings.TrimSpace(p.OGTags["og:url"])
	if can == "" || ogURL == "" {
		return nil
	}
	if !strings.EqualFold(strings.TrimRight(can, "/"), strings.TrimRight(ogURL, "/")) {
		return []models.CheckResult{{
			ID:       "canonical.conflict_og_url",
			Category: "Canonical",
			Severity: models.SeverityWarning,
			Message:  "Canonical URL conflicts with og:url",
			URL:      p.URL,
			Details:  "canonical=" + can + " og:url=" + ogURL,
		}}
	}
	return nil
}
