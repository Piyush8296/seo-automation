package server

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/cars24/seo-automation/internal/crawler"
	"github.com/cars24/seo-automation/internal/models"
	"github.com/temoto/robotstxt"
)

type CrawlerEvidenceCheck struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Category    string `json:"category"`
	Priority    string `json:"priority"`
	CrawlerRole string `json:"crawler_role"`
	Notes       string `json:"notes"`
}

type CrawlerEvidenceWorkspaceResponse struct {
	Checks        []CrawlerEvidenceCheck       `json:"checks"`
	DefaultConfig CrawlerEvidenceDefaultConfig `json:"default_config"`
}

type CrawlerEvidenceDefaultConfig struct {
	URL          string   `json:"url"`
	MaxPages     int      `json:"max_pages"`
	MaxDepth     int      `json:"max_depth"`
	Concurrency  int      `json:"concurrency"`
	Timeout      string   `json:"timeout"`
	SitemapMode  string   `json:"sitemap_mode"`
	ImageCDNHost []string `json:"image_cdn_hosts"`
}

type CrawlerEvidenceRunRequest struct {
	URL                    string   `json:"url"`
	SitemapURL             string   `json:"sitemap_url"`
	MaxPages               int      `json:"max_pages"`
	MaxDepth               int      `json:"max_depth"`
	Concurrency            int      `json:"concurrency"`
	Timeout                string   `json:"timeout"`
	SitemapMode            string   `json:"sitemap_mode"`
	RespectRobots          *bool    `json:"respect_robots,omitempty"`
	ExpectedInventoryURLs  []string `json:"expected_inventory_urls"`
	ExpectedParameterNames []string `json:"expected_parameter_names"`
	AllowedImageCDNHosts   []string `json:"allowed_image_cdn_hosts"`
	RequiredLiveText       []string `json:"required_live_text"`
}

type CrawlerEvidenceRunResponse struct {
	Status     string                 `json:"status"`
	StartedAt  time.Time              `json:"started_at"`
	FinishedAt time.Time              `json:"finished_at"`
	Config     CrawlerEvidenceConfig  `json:"config"`
	Summary    CrawlerEvidenceSummary `json:"summary"`
	Report     []CrawlerEvidenceItem  `json:"report"`
	Pages      []CrawlerEvidencePage  `json:"pages"`
}

type CrawlerEvidenceConfig struct {
	URL                    string   `json:"url"`
	SitemapURL             string   `json:"sitemap_url,omitempty"`
	MaxPages               int      `json:"max_pages"`
	MaxDepth               int      `json:"max_depth"`
	Concurrency            int      `json:"concurrency"`
	Timeout                string   `json:"timeout"`
	SitemapMode            string   `json:"sitemap_mode"`
	RespectRobots          bool     `json:"respect_robots"`
	ExpectedInventoryCount int      `json:"expected_inventory_count"`
	ExpectedParameterNames []string `json:"expected_parameter_names"`
	AllowedImageCDNHosts   []string `json:"allowed_image_cdn_hosts"`
	RequiredLiveText       []string `json:"required_live_text"`
}

type CrawlerEvidenceSummary struct {
	PagesCrawled    int `json:"pages_crawled"`
	SitemapURLs     int `json:"sitemap_urls"`
	ImagesChecked   int `json:"images_checked"`
	Pass            int `json:"pass"`
	Warning         int `json:"warning"`
	Fail            int `json:"fail"`
	NeedsInput      int `json:"needs_input"`
	Info            int `json:"info"`
	ParameterURLCnt int `json:"parameter_url_count"`
}

type CrawlerEvidenceItem = models.EvidenceCheckResult

type CrawlerEvidencePage struct {
	URL          string `json:"url"`
	StatusCode   int    `json:"status_code"`
	Title        string `json:"title"`
	ContentHash  string `json:"content_hash"`
	InSitemap    bool   `json:"in_sitemap"`
	ImageCount   int    `json:"image_count"`
	HTMLSizeByte int    `json:"html_size_bytes"`
}

type robotsEvidence struct {
	URL        string
	Exists     bool
	ParseOK    bool
	ParseError string
	Raw        string
}

