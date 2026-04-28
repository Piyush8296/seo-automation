package url_structure

import (
	"encoding/json"
	"fmt"
	"net/url"
	"regexp"
	"sort"
	"strings"
	"unicode"

	"github.com/cars24/seo-automation/internal/models"
)

var sessionParams = []string{"jsessionid", "sessionid", "phpsessid", "aspsessionid", "sid"}

var nonDescriptivePattern = regexp.MustCompile(`^/(p|page|post|node|item|product|article|id|detail)s?/\d+/?$`)

const maxURLLength = 2048

// Common English stop words that add no SEO value to URLs.
var stopWords = map[string]bool{
	"a": true, "an": true, "the": true, "and": true, "or": true,
	"but": true, "is": true, "in": true, "on": true, "at": true,
	"to": true, "for": true, "of": true, "with": true, "by": true,
	"from": true, "as": true, "into": true, "about": true, "that": true,
	"this": true, "it": true, "not": true, "are": true, "was": true,
	"were": true, "been": true, "be": true, "has": true, "have": true,
	"had": true, "do": true, "does": true, "did": true, "will": true,
	"would": true, "could": true, "should": true, "may": true, "might": true,
}

var genericURLWords = map[string]bool{
	"amp": true, "blog": true, "buy": true, "car": true, "cars": true,
	"category": true, "detail": true, "details": true, "filter": true, "home": true,
	"html": true, "htm": true, "index": true, "india": true, "listing": true,
	"new": true, "news": true, "offer": true, "offers": true, "page": true,
	"price": true, "prices": true, "product": true, "products": true, "sale": true,
	"search": true, "sell": true, "used": true, "view": true,
}

// PageChecks returns per-page URL structure checks.
func PageChecks() []models.PageCheck {
	return []models.PageCheck{
		&urlTooLong{},
		&urlHasUnderscores{},
		&urlHasUppercase{},
		&urlHasSpaces{},
		&urlHasSessionParams{},
		&urlTooManyParams{},
		&urlDoubleSlash{},
		&urlNonDescriptive{},
		&urlPathDepthTooDeep{},
		&urlContainsStopWords{},
		&urlKeywordTopicMismatch{},
		&urlBreadcrumbMismatch{},
		&urlPrintCanonicalConflict{},
		&urlNonASCII{},
	}
}

// SiteChecks returns crawl-level URL structure checks.
func SiteChecks() []models.SiteCheck {
	return []models.SiteCheck{
		&urlConsistentStructure{},
		&urlTrailingSlashInconsistent{},
	}
}

type urlTooLong struct{}

func (c *urlTooLong) Run(p *models.PageData) []models.CheckResult {
	if len(p.URL) > maxURLLength {
		return []models.CheckResult{{
			ID:       "url.too_long",
			Category: "URL Structure",
			Severity: models.SeverityWarning,
			Message:  fmt.Sprintf("URL too long (%d chars, max %d)", len(p.URL), maxURLLength),
			URL:      p.URL,
		}}
	}
	return nil
}

type urlHasUnderscores struct{}

func (c *urlHasUnderscores) Run(p *models.PageData) []models.CheckResult {
	parsed, err := url.Parse(p.URL)
	if err != nil {
		return nil
	}
	if strings.Contains(parsed.Path, "_") {
		return []models.CheckResult{{
			ID:       "url.has_underscores",
			Category: "URL Structure",
			Severity: models.SeverityNotice,
			Message:  "URL path contains underscores (prefer hyphens)",
			URL:      p.URL,
		}}
	}
	return nil
}

type urlKeywordTopicMismatch struct{}

