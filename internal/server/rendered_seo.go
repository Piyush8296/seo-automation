package server

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"time"

	"github.com/cars24/seo-automation/internal/crawler"
	"github.com/cars24/seo-automation/internal/models"
)

const (
	defaultRenderedSampleLimit = 5
	maxRenderedSampleLimit     = 20
	defaultRenderedTimeout     = 20 * time.Second
)

type renderedSEOProbeInput struct {
	Pages     []renderedSEOPageInput `json:"pages"`
	TimeoutMS int                    `json:"timeout_ms"`
	UserAgent string                 `json:"user_agent,omitempty"`
}

type renderedSEOPageInput struct {
	URL                  string   `json:"url"`
	RawTitle             string   `json:"raw_title,omitempty"`
	RawH1s               []string `json:"raw_h1s,omitempty"`
	RawCanonical         string   `json:"raw_canonical,omitempty"`
	RawWordCount         int      `json:"raw_word_count,omitempty"`
	RawInternalLinkCount int      `json:"raw_internal_link_count,omitempty"`
	RawSchemaCount       int      `json:"raw_schema_count,omitempty"`
	RawScriptCount       int      `json:"raw_script_count,omitempty"`
	RawHashRouteLinks    int      `json:"raw_hash_route_links,omitempty"`
	RawLazyImages        int      `json:"raw_lazy_images,omitempty"`
}

type renderedSEOProbeResponse struct {
	Renderer string                 `json:"renderer,omitempty"`
	Pages    []renderedSEOProbePage `json:"pages"`
	Errors   []string               `json:"errors,omitempty"`
	Debug    map[string]interface{} `json:"debug,omitempty"`
}

type renderedSEOProbePage struct {
	URL                       string   `json:"url"`
	FinalURL                  string   `json:"final_url,omitempty"`
	Error                     string   `json:"error,omitempty"`
	RenderTimeMS              int      `json:"render_time_ms,omitempty"`
	Title                     string   `json:"title,omitempty"`
	MetaDescription           string   `json:"meta_description,omitempty"`
	Canonical                 string   `json:"canonical,omitempty"`
	H1s                       []string `json:"h1s,omitempty"`
	WordCount                 int      `json:"word_count,omitempty"`
	InternalLinkCount         int      `json:"internal_link_count,omitempty"`
	SchemaCount               int      `json:"schema_count,omitempty"`
	InvalidSchemaCount        int      `json:"invalid_schema_count,omitempty"`
	HashRouteLinkCount        int      `json:"hash_route_link_count,omitempty"`
	LoadMoreButtonCount       int      `json:"load_more_button_count,omitempty"`
	BeforeScrollWordCount     int      `json:"before_scroll_word_count,omitempty"`
	AfterScrollWordCount      int      `json:"after_scroll_word_count,omitempty"`
	BeforeScrollInternalLinks int      `json:"before_scroll_internal_links,omitempty"`
	AfterScrollInternalLinks  int      `json:"after_scroll_internal_links,omitempty"`
	BeforeScrollImages        int      `json:"before_scroll_images,omitempty"`
	AfterScrollImages         int      `json:"after_scroll_images,omitempty"`
	ThirdPartyScriptCount     int      `json:"third_party_script_count,omitempty"`
	ThirdPartyScriptHosts     []string `json:"third_party_script_hosts,omitempty"`
	AnalyticsHits             []string `json:"analytics_hits,omitempty"`
	GA4MeasurementIDs         []string `json:"ga4_measurement_ids,omitempty"`
	RequestFailures           []string `json:"request_failures,omitempty"`
	ConsoleErrors             []string `json:"console_errors,omitempty"`
}

type renderedSEOPagePair struct {
	Raw      renderedSEOPageInput
	Rendered renderedSEOProbePage
}

func renderedSEOForAudit(ctx context.Context, audit *models.SiteAudit, req StartAuditRequest, auditTimeout time.Duration) []models.EvidenceCheckResult {
	inputs := renderedSEOPageInputs(audit.Pages, req.RenderedSampleLimit)
	if len(inputs) == 0 {
		return renderedSEOUnavailable("No HTML pages were available for rendered SEO sampling.")
	}

	timeout := renderedTimeout(req.RenderedTimeout, auditTimeout)
	probeCtx, cancel := context.WithTimeout(ctx, timeout*time.Duration(len(inputs)+1))
	defer cancel()

	probe, err := runRenderedSEOProbe(probeCtx, inputs, timeout, req.UserAgent)
	if err != nil {
		return renderedSEOUnavailable(err.Error())
	}
	return analyzeRenderedSEO(inputs, probe)
}

