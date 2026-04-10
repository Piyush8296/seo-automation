// Package pagination checks for paginated series SEO issues.
// A critical area often missed: paginated pages (page 2, 3...) need careful
// canonical strategy. Google deprecated rel=prev/next in 2019 but canonical
// handling of paginated series remains one of the top technical SEO issues.
package pagination

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"

	"github.com/cars24/seo-automation/internal/models"
)

var paginationPatterns = []*regexp.Regexp{
	regexp.MustCompile(`[?&](page|p|pg|paged|start|offset|from)\s*=\s*\d+`),
	regexp.MustCompile(`/page/\d+/?$`),
	regexp.MustCompile(`/p/\d+/?$`),
	regexp.MustCompile(`-\d+/?$`), // trailing number segment
}

// isPaginatedURL returns true if the URL looks like a paginated page.
func isPaginatedURL(rawURL string) bool {
	for _, re := range paginationPatterns {
		if re.MatchString(rawURL) {
			return true
		}
	}
	return false
}

// getPageNum extracts the page number from a URL (returns 0 if not found).
func getPageNum(rawURL string) int {
	for _, re := range paginationPatterns[:2] { // param-based
		m := re.FindString(rawURL)
		if m == "" {
			continue
		}
		// extract number
		numRe := regexp.MustCompile(`\d+`)
		n := numRe.FindString(m)
		if n == "" {
			continue
		}
		num := 0
		for _, c := range n {
			num = num*10 + int(c-'0')
		}
		if num > 0 {
			return num
		}
	}
	return 0
}

// PageChecks returns per-page pagination SEO checks.
func PageChecks() []models.PageCheck {
	return []models.PageCheck{
		&paginatedSelfCanonical{},
		&paginatedCanonicalToFirstPage{},
		&paginatedThinContent{},
		&paginatedNoIndex{},
		&paginatedMissingTitle{},
	}
}

// SiteChecks returns site-wide pagination checks.
func SiteChecks() []models.SiteCheck {
	return []models.SiteCheck{
		&paginatedSeriesCanonicalInconsistency{},
	}
}

// paginatedSelfCanonical: paginated pages should have self-referencing canonical
// (NOT point to page 1) — best practice since Google deprecated rel=prev/next.
type paginatedSelfCanonical struct{}

func (c *paginatedSelfCanonical) Run(p *models.PageData) []models.CheckResult {
	if !isPaginatedURL(p.URL) {
		return nil
	}
	pageNum := getPageNum(p.URL)
	if pageNum <= 1 {
		return nil
	}
	// Check if canonical points to page 1 (bad practice: duplicate content signal)
	if p.Canonical != "" {
		if !isPaginatedURL(p.Canonical) {
			// Canonical points to a non-paginated URL (likely page 1)
			return []models.CheckResult{{
				ID:       "pagination.canonical_first_page",
				Category: "Pagination",
				Severity: models.SeverityWarning,
				Message:  fmt.Sprintf("Paginated page %d has canonical pointing to non-paginated URL (consolidates to page 1)", pageNum),
				URL:      p.URL,
				Details:  "canonical=" + p.Canonical + " — consider self-canonical instead",
			}}
		}
	}
	return nil
}

// paginatedCanonicalToFirstPage: explicitly flags pages 2+ that self-canonical
// (which is correct behavior and should be noted as passing).
type paginatedCanonicalToFirstPage struct{}

func (c *paginatedCanonicalToFirstPage) Run(p *models.PageData) []models.CheckResult {
	if !isPaginatedURL(p.URL) {
		return nil
	}
	pageNum := getPageNum(p.URL)
	if pageNum <= 1 {
		return nil
	}
	// If canonical is set but is empty or missing, flag it
	if p.Canonical == "" {
		return []models.CheckResult{{
			ID:       "pagination.missing_canonical",
			Category: "Pagination",
			Severity: models.SeverityWarning,
			Message:  fmt.Sprintf("Paginated page %d has no canonical tag", pageNum),
			URL:      p.URL,
			Details:  "Add self-referencing canonical or canonical to page 1 as appropriate",
		}}
	}
	return nil
}

// paginatedThinContent: page 2+ with very low word count vs typical listing pages.
type paginatedThinContent struct{}