func (c *urlKeywordTopicMismatch) Run(p *models.PageData) []models.CheckResult {
	words := meaningfulURLWords(pageURL(p))
	if len(words) < 2 {
		return nil
	}
	text := normalizeText(strings.Join(append([]string{p.Title, p.BodyText}, p.H1s...), " "))
	if text == "" {
		return nil
	}
	matched := 0
	for _, word := range words {
		if strings.Contains(text, word) {
			matched++
		}
	}
	if matched > 0 {
		return nil
	}
	return []models.CheckResult{{
		ID:       "url.keyword_topic_mismatch",
		Category: "URL Structure",
		Severity: models.SeverityNotice,
		Message:  "URL keywords do not appear to match the page topic",
		URL:      p.URL,
		Details:  "URL keywords: " + strings.Join(words, ", "),
	}}
}

type urlBreadcrumbMismatch struct{}

func (c *urlBreadcrumbMismatch) Run(p *models.PageData) []models.CheckResult {
	current, err := url.Parse(pageURL(p))
	if err != nil || current == nil {
		return nil
	}
	currentPath := normalizedURLPath(current)
	if currentPath == "" || currentPath == "/" {
		return nil
	}
	trails := breadcrumbTrails(p.SchemaJSONRaw)
	if len(trails) == 0 {
		return nil
	}
	pageText := normalizeText(strings.Join(append([]string{p.Title}, p.H1s...), " "))
	slugWords := meaningfulURLWords(current.String())
	for _, trail := range trails {
		if mismatch := breadcrumbTrailMismatch(trail, current, currentPath, pageText, slugWords); mismatch != "" {
			return []models.CheckResult{{
				ID:       "url.breadcrumb_mismatch",
				Category: "URL Structure",
				Severity: models.SeverityNotice,
				Message:  "Breadcrumb hierarchy does not match the current URL",
				URL:      p.URL,
				Details:  mismatch,
			}}
		}
	}
	return nil
}

type urlPrintCanonicalConflict struct{}

func (c *urlPrintCanonicalConflict) Run(p *models.PageData) []models.CheckResult {
	current := pageURL(p)
	if !looksLikePrintURL(current) {
		return nil
	}
	canonical := strings.TrimSpace(p.Canonical)
	if canonical == "" {
		return []models.CheckResult{{
			ID:       "url.print_canonical_conflict",
			Category: "URL Structure",
			Severity: models.SeverityWarning,
			Message:  "Print-friendly URL is missing a canonical URL to the main page",
			URL:      p.URL,
		}}
	}
	resolvedCanonical := resolveReference(canonical, current)
	if sameURLIgnoringTrailingSlash(resolvedCanonical, current) || looksLikePrintURL(resolvedCanonical) {
		return []models.CheckResult{{
			ID:       "url.print_canonical_conflict",
			Category: "URL Structure",
			Severity: models.SeverityWarning,
			Message:  "Print-friendly URL canonical points to a print URL instead of the main page",
			URL:      p.URL,
			Details:  "canonical=" + canonical,
		}}
	}
	return nil
}

type urlNonASCII struct{}

func (c *urlNonASCII) Run(p *models.PageData) []models.CheckResult {
	raw := pageURL(p)
	if raw == "" {
		return nil
	}
	if !urlHasNonASCII(raw) {
		return nil
	}
	return []models.CheckResult{{
		ID:       "url.non_ascii",
		Category: "URL Structure",
		Severity: models.SeverityWarning,
		Message:  "URL contains non-ASCII characters or encoded non-ASCII bytes",
		URL:      p.URL,
		Details:  "Use ASCII slugs to avoid encoding inconsistencies across crawlers, browsers, and reporting tools.",
	}}
}

type urlConsistentStructure struct{}