func (h *Handlers) getCrawlerEvidence(w http.ResponseWriter, r *http.Request) {
	cfg := h.manager.GetSettings()
	writeJSON(w, http.StatusOK, CrawlerEvidenceWorkspaceResponse{
		Checks: crawlerEvidenceChecks(),
		DefaultConfig: CrawlerEvidenceDefaultConfig{
			URL:         cfg.SiteProfile.SiteURL,
			MaxPages:    40,
			MaxDepth:    2,
			Concurrency: 5,
			Timeout:     "20s",
			SitemapMode: string(models.SitemapModeDiscover),
		},
	})
}

func (h *Handlers) runCrawlerEvidence(w http.ResponseWriter, r *http.Request) {
	var req CrawlerEvidenceRunRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	req.URL = strings.TrimSpace(req.URL)
	if req.URL == "" {
		writeErr(w, http.StatusBadRequest, "url is required")
		return
	}
	if !strings.HasPrefix(strings.ToLower(req.URL), "http://") && !strings.HasPrefix(strings.ToLower(req.URL), "https://") {
		writeErr(w, http.StatusBadRequest, "url must start with http:// or https://")
		return
	}
	if req.MaxPages <= 0 {
		req.MaxPages = 40
	}
	if req.MaxPages > 250 {
		req.MaxPages = 250
	}
	if req.MaxDepth == 0 {
		req.MaxDepth = 2
	}
	if req.Concurrency <= 0 {
		req.Concurrency = 5
	}
	if req.Concurrency > 12 {
		req.Concurrency = 12
	}
	if strings.TrimSpace(req.Timeout) == "" {
		req.Timeout = "20s"
	}
	timeout, err := time.ParseDuration(req.Timeout)
	if err != nil {
		writeErr(w, http.StatusBadRequest, "invalid timeout")
		return
	}
	sitemapMode := models.SitemapMode(strings.ToLower(strings.TrimSpace(req.SitemapMode)))
	if sitemapMode == "" {
		sitemapMode = models.SitemapModeDiscover
	}
	switch sitemapMode {
	case models.SitemapModeDiscover, models.SitemapModeSeed, models.SitemapModeOff:
	default:
		writeErr(w, http.StatusBadRequest, "sitemap_mode must be discover, seed, or off")
		return
	}
	respectRobots := true
	if req.RespectRobots != nil {
		respectRobots = *req.RespectRobots
	}

	started := time.Now()
	ctx, cancel := context.WithTimeout(r.Context(), timeout*time.Duration(maxInt(req.MaxPages, 1)))
	defer cancel()

	settings := h.manager.GetSettings()
	crawlCfg := &models.CrawlConfig{
		SeedURL:                  req.URL,
		SitemapURL:               strings.TrimSpace(req.SitemapURL),
		Scope:                    models.CrawlScopeHost,
		SitemapMode:              sitemapMode,
		MaxDepth:                 req.MaxDepth,
		MaxPages:                 req.MaxPages,
		Concurrency:              req.Concurrency,
		Timeout:                  timeout,
		NoMobileCheck:            true,
		UserAgent:                defaultUserAgent,
		RespectRobots:            respectRobots,
		MaxRedirects:             defaultMaxRedirects,
		MaxPageSizeBytes:         defaultMaxPageSizeKB * 1024,
		MaxURLLength:             240,
		MaxQueryParams:           8,
		MaxLinksPerPage:          120,
		FollowNofollowLinks:      false,
		ExpandNoindexPages:       true,
		ExpandCanonicalizedPages: true,
		RenderMode:               "html-only",
		DiscoverResources:        true,
		SkipLinkHosts:            settings.SkipLinkHosts,
	}

	audit, err := crawler.NewCrawler(crawlCfg).Crawl(ctx)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}

	robots := fetchRobotsEvidence(ctx, req.URL, timeout)
	report := analyzeCrawlerEvidence(audit, robots, req)
	summary := summarizeCrawlerEvidence(audit, report)
	pages := summarizeCrawlerEvidencePages(audit.Pages)

	writeJSON(w, http.StatusOK, CrawlerEvidenceRunResponse{
		Status:     "complete",
		StartedAt:  started,
		FinishedAt: time.Now(),
		Config: CrawlerEvidenceConfig{
			URL:                    req.URL,
			SitemapURL:             strings.TrimSpace(req.SitemapURL),
			MaxPages:               req.MaxPages,
			MaxDepth:               req.MaxDepth,
			Concurrency:            req.Concurrency,
			Timeout:                req.Timeout,
			SitemapMode:            string(sitemapMode),
			RespectRobots:          respectRobots,
			ExpectedInventoryCount: len(normalizeURLList(req.ExpectedInventoryURLs)),
			ExpectedParameterNames: normalizeTokenList(req.ExpectedParameterNames),
			AllowedImageCDNHosts:   normalizeHostList(req.AllowedImageCDNHosts),
			RequiredLiveText:       normalizeTokenList(req.RequiredLiveText),
		},
		Summary: summary,
		Report:  report,
		Pages:   pages,
	})
}

