// Package eeat implements E-E-A-T (Experience, Expertise, Authoritativeness, Trustworthiness) checks.
// E-E-A-T is Google's quality rater framework and has become increasingly important
// as a ranking factor, especially after the Helpful Content updates (2022-2023).
// While E-E-A-T isn't directly measurable via HTML alone, many signals are detectable.
package eeat

import (
	"strings"

	"github.com/cars24/seo-automation/internal/models"
)

// PageChecks returns per-page E-E-A-T signal checks.
func PageChecks() []models.PageCheck {
	return []models.PageCheck{
		&authorInfoMissing{},
		&reviewsDatesMissing{},
		&aboutPageMissing{},
		&contactPageMissing{},
		&privacyPolicyMissing{},
		&breadcrumbsMissing{},
	}
}

// SiteChecks returns site-wide E-E-A-T checks.
func SiteChecks() []models.SiteCheck {
	return []models.SiteCheck{
		&siteHasAboutPage{},
		&siteHasContactPage{},
	}
}

// authorInfoMissing: Article/blog pages without author information reduce
// Expertise and Experience signals. Critical for YMYL content.
type authorInfoMissing struct{}

func (c *authorInfoMissing) Run(p *models.PageData) []models.CheckResult {
	// Only check article-type pages
	hasArticleSchema := false
	for _, raw := range p.SchemaJSONRaw {
		rawLower := strings.ToLower(raw)
		if strings.Contains(rawLower, `"article"`) ||
			strings.Contains(rawLower, `"blogposting"`) ||
			strings.Contains(rawLower, `"newsarticle"`) {
			hasArticleSchema = true
			break
		}
	}

	if !hasArticleSchema && p.WordCount < 300 {
		return nil // Not an article page
	}

	// Check for author byline signals in HTML
	htmlLower := strings.ToLower(p.RawHTML)
	hasAuthor := strings.Contains(htmlLower, `class="author"`) ||
		strings.Contains(htmlLower, `rel="author"`) ||
		strings.Contains(htmlLower, `itemprop="author"`) ||
		strings.Contains(htmlLower, `class="byline"`) ||
		strings.Contains(htmlLower, "written by") ||
		strings.Contains(htmlLower, "author:") ||
		strings.Contains(htmlLower, "by ")

	if !hasAuthor && hasArticleSchema {
		return []models.CheckResult{{
			ID:       "eeat.author_info.missing",
			Category: "E-E-A-T",
			Severity: models.SeverityWarning,
			Message:  "Article page missing author attribution (E-E-A-T experience/expertise signal)",
			URL:      p.URL,
			Details:  "Add author name, bio, credentials, and profile link. Use Article schema with author property.",
		}}
	}
	return nil
}

// reviewsDatesMissing: Review/article pages without publication dates reduce
// freshness signals. Google shows dates in snippets which improves CTR.
type reviewsDatesMissing struct{}

func (c *reviewsDatesMissing) Run(p *models.PageData) []models.CheckResult {
	for _, raw := range p.SchemaJSONRaw {
		rawLower := strings.ToLower(raw)
		if (strings.Contains(rawLower, `"article"`) || strings.Contains(rawLower, `"review"`)) &&
			!strings.Contains(rawLower, "datepublished") {
			return []models.CheckResult{{
				ID:       "eeat.dates.missing",
				Category: "E-E-A-T",
				Severity: models.SeverityWarning,
				Message:  "Article/Review schema missing datePublished (reduces freshness signal)",
				URL:      p.URL,
				Details:  "Add datePublished and dateModified to Article schema. Google uses these for snippet dates.",
			}}
		}
	}
	return nil
}

// aboutPageMissing: site-level check, but per-page we check if this IS an
// about page and validate it has appropriate signals.
type aboutPageMissing struct{}

func (c *aboutPageMissing) Run(p *models.PageData) []models.CheckResult {
	urlLower := strings.ToLower(p.URL)
	if !strings.Contains(urlLower, "/about") {
		return nil
	}
	// It's an about page — check it has substantial content
	if p.WordCount < 100 {
		return []models.CheckResult{{
			ID:       "eeat.about_page.thin",
			Category: "E-E-A-T",
			Severity: models.SeverityWarning,
			Message:  "About page has thin content — missed opportunity for E-E-A-T signals",
			URL:      p.URL,
			Details:  "Include: company history, team members, credentials, awards, press mentions, and mission statement.",
		}}
	}
	return nil
}

// contactPageMissing: per-page validation of contact pages.
type contactPageMissing struct{}

