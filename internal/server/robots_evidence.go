package server

import (
	"fmt"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/cars24/seo-automation/internal/models"
)

const maxRobotsTxtBytes = 500 * 1024

type robotsDirectiveLine struct {
	Agents    []string
	Directive string
	Value     string
	Line      int
	Raw       string
}

func robotsEvidenceChecks() []CrawlerEvidenceCheck {
	return []CrawlerEvidenceCheck{
		robotCheck("ROBOTS-001", "robots.txt file exists and is accessible", "Critical", "Fetch /robots.txt and verify it returns HTTP 200."),
		robotCheck("ROBOTS-002", "robots.txt returns correct content-type", "High", "Validate the robots.txt response Content-Type header."),
		robotCheck("ROBOTS-003", "robots.txt file size under 500KB", "High", "Measure robots.txt bytes fetched from the live host."),
		robotCheck("ROBOTS-004", "Sitemap URL declared in robots.txt", "High", "Parse Sitemap directives from robots.txt."),
		robotCheck("ROBOTS-005", "No critical pages blocked by robots.txt", "Critical", "Use crawl evidence to flag internally linked pages blocked by robots.txt."),
		robotCheck("ROBOTS-006", "CSS and JS files not blocked in robots.txt", "High", "Test discovered internal CSS/JS resource URLs against robots.txt."),
		robotCheck("ROBOTS-007", "Googlebot user-agent rules are correctly written", "High", "Parse Googlebot or wildcard policy and verify Googlebot can crawl the root."),
		robotCheck("ROBOTS-008", "No conflicting Allow/Disallow rules", "Medium", "Detect identical Allow and Disallow paths within the same robots group."),
		robotCheck("ROBOTS-009", "Staging/dev environments blocked in robots.txt", "Medium", "Check staging-like hosts directly and request evidence for production hosts."),
		robotCheck("ROBOTS-010", "robots.txt tested with crawler policy checks", "High", "Fetch robots.txt, parse rules, and verify the seed URL is crawlable."),
		robotCheck("ROBOTS-011", "Crawl-delay not set too high for Googlebot", "Medium", "Inspect Crawl-delay values in Googlebot and wildcard robots groups."),
		robotCheck("ROBOTS-012", "Internal search result pages blocked", "High", "Look for search-result URL patterns and matching Disallow rules."),
		robotCheck("ROBOTS-013", "Faceted/filtered URLs blocked if not indexable", "High", "Compare discovered facet/filter parameters with robots.txt disallow coverage."),
		robotCheck("ROBOTS-014", "Admin/login/dashboard pages blocked", "Medium", "Look for protected path links and common protected-path Disallow rules."),
		robotCheck("ROBOTS-015", "API endpoints blocked", "Medium", "Look for internal API endpoint links and common API Disallow rules."),
		robotCheck("ROBOTS-016", "robots.txt syntax is valid", "High", "Parse robots.txt and report syntax errors."),
		robotCheck("ROBOTS-017", "Wildcard patterns in robots.txt are correct", "Medium", "Inspect wildcard and end-anchor usage for suspicious patterns."),
		robotCheck("ROBOTS-018", "robots.txt allows image crawling", "Medium", "Test discovered image URLs and the image crawl path with Googlebot-Image."),
		robotCheck("ROBOTS-019", "CDN does not overwrite robots.txt rules", "Medium", "Verify robots.txt is not served as HTML/error content by an edge layer."),
		robotCheck("ROBOTS-020", "Disallow rules for parameter URLs kept updated", "High", "Discover parameter URLs and compare them to robots.txt disallow coverage."),
	}
}

func robotCheck(id, name, priority, role string) CrawlerEvidenceCheck {
	return CrawlerEvidenceCheck{
		ID:          id,
		Name:        name,
		Category:    "Technical SEO",
		Priority:    priority,
		CrawlerRole: role,
		Notes:       "Runs during the normal crawler evidence pass.",
	}
}