func crawlerEvidenceChecks() []CrawlerEvidenceCheck {
	return []CrawlerEvidenceCheck{
		{
			ID:          "ROBOTS-010",
			Name:        "robots.txt tested with crawler policy checks",
			Category:    "Technical SEO",
			Priority:    "High",
			CrawlerRole: "Fetch robots.txt, parse rules, and verify the seed URL is crawlable.",
			Notes:       "GSC tester evidence can be attached later; this pack provides crawler-side policy proof.",
		},
		{
			ID:          "ROBOTS-020",
			Name:        "Disallow rules for parameter URLs kept updated",
			Category:    "Technical SEO",
			Priority:    "High",
			CrawlerRole: "Discover parameter URLs and compare them to robots.txt disallow coverage.",
			Notes:       "Useful for filter, sort, city, and tracking parameters.",
		},
		{
			ID:          "SITEMAP-022",
			Name:        "CDP URLs updated in sitemap on new inventory",
			Category:    "Technical SEO",
			Priority:    "Critical",
			CrawlerRole: "Compare sitemap URLs with expected inventory URLs or crawler-discovered priority URLs.",
			Notes:       "Expected inventory can later come from CMS/feed/API.",
		},
		{
			ID:          "IMG-013",
			Name:        "Image CDN configured correctly",
			Category:    "On-Page SEO",
			Priority:    "High",
			CrawlerRole: "Inspect discovered image URLs, status codes, content types, and allowed CDN hosts.",
			Notes:       "Cache header inspection can be added after image HEAD metadata is expanded.",
		},
		{
			ID:          "INDEX-015",
			Name:        "Post-cache-purge content updates reflect live",
			Category:    "Technical SEO",
			Priority:    "High",
			CrawlerRole: "Refetch live pages, capture content hashes, and verify required text snippets.",
			Notes:       "Use required snippets after deploy/cache purge; otherwise this produces a snapshot baseline.",
		},
	}
}

func analyzeCrawlerEvidence(audit *models.SiteAudit, robots robotsEvidence, req CrawlerEvidenceRunRequest) []CrawlerEvidenceItem {
	return []CrawlerEvidenceItem{
		analyzeRobotsTester(audit, robots, req.URL),
		analyzeParameterDisallows(audit, robots, req),
		analyzeSitemapInventory(audit, req.ExpectedInventoryURLs),
		analyzeImageCDN(audit, req.AllowedImageCDNHosts),
		analyzeLiveContent(audit, req.RequiredLiveText),
	}
}

func crawlerEvidenceForAudit(ctx context.Context, audit *models.SiteAudit, req StartAuditRequest, timeout time.Duration) []models.EvidenceCheckResult {
	evidenceReq := CrawlerEvidenceRunRequest{
		URL:                    req.URL,
		SitemapURL:             req.SitemapURL,
		ExpectedInventoryURLs:  req.ExpectedInventoryURLs,
		ExpectedParameterNames: req.ExpectedParameterNames,
		AllowedImageCDNHosts:   req.AllowedImageCDNHosts,
		RequiredLiveText:       req.RequiredLiveText,
	}
	robots := fetchRobotsEvidence(ctx, req.URL, timeout)
	items := analyzeCrawlerEvidence(audit, robots, evidenceReq)
	out := make([]models.EvidenceCheckResult, 0, len(items))
	for _, item := range items {
		out = append(out, models.EvidenceCheckResult(item))
	}
	return out
}

