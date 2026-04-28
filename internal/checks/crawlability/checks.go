package crawlability

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"

	"github.com/cars24/seo-automation/internal/models"
)

const redirectCategory = "Crawlability"

var (
	jsRedirectPattern     = regexp.MustCompile(`(?is)\b(?:window\.)?location(?:\.href)?\s*=|\b(?:window\.)?location\.(?:replace|assign)\s*\(`)
	metaRefreshPattern    = regexp.MustCompile(`(?is)<meta\b[^>]*http-equiv\s*=\s*["']?refresh["']?[^>]*>`)
	metaRefreshAltPattern = regexp.MustCompile(`(?is)<meta\b[^>]*content\s*=\s*["'][^"']*;\s*url\s*=[^"']+["'][^>]*>`)
)

// PageChecks returns per-page crawlability checks.
func PageChecks() []models.PageCheck {
	return []models.PageCheck{
		&response4xx{},
		&response5xx{},
		&responseTimeout{},
		&redirectChain{},
		&redirectLoop{},
		&redirect302Permanent{},
		&redirectHTTPToHTTPSPermanent{},
		&redirectWWWVariantPermanent{},
		&redirectTrailingSlashPermanent{},
		&redirectJavascript{},
		&redirectMetaRefresh{},
		&redirectDestinationIndexable{},
		&noindexHasInlinks{},
		&pageDepthTooDeep{},
		&robotsNofollowPage{},
		&robotsNoarchive{},
		&robotsNosnippet{},
		&robotsXRobotsTag{},
		&robotsDirectiveConflict{},
	}
}

// SiteChecks returns site-wide crawlability checks.
func SiteChecks() []models.SiteCheck {
	return []models.SiteCheck{
		&robotsTxtMissing{},
		&robotsBlocksAll{},
		&robotsMissingSitemapDirective{},
		&noindexInSitemapSite{},
		&orphanExternalOnly{},
		&robotsPageBlockedButLinked{},
		&redirectTrailingSlashSiteConsistency{},
	}
}

// --- Per-page checks ---

type response4xx struct{}

func (r *response4xx) Run(p *models.PageData) []models.CheckResult {
	if p.StatusCode >= 400 && p.StatusCode < 500 {
		return []models.CheckResult{{
			ID:       "crawl.response.4xx",
			Category: "Crawlability",
			Severity: models.SeverityError,
			Message:  fmt.Sprintf("4xx response code (%d)", p.StatusCode),
			URL:      p.URL,
		}}
	}
	return nil
}

type response5xx struct{}

func (r *response5xx) Run(p *models.PageData) []models.CheckResult {
	if p.StatusCode >= 500 {
		return []models.CheckResult{{
			ID:       "crawl.response.5xx",
			Category: "Crawlability",
			Severity: models.SeverityError,
			Message:  fmt.Sprintf("5xx response code (%d)", p.StatusCode),
			URL:      p.URL,
		}}
	}
	return nil
}

type responseTimeout struct{}

func (r *responseTimeout) Run(p *models.PageData) []models.CheckResult {
	if strings.Contains(strings.ToLower(p.Error), "timeout") ||
		strings.Contains(strings.ToLower(p.Error), "deadline exceeded") {
		return []models.CheckResult{{
			ID:       "crawl.response.timeout",
			Category: "Crawlability",
			Severity: models.SeverityError,
			Message:  "Request timed out",
			URL:      p.URL,
			Details:  p.Error,
		}}
	}
	return nil
}

type redirectChain struct{}

func (r *redirectChain) Run(p *models.PageData) []models.CheckResult {
	if len(p.RedirectChain) > 3 {
		return []models.CheckResult{{
			ID:       "crawl.redirect.chain",
			Category: redirectCategory,
			Severity: models.SeverityWarning,
			Message:  fmt.Sprintf("Redirect chain too long (%d hops)", len(p.RedirectChain)),
			URL:      p.URL,
			Details:  redirectChainDetails(p),
		}}
	}
	return nil
}

type redirectLoop struct{}