func (c *urlConsistentStructure) Run(pages []*models.PageData) []models.CheckResult {
	type sample struct {
		url   string
		depth int
	}
	sections := map[string][]sample{}
	for _, p := range pages {
		u := parsePageURL(p)
		if u == nil || !isIndexableStatus(p) || isStaticURLPath(u.Path) {
			continue
		}
		segments := pathSegments(u.Path)
		if len(segments) < 2 {
			continue
		}
		section := strings.ToLower(segments[0])
		sections[section] = append(sections[section], sample{url: u.String(), depth: len(segments)})
	}
	for section, samples := range sections {
		if len(samples) < 4 {
			continue
		}
		byDepth := map[int][]string{}
		minDepth, maxDepth := samples[0].depth, samples[0].depth
		for _, item := range samples {
			byDepth[item.depth] = append(byDepth[item.depth], item.url)
			if item.depth < minDepth {
				minDepth = item.depth
			}
			if item.depth > maxDepth {
				maxDepth = item.depth
			}
		}
		if maxDepth-minDepth < 2 {
			continue
		}
		shallow := byDepth[minDepth]
		deep := byDepth[maxDepth]
		if len(shallow) == 0 || len(deep) == 0 {
			continue
		}
		return []models.CheckResult{{
			ID:       "url.consistent_structure",
			Category: "URL Structure",
			Severity: models.SeverityNotice,
			Message:  fmt.Sprintf("URL section /%s uses inconsistent depth patterns", section),
			URL:      samples[0].url,
			Details:  fmt.Sprintf("depths=%d-%d examples=%s", minDepth, maxDepth, strings.Join(sampleStrings(append(shallow, deep...), 4), " | ")),
		}}
	}
	return nil
}

type urlTrailingSlashInconsistent struct{}

func (c *urlTrailingSlashInconsistent) Run(pages []*models.PageData) []models.CheckResult {
	byKey := map[string][]string{}
	withSlash := 0
	withoutSlash := 0
	for _, p := range pages {
		u := parsePageURL(p)
		if u == nil || !isIndexableStatus(p) || u.Path == "" || u.Path == "/" || isStaticURLPath(u.Path) {
			continue
		}
		if strings.HasSuffix(u.Path, "/") {
			withSlash++
		} else {
			withoutSlash++
		}
		key := trailingSlashStructureKey(u)
		if key != "" {
			byKey[key] = append(byKey[key], u.String())
		}
	}
	for _, variants := range byKey {
		variants = uniqueStrings(variants)
		if len(variants) > 1 {
			return []models.CheckResult{{
				ID:       "url.trailing_slash_inconsistent",
				Category: "URL Structure",
				Severity: models.SeverityWarning,
				Message:  "Same URL path is available with and without a trailing slash",
				URL:      variants[0],
				Details:  strings.Join(variants, " | "),
			}}
		}
	}
	if withSlash >= 2 && withoutSlash >= 2 {
		return []models.CheckResult{{
			ID:       "url.trailing_slash_inconsistent",
			Category: "URL Structure",
			Severity: models.SeverityNotice,
			Message:  "Site mixes trailing-slash and non-trailing-slash URL patterns",
			URL:      firstPageURL(pages),
			Details:  fmt.Sprintf("with_slash=%d without_slash=%d", withSlash, withoutSlash),
		}}
	}
	return nil
}

type breadcrumbTrail struct {
	Names []string
	URLs  []string
}

func pageURL(p *models.PageData) string {
	if p == nil {
		return ""
	}
	if strings.TrimSpace(p.FinalURL) != "" {
		return strings.TrimSpace(p.FinalURL)
	}
	return strings.TrimSpace(p.URL)
}

func parsePageURL(p *models.PageData) *url.URL {
	if p == nil {
		return nil
	}
	u, err := url.Parse(pageURL(p))
	if err != nil {
		return nil
	}
	return u
}

func meaningfulURLWords(raw string) []string {
	u, err := url.Parse(strings.TrimSpace(raw))
	if err != nil {
		return nil
	}
	decodedPath := u.Path
	if unescaped, err := url.PathUnescape(u.EscapedPath()); err == nil {
		decodedPath = unescaped
	}
	tokens := strings.FieldsFunc(strings.ToLower(decodedPath), func(r rune) bool {
		return !(unicode.IsLetter(r) || unicode.IsDigit(r))
	})
	seen := map[string]bool{}
	var words []string
	for _, token := range tokens {
		token = strings.TrimSpace(token)
		if !isMeaningfulURLWord(token) || seen[token] {
			continue
		}
		seen[token] = true
		words = append(words, token)
	}
	return words
}