func crawlerEvidenceFindings(items []models.EvidenceCheckResult) []models.CheckResult {
	out := []models.CheckResult{}
	for _, item := range items {
		severity, ok := crawlerEvidenceSeverity(item.Status)
		if !ok {
			continue
		}
		out = append(out, models.CheckResult{
			ID:       item.ID,
			Category: "Crawler Evidence",
			Severity: severity,
			Message:  item.Message,
			URL:      "(site-wide)",
			Details:  crawlerEvidenceDetails(item),
			Platform: models.PlatformBoth,
		})
	}
	return out
}

func crawlerEvidenceSeverity(status string) (models.Severity, bool) {
	switch status {
	case "fail":
		return models.SeverityError, true
	case "warning":
		return models.SeverityWarning, true
	default:
		return "", false
	}
}

func crawlerEvidenceDetails(item models.EvidenceCheckResult) string {
	parts := []string{}
	if item.Name != "" {
		parts = append(parts, item.Name)
	}
	if item.Details != "" {
		parts = append(parts, item.Details)
	}
	if len(item.Evidence) > 0 {
		parts = append(parts, "Evidence: "+strings.Join(firstN(item.Evidence, 8), " | "))
	}
	return strings.Join(parts, " — ")
}

func analyzeRobotsTester(audit *models.SiteAudit, robots robotsEvidence, seedURL string) CrawlerEvidenceItem {
	item := CrawlerEvidenceItem{
		ID:       "ROBOTS-010",
		Name:     "robots.txt tested with crawler policy checks",
		Category: "Crawler Evidence",
		Status:   "pass",
		Message:  "robots.txt was fetched, parsed, and does not block all crawling.",
		Evidence: []string{
			"robots URL: " + robots.URL,
			fmt.Sprintf("pages crawled: %d", audit.PagesCrawled),
		},
	}
	if !robots.Exists {
		item.Status = "warning"
		item.Message = "robots.txt could not be fetched; crawler proceeded with default allow behavior."
		return item
	}
	if !robots.ParseOK {
		item.Status = "fail"
		item.Message = "robots.txt was fetched but could not be parsed."
		item.Details = robots.ParseError
		return item
	}
	if audit.RobotsBlocksAll {
		item.Status = "fail"
		item.Message = "robots.txt appears to block all crawling."
		return item
	}
	if len(audit.Pages) == 0 {
		item.Status = "warning"
		item.Message = "robots.txt parsed, but no pages were crawled."
		return item
	}
	if seedURL != "" {
		item.Evidence = append(item.Evidence, "seed tested: "+seedURL)
	}
	if audit.RobotsSitemapDir {
		item.Evidence = append(item.Evidence, "robots.txt includes Sitemap directive")
	}
	return item
}

func analyzeParameterDisallows(audit *models.SiteAudit, robots robotsEvidence, req CrawlerEvidenceRunRequest) CrawlerEvidenceItem {
	params := discoveredParameterNames(audit.Pages)
	for _, name := range normalizeTokenList(req.ExpectedParameterNames) {
		params[name] = true
	}
	names := sortedKeys(params)
	item := CrawlerEvidenceItem{
		ID:       "ROBOTS-020",
		Name:     "Disallow rules for parameter URLs kept updated",
		Category: "Crawler Evidence",
		Status:   "info",
		Message:  "No parameter URLs or expected parameter names were supplied.",
		Evidence: names,
	}
	if len(names) == 0 {
		return item
	}
	if !robots.Exists {
		item.Status = "warning"
		item.Message = "Parameter URLs exist, but robots.txt could not be fetched."
		return item
	}
	covered, missing := parameterCoverage(robots.Raw, names)
	if len(missing) == 0 {
		item.Status = "pass"
		item.Message = "Parameter patterns are covered by robots.txt disallow rules."
		item.Evidence = covered
		return item
	}
	item.Status = "warning"
	item.Message = "Some discovered or expected URL parameters are not covered by robots.txt disallow rules."
	item.Evidence = append([]string{"covered: " + strings.Join(covered, ", ")}, "missing: "+strings.Join(missing, ", "))
	return item
}

