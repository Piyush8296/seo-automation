package canonical

import (
	"fmt"
	"net/url"
	"sort"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/cars24/seo-automation/internal/models"
)

// PageChecks returns per-page canonical checks.
func PageChecks() []models.PageCheck {
	return []models.PageCheck{
		&canonicalMissing{},
		&canonicalMultiple{},
		&canonicalNotAbsolute{},
		&canonicalInsecure{},
		&canonicalInBody{},
		&canonicalHasFragment{},
		&canonicalHeaderMismatch{},
		&canonicalPointsElsewhere{},
		&canonicalConflictOGURL{},
		&canonicalParamsSelfReference{},
		&canonicalCountryFolderMismatch{},
	}
}

// SiteChecks returns site-wide canonical relationship checks.
func SiteChecks() []models.SiteCheck {
	return []models.SiteCheck{
		&canonicalTargetNon200{},
		&canonicalLoop{},
		&canonicalChain{},
		&canonicalWWWVariant{},
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

type canonicalMultiple struct{}

func (c *canonicalMultiple) Run(p *models.PageData) []models.CheckResult {
	links := canonicalLinks(p)
	if len(links) <= 1 {
		return nil
	}
	return []models.CheckResult{{
		ID:       "canonical.multiple",
		Category: "Canonical",
		Severity: models.SeverityError,
		Message:  "Page has more than one canonical tag",
		URL:      p.URL,
		Details:  strings.Join(canonicalLinkHrefs(links), " | "),
	}}
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

type canonicalInBody struct{}

func (c *canonicalInBody) Run(p *models.PageData) []models.CheckResult {
	for _, link := range canonicalLinks(p) {
		if !link.InHead {
			return []models.CheckResult{{
				ID:       "canonical.in_body",
				Category: "Canonical",
				Severity: models.SeverityError,
				Message:  "Canonical tag is outside the document head",
				URL:      p.URL,
				Details:  link.Href,
			}}
		}
	}
	return nil
}

type canonicalHasFragment struct{}

func (c *canonicalHasFragment) Run(p *models.PageData) []models.CheckResult {
	can := strings.TrimSpace(p.Canonical)
	if can == "" {
		return nil
	}
	parsed, err := url.Parse(can)
	if err != nil || parsed.Fragment == "" {
		return nil
	}
	return []models.CheckResult{{
		ID:       "canonical.has_fragment",
		Category: "Canonical",
		Severity: models.SeverityWarning,
		Message:  "Canonical URL contains a fragment",
		URL:      p.URL,
		Details:  can,
	}}
}

type canonicalHeaderMismatch struct{}

func (c *canonicalHeaderMismatch) Run(p *models.PageData) []models.CheckResult {
	htmlCanonical := strings.TrimSpace(p.Canonical)
	headerCanonical := canonicalFromLinkHeader(p.Headers["link"])
	if htmlCanonical == "" || headerCanonical == "" {
		return nil
	}
	if sameCanonicalURL(htmlCanonical, headerCanonical, p.FinalURL) {
		return nil
	}
	return []models.CheckResult{{
		ID:       "canonical.header_mismatch",
		Category: "Canonical",
		Severity: models.SeverityError,
		Message:  "HTTP Link canonical does not match HTML canonical",
		URL:      p.URL,
		Details:  "html=" + htmlCanonical + " header=" + headerCanonical,
	}}
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

type canonicalParamsSelfReference struct{}

func (c *canonicalParamsSelfReference) Run(p *models.PageData) []models.CheckResult {
	can := strings.TrimSpace(p.Canonical)
	if can == "" {
		return nil
	}
	pageURL, err := url.Parse(p.FinalURL)
	if err != nil || pageURL.RawQuery == "" {
		return nil
	}
	canonicalURL, err := resolveCanonical(can, p.FinalURL)
	if err != nil || canonicalURL.RawQuery == "" {
		return nil
	}
	if canonicalKey(canonicalURL.String()) != canonicalKey(p.FinalURL) {
		return nil
	}
	return []models.CheckResult{{
		ID:       "canonical.params_self_reference",
		Category: "Canonical",
		Severity: models.SeverityWarning,
		Message:  "Parameterized URL is self-canonical instead of consolidating to a clean URL",
		URL:      p.URL,
		Details:  can,
	}}
}

type canonicalCountryFolderMismatch struct{}

func (c *canonicalCountryFolderMismatch) Run(p *models.PageData) []models.CheckResult {
	can := strings.TrimSpace(p.Canonical)
	if can == "" {
		return nil
	}
	pageURL, err := url.Parse(p.FinalURL)
	if err != nil {
		return nil
	}
	canonicalURL, err := resolveCanonical(can, p.FinalURL)
	if err != nil {
		return nil
	}
	pageFolder := firstCountryLikePathSegment(pageURL.Path)
	canonicalFolder := firstCountryLikePathSegment(canonicalURL.Path)
	if pageFolder == "" || canonicalFolder == "" || strings.EqualFold(pageFolder, canonicalFolder) {
		return nil
	}
	return []models.CheckResult{{
		ID:       "canonical.country_folder_mismatch",
		Category: "Canonical",
		Severity: models.SeverityWarning,
		Message:  "Canonical URL points to a different country or locale folder",
		URL:      p.URL,
		Details:  "page folder=" + pageFolder + " canonical folder=" + canonicalFolder,
	}}
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

type canonicalTargetNon200 struct{}

func (c *canonicalTargetNon200) Run(pages []*models.PageData) []models.CheckResult {
	pageByURL := canonicalPageMap(pages)
	out := []models.CheckResult{}
	for _, page := range pages {
		target := resolvedCanonicalString(page)
		if target == "" {
			continue
		}
		targetPage := pageByURL[canonicalKey(target)]
		if targetPage == nil {
			continue
		}
		if targetPage.StatusCode != httpStatusOK {
			out = append(out, models.CheckResult{
				ID:       "canonical.target_non_200",
				Category: "Canonical",
				Severity: models.SeverityError,
				Message:  "Canonical target does not return HTTP 200 in the crawl",
				URL:      page.URL,
				Details:  fmt.Sprintf("canonical=%s status=%d", target, targetPage.StatusCode),
			})
		}
	}
	return out
}

type canonicalLoop struct{}

func (c *canonicalLoop) Run(pages []*models.PageData) []models.CheckResult {
	pageByURL := canonicalPageMap(pages)
	out := []models.CheckResult{}
	for _, page := range pages {
		target := resolvedCanonicalString(page)
		if target == "" || sameCanonicalURL(target, page.FinalURL, page.FinalURL) {
			continue
		}
		targetPage := pageByURL[canonicalKey(target)]
		if targetPage == nil {
			continue
		}
		secondTarget := resolvedCanonicalString(targetPage)
		if secondTarget != "" && sameCanonicalURL(secondTarget, page.FinalURL, targetPage.FinalURL) {
			out = append(out, models.CheckResult{
				ID:       "canonical.loop",
				Category: "Canonical",
				Severity: models.SeverityError,
				Message:  "Canonical loop detected",
				URL:      page.URL,
				Details:  fmt.Sprintf("%s -> %s -> %s", page.FinalURL, target, secondTarget),
			})
		}
	}
	return out
}

type canonicalChain struct{}

func (c *canonicalChain) Run(pages []*models.PageData) []models.CheckResult {
	pageByURL := canonicalPageMap(pages)
	out := []models.CheckResult{}
	for _, page := range pages {
		target := resolvedCanonicalString(page)
		if target == "" || sameCanonicalURL(target, page.FinalURL, page.FinalURL) {
			continue
		}
		targetPage := pageByURL[canonicalKey(target)]
		if targetPage == nil {
			continue
		}
		secondTarget := resolvedCanonicalString(targetPage)
		if secondTarget == "" || sameCanonicalURL(secondTarget, target, targetPage.FinalURL) || sameCanonicalURL(secondTarget, page.FinalURL, targetPage.FinalURL) {
			continue
		}
		out = append(out, models.CheckResult{
			ID:       "canonical.chain",
			Category: "Canonical",
			Severity: models.SeverityWarning,
			Message:  "Canonical target points to another canonical target",
			URL:      page.URL,
			Details:  fmt.Sprintf("%s -> %s -> %s", page.FinalURL, target, secondTarget),
		})
	}
	return out
}

type canonicalWWWVariant struct{}

func (c *canonicalWWWVariant) Run(pages []*models.PageData) []models.CheckResult {
	hostVariants := map[string]map[string]bool{}
	pageMismatches := []string{}
	for _, page := range pages {
		target := resolvedCanonicalString(page)
		if target == "" {
			continue
		}
		pageHost := hostLower(page.FinalURL)
		targetHost := hostLower(target)
		if pageHost == "" || targetHost == "" {
			continue
		}
		base := stripWWW(pageHost)
		if base != stripWWW(targetHost) {
			continue
		}
		if hostVariants[base] == nil {
			hostVariants[base] = map[string]bool{}
		}
		hostVariants[base][targetHost] = true
		if pageHost != targetHost {
			pageMismatches = append(pageMismatches, fmt.Sprintf("%s canonical=%s", page.FinalURL, target))
		}
	}
	for _, variants := range hostVariants {
		if len(variants) > 1 {
			return []models.CheckResult{{
				ID:       "canonical.www_variant",
				Category: "Canonical",
				Severity: models.SeverityWarning,
				Message:  "Canonical host alternates between www and non-www variants",
				URL:      "(site-wide)",
				Details:  strings.Join(firstN(sortedKeys(variants), 6), ", "),
			}}
		}
	}
	if len(pageMismatches) > 0 {
		return []models.CheckResult{{
			ID:       "canonical.www_variant",
			Category: "Canonical",
			Severity: models.SeverityNotice,
			Message:  "Some crawled pages canonicalize to the opposite www/non-www host variant",
			URL:      "(site-wide)",
			Details:  strings.Join(firstN(pageMismatches, 8), " | "),
		}}
	}
	return nil
}

const httpStatusOK = 200

type canonicalLink struct {
	Href   string
	InHead bool
}

func canonicalLinks(p *models.PageData) []canonicalLink {
	if strings.TrimSpace(p.RawHTML) == "" {
		if strings.TrimSpace(p.Canonical) == "" {
			return nil
		}
		return []canonicalLink{{Href: strings.TrimSpace(p.Canonical), InHead: true}}
	}
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(p.RawHTML))
	if err != nil {
		return nil
	}
	out := []canonicalLink{}
	doc.Find("link[rel]").Each(func(_ int, s *goquery.Selection) {
		if !relContains(s.AttrOr("rel", ""), "canonical") {
			return
		}
		out = append(out, canonicalLink{
			Href:   strings.TrimSpace(s.AttrOr("href", "")),
			InHead: s.ParentsFiltered("head").Length() > 0,
		})
	})
	return out
}

func canonicalLinkHrefs(links []canonicalLink) []string {
	out := make([]string, 0, len(links))
	for _, link := range links {
		out = append(out, link.Href)
	}
	return out
}

func canonicalFromLinkHeader(header string) string {
	for _, part := range strings.Split(header, ",") {
		lower := strings.ToLower(part)
		if !strings.Contains(lower, "rel=\"canonical\"") && !strings.Contains(lower, "rel=canonical") {
			continue
		}
		part = strings.TrimSpace(part)
		if start := strings.Index(part, "<"); start >= 0 {
			if end := strings.Index(part[start+1:], ">"); end >= 0 {
				return strings.TrimSpace(part[start+1 : start+1+end])
			}
		}
	}
	return ""
}

func resolvedCanonicalString(p *models.PageData) string {
	can := strings.TrimSpace(p.Canonical)
	if can == "" {
		return ""
	}
	resolved, err := resolveCanonical(can, p.FinalURL)
	if err != nil {
		return ""
	}
	resolved.Fragment = ""
	return resolved.String()
}

func resolveCanonical(canonicalURL, baseURL string) (*url.URL, error) {
	base, err := url.Parse(baseURL)
	if err != nil {
		return nil, err
	}
	ref, err := url.Parse(strings.TrimSpace(canonicalURL))
	if err != nil {
		return nil, err
	}
	resolved := base.ResolveReference(ref)
	resolved.Scheme = strings.ToLower(resolved.Scheme)
	resolved.Host = strings.ToLower(resolved.Host)
	return resolved, nil
}

func sameCanonicalURL(a, b, base string) bool {
	aURL, err := resolveCanonical(a, base)
	if err != nil {
		return false
	}
	bURL, err := resolveCanonical(b, base)
	if err != nil {
		return false
	}
	return canonicalKey(aURL.String()) == canonicalKey(bURL.String())
}

func canonicalPageMap(pages []*models.PageData) map[string]*models.PageData {
	out := map[string]*models.PageData{}
	for _, page := range pages {
		for _, raw := range []string{page.URL, page.FinalURL} {
			key := canonicalKey(raw)
			if key != "" {
				out[key] = page
			}
		}
	}
	return out
}

func canonicalKey(raw string) string {
	parsed, err := url.Parse(strings.TrimSpace(raw))
	if err != nil {
		return strings.ToLower(strings.TrimSpace(raw))
	}
	parsed.Fragment = ""
	parsed.Scheme = strings.ToLower(parsed.Scheme)
	parsed.Host = strings.ToLower(parsed.Host)
	if len(parsed.Path) > 1 {
		parsed.Path = strings.TrimRight(parsed.Path, "/")
	}
	return parsed.String()
}

func relContains(rel string, token string) bool {
	token = strings.ToLower(token)
	for _, part := range strings.Fields(strings.ToLower(rel)) {
		if part == token {
			return true
		}
	}
	return false
}

func firstCountryLikePathSegment(path string) string {
	path = strings.Trim(path, "/")
	if path == "" {
		return ""
	}
	segment := strings.ToLower(strings.Split(path, "/")[0])
	switch {
	case len(segment) == 2 && isAlpha(segment):
		return segment
	case len(segment) == 5 && segment[2] == '-' && isAlpha(segment[:2]) && isAlpha(segment[3:]):
		return segment
	default:
		return ""
	}
}

func isAlpha(value string) bool {
	for _, r := range value {
		if r < 'a' || r > 'z' {
			return false
		}
	}
	return value != ""
}

func hostLower(raw string) string {
	parsed, err := url.Parse(strings.TrimSpace(raw))
	if err != nil {
		return ""
	}
	return strings.ToLower(parsed.Hostname())
}

func stripWWW(host string) string {
	return strings.TrimPrefix(strings.ToLower(host), "www.")
}

func sortedKeys(m map[string]bool) []string {
	out := make([]string, 0, len(m))
	for key := range m {
		out = append(out, key)
	}
	sort.Strings(out)
	return out
}

func firstN(values []string, limit int) []string {
	if limit <= 0 || len(values) <= limit {
		return values
	}
	return values[:limit]
}