func isMeaningfulURLWord(token string) bool {
	if len(token) < 3 || stopWords[token] || genericURLWords[token] {
		return false
	}
	allDigits := true
	hasVowel := false
	for _, r := range token {
		if !unicode.IsDigit(r) {
			allDigits = false
		}
		if strings.ContainsRune("aeiou", r) {
			hasVowel = true
		}
	}
	if allDigits {
		return false
	}
	if len(token) >= 8 && !hasVowel {
		return false
	}
	return true
}

func normalizeText(raw string) string {
	return strings.Join(strings.Fields(strings.ToLower(raw)), " ")
}

func breadcrumbTrails(rawSchemas []string) []breadcrumbTrail {
	var trails []breadcrumbTrail
	for _, raw := range rawSchemas {
		var data interface{}
		if err := json.Unmarshal([]byte(raw), &data); err != nil {
			continue
		}
		collectBreadcrumbTrails(data, &trails)
	}
	return trails
}

func collectBreadcrumbTrails(value interface{}, trails *[]breadcrumbTrail) {
	switch v := value.(type) {
	case []interface{}:
		for _, item := range v {
			collectBreadcrumbTrails(item, trails)
		}
	case map[string]interface{}:
		if schemaTypeMatches(v["@type"], "breadcrumblist") {
			if trail := parseBreadcrumbTrail(v["itemListElement"]); len(trail.Names) > 0 || len(trail.URLs) > 0 {
				*trails = append(*trails, trail)
			}
		}
		for _, child := range v {
			collectBreadcrumbTrails(child, trails)
		}
	}
}

func schemaTypeMatches(raw interface{}, expected string) bool {
	switch v := raw.(type) {
	case string:
		return strings.EqualFold(v, expected)
	case []interface{}:
		for _, item := range v {
			if schemaTypeMatches(item, expected) {
				return true
			}
		}
	}
	return false
}

func parseBreadcrumbTrail(raw interface{}) breadcrumbTrail {
	items, ok := raw.([]interface{})
	if !ok {
		if raw != nil {
			items = []interface{}{raw}
		}
	}
	sort.SliceStable(items, func(i, j int) bool {
		return breadcrumbPosition(items[i]) < breadcrumbPosition(items[j])
	})
	var trail breadcrumbTrail
	for _, item := range items {
		name, itemURL := breadcrumbItem(item)
		if name != "" {
			trail.Names = append(trail.Names, name)
		}
		if itemURL != "" {
			trail.URLs = append(trail.URLs, itemURL)
		}
	}
	return trail
}

func breadcrumbPosition(raw interface{}) float64 {
	item, ok := raw.(map[string]interface{})
	if !ok {
		return 0
	}
	switch v := item["position"].(type) {
	case float64:
		return v
	case int:
		return float64(v)
	}
	return 0
}

func breadcrumbItem(raw interface{}) (string, string) {
	item, ok := raw.(map[string]interface{})
	if !ok {
		return "", ""
	}
	name := stringProperty(item, "name")
	itemValue := item["item"]
	itemURL := ""
	switch v := itemValue.(type) {
	case string:
		if strings.Contains(v, "/") || strings.HasPrefix(strings.ToLower(v), "http") {
			itemURL = v
		} else if name == "" {
			name = v
		}
	case map[string]interface{}:
		itemURL = firstNonEmptyString(
			stringProperty(v, "@id"),
			stringProperty(v, "id"),
			stringProperty(v, "url"),
		)
		if name == "" {
			name = stringProperty(v, "name")
		}
	}
	return strings.TrimSpace(name), strings.TrimSpace(itemURL)
}