func analyzeRobotsEvidence(audit *models.SiteAudit, robots robotsEvidence, req CrawlerEvidenceRunRequest) []CrawlerEvidenceItem {
	return []CrawlerEvidenceItem{
		analyzeRobotsExists(robots),
		analyzeRobotsContentType(robots),
		analyzeRobotsSize(robots),
		analyzeRobotsSitemapDirective(robots),
		analyzeCriticalPagesNotBlocked(audit, robots),
		analyzeRobotsResourceAccess(audit, robots),
		analyzeGooglebotRules(robots),
		analyzeRobotsRuleConflicts(robots),
		analyzeStagingRobots(req.URL, robots),
		analyzeRobotsTester(audit, robots, req.URL),
		analyzeCrawlDelay(robots),
		analyzeInternalSearchBlocked(audit, robots),
		analyzeFacetedURLsBlocked(audit, robots),
		analyzeProtectedPathsBlocked(audit, robots),
		analyzeAPIEndpointsBlocked(audit, robots),
		analyzeRobotsSyntax(robots),
		analyzeWildcardRules(robots),
		analyzeImageCrawling(audit, robots),
		analyzeRobotsCDNOverwrite(robots),
		analyzeParameterDisallows(audit, robots, req),
	}
}

func analyzeRobotsExists(robots robotsEvidence) CrawlerEvidenceItem {
	item := robotsItem("ROBOTS-001", "robots.txt file exists and is accessible", "pass", "robots.txt is accessible.")
	item.Evidence = []string{
		"robots URL: " + robots.URL,
		fmt.Sprintf("status: %d", robots.StatusCode),
	}
	if !robots.Exists {
		item.Status = "fail"
		item.Message = "robots.txt is not accessible with HTTP 200."
		if robots.ParseError != "" {
			item.Evidence = append(item.Evidence, robots.ParseError)
		}
	}
	return item
}

func analyzeRobotsContentType(robots robotsEvidence) CrawlerEvidenceItem {
	item := robotsItem("ROBOTS-002", "robots.txt returns correct content-type", "pass", "robots.txt is served with text/plain content type.")
	item.Evidence = []string{"content-type: " + emptyDash(robots.ContentType)}
	if !robots.Exists {
		item.Status = "warning"
		item.Message = "robots.txt content type could not be verified because the file is not accessible."
		return item
	}
	contentType := strings.ToLower(robots.ContentType)
	if !strings.Contains(contentType, "text/plain") {
		item.Status = "fail"
		item.Message = "robots.txt is not served as text/plain."
	}
	return item
}

func analyzeRobotsSize(robots robotsEvidence) CrawlerEvidenceItem {
	item := robotsItem("ROBOTS-003", "robots.txt file size under 500KB", "pass", "robots.txt is under 500KB.")
	item.Evidence = []string{fmt.Sprintf("size: %d bytes", robots.SizeBytes)}
	if !robots.Exists {
		item.Status = "warning"
		item.Message = "robots.txt size could not be verified because the file is not accessible."
		return item
	}
	if robots.SizeBytes > maxRobotsTxtBytes {
		item.Status = "fail"
		item.Message = "robots.txt exceeds 500KB."
	}
	return item
}

func analyzeRobotsSitemapDirective(robots robotsEvidence) CrawlerEvidenceItem {
	sitemaps := robotsSitemaps(robots)
	item := robotsItem("ROBOTS-004", "Sitemap URL declared in robots.txt", "pass", "robots.txt declares sitemap URLs.")
	item.Evidence = firstN(sitemaps, 8)
	if !robots.Exists {
		item.Status = "warning"
		item.Message = "robots.txt is not accessible, so sitemap directives could not be verified."
		return item
	}
	if len(sitemaps) == 0 {
		item.Status = "warning"
		item.Message = "robots.txt does not declare a Sitemap directive."
	}
	return item
}