func (r *redirectLoop) Run(p *models.PageData) []models.CheckResult {
	if strings.Contains(strings.ToLower(p.Error), "redirect loop") ||
		strings.Contains(strings.ToLower(p.Error), "stopped after") {
		return []models.CheckResult{{
			ID:       "crawl.redirect.loop",
			Category: redirectCategory,
			Severity: models.SeverityError,
			Message:  "Redirect loop detected",
			URL:      p.URL,
			Details:  p.Error,
		}}
	}
	// Check for repeated URL in chain
	seen := map[string]bool{}
	for _, hop := range p.RedirectChain {
		if seen[hop.URL] {
			return []models.CheckResult{{
				ID:       "crawl.redirect.loop",
				Category: redirectCategory,
				Severity: models.SeverityError,
				Message:  "Redirect loop detected",
				URL:      p.URL,
			}}
		}
		seen[hop.URL] = true
	}
	return nil
}

type redirect302Permanent struct{}

func (r *redirect302Permanent) Run(p *models.PageData) []models.CheckResult {
	for _, transition := range redirectTransitions(p) {
		if isTemporaryRedirectStatus(transition.StatusCode) {
			return []models.CheckResult{{
				ID:       "crawl.redirect.302_permanent",
				Category: redirectCategory,
				Severity: models.SeverityWarning,
				Message:  "Temporary redirect used for a likely permanent move",
				URL:      p.URL,
				Details:  transition.String(),
			}}
		}
	}
	if isTemporaryRedirectStatus(p.StatusCode) && p.URL != p.FinalURL {
		return []models.CheckResult{{
			ID:       "crawl.redirect.302_permanent",
			Category: redirectCategory,
			Severity: models.SeverityWarning,
			Message:  "Temporary redirect used for a likely permanent move",
			URL:      p.URL,
			Details:  fmt.Sprintf("Redirects to: %s", p.FinalURL),
		}}
	}
	return nil
}

type redirectHTTPToHTTPSPermanent struct{}

func (r *redirectHTTPToHTTPSPermanent) Run(p *models.PageData) []models.CheckResult {
	start, final := parsedURL(p.URL), parsedURL(firstNonEmpty(p.FinalURL, p.URL))
	if start == nil || start.Scheme != "http" {
		return nil
	}
	if final == nil || final.Scheme != "https" {
		return []models.CheckResult{{
			ID:       "crawl.redirect.http_to_https_not_301",
			Category: redirectCategory,
			Severity: models.SeverityError,
			Message:  "HTTP URL does not permanently redirect to HTTPS",
			URL:      p.URL,
			Details:  redirectChainDetails(p),
		}}
	}
	for _, transition := range redirectTransitions(p) {
		from, to := parsedURL(transition.From), parsedURL(transition.To)
		if from == nil || to == nil || from.Scheme != "http" || to.Scheme != "https" {
			continue
		}
		if !isPermanentRedirectStatus(transition.StatusCode) {
			return []models.CheckResult{{
				ID:       "crawl.redirect.http_to_https_not_301",
				Category: redirectCategory,
				Severity: models.SeverityWarning,
				Message:  "HTTP to HTTPS redirect is not permanent",
				URL:      p.URL,
				Details:  transition.String(),
			}}
		}
	}
	return nil
}

type redirectWWWVariantPermanent struct{}

func (r *redirectWWWVariantPermanent) Run(p *models.PageData) []models.CheckResult {
	for _, transition := range redirectTransitions(p) {
		from, to := parsedURL(transition.From), parsedURL(transition.To)
		if !isWWWVariantChange(from, to) {
			continue
		}
		if !isPermanentRedirectStatus(transition.StatusCode) {
			return []models.CheckResult{{
				ID:       "crawl.redirect.www_variant_not_301",
				Category: redirectCategory,
				Severity: models.SeverityWarning,
				Message:  "WWW/non-WWW redirect is not permanent",
				URL:      p.URL,
				Details:  transition.String(),
			}}
		}
	}
	return nil
}

type redirectTrailingSlashPermanent struct{}