func renderedSEOUnavailable(reason string) []models.EvidenceCheckResult {
	items := []models.EvidenceCheckResult{}
	for _, check := range renderedSEOChecks() {
		status := "warning"
		if check.ID == "JS-010" {
			status = "needs_input"
		}
		items = append(items, models.EvidenceCheckResult{
			ID:       check.ID,
			Name:     check.Name,
			Category: "Rendered SEO",
			Status:   status,
			Message:  "Rendered crawl evidence could not be captured.",
			Details:  reason,
		})
	}
	return items
}

type renderedSEOCheck struct {
	ID   string
	Name string
}

func renderedSEOChecks() []renderedSEOCheck {
	return []renderedSEOCheck{
		{ID: "JS-001", Name: "Critical SEO content not JavaScript-rendered only"},
		{ID: "JS-002", Name: "Internal links in JavaScript are crawlable"},
		{ID: "JS-003", Name: "React/Vue/Angular site has SSR or pre-rendering"},
		{ID: "JS-004", Name: "Google rendering of page verified"},
		{ID: "JS-005", Name: "Structured data in JavaScript is rendered correctly"},
		{ID: "JS-006", Name: "Infinite scroll loads content accessible to crawlers"},
		{ID: "JS-007", Name: "SPA navigation uses History API, not hash URLs"},
		{ID: "JS-008", Name: "Lazy-loaded content is crawlable"},
		{ID: "JS-009", Name: "Third-party JS not blocking critical rendering"},
		{ID: "JS-010", Name: "JS rendering tested with URL inspection evidence"},
	}
}

func renderedSEOPageInputs(pages []*models.PageData, requestedLimit int) []renderedSEOPageInput {
	limit := requestedLimit
	if limit <= 0 {
		limit = defaultRenderedSampleLimit
	}
	if limit > maxRenderedSampleLimit {
		limit = maxRenderedSampleLimit
	}

	type candidate struct {
		page  *models.PageData
		score int
	}
	candidates := []candidate{}
	for _, page := range pages {
		if page == nil || page.StatusCode >= 400 || !strings.Contains(strings.ToLower(page.ContentType), "html") {
			continue
		}
		pageURL := page.FinalURL
		if pageURL == "" {
			pageURL = page.URL
		}
		if pageURL == "" {
			continue
		}
		score := 0
		if page.Depth == 0 {
			score += 100
		}
		score += page.ExternalScriptCount * 5
		score += page.RenderBlockingScripts * 8
		score += len(page.SchemaJSONRaw) * 4
		if page.WordCount < 120 {
			score += 12
		}
		score += countHashRouteLinks(page.Links) * 8
		score += countLazyImages(page.Images) * 2
		candidates = append(candidates, candidate{page: page, score: score})
	}

	sort.SliceStable(candidates, func(i, j int) bool {
		if candidates[i].score == candidates[j].score {
			return candidates[i].page.Depth < candidates[j].page.Depth
		}
		return candidates[i].score > candidates[j].score
	})

	out := []renderedSEOPageInput{}
	seen := map[string]bool{}
	for _, c := range candidates {
		page := c.page
		pageURL := page.FinalURL
		if pageURL == "" {
			pageURL = page.URL
		}
		key := crawler.DedupeKey(pageURL)
		if seen[key] {
			continue
		}
		seen[key] = true
		out = append(out, renderedSEOPageInput{
			URL:                  pageURL,
			RawTitle:             page.Title,
			RawH1s:               page.H1s,
			RawCanonical:         page.Canonical,
			RawWordCount:         page.WordCount,
			RawInternalLinkCount: countInternalLinks(page.Links),
			RawSchemaCount:       len(page.SchemaJSONRaw),
			RawScriptCount:       page.ExternalScriptCount,
			RawHashRouteLinks:    countHashRouteLinks(page.Links),
			RawLazyImages:        countLazyImages(page.Images),
		})
		if len(out) >= limit {
			break
		}
	}
	return out
}