func analyzeCriticalPagesNotBlocked(audit *models.SiteAudit, robots robotsEvidence) CrawlerEvidenceItem {
	blocked := blockedRobotsPages(audit.Pages)
	item := robotsItem("ROBOTS-005", "No critical pages blocked by robots.txt", "pass", "No crawled internal pages were blocked by robots.txt.")
	item.Evidence = []string{fmt.Sprintf("pages crawled: %d", audit.PagesCrawled)}
	if !robots.Exists {
		item.Status = "warning"
		item.Message = "robots.txt is not accessible, so blocked critical pages could not be fully verified."
		return item
	}
	if len(blocked) > 0 {
		item.Status = "fail"
		item.Message = "Some internally discovered pages are blocked by robots.txt."
		item.Evidence = firstN(blocked, 10)
	}
	return item
}

func analyzeRobotsResourceAccess(audit *models.SiteAudit, robots robotsEvidence) CrawlerEvidenceItem {
	resources := discoveredInternalResourceURLs(audit.Pages)
	item := robotsItem("ROBOTS-006", "CSS and JS files not blocked in robots.txt", "pass", "Discovered internal CSS/JS resources are allowed for Googlebot.")
	item.Evidence = []string{fmt.Sprintf("internal CSS/JS resources checked: %d", len(resources))}
	if len(resources) == 0 {
		item.Status = "info"
		item.Message = "No internal CSS/JS resources were discovered in the bounded crawl."
		return item
	}
	if !robots.ParseOK {
		item.Status = "warning"
		item.Message = "robots.txt could not be parsed, so CSS/JS access could not be verified."
		return item
	}
	blocked := []string{}
	for _, resourceURL := range resources {
		if !robotsAllowsURL(robots, resourceURL, "Googlebot") {
			blocked = append(blocked, resourceURL)
		}
	}
	if len(blocked) > 0 {
		item.Status = "fail"
		item.Message = "Some internal CSS/JS resources are blocked for Googlebot."
		item.Evidence = firstN(blocked, 10)
	}
	return item
}

func analyzeGooglebotRules(robots robotsEvidence) CrawlerEvidenceItem {
	item := robotsItem("ROBOTS-007", "Googlebot user-agent rules are correctly written", "pass", "Googlebot robots policy parsed successfully and allows root crawling.")
	if !robots.Exists {
		item.Status = "warning"
		item.Message = "robots.txt is not accessible, so Googlebot rules could not be verified."
		return item
	}
	if !robots.ParseOK {
		item.Status = "fail"
		item.Message = "robots.txt could not be parsed for Googlebot rules."
		item.Details = robots.ParseError
		return item
	}
	groupLabel := "wildcard fallback"
	if robotsHasUserAgent(robots.Raw, "googlebot") {
		groupLabel = "explicit Googlebot group"
	}
	item.Evidence = []string{groupLabel}
	if !robotsAllowsPath(robots, "/", "Googlebot") {
		item.Status = "fail"
		item.Message = "Googlebot is blocked from the site root."
	}
	return item
}

func analyzeRobotsRuleConflicts(robots robotsEvidence) CrawlerEvidenceItem {
	item := robotsItem("ROBOTS-008", "No conflicting Allow/Disallow rules", "pass", "No identical Allow/Disallow conflicts were found in robots.txt.")
	if !robots.Exists {
		item.Status = "warning"
		item.Message = "robots.txt is not accessible, so Allow/Disallow conflicts could not be verified."
		return item
	}
	conflicts := robotsRuleConflicts(robots.Raw)
	if len(conflicts) > 0 {
		item.Status = "warning"
		item.Message = "Potential conflicting Allow/Disallow rules were found."
		item.Evidence = firstN(conflicts, 10)
	}
	return item
}

func analyzeStagingRobots(seedURL string, robots robotsEvidence) CrawlerEvidenceItem {
	item := robotsItem("ROBOTS-009", "Staging/dev environments blocked in robots.txt", "needs_input", "Production host cannot prove staging/dev robots behavior without staging URLs.")
	item.Evidence = []string{"host: " + hostname(seedURL)}
	if !isLikelyNonProductionHost(seedURL) {
		return item
	}
	if !robots.ParseOK {
		item.Status = "warning"
		item.Message = "Non-production-looking host found, but robots.txt could not be parsed."
		return item
	}
	item.Status = "pass"
	item.Message = "Non-production-looking host blocks root crawling."
	if robotsAllowsPath(robots, "/", "*") || robotsAllowsPath(robots, "/", "Googlebot") {
		item.Status = "fail"
		item.Message = "Non-production-looking host does not block root crawling."
	}
	return item
}