func breadcrumbTrailMismatch(trail breadcrumbTrail, current *url.URL, currentPath, pageText string, slugWords []string) string {
	for _, raw := range trail.URLs {
		crumbURL := resolveReference(raw, current.String())
		parsed, err := url.Parse(crumbURL)
		if err != nil || parsed == nil {
			continue
		}
		crumbPath := normalizedURLPath(parsed)
		if crumbPath == "" || crumbPath == "/" {
			continue
		}
		if crumbPath == currentPath {
			continue
		}
		if !strings.HasPrefix(currentPath+"/", strings.TrimRight(crumbPath, "/")+"/") {
			return fmt.Sprintf("breadcrumb URL %s is not in current path %s", crumbPath, currentPath)
		}
	}
	if len(trail.URLs) > 0 {
		lastURL := resolveReference(trail.URLs[len(trail.URLs)-1], current.String())
		if parsed, err := url.Parse(lastURL); err == nil && parsed != nil {
			lastPath := normalizedURLPath(parsed)
			if lastPath != "" && lastPath != currentPath && !strings.HasPrefix(currentPath+"/", strings.TrimRight(lastPath, "/")+"/") {
				return fmt.Sprintf("final breadcrumb URL %s does not match current path %s", lastPath, currentPath)
			}
		}
	}
	if len(trail.Names) == 0 || pageText == "" {
		return ""
	}
	finalName := normalizeText(trail.Names[len(trail.Names)-1])
	if finalName == "" || strings.Contains(pageText, finalName) {
		return ""
	}
	for _, word := range slugWords {
		if strings.Contains(finalName, word) || strings.Contains(pageText, word) {
			return ""
		}
	}
	return fmt.Sprintf("final breadcrumb label %q does not match page title/H1", trail.Names[len(trail.Names)-1])
}

func normalizedURLPath(u *url.URL) string {
	if u == nil {
		return ""
	}
	path := u.EscapedPath()
	if path == "" {
		path = u.Path
	}
	if decoded, err := url.PathUnescape(path); err == nil {
		path = decoded
	}
	if path == "" {
		return "/"
	}
	if path != "/" {
		path = "/" + strings.Trim(path, "/")
	}
	return strings.ToLower(path)
}

func looksLikePrintURL(raw string) bool {
	u, err := url.Parse(strings.TrimSpace(raw))
	if err != nil || u == nil {
		return false
	}
	path := strings.ToLower(u.Path)
	if strings.Contains(path, "/print/") || strings.HasSuffix(path, "/print") ||
		strings.Contains(path, "/printer/") || strings.Contains(path, "/printable/") {
		return true
	}
	for key, values := range u.Query() {
		key = strings.ToLower(key)
		if key == "print" || key == "printer" || key == "printable" {
			return true
		}
		for _, value := range values {
			value = strings.ToLower(value)
			if value == "print" || value == "printer" || value == "printable" {
				return true
			}
		}
	}
	return false
}

func resolveReference(raw, baseRaw string) string {
	ref, err := url.Parse(strings.TrimSpace(raw))
	if err != nil {
		return strings.TrimSpace(raw)
	}
	base, err := url.Parse(strings.TrimSpace(baseRaw))
	if err != nil || base == nil {
		return ref.String()
	}
	return base.ResolveReference(ref).String()
}

func sameURLIgnoringTrailingSlash(a, b string) bool {
	au, errA := url.Parse(strings.TrimSpace(a))
	bu, errB := url.Parse(strings.TrimSpace(b))
	if errA != nil || errB != nil || au == nil || bu == nil {
		return strings.TrimRight(a, "/") == strings.TrimRight(b, "/")
	}
	return strings.EqualFold(au.Scheme, bu.Scheme) &&
		strings.EqualFold(au.Host, bu.Host) &&
		strings.TrimRight(au.EscapedPath(), "/") == strings.TrimRight(bu.EscapedPath(), "/") &&
		au.RawQuery == bu.RawQuery
}