func (r *redirectTrailingSlashPermanent) Run(p *models.PageData) []models.CheckResult {
	for _, transition := range redirectTransitions(p) {
		from, to := parsedURL(transition.From), parsedURL(transition.To)
		if !isTrailingSlashChange(from, to) {
			continue
		}
		if !isPermanentRedirectStatus(transition.StatusCode) {
			return []models.CheckResult{{
				ID:       "crawl.redirect.trailing_slash_inconsistent",
				Category: redirectCategory,
				Severity: models.SeverityWarning,
				Message:  "Trailing slash redirect is not permanent",
				URL:      p.URL,
				Details:  transition.String(),
			}}
		}
	}
	return nil
}

type redirectJavascript struct{}

func (r *redirectJavascript) Run(p *models.PageData) []models.CheckResult {
	if strings.TrimSpace(p.RawHTML) == "" || !jsRedirectPattern.MatchString(p.RawHTML) {
		return nil
	}
	return []models.CheckResult{{
		ID:       "crawl.redirect.javascript",
		Category: redirectCategory,
		Severity: models.SeverityWarning,
		Message:  "JavaScript-based redirect detected",
		URL:      p.URL,
		Details:  snippetAroundMatch(p.RawHTML, jsRedirectPattern),
	}}
}

type redirectMetaRefresh struct{}

func (r *redirectMetaRefresh) Run(p *models.PageData) []models.CheckResult {
	if strings.TrimSpace(p.RawHTML) == "" {
		return nil
	}
	match := metaRefreshPattern.FindString(p.RawHTML)
	if match == "" {
		return nil
	}
	if !strings.Contains(strings.ToLower(match), "url=") && !metaRefreshAltPattern.MatchString(match) {
		return nil
	}
	return []models.CheckResult{{
		ID:       "crawl.redirect.meta_refresh",
		Category: redirectCategory,
		Severity: models.SeverityWarning,
		Message:  "Meta refresh redirect detected",
		URL:      p.URL,
		Details:  compactSnippet(match),
	}}
}

type redirectDestinationIndexable struct{}

func (r *redirectDestinationIndexable) Run(p *models.PageData) []models.CheckResult {
	if !hasRedirect(p) {
		return nil
	}
	if p.StatusCode != 0 && p.StatusCode != 200 {
		return []models.CheckResult{{
			ID:       "crawl.redirect.destination_not_indexable",
			Category: redirectCategory,
			Severity: models.SeverityError,
			Message:  "Redirect destination does not return HTTP 200",
			URL:      p.URL,
			Details:  fmt.Sprintf("final_url=%s status=%d", p.FinalURL, p.StatusCode),
		}}
	}
	if robotsHasDirective(p, "noindex") {
		return []models.CheckResult{{
			ID:       "crawl.redirect.destination_not_indexable",
			Category: redirectCategory,
			Severity: models.SeverityWarning,
			Message:  "Redirect destination is marked noindex",
			URL:      p.URL,
			Details:  fmt.Sprintf("final_url=%s robots=%s x_robots=%s", p.FinalURL, p.RobotsTag, p.XRobotsTag),
		}}
	}
	return nil
}

type noindexHasInlinks struct{}

func (n *noindexHasInlinks) Run(p *models.PageData) []models.CheckResult {
	// Per-page: flag if noindex is set (site-wide check will verify inlinks)
	if strings.Contains(p.RobotsTag, "noindex") && p.InSitemap {
		return []models.CheckResult{{
			ID:       "crawl.noindex.in_sitemap",
			Category: "Crawlability",
			Severity: models.SeverityWarning,
			Message:  "Noindex page is listed in sitemap",
			URL:      p.URL,
		}}
	}
	return nil
}

type pageDepthTooDeep struct{}

func (pd *pageDepthTooDeep) Run(p *models.PageData) []models.CheckResult {
	if p.Depth > 3 {
		return []models.CheckResult{{
			ID:       "crawl.page_depth.too_deep",
			Category: "Crawlability",
			Severity: models.SeverityWarning,
			Message:  fmt.Sprintf("Page is too deep in site structure (depth: %d)", p.Depth),
			URL:      p.URL,
		}}
	}
	return nil
}

// --- Robots directive checks (Feature 1.1) ---

type robotsNofollowPage struct{}