func analyzeSitemapInventory(audit *models.SiteAudit, expected []string) CrawlerEvidenceItem {
	expectedURLs := normalizeURLList(expected)
	sitemapSet := map[string]bool{}
	for _, sitemapURL := range audit.SitemapURLs {
		sitemapSet[crawler.DedupeKey(sitemapURL)] = true
	}
	item := CrawlerEvidenceItem{
		ID:       "SITEMAP-022",
		Name:     "CDP URLs updated in sitemap on new inventory",
		Category: "Crawler Evidence",
		Status:   "needs_input",
		Message:  "Sitemap was discovered, but expected inventory URLs were not supplied.",
		Evidence: []string{
			fmt.Sprintf("sitemap URLs discovered: %d", len(sitemapSet)),
			fmt.Sprintf("crawled pages in sitemap: %d", countPagesInSitemap(audit.Pages)),
		},
	}
	if len(sitemapSet) == 0 {
		item.Status = "warning"
		item.Message = "No sitemap URLs were discovered during the crawl."
	}
	if len(expectedURLs) == 0 {
		return item
	}
	missing := []string{}
	for _, expectedURL := range expectedURLs {
		if !sitemapSet[crawler.DedupeKey(expectedURL)] {
			missing = append(missing, expectedURL)
		}
	}
	if len(missing) == 0 {
		item.Status = "pass"
		item.Message = "All supplied expected inventory URLs are present in the sitemap."
		item.Evidence = []string{fmt.Sprintf("expected inventory URLs checked: %d", len(expectedURLs))}
		return item
	}
	item.Status = "fail"
	item.Message = "Some expected inventory URLs are missing from the sitemap."
	item.Evidence = firstN(missing, 12)
	if len(missing) > 12 {
		item.Details = fmt.Sprintf("%d more missing URLs omitted", len(missing)-12)
	}
	return item
}

func analyzeImageCDN(audit *models.SiteAudit, allowedHosts []string) CrawlerEvidenceItem {
	allowed := map[string]bool{}
	for _, host := range normalizeHostList(allowedHosts) {
		allowed[host] = true
	}
	total := 0
	broken := []string{}
	nonImage := []string{}
	hostMismatch := []string{}
	for _, page := range audit.Pages {
		for _, img := range page.Images {
			if strings.TrimSpace(img.Src) == "" {
				continue
			}
			total++
			if img.StatusCode == 0 || img.StatusCode >= 400 {
				broken = append(broken, img.Src)
			}
			if img.ContentType != "" && !strings.Contains(strings.ToLower(img.ContentType), "image/") {
				nonImage = append(nonImage, img.Src+" content-type="+img.ContentType)
			}
			if len(allowed) > 0 {
				host := hostname(img.Src)
				if host != "" && !allowed[host] {
					hostMismatch = append(hostMismatch, img.Src)
				}
			}
		}
	}
	item := CrawlerEvidenceItem{
		ID:       "IMG-013",
		Name:     "Image CDN configured correctly",
		Category: "Crawler Evidence",
		Status:   "pass",
		Message:  fmt.Sprintf("%d images checked successfully.", total),
	}
	if total == 0 {
		item.Status = "info"
		item.Message = "No images were discovered in the bounded crawl."
		return item
	}
	if len(allowed) == 0 {
		item.Status = "needs_input"
		item.Message = "Images were checked, but no allowed CDN hosts were supplied."
		item.Evidence = discoveredImageHosts(audit.Pages)
		return item
	}
	if len(broken) > 0 || len(nonImage) > 0 || len(hostMismatch) > 0 {
		item.Status = "fail"
		item.Message = "Image CDN evidence found broken images, non-image responses, or host mismatches."
		item.Evidence = appendEvidenceSections(
			"broken", broken,
			"non_image", nonImage,
			"host_mismatch", hostMismatch,
		)
		return item
	}
	item.Evidence = []string{"allowed hosts: " + strings.Join(sortedKeys(allowed), ", ")}
	return item
}