func urlHasNonASCII(raw string) bool {
	if containsNonASCII(raw) {
		return true
	}
	u, err := url.Parse(strings.TrimSpace(raw))
	if err != nil || u == nil {
		return false
	}
	if decoded, err := url.PathUnescape(u.EscapedPath()); err == nil && containsNonASCII(decoded) {
		return true
	}
	if decoded, err := url.QueryUnescape(u.RawQuery); err == nil && containsNonASCII(decoded) {
		return true
	}
	return false
}

func containsNonASCII(raw string) bool {
	for _, r := range raw {
		if r > unicode.MaxASCII {
			return true
		}
	}
	return false
}

func isIndexableStatus(p *models.PageData) bool {
	if p == nil || p.Error != "" {
		return false
	}
	return p.StatusCode == 0 || (p.StatusCode >= 200 && p.StatusCode < 300)
}

func isStaticURLPath(path string) bool {
	path = strings.ToLower(path)
	staticExts := []string{".avif", ".css", ".gif", ".ico", ".jpeg", ".jpg", ".js", ".json", ".pdf", ".png", ".svg", ".webp", ".xml"}
	for _, ext := range staticExts {
		if strings.HasSuffix(path, ext) {
			return true
		}
	}
	return false
}

func pathSegments(path string) []string {
	path = strings.Trim(path, "/")
	if path == "" {
		return nil
	}
	rawSegments := strings.Split(path, "/")
	segments := make([]string, 0, len(rawSegments))
	for _, segment := range rawSegments {
		segment = strings.TrimSpace(segment)
		if segment != "" {
			segments = append(segments, strings.ToLower(segment))
		}
	}
	return segments
}

func trailingSlashStructureKey(u *url.URL) string {
	if u == nil || u.Path == "" || u.Path == "/" {
		return ""
	}
	path := "/" + strings.Trim(u.Path, "/")
	if path == "/" {
		return ""
	}
	return strings.ToLower(u.Scheme) + "://" + strings.ToLower(u.Host) + path
}

func firstPageURL(pages []*models.PageData) string {
	for _, p := range pages {
		if raw := pageURL(p); raw != "" {
			return raw
		}
	}
	return ""
}

func stringProperty(item map[string]interface{}, key string) string {
	if value, ok := item[key]; ok {
		if str, ok := value.(string); ok {
			return strings.TrimSpace(str)
		}
	}
	return ""
}