func analyzeCrawlDelay(robots robotsEvidence) CrawlerEvidenceItem {
	item := robotsItem("ROBOTS-011", "Crawl-delay not set too high for Googlebot", "pass", "No high Crawl-delay was found for Googlebot or wildcard groups.")
	if !robots.ParseOK {
		item.Status = "warning"
		item.Message = "robots.txt could not be parsed, so Crawl-delay could not be verified."
		return item
	}
	delays := robotsCrawlDelays(robots)
	item.Evidence = delays
	tooHigh := []string{}
	for _, delay := range delays {
		if robotDelaySeconds(delay) > 10 {
			tooHigh = append(tooHigh, delay)
		}
	}
	if len(tooHigh) > 0 {
		item.Status = "warning"
		item.Message = "Crawl-delay is above 10 seconds."
		item.Evidence = tooHigh
	}
	return item
}

func analyzeInternalSearchBlocked(audit *models.SiteAudit, robots robotsEvidence) CrawlerEvidenceItem {
	searchURLs := matchingInternalURLs(audit.Pages, internalSearchURL)
	patterns := []string{"/search", "/search/", "?q=", "&q=", "?query=", "&query=", "?keyword=", "&keyword=", "?search=", "&search="}
	item := robotsItem("ROBOTS-012", "Internal search result pages blocked", "needs_input", "No internal search URLs were discovered in the bounded crawl.")
	if len(searchURLs) == 0 && len(robotsDisallowMatches(robots.Raw, patterns)) == 0 {
		return item
	}
	if !robots.ParseOK {
		item.Status = "warning"
		item.Message = "Internal search signals exist, but robots.txt could not be parsed."
		item.Evidence = firstN(searchURLs, 8)
		return item
	}
	blocked := robotsDisallowMatches(robots.Raw, patterns)
	if len(blocked) > 0 {
		item.Status = "pass"
		item.Message = "Internal search patterns are covered by robots.txt disallow rules."
		item.Evidence = blocked
		return item
	}
	item.Status = "warning"
	item.Message = "Internal search URLs were discovered but no matching robots.txt disallow rule was found."
	item.Evidence = firstN(searchURLs, 8)
	return item
}

func analyzeFacetedURLsBlocked(audit *models.SiteAudit, robots robotsEvidence) CrawlerEvidenceItem {
	facetNames := discoveredFacetParameterNames(audit.Pages)
	patterns := []string{"/filter", "/filters", "/facet", "?filter=", "&filter=", "?sort=", "&sort=", "?make=", "&make=", "?model=", "&model=", "?fuel=", "&fuel=", "?body=", "&body="}
	item := robotsItem("ROBOTS-013", "Faceted/filtered URLs blocked if not indexable", "needs_input", "No faceted or filtered URL patterns were discovered in the bounded crawl.")
	if len(facetNames) == 0 && len(robotsDisallowMatches(robots.Raw, patterns)) == 0 {
		return item
	}
	if !robots.ParseOK {
		item.Status = "warning"
		item.Message = "Facet/filter signals exist, but robots.txt could not be parsed."
		item.Evidence = sortedKeys(facetNames)
		return item
	}
	names := sortedKeys(facetNames)
	covered, missing := parameterCoverage(robots.Raw, names)
	if len(missing) == 0 && len(names) > 0 {
		item.Status = "pass"
		item.Message = "Discovered facet/filter parameters are covered by robots.txt disallow rules."
		item.Evidence = covered
		return item
	}
	blocked := robotsDisallowMatches(robots.Raw, patterns)
	if len(blocked) > 0 && len(names) == 0 {
		item.Status = "pass"
		item.Message = "robots.txt includes disallow rules for common facet/filter patterns."
		item.Evidence = blocked
		return item
	}
	item.Status = "warning"
	item.Message = "Some discovered facet/filter parameters are not covered by robots.txt disallow rules."
	item.Evidence = append([]string{"covered: " + strings.Join(covered, ", ")}, "missing: "+strings.Join(missing, ", "))
	return item
}