func (c *robotsNofollowPage) Run(p *models.PageData) []models.CheckResult {
	for _, d := range p.RobotsDirectives {
		if d == "nofollow" {
			return []models.CheckResult{{
				ID:       "crawl.robots.nofollow_page",
				Category: "Crawlability",
				Severity: models.SeverityWarning,
				Message:  "Page has nofollow robots directive — outgoing links will not pass PageRank",
				URL:      p.URL,
				Details:  "Source: " + directiveSource(p, "nofollow"),
			}}
		}
	}
	return nil
}

type robotsNoarchive struct{}

func (c *robotsNoarchive) Run(p *models.PageData) []models.CheckResult {
	for _, d := range p.RobotsDirectives {
		if d == "noarchive" {
			return []models.CheckResult{{
				ID:       "crawl.robots.noarchive",
				Category: "Crawlability",
				Severity: models.SeverityNotice,
				Message:  "Page has noarchive directive — cached version unavailable in search",
				URL:      p.URL,
				Details:  "Source: " + directiveSource(p, "noarchive"),
			}}
		}
	}
	return nil
}

type robotsNosnippet struct{}

func (c *robotsNosnippet) Run(p *models.PageData) []models.CheckResult {
	for _, d := range p.RobotsDirectives {
		if d == "nosnippet" {
			return []models.CheckResult{{
				ID:       "crawl.robots.nosnippet",
				Category: "Crawlability",
				Severity: models.SeverityWarning,
				Message:  "Page has nosnippet directive — no text snippet will appear in search results",
				URL:      p.URL,
				Details:  "Source: " + directiveSource(p, "nosnippet"),
			}}
		}
	}
	return nil
}

type robotsXRobotsTag struct{}

func (c *robotsXRobotsTag) Run(p *models.PageData) []models.CheckResult {
	if p.XRobotsTag != "" {
		return []models.CheckResult{{
			ID:       "crawl.robots.x_robots_tag",
			Category: "Crawlability",
			Severity: models.SeverityNotice,
			Message:  "X-Robots-Tag HTTP header is present",
			URL:      p.URL,
			Details:  p.XRobotsTag,
		}}
	}
	return nil
}

type robotsDirectiveConflict struct{}

func (c *robotsDirectiveConflict) Run(p *models.PageData) []models.CheckResult {
	set := make(map[string]bool, len(p.RobotsDirectives))
	for _, d := range p.RobotsDirectives {
		set[d] = true
	}
	var conflicts []string
	if set["index"] && set["noindex"] {
		conflicts = append(conflicts, "index + noindex")
	}
	if set["follow"] && set["nofollow"] {
		conflicts = append(conflicts, "follow + nofollow")
	}
	if len(conflicts) > 0 {
		return []models.CheckResult{{
			ID:       "crawl.robots.directive_conflict",
			Category: "Crawlability",
			Severity: models.SeverityError,
			Message:  "Conflicting robots directives detected",
			URL:      p.URL,
			Details:  strings.Join(conflicts, "; "),
		}}
	}
	return nil
}

func directiveSource(p *models.PageData, directive string) string {
	inMeta := strings.Contains(p.RobotsTag, directive)
	inHeader := p.XRobotsTag != "" && strings.Contains(strings.ToLower(p.XRobotsTag), directive)
	switch {
	case inMeta && inHeader:
		return "meta robots + X-Robots-Tag"
	case inHeader:
		return "X-Robots-Tag"
	default:
		return "meta robots"
	}
}

// --- Site-wide checks ---

type robotsTxtMissing struct{}

func (r *robotsTxtMissing) Run(pages []*models.PageData) []models.CheckResult {
	if len(pages) == 0 {
		return nil
	}
	// Check the audit-level flag passed via first page's parent audit
	// We detect this via absence of robots data — use a heuristic: if the site
	// served pages fine but we have no robots info, we can't tell from pages alone.
	// This will be set by the crawler in audit.RobotsTxtMissing; for now skip.
	return nil
}

type robotsBlocksAll struct{}

func (r *robotsBlocksAll) Run(pages []*models.PageData) []models.CheckResult {
	return nil
}

type robotsMissingSitemapDirective struct{}

func (r *robotsMissingSitemapDirective) Run(pages []*models.PageData) []models.CheckResult {
	return nil
}

type noindexInSitemapSite struct{}