func runRenderedSEOProbe(ctx context.Context, pages []renderedSEOPageInput, timeout time.Duration, userAgent string) (renderedSEOProbeResponse, error) {
	scriptPath, err := renderedSEOProbeScriptPath()
	if err != nil {
		return renderedSEOProbeResponse{}, err
	}
	nodePath := os.Getenv("SEO_RENDER_NODE")
	if strings.TrimSpace(nodePath) == "" {
		nodePath, err = exec.LookPath("node")
		if err != nil {
			return renderedSEOProbeResponse{}, errors.New("node runtime not found in PATH; set SEO_RENDER_NODE to enable rendered crawl checks")
		}
	}
	timeoutMS := int(timeout / time.Millisecond)
	if timeoutMS <= 0 {
		timeoutMS = int(defaultRenderedTimeout / time.Millisecond)
	}
	payload, err := json.Marshal(renderedSEOProbeInput{
		Pages:     pages,
		TimeoutMS: timeoutMS,
		UserAgent: userAgent,
	})
	if err != nil {
		return renderedSEOProbeResponse{}, err
	}
	cmd := exec.CommandContext(ctx, nodePath, scriptPath)
	cmd.Stdin = bytes.NewReader(payload)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, "NODE_PATH="+filepath.Join(projectRootDir(), "node_modules"))
	if err := cmd.Run(); err != nil {
		msg := strings.TrimSpace(stderr.String())
		if msg == "" {
			msg = err.Error()
		}
		return renderedSEOProbeResponse{}, fmt.Errorf("rendered SEO probe failed: %s", msg)
	}
	var response renderedSEOProbeResponse
	if err := json.Unmarshal(stdout.Bytes(), &response); err != nil {
		return renderedSEOProbeResponse{}, fmt.Errorf("rendered SEO probe returned invalid JSON: %w", err)
	}
	if len(response.Pages) == 0 && len(response.Errors) > 0 {
		return response, fmt.Errorf("rendered SEO probe failed: %s", strings.Join(response.Errors, "; "))
	}
	return response, nil
}

func renderedSEOProbeScriptPath() (string, error) {
	candidates := []string{}
	if cwd, err := os.Getwd(); err == nil {
		candidates = append(candidates, filepath.Join(cwd, "scripts", "rendered-seo-probe.mjs"))
	}
	candidates = append(candidates, filepath.Join(projectRootDir(), "scripts", "rendered-seo-probe.mjs"))
	for _, path := range candidates {
		if _, err := os.Stat(path); err == nil {
			return path, nil
		}
	}
	return "", fmt.Errorf("rendered SEO probe script missing; checked %s", strings.Join(candidates, ", "))
}

func projectRootDir() string {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		return "."
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
}

func renderedTimeout(raw string, auditTimeout time.Duration) time.Duration {
	if parsed, err := time.ParseDuration(strings.TrimSpace(raw)); err == nil && parsed > 0 {
		return parsed
	}
	if auditTimeout > 0 {
		if auditTimeout < 10*time.Second {
			return 10 * time.Second
		}
		if auditTimeout > 45*time.Second {
			return 45 * time.Second
		}
		return auditTimeout
	}
	return defaultRenderedTimeout
}

func analyzeRenderedSEO(rawPages []renderedSEOPageInput, probe renderedSEOProbeResponse) []models.EvidenceCheckResult {
	pairs := pairRenderedPages(rawPages, probe.Pages)
	if len(pairs) == 0 {
		return renderedSEOUnavailable("Rendered probe finished, but no page evidence was returned.")
	}
	return []models.EvidenceCheckResult{
		analyzeCriticalSEOContent(pairs),
		analyzeRenderedInternalLinks(pairs),
		analyzeSSRPrerendering(pairs),
		analyzeRenderingVerified(pairs, probe),
		analyzeRenderedStructuredData(pairs),
		analyzeInfiniteScroll(pairs),
		analyzeHashRouting(pairs),
		analyzeLazyLoadedContent(pairs),
		analyzeThirdPartyJS(pairs),
		analyzeURLInspectionEvidence(pairs, probe),
	}
}

func pairRenderedPages(rawPages []renderedSEOPageInput, rendered []renderedSEOProbePage) []renderedSEOPagePair {
	byURL := map[string]renderedSEOPageInput{}
	for _, page := range rawPages {
		byURL[crawler.DedupeKey(page.URL)] = page
	}
	pairs := []renderedSEOPagePair{}
	for _, page := range rendered {
		raw, ok := byURL[crawler.DedupeKey(page.URL)]
		if !ok && page.FinalURL != "" {
			raw, ok = byURL[crawler.DedupeKey(page.FinalURL)]
		}
		if !ok {
			raw = renderedSEOPageInput{URL: page.URL}
		}
		pairs = append(pairs, renderedSEOPagePair{Raw: raw, Rendered: page})
	}
	return pairs
}