func analyzeProtectedPathsBlocked(audit *models.SiteAudit, robots robotsEvidence) CrawlerEvidenceItem {
	patterns := []string{"/admin", "/login", "/dashboard", "/account", "/auth", "/signin", "/sign-in"}
	found := matchingInternalURLs(audit.Pages, func(raw string) bool { return urlContainsAny(raw, patterns) })
	matches := robotsDisallowMatches(robots.Raw, patterns)
	item := robotsItem("ROBOTS-014", "Admin/login/dashboard pages blocked", "needs_input", "No admin/login/dashboard URLs were discovered in the bounded crawl.")
	if len(matches) > 0 {
		item.Status = "pass"
		item.Message = "robots.txt includes disallow rules for protected path patterns."
		item.Evidence = matches
		return item
	}
	if len(found) > 0 {
		item.Status = "warning"
		item.Message = "Protected-looking URLs were discovered but no matching robots.txt disallow rule was found."
		item.Evidence = firstN(found, 8)
	}
	return item
}

func analyzeAPIEndpointsBlocked(audit *models.SiteAudit, robots robotsEvidence) CrawlerEvidenceItem {
	patterns := []string{"/api", "/api/", "/graphql", "/rest/"}
	found := matchingInternalURLs(audit.Pages, func(raw string) bool { return urlContainsAny(raw, patterns) })
	matches := robotsDisallowMatches(robots.Raw, patterns)
	item := robotsItem("ROBOTS-015", "API endpoints blocked", "needs_input", "No internal API endpoint URLs were discovered in the bounded crawl.")
	if len(matches) > 0 {
		item.Status = "pass"
		item.Message = "robots.txt includes disallow rules for API endpoint patterns."
		item.Evidence = matches
		return item
	}
	if len(found) > 0 {
		item.Status = "warning"
		item.Message = "API endpoint URLs were discovered but no matching robots.txt disallow rule was found."
		item.Evidence = firstN(found, 8)
	}
	return item
}

func analyzeRobotsSyntax(robots robotsEvidence) CrawlerEvidenceItem {
	item := robotsItem("ROBOTS-016", "robots.txt syntax is valid", "pass", "robots.txt parsed successfully.")
	if !robots.Exists {
		item.Status = "warning"
		item.Message = "robots.txt is not accessible, so syntax could not be verified."
		return item
	}
	if !robots.ParseOK {
		item.Status = "fail"
		item.Message = "robots.txt syntax could not be parsed."
		item.Details = robots.ParseError
	}
	return item
}

func analyzeWildcardRules(robots robotsEvidence) CrawlerEvidenceItem {
	item := robotsItem("ROBOTS-017", "Wildcard patterns in robots.txt are correct", "pass", "No suspicious wildcard or end-anchor patterns were found.")
	if !robots.Exists {
		item.Status = "warning"
		item.Message = "robots.txt is not accessible, so wildcard patterns could not be verified."
		return item
	}
	if !robots.ParseOK {
		item.Status = "fail"
		item.Message = "robots.txt could not be parsed before wildcard inspection."
		item.Details = robots.ParseError
		return item
	}
	suspicious := suspiciousWildcardRules(robots.Raw)
	if len(suspicious) > 0 {
		item.Status = "warning"
		item.Message = "Suspicious wildcard or end-anchor robots patterns were found."
		item.Evidence = firstN(suspicious, 10)
	}
	return item
}