func (n *noindexInSitemapSite) Run(pages []*models.PageData) []models.CheckResult {
	var results []models.CheckResult
	for _, p := range pages {
		if strings.Contains(p.RobotsTag, "noindex") && p.InSitemap {
			results = append(results, models.CheckResult{
				ID:       "crawl.noindex.in_sitemap",
				Category: "Crawlability",
				Severity: models.SeverityWarning,
				Message:  "Noindex page listed in XML sitemap",
				URL:      p.URL,
			})
		}
	}
	return results
}

type orphanExternalOnly struct{}

func (o *orphanExternalOnly) Run(pages []*models.PageData) []models.CheckResult {
	// Build inlink map
	inlinks := map[string]int{}
	for _, p := range pages {
		for _, link := range p.Links {
			if link.IsInternal {
				inlinks[link.URL]++
			}
		}
	}
	var results []models.CheckResult
	for _, p := range pages {
		if p.Depth == 0 {
			continue // seed page is never orphan
		}
		if inlinks[p.URL] == 0 && inlinks[p.FinalURL] == 0 {
			results = append(results, models.CheckResult{
				ID:       "crawl.page_depth.orphan_external_only",
				Category: "Crawlability",
				Severity: models.SeverityWarning,
				Message:  "Page has no internal inlinks (orphan page)",
				URL:      p.URL,
			})
		}
	}
	return results
}

type robotsPageBlockedButLinked struct{}

func (r *robotsPageBlockedButLinked) Run(pages []*models.PageData) []models.CheckResult {
	blocked := map[string]bool{}
	for _, p := range pages {
		if strings.Contains(p.Error, "robots") {
			blocked[p.URL] = true
		}
	}
	if len(blocked) == 0 {
		return nil
	}
	linked := map[string]bool{}
	for _, p := range pages {
		for _, link := range p.Links {
			if link.IsInternal {
				linked[link.URL] = true
			}
		}
	}
	var results []models.CheckResult
	for url := range blocked {
		if linked[url] {
			results = append(results, models.CheckResult{
				ID:       "crawl.robots.page_blocked_but_linked",
				Category: "Crawlability",
				Severity: models.SeverityWarning,
				Message:  "Page is blocked by robots.txt but has internal inlinks",
				URL:      url,
			})
		}
	}
	return results
}

type redirectTrailingSlashSiteConsistency struct{}

func (r *redirectTrailingSlashSiteConsistency) Run(pages []*models.PageData) []models.CheckResult {
	byKey := map[string][]string{}
	for _, page := range pages {
		if page == nil || page.StatusCode != 200 || hasRedirect(page) {
			continue
		}
		u := parsedURL(firstNonEmpty(page.FinalURL, page.URL))
		if u == nil || u.Path == "" || u.Path == "/" {
			continue
		}
		key := trailingSlashKey(u)
		if key == "" {
			continue
		}
		byKey[key] = append(byKey[key], u.String())
	}
	for _, variants := range byKey {
		if len(uniqueStrings(variants)) > 1 {
			return []models.CheckResult{{
				ID:       "crawl.redirect.trailing_slash_inconsistent",
				Category: redirectCategory,
				Severity: models.SeverityWarning,
				Message:  "Trailing slash URL variants both return 200",
				URL:      variants[0],
				Details:  strings.Join(uniqueStrings(variants), " | "),
			}}
		}
	}
	return nil
}

type redirectTransition struct {
	From       string
	To         string
	StatusCode int
}

func (t redirectTransition) String() string {
	status := "unknown"
	if t.StatusCode > 0 {
		status = fmt.Sprintf("%d", t.StatusCode)
	}
	return fmt.Sprintf("%s --%s--> %s", t.From, status, t.To)
}

func redirectTransitions(p *models.PageData) []redirectTransition {
	if p == nil || len(p.RedirectChain) == 0 {
		return nil
	}
	var transitions []redirectTransition
	for i, hop := range p.RedirectChain {
		to := firstNonEmpty(p.FinalURL, p.URL)
		if i+1 < len(p.RedirectChain) {
			to = p.RedirectChain[i+1].URL
		}
		transitions = append(transitions, redirectTransition{
			From:       hop.URL,
			To:         to,
			StatusCode: hop.StatusCode,
		})
	}
	return transitions
}