func analyzeLiveContent(audit *models.SiteAudit, requiredTerms []string) CrawlerEvidenceItem {
	terms := normalizeTokenList(requiredTerms)
	item := CrawlerEvidenceItem{
		ID:       "INDEX-015",
		Name:     "Post-cache-purge content updates reflect live",
		Category: "Crawler Evidence",
		Status:   "needs_input",
		Message:  "Live pages were fetched and content hashes were captured. Add required snippets to verify a deploy or purge.",
		Evidence: firstPageHashes(audit.Pages, 8),
	}
	if len(terms) == 0 {
		return item
	}
	missing := []string{}
	combined := strings.ToLower(combinePageText(audit.Pages))
	for _, term := range terms {
		if !strings.Contains(combined, strings.ToLower(term)) {
			missing = append(missing, term)
		}
	}
	if len(missing) == 0 {
		item.Status = "pass"
		item.Message = "All required live text snippets were found in the crawled pages."
		item.Evidence = terms
		return item
	}
	item.Status = "fail"
	item.Message = "Some required live text snippets were not found in the crawled pages."
	item.Evidence = missing
	return item
}

func summarizeCrawlerEvidence(audit *models.SiteAudit, report []CrawlerEvidenceItem) CrawlerEvidenceSummary {
	s := CrawlerEvidenceSummary{
		PagesCrawled:    audit.PagesCrawled,
		SitemapURLs:     len(audit.SitemapURLs),
		ImagesChecked:   countImages(audit.Pages),
		ParameterURLCnt: len(discoveredParameterURLs(audit.Pages)),
	}
	for _, item := range report {
		switch item.Status {
		case "pass":
			s.Pass++
		case "warning":
			s.Warning++
		case "fail":
			s.Fail++
		case "needs_input":
			s.NeedsInput++
		default:
			s.Info++
		}
	}
	return s
}

func summarizeCrawlerEvidencePages(pages []*models.PageData) []CrawlerEvidencePage {
	out := make([]CrawlerEvidencePage, 0, minInt(len(pages), 40))
	for _, page := range pages {
		out = append(out, CrawlerEvidencePage{
			URL:          page.FinalURL,
			StatusCode:   page.StatusCode,
			Title:        page.Title,
			ContentHash:  pageContentHash(page),
			InSitemap:    page.InSitemap,
			ImageCount:   len(page.Images),
			HTMLSizeByte: page.HTMLSizeBytes,
		})
		if len(out) >= 40 {
			break
		}
	}
	return out
}

func fetchRobotsEvidence(ctx context.Context, siteURL string, timeout time.Duration) robotsEvidence {
	robotsURL := strings.TrimRight(crawler.OriginOf(siteURL), "/") + "/robots.txt"
	out := robotsEvidence{URL: robotsURL}
	client := http.Client{Timeout: timeout}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, robotsURL, nil)
	if err != nil {
		out.ParseError = err.Error()
		return out
	}
	req.Header.Set("User-Agent", defaultUserAgent)
	resp, err := client.Do(req)
	if err != nil {
		out.ParseError = err.Error()
		return out
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		out.ParseError = fmt.Sprintf("robots.txt returned HTTP %d", resp.StatusCode)
		return out
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, 1024*1024))
	if err != nil {
		out.ParseError = err.Error()
		return out
	}
	out.Exists = true
	out.Raw = string(body)
	_, err = robotstxt.FromString(out.Raw)
	if err != nil {
		out.ParseError = err.Error()
		return out
	}
	out.ParseOK = true
	return out
}

func discoveredParameterNames(pages []*models.PageData) map[string]bool {
	out := map[string]bool{}
	for _, rawURL := range discoveredParameterURLs(pages) {
		parsed, err := url.Parse(rawURL)
		if err != nil {
			continue
		}
		for key := range parsed.Query() {
			key = strings.ToLower(strings.TrimSpace(key))
			if key != "" {
				out[key] = true
			}
		}
	}
	return out
}

func discoveredParameterURLs(pages []*models.PageData) []string {
	seen := map[string]bool{}
	for _, page := range pages {
		for _, rawURL := range []string{page.URL, page.FinalURL} {
			if strings.Contains(rawURL, "?") {
				seen[rawURL] = true
			}
		}
		for _, link := range page.Links {
			if link.IsInternal && strings.Contains(link.URL, "?") {
				seen[link.URL] = true
			}
		}
	}
	return sortedKeys(seen)
}