func analyzeImageCrawling(audit *models.SiteAudit, robots robotsEvidence) CrawlerEvidenceItem {
	imageURLs := discoveredImageURLs(audit.Pages)
	item := robotsItem("ROBOTS-018", "robots.txt allows image crawling", "pass", "Googlebot-Image is allowed to crawl discovered images.")
	item.Evidence = []string{fmt.Sprintf("images checked: %d", len(imageURLs))}
	if !robots.ParseOK {
		item.Status = "warning"
		item.Message = "robots.txt could not be parsed, so image crawling access could not be verified."
		return item
	}
	if len(imageURLs) == 0 {
		if !robotsAllowsPath(robots, "/images/", "Googlebot-Image") {
			item.Status = "warning"
			item.Message = "No images were discovered, and /images/ appears blocked for Googlebot-Image."
		}
		return item
	}
	blocked := []string{}
	for _, imageURL := range imageURLs {
		if !robotsAllowsURL(robots, imageURL, "Googlebot-Image") {
			blocked = append(blocked, imageURL)
		}
	}
	if len(blocked) > 0 {
		item.Status = "fail"
		item.Message = "Some discovered image URLs are blocked for Googlebot-Image."
		item.Evidence = firstN(blocked, 10)
	}
	return item
}

func analyzeRobotsCDNOverwrite(robots robotsEvidence) CrawlerEvidenceItem {
	item := robotsItem("ROBOTS-019", "CDN does not overwrite robots.txt rules", "pass", "robots.txt is parsable and does not look like CDN-generated HTML/error content.")
	item.Evidence = []string{
		"content-type: " + emptyDash(robots.ContentType),
		"server: " + emptyDash(robots.Server),
		"via: " + emptyDash(robots.Via),
	}
	if !robots.Exists {
		item.Status = "warning"
		item.Message = "robots.txt is not accessible, so CDN overwrite behavior could not be verified."
		return item
	}
	if looksLikeHTML(robots.Raw) || strings.Contains(strings.ToLower(robots.ContentType), "text/html") {
		item.Status = "fail"
		item.Message = "robots.txt appears to be served as HTML, which may indicate CDN or edge overwrite."
		return item
	}
	if !robots.ParseOK {
		item.Status = "fail"
		item.Message = "robots.txt is served but not parsable, which may indicate an edge rewrite or invalid content."
		item.Details = robots.ParseError
	}
	return item
}

func robotsItem(id, name, status, message string) CrawlerEvidenceItem {
	return CrawlerEvidenceItem{
		ID:       id,
		Name:     name,
		Category: "Crawler Evidence",
		Status:   status,
		Message:  message,
	}
}

func robotsSitemaps(robots robotsEvidence) []string {
	seen := map[string]bool{}
	if robots.Data != nil {
		for _, sitemapURL := range robots.Data.Sitemaps {
			if strings.TrimSpace(sitemapURL) != "" {
				seen[strings.TrimSpace(sitemapURL)] = true
			}
		}
	}
	for _, line := range strings.Split(robots.Raw, "\n") {
		key, value, ok := splitRobotsLine(line)
		if ok && key == "sitemap" && value != "" {
			seen[value] = true
		}
	}
	return sortedKeys(seen)
}

func blockedRobotsPages(pages []*models.PageData) []string {
	out := []string{}
	for _, page := range pages {
		if strings.Contains(strings.ToLower(page.Error), "robots") {
			out = append(out, page.URL)
		}
	}
	return sortedUnique(out)
}

func discoveredInternalResourceURLs(pages []*models.PageData) []string {
	seen := map[string]bool{}
	for _, page := range pages {
		for _, resource := range page.Resources {
			if !resource.IsInternal {
				continue
			}
			if resource.Type == models.ResourceCSS || resource.Type == models.ResourceScript {
				seen[resource.URL] = true
			}
		}
	}
	return sortedKeys(seen)
}

func discoveredImageURLs(pages []*models.PageData) []string {
	seen := map[string]bool{}
	for _, page := range pages {
		for _, img := range page.Images {
			if strings.TrimSpace(img.Src) != "" {
				seen[img.Src] = true
			}
		}
	}
	return sortedKeys(seen)
}

func robotsAllowsURL(robots robotsEvidence, rawURL string, agent string) bool {
	path := "/"
	if parsed, err := url.Parse(rawURL); err == nil && parsed.RequestURI() != "" {
		path = parsed.RequestURI()
	}
	return robotsAllowsPath(robots, path, agent)
}

func robotsAllowsPath(robots robotsEvidence, path string, agent string) bool {
	if robots.Data == nil {
		return true
	}
	if strings.TrimSpace(path) == "" {
		path = "/"
	}
	return robots.Data.TestAgent(path, agent)
}