func (c *contactPageMissing) Run(p *models.PageData) []models.CheckResult {
	urlLower := strings.ToLower(p.URL)
	if !strings.Contains(urlLower, "/contact") {
		return nil
	}
	htmlLower := strings.ToLower(p.RawHTML)
	// Check for actual contact methods
	hasEmail := strings.Contains(htmlLower, "@") && strings.Contains(htmlLower, ".")
	hasPhone := strings.Contains(htmlLower, "tel:") || strings.Contains(htmlLower, "phone")
	hasForm := strings.Contains(htmlLower, "<form")
	hasAddress := strings.Contains(htmlLower, "address") || strings.Contains(htmlLower, "street")

	signals := 0
	if hasEmail {
		signals++
	}
	if hasPhone {
		signals++
	}
	if hasForm {
		signals++
	}
	if hasAddress {
		signals++
	}

	if signals < 2 {
		return []models.CheckResult{{
			ID:       "eeat.contact_page.insufficient",
			Category: "E-E-A-T",
			Severity: models.SeverityNotice,
			Message:  "Contact page has limited contact information (low trust signal)",
			URL:      p.URL,
			Details:  "Include multiple contact methods: email, phone, address, and contact form for maximum trust signals.",
		}}
	}
	return nil
}

// privacyPolicyMissing: per-page check of privacy policy page.
type privacyPolicyMissing struct{}

func (c *privacyPolicyMissing) Run(p *models.PageData) []models.CheckResult {
	urlLower := strings.ToLower(p.URL)
	if !strings.Contains(urlLower, "/privacy") {
		return nil
	}
	if p.WordCount < 200 {
		return []models.CheckResult{{
			ID:       "eeat.privacy_policy.thin",
			Category: "E-E-A-T",
			Severity: models.SeverityWarning,
			Message:  "Privacy policy page has thin content — YMYL trust signal",
			URL:      p.URL,
		}}
	}
	return nil
}

// breadcrumbsMissing: breadcrumbs help users understand site hierarchy and
// appear in search results as rich snippets (direct CTR improvement).
type breadcrumbsMissing struct{}

func (c *breadcrumbsMissing) Run(p *models.PageData) []models.CheckResult {
	if p.Depth < 2 {
		return nil
	}
	// Check for breadcrumb schema
	hasBreadcrumb := false
	for _, raw := range p.SchemaJSONRaw {
		if strings.Contains(strings.ToLower(raw), "breadcrumblist") {
			hasBreadcrumb = true
			break
		}
	}
	// Check for breadcrumb HTML patterns
	htmlLower := strings.ToLower(p.RawHTML)
	hasBreadcrumbHTML := strings.Contains(htmlLower, `itemtype="https://schema.org/breadcrumblist"`) ||
		strings.Contains(htmlLower, `class="breadcrumb"`) ||
		strings.Contains(htmlLower, `class="breadcrumbs"`) ||
		strings.Contains(htmlLower, `aria-label="breadcrumb"`)

	if !hasBreadcrumb && !hasBreadcrumbHTML {
		return []models.CheckResult{{
			ID:       "eeat.breadcrumbs.missing",
			Category: "E-E-A-T",
			Severity: models.SeverityNotice,
			Message:  "Deep page missing breadcrumb navigation (schema + HTML)",
			URL:      p.URL,
			Details:  "Add BreadcrumbList schema and visible breadcrumb HTML. Improves rich snippets and user experience.",
		}}
	}
	return nil
}

// Site-wide checks

type siteHasAboutPage struct{}

func (c *siteHasAboutPage) Run(pages []*models.PageData) []models.CheckResult {
	for _, p := range pages {
		if strings.Contains(strings.ToLower(p.URL), "/about") {
			return nil
		}
	}
	if len(pages) == 0 {
		return nil
	}
	return []models.CheckResult{{
		ID:       "eeat.about_page.missing",
		Category: "E-E-A-T",
		Severity: models.SeverityWarning,
		Message:  "Site has no About page (critical E-E-A-T trust signal)",
		URL:      pages[0].URL,
		Details:  "Create /about page with team info, company history, mission, and credentials.",
	}}
}

type siteHasContactPage struct{}

func (c *siteHasContactPage) Run(pages []*models.PageData) []models.CheckResult {
	for _, p := range pages {
		if strings.Contains(strings.ToLower(p.URL), "/contact") {
			return nil
		}
	}
	if len(pages) == 0 {
		return nil
	}
	return []models.CheckResult{{
		ID:       "eeat.contact_page.missing",
		Category: "E-E-A-T",
		Severity: models.SeverityWarning,
		Message:  "Site has no Contact page (reduces Trustworthiness signal)",
		URL:      pages[0].URL,
		Details:  "Create /contact page. Google quality raters look for clear contact information.",
	}}
}