func analyzeCriticalSEOContent(pairs []renderedSEOPagePair) models.EvidenceCheckResult {
	missingRaw := []string{}
	renderErrors := []string{}
	for _, pair := range pairs {
		if pair.Rendered.Error != "" {
			renderErrors = append(renderErrors, pair.Raw.URL+": "+pair.Rendered.Error)
			continue
		}
		missing := []string{}
		if strings.TrimSpace(pair.Raw.RawTitle) == "" && strings.TrimSpace(pair.Rendered.Title) != "" {
			missing = append(missing, "title")
		}
		if len(pair.Raw.RawH1s) == 0 && len(pair.Rendered.H1s) > 0 {
			missing = append(missing, "h1")
		}
		if strings.TrimSpace(pair.Raw.RawCanonical) == "" && strings.TrimSpace(pair.Rendered.Canonical) != "" {
			missing = append(missing, "canonical")
		}
		if len(missing) > 0 {
			missingRaw = append(missingRaw, pair.Raw.URL+" rendered-only: "+strings.Join(missing, ", "))
		}
	}
	item := evidenceItem("JS-001", "Critical SEO content not JavaScript-rendered only", "pass", "Critical title, H1, and canonical signals were present in raw HTML for sampled pages.")
	item.Evidence = sampledURLs(pairs)
	if len(renderErrors) > 0 {
		item.Status = "warning"
		item.Message = "Some sampled pages could not be rendered, so critical SEO content could not be fully verified."
		item.Evidence = firstN(renderErrors, 8)
		return item
	}
	if len(missingRaw) > 0 {
		item.Status = "fail"
		item.Message = "Some critical SEO elements only appeared after JavaScript rendering."
		item.Evidence = firstN(missingRaw, 8)
	}
	return item
}

func analyzeRenderedInternalLinks(pairs []renderedSEOPagePair) models.EvidenceCheckResult {
	risky := []string{}
	for _, pair := range pairs {
		if pair.Rendered.Error != "" {
			continue
		}
		rawCount := pair.Raw.RawInternalLinkCount
		renderedCount := pair.Rendered.InternalLinkCount
		if rawCount == 0 && renderedCount >= 5 {
			risky = append(risky, fmt.Sprintf("%s raw_links=%d rendered_links=%d", pair.Raw.URL, rawCount, renderedCount))
			continue
		}
		if rawCount > 0 && renderedCount > rawCount+(rawCount/2) && renderedCount-rawCount >= 10 {
			risky = append(risky, fmt.Sprintf("%s raw_links=%d rendered_links=%d", pair.Raw.URL, rawCount, renderedCount))
		}
	}
	item := evidenceItem("JS-002", "Internal links in JavaScript are crawlable", "pass", "Sampled pages expose internal links in raw HTML without heavy rendered-only deltas.")
	if len(risky) > 0 {
		item.Status = "warning"
		item.Message = "Some pages expose many internal links only after JavaScript rendering."
		item.Evidence = firstN(risky, 8)
		return item
	}
	item.Evidence = rawRenderedLinkEvidence(pairs)
	return item
}

func analyzeSSRPrerendering(pairs []renderedSEOPagePair) models.EvidenceCheckResult {
	weak := []string{}
	for _, pair := range pairs {
		if pair.Rendered.Error != "" {
			continue
		}
		if pair.Rendered.WordCount >= 80 && pair.Raw.RawWordCount < pair.Rendered.WordCount/2 {
			weak = append(weak, fmt.Sprintf("%s raw_words=%d rendered_words=%d", pair.Raw.URL, pair.Raw.RawWordCount, pair.Rendered.WordCount))
		}
	}
	item := evidenceItem("JS-003", "React/Vue/Angular site has SSR or pre-rendering", "pass", "Raw HTML content is broadly aligned with rendered content on sampled pages.")
	item.Evidence = rawRenderedWordEvidence(pairs)
	if len(weak) > 0 {
		item.Status = "warning"
		item.Message = "Some sampled pages look thin in raw HTML compared with the rendered DOM."
		item.Evidence = firstN(weak, 8)
	}
	return item
}