func firstNonEmptyString(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func uniqueStrings(values []string) []string {
	seen := map[string]bool{}
	var out []string
	for _, value := range values {
		if value == "" || seen[value] {
			continue
		}
		seen[value] = true
		out = append(out, value)
	}
	return out
}

func sampleStrings(values []string, max int) []string {
	values = uniqueStrings(values)
	if len(values) <= max {
		return values
	}
	return values[:max]
}

type urlHasUppercase struct{}

func (c *urlHasUppercase) Run(p *models.PageData) []models.CheckResult {
	parsed, err := url.Parse(p.URL)
	if err != nil {
		return nil
	}
	if strings.ToLower(parsed.Path) != parsed.Path {
		return []models.CheckResult{{
			ID:       "url.has_uppercase",
			Category: "URL Structure",
			Severity: models.SeverityWarning,
			Message:  "URL path contains uppercase characters",
			URL:      p.URL,
		}}
	}
	return nil
}

type urlHasSpaces struct{}

func (c *urlHasSpaces) Run(p *models.PageData) []models.CheckResult {
	if strings.Contains(p.URL, "%20") || strings.Contains(p.URL, "+") {
		return []models.CheckResult{{
			ID:       "url.has_spaces",
			Category: "URL Structure",
			Severity: models.SeverityError,
			Message:  "URL contains spaces (encoded)",
			URL:      p.URL,
		}}
	}
	return nil
}

type urlHasSessionParams struct{}

func (c *urlHasSessionParams) Run(p *models.PageData) []models.CheckResult {
	lowerURL := strings.ToLower(p.URL)
	for _, param := range sessionParams {
		if strings.Contains(lowerURL, param) {
			return []models.CheckResult{{
				ID:       "url.has_session_params",
				Category: "URL Structure",
				Severity: models.SeverityError,
				Message:  fmt.Sprintf("URL contains session parameter: %s", param),
				URL:      p.URL,
			}}
		}
	}
	return nil
}

type urlTooManyParams struct{}

func (c *urlTooManyParams) Run(p *models.PageData) []models.CheckResult {
	parsed, err := url.Parse(p.URL)
	if err != nil {
		return nil
	}
	params := parsed.Query()
	if len(params) > 3 {
		return []models.CheckResult{{
			ID:       "url.too_many_params",
			Category: "URL Structure",
			Severity: models.SeverityWarning,
			Message:  fmt.Sprintf("URL has too many query parameters (%d, max 3)", len(params)),
			URL:      p.URL,
		}}
	}
	return nil
}

type urlDoubleSlash struct{}

func (c *urlDoubleSlash) Run(p *models.PageData) []models.CheckResult {
	parsed, err := url.Parse(p.URL)
	if err != nil {
		return nil
	}
	// Skip protocol's double slash
	path := parsed.Path
	if strings.Contains(path, "//") {
		return []models.CheckResult{{
			ID:       "url.double_slash",
			Category: "URL Structure",
			Severity: models.SeverityWarning,
			Message:  "URL path contains double slashes",
			URL:      p.URL,
		}}
	}
	return nil
}

type urlNonDescriptive struct{}

func (c *urlNonDescriptive) Run(p *models.PageData) []models.CheckResult {
	parsed, err := url.Parse(p.URL)
	if err != nil {
		return nil
	}
	path := parsed.Path
	if path == "/" || path == "" {
		return nil
	}
	if nonDescriptivePattern.MatchString(path) {
		return []models.CheckResult{{
			ID:       "url.non_descriptive",
			Category: "URL Structure",
			Severity: models.SeverityNotice,
			Message:  "URL path is non-descriptive (contains only numeric ID)",
			URL:      p.URL,
		}}
	}
	return nil
}

type urlPathDepthTooDeep struct{}

func (c *urlPathDepthTooDeep) Run(p *models.PageData) []models.CheckResult {
	parsed, err := url.Parse(p.URL)
	if err != nil {
		return nil
	}
	path := strings.Trim(parsed.Path, "/")
	if path == "" {
		return nil
	}
	segments := strings.Split(path, "/")
	if len(segments) > 4 {
		return []models.CheckResult{{
			ID:       "url.path_depth_too_deep",
			Category: "URL Structure",
			Severity: models.SeverityWarning,
			Message:  fmt.Sprintf("URL path too deep (%d segments, max 4)", len(segments)),
			URL:      p.URL,
		}}
	}
	return nil
}

type urlContainsStopWords struct{}

func (c *urlContainsStopWords) Run(p *models.PageData) []models.CheckResult {
	parsed, err := url.Parse(p.URL)
	if err != nil {
		return nil
	}
	path := strings.Trim(parsed.Path, "/")
	if path == "" {
		return nil
	}
	// Split path into segments, then split each segment by hyphens/underscores
	segments := strings.Split(path, "/")
	var found []string
	for _, seg := range segments {
		words := strings.FieldsFunc(seg, func(r rune) bool {
			return r == '-' || r == '_'
		})
		for _, w := range words {
			if stopWords[strings.ToLower(w)] {
				found = append(found, strings.ToLower(w))
			}
		}
	}
	if len(found) > 0 {
		return []models.CheckResult{{
			ID:       "url.contains_stop_words",
			Category: "URL Structure",
			Severity: models.SeverityNotice,
			Message:  fmt.Sprintf("URL contains stop words: %s", strings.Join(found, ", ")),
			URL:      p.URL,
		}}
	}
	return nil
}