func robotsHasUserAgent(raw, agent string) bool {
	agent = strings.ToLower(strings.TrimSpace(agent))
	for _, line := range strings.Split(raw, "\n") {
		key, value, ok := splitRobotsLine(line)
		if ok && key == "user-agent" && strings.EqualFold(value, agent) {
			return true
		}
	}
	return false
}

func robotsRuleConflicts(raw string) []string {
	seen := map[string]map[string]robotsDirectiveLine{}
	out := []string{}
	for _, rule := range parseRobotsDirectiveLines(raw) {
		if rule.Directive != "allow" && rule.Directive != "disallow" {
			continue
		}
		if rule.Value == "" {
			continue
		}
		group := strings.Join(rule.Agents, ",")
		key := group + "|" + rule.Value
		if _, ok := seen[key]; !ok {
			seen[key] = map[string]robotsDirectiveLine{}
		}
		seen[key][rule.Directive] = rule
		if allow, hasAllow := seen[key]["allow"]; hasAllow {
			if disallow, hasDisallow := seen[key]["disallow"]; hasDisallow {
				out = append(out, fmt.Sprintf("agents=%s path=%s allow_line=%d disallow_line=%d", group, rule.Value, allow.Line, disallow.Line))
			}
		}
	}
	return sortedUnique(out)
}

func robotsCrawlDelays(robots robotsEvidence) []string {
	if robots.Data == nil {
		return nil
	}
	out := []string{}
	for _, agent := range []string{"Googlebot", "*"} {
		group := robots.Data.FindGroup(agent)
		if group != nil && group.CrawlDelay > 0 {
			out = append(out, fmt.Sprintf("%s: %s", agent, group.CrawlDelay.Round(time.Second)))
		}
	}
	if len(out) == 0 {
		return []string{"crawl-delay: not set"}
	}
	return sortedUnique(out)
}

func robotDelaySeconds(value string) int {
	idx := strings.LastIndex(value, ":")
	if idx >= 0 {
		value = strings.TrimSpace(value[idx+1:])
	}
	duration, err := time.ParseDuration(value)
	if err != nil {
		return 0
	}
	return int(duration.Seconds())
}

func matchingInternalURLs(pages []*models.PageData, match func(string) bool) []string {
	seen := map[string]bool{}
	for _, page := range pages {
		for _, raw := range []string{page.URL, page.FinalURL} {
			if match(raw) {
				seen[raw] = true
			}
		}
		for _, link := range page.Links {
			if link.IsInternal && match(link.URL) {
				seen[link.URL] = true
			}
		}
	}
	return sortedKeys(seen)
}

func internalSearchURL(raw string) bool {
	lower := strings.ToLower(raw)
	return strings.Contains(lower, "/search") ||
		strings.Contains(lower, "?q=") ||
		strings.Contains(lower, "&q=") ||
		strings.Contains(lower, "?query=") ||
		strings.Contains(lower, "&query=") ||
		strings.Contains(lower, "?keyword=") ||
		strings.Contains(lower, "&keyword=") ||
		strings.Contains(lower, "?search=") ||
		strings.Contains(lower, "&search=")
}

func discoveredFacetParameterNames(pages []*models.PageData) map[string]bool {
	facetNames := map[string]bool{}
	facetTokens := map[string]bool{
		"filter": true, "sort": true, "make": true, "model": true, "variant": true,
		"fuel": true, "body": true, "transmission": true, "price": true, "min": true,
		"max": true, "year": true, "color": true,
	}
	for name := range discoveredParameterNames(pages) {
		if facetTokens[name] {
			facetNames[name] = true
		}
	}
	return facetNames
}

func robotsDisallowMatches(raw string, patterns []string) []string {
	patterns = lowerTrimmed(patterns)
	out := []string{}
	for _, rule := range parseRobotsDirectiveLines(raw) {
		if rule.Directive != "disallow" || rule.Value == "" {
			continue
		}
		value := strings.ToLower(rule.Value)
		for _, pattern := range patterns {
			if strings.Contains(value, pattern) || strings.Contains(pattern, strings.TrimRight(value, "*")) {
				out = append(out, rule.Raw)
				break
			}
		}
	}
	return sortedUnique(out)
}

