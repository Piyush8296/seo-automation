package social

import (
	"strings"

	"github.com/cars24/seo-automation/internal/models"
)

// PageChecks returns per-page social/OG/Twitter checks.
func PageChecks() []models.PageCheck {
	return []models.PageCheck{
		&ogTitleMissing{},
		&ogDescriptionMissing{},
		&ogImageMissing{},
		&ogURLMissing{},
		&ogURLMismatchCanonical{},
		&twitterCardMissing{},
		&twitterTitleMissing{},
		&twitterImageMissing{},
	}
}

type ogTitleMissing struct{}

func (c *ogTitleMissing) Run(p *models.PageData) []models.CheckResult {
	if p.OGTags["og:title"] == "" {
		return []models.CheckResult{{
			ID:       "og.title.missing",
			Category: "Social",
			Severity: models.SeverityWarning,
			Message:  "Missing og:title meta tag",
			URL:      p.URL,
		}}
	}
	return nil
}

type ogDescriptionMissing struct{}

func (c *ogDescriptionMissing) Run(p *models.PageData) []models.CheckResult {
	if p.OGTags["og:description"] == "" {
		return []models.CheckResult{{
			ID:       "og.description.missing",
			Category: "Social",
			Severity: models.SeverityWarning,
			Message:  "Missing og:description meta tag",
			URL:      p.URL,
		}}
	}
	return nil
}

type ogImageMissing struct{}

func (c *ogImageMissing) Run(p *models.PageData) []models.CheckResult {
	if p.OGTags["og:image"] == "" {
		return []models.CheckResult{{
			ID:       "og.image.missing",
			Category: "Social",
			Severity: models.SeverityWarning,
			Message:  "Missing og:image meta tag",
			URL:      p.URL,
		}}
	}
	return nil
}

type ogURLMissing struct{}

func (c *ogURLMissing) Run(p *models.PageData) []models.CheckResult {
	if p.OGTags["og:url"] == "" {
		return []models.CheckResult{{
			ID:       "og.url.missing",
			Category: "Social",
			Severity: models.SeverityNotice,
			Message:  "Missing og:url meta tag",
			URL:      p.URL,
		}}
	}
	return nil
}

type ogURLMismatchCanonical struct{}

func (c *ogURLMismatchCanonical) Run(p *models.PageData) []models.CheckResult {
	ogURL := strings.TrimRight(strings.TrimSpace(p.OGTags["og:url"]), "/")
	canonical := strings.TrimRight(strings.TrimSpace(p.Canonical), "/")
	if ogURL != "" && canonical != "" && !strings.EqualFold(ogURL, canonical) {
		return []models.CheckResult{{
			ID:       "og.url.mismatch_canonical",
			Category: "Social",
			Severity: models.SeverityWarning,
			Message:  "og:url doesn't match canonical URL",
			URL:      p.URL,
			Details:  "og:url=" + ogURL + " canonical=" + canonical,
		}}
	}
	return nil
}

type twitterCardMissing struct{}

func (c *twitterCardMissing) Run(p *models.PageData) []models.CheckResult {
	if p.TwitterTags["twitter:card"] == "" {
		return []models.CheckResult{{
			ID:       "twitter.card.missing",
			Category: "Social",
			Severity: models.SeverityNotice,
			Message:  "Missing twitter:card meta tag",
			URL:      p.URL,
		}}
	}
	return nil
}

type twitterTitleMissing struct{}

func (c *twitterTitleMissing) Run(p *models.PageData) []models.CheckResult {
	if p.TwitterTags["twitter:title"] == "" && p.OGTags["og:title"] == "" {
		return []models.CheckResult{{
			ID:       "twitter.title.missing",
			Category: "Social",
			Severity: models.SeverityNotice,
			Message:  "Missing twitter:title (and no og:title fallback)",
			URL:      p.URL,
		}}
	}
	return nil
}

type twitterImageMissing struct{}

func (c *twitterImageMissing) Run(p *models.PageData) []models.CheckResult {
	if p.TwitterTags["twitter:image"] == "" && p.OGTags["og:image"] == "" {
		return []models.CheckResult{{
			ID:       "twitter.image.missing",
			Category: "Social",
			Severity: models.SeverityNotice,
			Message:  "Missing twitter:image (and no og:image fallback)",
			URL:      p.URL,
		}}
	}
	return nil
}