func hasRedirect(p *models.PageData) bool {
	if p == nil {
		return false
	}
	return len(p.RedirectChain) > 0 || (p.URL != "" && p.FinalURL != "" && !sameURL(p.URL, p.FinalURL))
}

func redirectChainDetails(p *models.PageData) string {
	transitions := redirectTransitions(p)
	if len(transitions) == 0 {
		return firstNonEmpty(p.FinalURL, p.URL)
	}
	parts := make([]string, 0, len(transitions))
	for _, transition := range transitions {
		parts = append(parts, transition.String())
	}
	return strings.Join(parts, " | ")
}

func isPermanentRedirectStatus(status int) bool {
	return status == 301 || status == 308
}

func isTemporaryRedirectStatus(status int) bool {
	return status == 302 || status == 303 || status == 307
}

func parsedURL(raw string) *url.URL {
	u, err := url.Parse(strings.TrimSpace(raw))
	if err != nil {
		return nil
	}
	return u
}

func isWWWVariantChange(from *url.URL, to *url.URL) bool {
	if from == nil || to == nil {
		return false
	}
	fromHost := strings.ToLower(from.Hostname())
	toHost := strings.ToLower(to.Hostname())
	return fromHost != toHost && stripWWW(fromHost) == stripWWW(toHost)
}

func stripWWW(host string) string {
	return strings.TrimPrefix(strings.ToLower(host), "www.")
}

func isTrailingSlashChange(from *url.URL, to *url.URL) bool {
	if from == nil || to == nil {
		return false
	}
	if !strings.EqualFold(from.Scheme, to.Scheme) || !strings.EqualFold(from.Host, to.Host) {
		return false
	}
	if from.RawQuery != to.RawQuery {
		return false
	}
	fromPath := normalPath(from.Path)
	toPath := normalPath(to.Path)
	if fromPath == "/" || toPath == "/" || fromPath == toPath {
		return false
	}
	return strings.TrimRight(fromPath, "/") == strings.TrimRight(toPath, "/")
}

func trailingSlashKey(u *url.URL) string {
	if u == nil {
		return ""
	}
	path := normalPath(u.Path)
	if path == "/" {
		return ""
	}
	copyURL := *u
	copyURL.Path = strings.TrimRight(path, "/")
	copyURL.RawQuery = ""
	copyURL.Fragment = ""
	return strings.ToLower(copyURL.Scheme + "://" + copyURL.Host + copyURL.Path)
}

func normalPath(path string) string {
	if path == "" {
		return "/"
	}
	return path
}

func sameURL(a string, b string) bool {
	parsedA, parsedB := parsedURL(a), parsedURL(b)
	if parsedA == nil || parsedB == nil {
		return strings.TrimRight(a, "/") == strings.TrimRight(b, "/")
	}
	parsedA.Fragment = ""
	parsedB.Fragment = ""
	return strings.TrimRight(parsedA.String(), "/") == strings.TrimRight(parsedB.String(), "/")
}

func robotsHasDirective(p *models.PageData, directive string) bool {
	directive = strings.ToLower(directive)
	for _, value := range p.RobotsDirectives {
		if strings.ToLower(value) == directive {
			return true
		}
	}
	return strings.Contains(strings.ToLower(p.RobotsTag), directive) ||
		strings.Contains(strings.ToLower(p.XRobotsTag), directive)
}

func snippetAroundMatch(raw string, pattern *regexp.Regexp) string {
	loc := pattern.FindStringIndex(raw)
	if len(loc) != 2 {
		return ""
	}
	start := loc[0] - 60
	if start < 0 {
		start = 0
	}
	end := loc[1] + 80
	if end > len(raw) {
		end = len(raw)
	}
	return compactSnippet(raw[start:end])
}

func compactSnippet(raw string) string {
	return strings.Join(strings.Fields(raw), " ")
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

func uniqueStrings(values []string) []string {
	seen := map[string]bool{}
	var unique []string
	for _, value := range values {
		if seen[value] {
			continue
		}
		seen[value] = true
		unique = append(unique, value)
	}
	return unique
}