func (c *paginatedThinContent) Run(p *models.PageData) []models.CheckResult {
	if !isPaginatedURL(p.URL) {
		return nil
	}
	pageNum := getPageNum(p.URL)
	if pageNum <= 1 || p.WordCount == 0 {
		return nil
	}
	if p.WordCount < 50 {
		return []models.CheckResult{{
			ID:       "pagination.thin_content",
			Category: "Pagination",
			Severity: models.SeverityWarning,
			Message:  fmt.Sprintf("Paginated page %d has very thin content (%d words)", pageNum, p.WordCount),
			URL:      p.URL,
		}}
	}
	return nil
}

// paginatedNoIndex: paginated pages with noindex may prevent crawl of later pages.
type paginatedNoIndex struct{}

func (c *paginatedNoIndex) Run(p *models.PageData) []models.CheckResult {
	if !isPaginatedURL(p.URL) {
		return nil
	}
	if strings.Contains(p.RobotsTag, "noindex") {
		return []models.CheckResult{{
			ID:       "pagination.noindex",
			Category: "Pagination",
			Severity: models.SeverityWarning,
			Message:  "Paginated page has noindex — may prevent crawl of subsequent pages",
			URL:      p.URL,
		}}
	}
	return nil
}

// paginatedMissingTitle: paginated pages should have unique titles.
type paginatedMissingTitle struct{}

func (c *paginatedMissingTitle) Run(p *models.PageData) []models.CheckResult {
	if !isPaginatedURL(p.URL) {
		return nil
	}
	pageNum := getPageNum(p.URL)
	if pageNum <= 1 {
		return nil
	}
	// Title should contain page number or "page N" to be unique
	titleLower := strings.ToLower(p.Title)
	numStr := fmt.Sprintf("%d", pageNum)
	if !strings.Contains(titleLower, "page "+numStr) &&
		!strings.Contains(titleLower, "page"+numStr) &&
		!strings.Contains(p.Title, numStr) {
		return []models.CheckResult{{
			ID:       "pagination.title_not_unique",
			Category: "Pagination",
			Severity: models.SeverityNotice,
			Message:  fmt.Sprintf("Paginated page %d title doesn't indicate page number", pageNum),
			URL:      p.URL,
			Details:  "Add '— Page " + numStr + "' to title for uniqueness",
		}}
	}
	return nil
}

// paginatedSeriesCanonicalInconsistency: site-wide check for inconsistent
// canonical strategy across a paginated series.
type paginatedSeriesCanonicalInconsistency struct{}

func (c *paginatedSeriesCanonicalInconsistency) Run(pages []*models.PageData) []models.CheckResult {
	// Group paginated pages by their base URL (strip page param)
	type seriesInfo struct {
		selfCanon  int
		firstCanon int
		noCanon    int
	}
	series := map[string]*seriesInfo{}

	for _, p := range pages {
		if !isPaginatedURL(p.URL) {
			continue
		}
		pageNum := getPageNum(p.URL)
		if pageNum <= 1 {
			continue
		}

		// Determine series base URL
		base := stripPageParam(p.URL)
		if series[base] == nil {
			series[base] = &seriesInfo{}
		}
		s := series[base]

		if p.Canonical == "" {
			s.noCanon++
		} else if !isPaginatedURL(p.Canonical) {
			s.firstCanon++
		} else {
			s.selfCanon++
		}
	}

	var results []models.CheckResult
	for base, s := range series {
		total := s.selfCanon + s.firstCanon + s.noCanon
		if total < 2 {
			continue
		}
		// Mixed strategy: some pointing to first page, some self-canonical
		if s.selfCanon > 0 && s.firstCanon > 0 {
			results = append(results, models.CheckResult{
				ID:       "pagination.inconsistent_canonical_strategy",
				Category: "Pagination",
				Severity: models.SeverityError,
				Message: fmt.Sprintf("Inconsistent canonical strategy in paginated series: %d self-canonical, %d point to page 1",
					s.selfCanon, s.firstCanon),
				URL:     base,
				Details: "Pick one strategy: either all pages self-canonical OR all pages canonical to page 1",
			})
		}
	}
	return results
}

func stripPageParam(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return rawURL
	}
	q := u.Query()
	for _, p := range []string{"page", "p", "pg", "paged", "start", "offset", "from"} {
		q.Del(p)
	}
	u.RawQuery = q.Encode()
	// Strip trailing /page/N
	u.Path = regexp.MustCompile(`/page/\d+/?$`).ReplaceAllString(u.Path, "")
	return u.String()
}