func parameterCoverage(rawRobots string, names []string) ([]string, []string) {
	lower := strings.ToLower(rawRobots)
	covered := []string{}
	missing := []string{}
	for _, name := range names {
		if strings.Contains(lower, "?"+strings.ToLower(name)+"=") ||
			strings.Contains(lower, "&"+strings.ToLower(name)+"=") ||
			strings.Contains(lower, "*"+strings.ToLower(name)+"=") ||
			strings.Contains(lower, strings.ToLower(name)+"=") {
			covered = append(covered, name)
		} else {
			missing = append(missing, name)
		}
	}
	return covered, missing
}

func normalizeURLList(values []string) []string {
	seen := map[string]bool{}
	out := []string{}
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		key := crawler.DedupeKey(value)
		if !seen[key] {
			seen[key] = true
			out = append(out, value)
		}
	}
	return out
}

func normalizeTokenList(values []string) []string {
	seen := map[string]bool{}
	out := []string{}
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		key := strings.ToLower(value)
		if !seen[key] {
			seen[key] = true
			out = append(out, value)
		}
	}
	return out
}

func normalizeHostList(values []string) []string {
	seen := map[string]bool{}
	for _, value := range values {
		host := hostname(value)
		if host == "" {
			host = strings.ToLower(strings.TrimSpace(value))
		}
		host = strings.TrimPrefix(host, "www.")
		if host != "" {
			seen[host] = true
		}
	}
	return sortedKeys(seen)
}

func hostname(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	if !strings.Contains(raw, "://") {
		raw = "https://" + raw
	}
	parsed, err := url.Parse(raw)
	if err != nil {
		return ""
	}
	return strings.TrimPrefix(strings.ToLower(parsed.Hostname()), "www.")
}

func countPagesInSitemap(pages []*models.PageData) int {
	count := 0
	for _, page := range pages {
		if page.InSitemap {
			count++
		}
	}
	return count
}

func countImages(pages []*models.PageData) int {
	count := 0
	for _, page := range pages {
		count += len(page.Images)
	}
	return count
}

func discoveredImageHosts(pages []*models.PageData) []string {
	hosts := map[string]bool{}
	for _, page := range pages {
		for _, img := range page.Images {
			if host := hostname(img.Src); host != "" {
				hosts[host] = true
			}
		}
	}
	return sortedKeys(hosts)
}

func appendEvidenceSections(labelA string, a []string, labelB string, b []string, labelC string, c []string) []string {
	out := []string{}
	for _, section := range []struct {
		label string
		items []string
	}{
		{labelA, a},
		{labelB, b},
		{labelC, c},
	} {
		if len(section.items) == 0 {
			continue
		}
		out = append(out, section.label+":")
		out = append(out, firstN(section.items, 8)...)
	}
	return out
}

func firstPageHashes(pages []*models.PageData, limit int) []string {
	out := []string{}
	for _, page := range pages {
		if page.FinalURL == "" {
			continue
		}
		out = append(out, page.FinalURL+" hash="+pageContentHash(page))
		if len(out) >= limit {
			break
		}
	}
	return out
}

func combinePageText(pages []*models.PageData) string {
	var sb strings.Builder
	for _, page := range pages {
		sb.WriteString(page.Title)
		sb.WriteString(" ")
		sb.WriteString(strings.Join(page.H1s, " "))
		sb.WriteString(" ")
		sb.WriteString(page.BodyText)
		sb.WriteString(" ")
	}
	return sb.String()
}

func pageContentHash(page *models.PageData) string {
	source := page.RawHTML
	if source == "" {
		source = page.Title + "\n" + strings.Join(page.H1s, "\n") + "\n" + page.BodyText
	}
	sum := sha256.Sum256([]byte(source))
	return hex.EncodeToString(sum[:])[:16]
}

func sortedKeys(m map[string]bool) []string {
	out := make([]string, 0, len(m))
	for key := range m {
		out = append(out, key)
	}
	sort.Strings(out)
	return out
}

func firstN(values []string, n int) []string {
	if len(values) <= n {
		return append([]string(nil), values...)
	}
	return append([]string(nil), values[:n]...)
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