func analyzeRenderingVerified(pairs []renderedSEOPagePair, probe renderedSEOProbeResponse) models.EvidenceCheckResult {
	rendered := 0
	failed := []string{}
	for _, pair := range pairs {
		if pair.Rendered.Error != "" {
			failed = append(failed, pair.Raw.URL+": "+pair.Rendered.Error)
			continue
		}
		rendered++
	}
	item := evidenceItem("JS-004", "Google rendering of page verified", "pass", fmt.Sprintf("Browser-rendered evidence captured for %d sampled pages.", rendered))
	item.Evidence = renderedTimingEvidence(pairs)
	if rendered == 0 {
		item.Status = "fail"
		item.Message = "No sampled pages rendered successfully."
		item.Evidence = append(firstN(failed, 8), firstN(probe.Errors, 4)...)
		return item
	}
	if len(failed) > 0 {
		item.Status = "warning"
		item.Message = fmt.Sprintf("%d sampled pages rendered successfully; %d failed.", rendered, len(failed))
		item.Evidence = append(renderedTimingEvidence(pairs), firstN(failed, 4)...)
	}
	return item
}

func analyzeRenderedStructuredData(pairs []renderedSEOPagePair) models.EvidenceCheckResult {
	invalid := []string{}
	renderedOnly := []string{}
	totalRendered := 0
	totalRaw := 0
	for _, pair := range pairs {
		if pair.Rendered.Error != "" {
			continue
		}
		totalRendered += pair.Rendered.SchemaCount
		totalRaw += pair.Raw.RawSchemaCount
		if pair.Rendered.InvalidSchemaCount > 0 {
			invalid = append(invalid, fmt.Sprintf("%s invalid_jsonld=%d", pair.Raw.URL, pair.Rendered.InvalidSchemaCount))
		}
		if pair.Raw.RawSchemaCount == 0 && pair.Rendered.SchemaCount > 0 {
			renderedOnly = append(renderedOnly, fmt.Sprintf("%s rendered_schema=%d", pair.Raw.URL, pair.Rendered.SchemaCount))
		}
	}
	item := evidenceItem("JS-005", "Structured data in JavaScript is rendered correctly", "pass", "Rendered JSON-LD parsed successfully on sampled pages.")
	item.Evidence = []string{fmt.Sprintf("raw schema objects: %d", totalRaw), fmt.Sprintf("rendered schema objects: %d", totalRendered)}
	if totalRendered == 0 && totalRaw == 0 {
		item.Status = "info"
		item.Message = "No JSON-LD structured data was present in raw or rendered sampled pages."
		return item
	}
	if len(invalid) > 0 {
		item.Status = "fail"
		item.Message = "Rendered JSON-LD contains invalid JSON on sampled pages."
		item.Evidence = firstN(invalid, 8)
		return item
	}
	if len(renderedOnly) > 0 {
		item.Status = "warning"
		item.Message = "Some structured data appears only after JavaScript rendering."
		item.Evidence = firstN(renderedOnly, 8)
	}
	return item
}

func analyzeInfiniteScroll(pairs []renderedSEOPagePair) models.EvidenceCheckResult {
	suspects := []string{}
	for _, pair := range pairs {
		if pair.Rendered.Error != "" {
			continue
		}
		linkGrowth := pair.Rendered.AfterScrollInternalLinks - pair.Rendered.BeforeScrollInternalLinks
		wordGrowth := pair.Rendered.AfterScrollWordCount - pair.Rendered.BeforeScrollWordCount
		if pair.Rendered.LoadMoreButtonCount > 0 || linkGrowth >= 10 || wordGrowth >= 150 {
			suspects = append(suspects, fmt.Sprintf("%s load_more=%d link_growth=%d word_growth=%d", pair.Raw.URL, pair.Rendered.LoadMoreButtonCount, linkGrowth, wordGrowth))
		}
	}
	item := evidenceItem("JS-006", "Infinite scroll loads content accessible to crawlers", "pass", "No risky infinite-scroll or load-more pattern was detected in sampled pages.")
	if len(suspects) > 0 {
		item.Status = "warning"
		item.Message = "Some pages add meaningful content after scrolling or expose load-more controls; verify crawlable pagination exists."
		item.Evidence = firstN(suspects, 8)
	}
	return item
}