func suspiciousWildcardRules(raw string) []string {
	out := []string{}
	for _, rule := range parseRobotsDirectiveLines(raw) {
		if rule.Directive != "allow" && rule.Directive != "disallow" {
			continue
		}
		value := rule.Value
		if !strings.ContainsAny(value, "*$") {
			continue
		}
		switch {
		case strings.Contains(value, "**"):
			out = append(out, fmt.Sprintf("line %d: repeated wildcard in %q", rule.Line, rule.Raw))
		case strings.Contains(value, "$") && !strings.HasSuffix(value, "$"):
			out = append(out, fmt.Sprintf("line %d: end-anchor $ is not at end in %q", rule.Line, rule.Raw))
		case strings.Contains(value, "*") && strings.Count(value, "*") > 4:
			out = append(out, fmt.Sprintf("line %d: unusually many wildcards in %q", rule.Line, rule.Raw))
		}
	}
	return out
}

func parseRobotsDirectiveLines(raw string) []robotsDirectiveLine {
	out := []robotsDirectiveLine{}
	agents := []string{}
	seenDirective := false
	for idx, line := range strings.Split(raw, "\n") {
		key, value, ok := splitRobotsLine(line)
		if !ok {
			continue
		}
		if key == "user-agent" {
			if seenDirective {
				agents = []string{}
				seenDirective = false
			}
			agents = append(agents, strings.ToLower(value))
			continue
		}
		if key != "allow" && key != "disallow" && key != "sitemap" && key != "crawl-delay" {
			continue
		}
		if len(agents) == 0 {
			agents = []string{"*"}
		}
		seenDirective = true
		out = append(out, robotsDirectiveLine{
			Agents:    append([]string(nil), agents...),
			Directive: key,
			Value:     strings.ToLower(value),
			Line:      idx + 1,
			Raw:       strings.TrimSpace(stripRobotsComment(line)),
		})
	}
	return out
}

func splitRobotsLine(line string) (string, string, bool) {
	line = strings.TrimSpace(stripRobotsComment(line))
	if line == "" {
		return "", "", false
	}
	parts := strings.SplitN(line, ":", 2)
	if len(parts) != 2 {
		return "", "", false
	}
	key := strings.ToLower(strings.TrimSpace(parts[0]))
	value := strings.TrimSpace(parts[1])
	if key == "" {
		return "", "", false
	}
	return key, value, true
}

func stripRobotsComment(line string) string {
	if idx := strings.Index(line, "#"); idx >= 0 {
		return line[:idx]
	}
	return line
}

func isLikelyNonProductionHost(raw string) bool {
	host := hostname(raw)
	parts := strings.FieldsFunc(host, func(r rune) bool {
		return r == '.' || r == '-'
	})
	nonProd := map[string]bool{"dev": true, "staging": true, "stage": true, "uat": true, "qa": true, "test": true, "preprod": true, "sandbox": true}
	for _, part := range parts {
		if nonProd[part] {
			return true
		}
	}
	return false
}

func urlContainsAny(raw string, patterns []string) bool {
	lower := strings.ToLower(raw)
	for _, pattern := range patterns {
		if strings.Contains(lower, strings.ToLower(pattern)) {
			return true
		}
	}
	return false
}

func lowerTrimmed(values []string) []string {
	out := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.ToLower(strings.TrimSpace(value))
		if value != "" {
			out = append(out, value)
		}
	}
	return out
}

func sortedUnique(values []string) []string {
	seen := map[string]bool{}
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" {
			seen[value] = true
		}
	}
	out := sortedKeys(seen)
	sort.Strings(out)
	return out
}

func looksLikeHTML(raw string) bool {
	prefix := strings.ToLower(strings.TrimSpace(raw))
	if len(prefix) > 300 {
		prefix = prefix[:300]
	}
	return strings.Contains(prefix, "<html") || strings.Contains(prefix, "<!doctype html")
}

func emptyDash(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return "-"
	}
	return value
}