func analyzeHashRouting(pairs []renderedSEOPagePair) models.EvidenceCheckResult {
	hashRoutes := []string{}
	for _, pair := range pairs {
		total := pair.Raw.RawHashRouteLinks + pair.Rendered.HashRouteLinkCount
		if total > 0 {
			hashRoutes = append(hashRoutes, fmt.Sprintf("%s hash_route_links=%d", pair.Raw.URL, total))
		}
		if strings.Contains(pair.Rendered.FinalURL, "/#/") || strings.Contains(pair.Rendered.FinalURL, "#!") {
			hashRoutes = append(hashRoutes, pair.Raw.URL+" final_url="+pair.Rendered.FinalURL)
		}
	}
	item := evidenceItem("JS-007", "SPA navigation uses History API, not hash URLs", "pass", "No hash-route internal navigation was detected in sampled pages.")
	if len(hashRoutes) > 0 {
		item.Status = "fail"
		item.Message = "Hash-route navigation was detected; SEO URLs should use clean History API paths."
		item.Evidence = firstN(hashRoutes, 8)
	}
	return item
}

func analyzeLazyLoadedContent(pairs []renderedSEOPagePair) models.EvidenceCheckResult {
	lazy := []string{}
	for _, pair := range pairs {
		if pair.Rendered.Error != "" {
			continue
		}
		imageGrowth := pair.Rendered.AfterScrollImages - pair.Rendered.BeforeScrollImages
		if pair.Raw.RawLazyImages > 0 || imageGrowth > 0 {
			lazy = append(lazy, fmt.Sprintf("%s raw_lazy_images=%d image_growth_after_scroll=%d", pair.Raw.URL, pair.Raw.RawLazyImages, imageGrowth))
		}
	}
	item := evidenceItem("JS-008", "Lazy-loaded content is crawlable", "pass", "No crawler-risky lazy content pattern was detected in sampled pages.")
	if len(lazy) > 0 {
		item.Status = "pass"
		item.Message = "Lazy-loaded assets/content were detected and rendered during the browser pass."
		item.Evidence = firstN(lazy, 8)
	}
	return item
}

func analyzeThirdPartyJS(pairs []renderedSEOPagePair) models.EvidenceCheckResult {
	risky := []string{}
	failures := []string{}
	consoleErrors := []string{}
	hosts := map[string]bool{}
	analyticsHits := 0
	for _, pair := range pairs {
		if pair.Rendered.Error != "" {
			failures = append(failures, pair.Raw.URL+": "+pair.Rendered.Error)
			continue
		}
		for _, host := range pair.Rendered.ThirdPartyScriptHosts {
			hosts[host] = true
		}
		analyticsHits += len(pair.Rendered.AnalyticsHits)
		if pair.Rendered.ThirdPartyScriptCount > 12 || pair.Rendered.RenderTimeMS > 8000 {
			risky = append(risky, fmt.Sprintf("%s third_party_scripts=%d render_ms=%d", pair.Raw.URL, pair.Rendered.ThirdPartyScriptCount, pair.Rendered.RenderTimeMS))
		}
		if len(pair.Rendered.RequestFailures) > 0 {
			failures = append(failures, pair.Raw.URL+" failed_requests="+strings.Join(firstN(pair.Rendered.RequestFailures, 3), " | "))
		}
		if len(pair.Rendered.ConsoleErrors) > 0 {
			consoleErrors = append(consoleErrors, pair.Raw.URL+" console_errors="+strings.Join(firstN(pair.Rendered.ConsoleErrors, 3), " | "))
		}
	}
	item := evidenceItem("JS-009", "Third-party JS not blocking critical rendering", "pass", "Third-party scripts did not exceed risk thresholds in sampled rendering.")
	item.Evidence = []string{
		fmt.Sprintf("third-party script hosts: %s", strings.Join(sortedKeys(hosts), ", ")),
		fmt.Sprintf("analytics network hits observed: %d", analyticsHits),
	}
	if len(risky) > 0 || len(failures) > 0 || len(consoleErrors) > 0 {
		item.Status = "warning"
		item.Message = "Rendered crawl found third-party script volume, failed requests, slow renders, or console errors."
		item.Evidence = appendEvidenceSections("risk", risky, "request_failures", failures, "console_errors", consoleErrors)
	}
	return item
}

func analyzeURLInspectionEvidence(pairs []renderedSEOPagePair, probe renderedSEOProbeResponse) models.EvidenceCheckResult {
	rendered := 0
	ga4IDs := map[string]bool{}
	for _, pair := range pairs {
		if pair.Rendered.Error == "" {
			rendered++
		}
		for _, id := range pair.Rendered.GA4MeasurementIDs {
			ga4IDs[id] = true
		}
	}
	item := evidenceItem("JS-010", "JS rendering tested with URL inspection evidence", "pass", fmt.Sprintf("Rendered DOM evidence captured for %d sampled pages.", rendered))
	item.Details = "This is crawler-side browser evidence. Official Google URL Inspection evidence can be connected through GSC later."
	item.Evidence = []string{
		"renderer: " + firstNonEmpty(probe.Renderer, "playwright"),
		fmt.Sprintf("sampled pages: %d", len(pairs)),
	}
	if len(ga4IDs) > 0 {
		item.Evidence = append(item.Evidence, "GA4 measurement IDs observed: "+strings.Join(sortedKeys(ga4IDs), ", "))
	}
	if rendered == 0 {
		item.Status = "needs_input"
		item.Message = "No rendered DOM evidence was captured."
	}
	return item
}

func renderedSEOFindings(items []models.EvidenceCheckResult) []models.CheckResult {
	out := []models.CheckResult{}
	for _, item := range items {
		severity, ok := renderedSEOSeverity(item.Status)
		if !ok {
			continue
		}
		out = append(out, models.CheckResult{
			ID:       item.ID,
			Category: "Rendered SEO",
			Severity: severity,
			Message:  item.Message,
			URL:      "(site-wide)",
			Details:  crawlerEvidenceDetails(item),
			Platform: models.PlatformBoth,
		})
	}
	return out
}

func renderedSEOSeverity(status string) (models.Severity, bool) {
	switch status {
	case "fail":
		return models.SeverityError, true
	case "warning":
		return models.SeverityWarning, true
	default:
		return "", false
	}
}

func evidenceItem(id, name, status, message string) models.EvidenceCheckResult {
	return models.EvidenceCheckResult{
		ID:       id,
		Name:     name,
		Category: "Rendered SEO",
		Status:   status,
		Message:  message,
	}
}

func countInternalLinks(links []models.Link) int {
	count := 0
	for _, link := range links {
		if link.IsInternal {
			count++
		}
	}
	return count
}

func countHashRouteLinks(links []models.Link) int {
	count := 0
	for _, link := range links {
		if !link.IsInternal {
			continue
		}
		if strings.Contains(link.URL, "/#/") || strings.Contains(link.URL, "#!") {
			count++
			continue
		}
		parsed, err := url.Parse(link.URL)
		if err == nil && strings.HasPrefix(parsed.Fragment, "/") {
			count++
		}
	}
	return count
}

func countLazyImages(images []models.Image) int {
	count := 0
	for _, img := range images {
		if img.Loading == "lazy" {
			count++
		}
	}
	return count
}

func sampledURLs(pairs []renderedSEOPagePair) []string {
	out := []string{}
	for _, pair := range pairs {
		out = append(out, pair.Raw.URL)
	}
	return firstN(out, 8)
}

func rawRenderedLinkEvidence(pairs []renderedSEOPagePair) []string {
	out := []string{}
	for _, pair := range pairs {
		out = append(out, fmt.Sprintf("%s raw_links=%d rendered_links=%d", pair.Raw.URL, pair.Raw.RawInternalLinkCount, pair.Rendered.InternalLinkCount))
	}
	return firstN(out, 8)
}

func rawRenderedWordEvidence(pairs []renderedSEOPagePair) []string {
	out := []string{}
	for _, pair := range pairs {
		out = append(out, fmt.Sprintf("%s raw_words=%d rendered_words=%d", pair.Raw.URL, pair.Raw.RawWordCount, pair.Rendered.WordCount))
	}
	return firstN(out, 8)
}

func renderedTimingEvidence(pairs []renderedSEOPagePair) []string {
	out := []string{}
	for _, pair := range pairs {
		if pair.Rendered.Error != "" {
			continue
		}
		out = append(out, fmt.Sprintf("%s render_ms=%d final_url=%s", pair.Raw.URL, pair.Rendered.RenderTimeMS, firstNonEmpty(pair.Rendered.FinalURL, pair.Rendered.URL)))
	}
	return firstN(out, 8)
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}
