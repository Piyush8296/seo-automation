package onpage_keywords

import (
	"fmt"
	"net/url"
	"regexp"
	"sort"
	"strings"
	"unicode"

	"github.com/cars24/seo-automation/internal/models"
)

const category = "On-Page Keywords"

// PageChecks returns deterministic title/meta/H1/body alignment checks.
func PageChecks() []models.PageCheck {
	return []models.PageCheck{
		&titlePrimaryKeywordMissing{},
		&titleKeywordLate{},
		&titleBrandMissing{},
		&titleKeywordStuffing{},
		&titleTopicMismatch{},
		&titleDynamicPlaceholder{},
		&titleSpecialChars{},
		&homepageTitleNotBrandFocused{},
		&titleModelMissing{},
		&titleYearMissing{},
		&blogTitleKeywordMissing{},
		&titleCityModelMissing{},
		&metaPrimaryKeywordMissing{},
		&metaDescriptionCTAMissing{},
		&metaDynamicPlaceholder{},
		&h1PrimaryKeywordMissing{},
		&h2SecondaryKeywordMissing{},
		&h2KeywordStuffing{},
		&keywordFirst100Missing{},
		&keywordDensityOutOfRange{},
		&keywordSurfaceMismatch{},
	}
}

type titlePrimaryKeywordMissing struct{}

func (c *titlePrimaryKeywordMissing) Run(p *models.PageData) []models.CheckResult {
	t := topicForPage(p)
	if !t.hasPrimary() || strings.TrimSpace(p.Title) == "" || hasCoverage(t.primary, p.Title) {
		return nil
	}
	return one(p, "title.primary_keyword_missing", models.SeverityWarning, "Primary URL/topic terms are missing from the title", t.primaryDetails())
}

type titleKeywordLate struct{}

func (c *titleKeywordLate) Run(p *models.PageData) []models.CheckResult {
	t := topicForPage(p)
	title := strings.TrimSpace(p.Title)
	if !t.hasPrimary() || title == "" || !hasCoverage(t.primary, title) {
		return nil
	}
	pos := firstTokenPosition(title, t.primary)
	if pos > 35 || float64(pos) > float64(len(title))*0.45 {
		return one(p, "title.keyword_not_near_start", models.SeverityNotice, "Primary topic appears late in the title", fmt.Sprintf("first topic token starts at character %d; topic=%s", pos, t.primaryDetails()))
	}
	return nil
}

type titleBrandMissing struct{}

func (c *titleBrandMissing) Run(p *models.PageData) []models.CheckResult {
	t := topicForPage(p)
	if t.brand == "" || strings.TrimSpace(p.Title) == "" || tokenInText(t.brand, p.Title) {
		return nil
	}
	return one(p, "title.brand_missing", models.SeverityNotice, "Title tag does not include the site/brand name", "expected brand token: "+t.brand)
}

type titleKeywordStuffing struct{}

func (c *titleKeywordStuffing) Run(p *models.PageData) []models.CheckResult {
	token, count := repeatedToken(tokenize(p.Title), 3)
	if token == "" {
		return nil
	}
	return one(p, "title.keyword_stuffing", models.SeverityWarning, "Title appears keyword-stuffed with repeated terms", fmt.Sprintf("%q appears %d times", token, count))
}

type titleTopicMismatch struct{}

func (c *titleTopicMismatch) Run(p *models.PageData) []models.CheckResult {
	t := topicForPage(p)
	if !t.hasPrimary() || strings.TrimSpace(p.Title) == "" {
		return nil
	}
	supporting := strings.Join(append(append([]string{}, p.H1s...), p.BodyText), " ")
	if hasCoverage(t.primary, p.Title) && !hasAnyTopicToken(t.primary, supporting) {
		return one(p, "title.topic_mismatch", models.SeverityWarning, "Title topic is not supported by the H1 or visible body copy", t.primaryDetails())
	}
	return nil
}

type titleDynamicPlaceholder struct{}

func (c *titleDynamicPlaceholder) Run(p *models.PageData) []models.CheckResult {
	if containsPlaceholder(p.Title) {
		return one(p, "title.dynamic_placeholder", models.SeverityError, "Title contains an unpopulated dynamic placeholder", strings.TrimSpace(p.Title))
	}
	return nil
}

type titleSpecialChars struct{}

func (c *titleSpecialChars) Run(p *models.PageData) []models.CheckResult {
	if badSpecialChars(p.Title) {
		return one(p, "title.special_chars", models.SeverityNotice, "Title contains special characters that can display poorly in SERPs", strings.TrimSpace(p.Title))
	}
	return nil
}

type homepageTitleNotBrandFocused struct{}

func (c *homepageTitleNotBrandFocused) Run(p *models.PageData) []models.CheckResult {
	t := topicForPage(p)
	title := strings.TrimSpace(p.Title)
	if !isHomePage(p.URL) || title == "" || (t.brand != "" && tokenInText(t.brand, title) && !isGenericTitle(title)) {
		return nil
	}
	return one(p, "title.homepage_not_brand_focused", models.SeverityNotice, "Homepage title is not clearly brand-focused", "brand="+t.brand+" title="+title)
}

type titleModelMissing struct{}

func (c *titleModelMissing) Run(p *models.PageData) []models.CheckResult {
	modelTokens := carModelTokens(topicForPage(p).urlTokens)
	if len(modelTokens) == 0 || hasAnyTopicToken(modelTokens, p.Title) {
		return nil
	}
	return one(p, "title.model_missing", modelSeverity(modelTokens), "Car model/brand token from the URL is missing from the title", strings.Join(modelTokens, ", "))
}

type titleYearMissing struct{}

func (c *titleYearMissing) Run(p *models.PageData) []models.CheckResult {
	year := firstYear(topicForPage(p).urlTokens)
	if year == "" || strings.Contains(p.Title, year) {
		return nil
	}
	return one(p, "title.year_missing", models.SeverityNotice, "Year signal from the URL is missing from the title", year)
}

type blogTitleKeywordMissing struct{}

func (c *blogTitleKeywordMissing) Run(p *models.PageData) []models.CheckResult {
	t := topicForPage(p)
	if !isBlogPage(p.URL) || !t.hasPrimary() || hasCoverage(t.primary, p.Title) {
		return nil
	}
	return one(p, "title.blog_keyword_missing", models.SeverityWarning, "Blog/article title does not include the inferred target topic", t.primaryDetails())
}

type titleCityModelMissing struct{}

func (c *titleCityModelMissing) Run(p *models.PageData) []models.CheckResult {
	t := topicForPage(p)
	city := cityToken(t.urlTokens)
	modelTokens := carModelTokens(t.urlTokens)
	if city == "" || len(modelTokens) == 0 {
		return nil
	}
	missing := []string{}
	if !tokenInText(city, p.Title) {
		missing = append(missing, city)
	}
	if !hasAnyTopicToken(modelTokens, p.Title) {
		missing = append(missing, strings.Join(modelTokens, "/"))
	}
	if len(missing) == 0 {
		return nil
	}
	return one(p, "title.city_model_missing", models.SeverityWarning, "City + model opportunity is missing from the title", strings.Join(missing, ", "))
}

type metaPrimaryKeywordMissing struct{}

func (c *metaPrimaryKeywordMissing) Run(p *models.PageData) []models.CheckResult {
	t := topicForPage(p)
	if !t.hasPrimary() || strings.TrimSpace(p.MetaDesc) == "" || hasCoverage(t.primary, p.MetaDesc) {
		return nil
	}
	return one(p, "meta_desc.primary_keyword_missing", models.SeverityNotice, "Primary URL/topic terms are missing from the meta description", t.primaryDetails())
}

type metaDescriptionCTAMissing struct{}

func (c *metaDescriptionCTAMissing) Run(p *models.PageData) []models.CheckResult {
	if strings.TrimSpace(p.MetaDesc) == "" || !isTransactionalPage(topicForPage(p)) || hasCTA(p.MetaDesc) {
		return nil
	}
	return one(p, "meta_desc.cta_missing", models.SeverityNotice, "Meta description lacks a clear action phrase for a transactional page", strings.TrimSpace(p.MetaDesc))
}

type metaDynamicPlaceholder struct{}

func (c *metaDynamicPlaceholder) Run(p *models.PageData) []models.CheckResult {
	if containsPlaceholder(p.MetaDesc) {
		return one(p, "meta_desc.dynamic_placeholder", models.SeverityError, "Meta description contains an unpopulated dynamic placeholder", strings.TrimSpace(p.MetaDesc))
	}
	return nil
}

type h1PrimaryKeywordMissing struct{}

func (c *h1PrimaryKeywordMissing) Run(p *models.PageData) []models.CheckResult {
	t := topicForPage(p)
	h1 := strings.Join(p.H1s, " ")
	if !t.hasPrimary() || strings.TrimSpace(h1) == "" || hasCoverage(t.primary, h1) {
		return nil
	}
	return one(p, "headings.h1.primary_keyword_missing", models.SeverityWarning, "Primary URL/topic terms are missing from the H1", t.primaryDetails())
}

type h2SecondaryKeywordMissing struct{}

func (c *h2SecondaryKeywordMissing) Run(p *models.PageData) []models.CheckResult {
	t := topicForPage(p)
	h2 := strings.Join(p.H2s, " ")
	if p.WordCount < 250 || len(t.primary) < 2 || strings.TrimSpace(h2) == "" || hasAnyTopicToken(secondaryTokens(t), h2) {
		return nil
	}
	return one(p, "headings.h2.secondary_keywords_missing", models.SeverityNotice, "H2 headings do not reinforce secondary topic terms", strings.Join(secondaryTokens(t), ", "))
}

type h2KeywordStuffing struct{}

func (c *h2KeywordStuffing) Run(p *models.PageData) []models.CheckResult {
	token, count := repeatedToken(tokenize(strings.Join(p.H2s, " ")), 4)
	if token == "" {
		return nil
	}
	return one(p, "headings.h2.keyword_stuffing", models.SeverityWarning, "H2 headings appear overstuffed with repeated terms", fmt.Sprintf("%q appears %d times", token, count))
}

type keywordFirst100Missing struct{}

func (c *keywordFirst100Missing) Run(p *models.PageData) []models.CheckResult {
	t := topicForPage(p)
	if !t.hasPrimary() || p.WordCount < 100 || hasCoverage(t.primary, firstWords(p.BodyText, 100)) {
		return nil
	}
	return one(p, "keyword.first_100_missing", models.SeverityNotice, "Primary topic is missing from the first 100 visible words", t.primaryDetails())
}

type keywordDensityOutOfRange struct{}

func (c *keywordDensityOutOfRange) Run(p *models.PageData) []models.CheckResult {
	t := topicForPage(p)
	if !t.hasPrimary() || p.WordCount < 150 {
		return nil
	}
	density := keywordDensity(t.primary, p.BodyText)
	if density == 0 || density < 0.005 || density > 0.02 {
		return one(p, "keyword.density_out_of_range", models.SeverityNotice, "Primary topic density is outside the 0.5%-2% target range", fmt.Sprintf("density %.2f%%; topic=%s", density*100, t.primaryDetails()))
	}
	return nil
}

type keywordSurfaceMismatch struct{}

func (c *keywordSurfaceMismatch) Run(p *models.PageData) []models.CheckResult {
	t := topicForPage(p)
	if !t.hasPrimary() {
		return nil
	}
	surfaces := map[string]bool{
		"url":   len(t.urlTokens) > 0,
		"title": hasCoverage(t.primary, p.Title),
		"h1":    hasCoverage(t.primary, strings.Join(p.H1s, " ")),
		"body":  hasAnyTopicToken(t.primary, p.BodyText),
	}
	covered := 0
	missing := []string{}
	for _, key := range []string{"url", "title", "h1", "body"} {
		if surfaces[key] {
			covered++
		} else {
			missing = append(missing, key)
		}
	}
	if covered < 3 {
		return one(p, "keyword.surface_mismatch", models.SeverityWarning, "Keyword/topic alignment is weak across URL, title, H1, and body", "missing: "+strings.Join(missing, ", "))
	}
	return nil
}

type pageTopic struct {
	urlTokens   []string
	titleTokens []string
	h1Tokens    []string
	bodyTokens  []string
	primary     []string
	brand       string
}

func topicForPage(p *models.PageData) pageTopic {
	urlTokens := urlTopicTokens(p.URL)
	titleTokens := tokenize(p.Title)
	h1Tokens := tokenize(strings.Join(p.H1s, " "))
	bodyTokens := tokenize(p.BodyText)
	primary := compactTokens(urlTokens)
	if len(primary) == 0 {
		primary = compactTokens(append(titleTokens, h1Tokens...))
	}
	if len(primary) > 6 {
		primary = primary[:6]
	}
	return pageTopic{
		urlTokens:   urlTokens,
		titleTokens: titleTokens,
		h1Tokens:    h1Tokens,
		bodyTokens:  bodyTokens,
		primary:     primary,
		brand:       brandFromURL(p.URL),
	}
}

func (t pageTopic) hasPrimary() bool {
	return len(t.primary) > 0
}

func (t pageTopic) primaryDetails() string {
	return strings.Join(t.primary, ", ")
}

func one(p *models.PageData, id string, sev models.Severity, msg string, details string) []models.CheckResult {
	return []models.CheckResult{{
		ID:       id,
		Category: category,
		Severity: sev,
		Message:  msg,
		URL:      p.URL,
		Details:  details,
	}}
}

func tokenize(text string) []string {
	var b strings.Builder
	for _, r := range strings.ToLower(text) {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(r)
		} else {
			b.WriteByte(' ')
		}
	}
	parts := strings.Fields(b.String())
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		token := normalizeToken(part)
		if token == "" || stopWords[token] || looksLikeID(token) {
			continue
		}
		out = append(out, token)
	}
	return out
}

func normalizeToken(token string) string {
	token = strings.Trim(token, "_-")
	if token == "" {
		return ""
	}
	if _, ok := yearToken(token); ok {
		return token
	}
	if len(token) < 3 {
		return ""
	}
	if strings.HasSuffix(token, "ies") && len(token) > 4 {
		token = strings.TrimSuffix(token, "ies") + "y"
	} else if strings.HasSuffix(token, "s") && !strings.HasSuffix(token, "ss") && len(token) > 3 {
		token = strings.TrimSuffix(token, "s")
	}
	if len(token) < 3 {
		return ""
	}
	return token
}

func urlTopicTokens(raw string) []string {
	parsed, err := url.Parse(raw)
	if err != nil {
		return nil
	}
	segments := strings.Split(strings.Trim(parsed.EscapedPath(), "/"), "/")
	var tokens []string
	for _, segment := range segments {
		if unescaped, err := url.PathUnescape(segment); err == nil {
			segment = unescaped
		}
		segment = strings.TrimSuffix(segment, ".html")
		segment = strings.TrimSuffix(segment, ".php")
		for _, token := range tokenize(segment) {
			if !genericURLTokens[token] {
				tokens = append(tokens, token)
			}
		}
	}
	return compactTokens(tokens)
}

func compactTokens(tokens []string) []string {
	seen := map[string]bool{}
	out := make([]string, 0, len(tokens))
	for _, token := range tokens {
		if token == "" || seen[token] {
			continue
		}
		seen[token] = true
		out = append(out, token)
	}
	return out
}

func hasCoverage(topic []string, text string) bool {
	if len(topic) == 0 {
		return false
	}
	textTokens := tokenSet(tokenize(text))
	matches := 0
	for _, token := range topic {
		if textTokens[token] {
			matches++
		}
	}
	return matches >= requiredMatches(len(topic))
}

func requiredMatches(total int) int {
	if total <= 1 {
		return 1
	}
	need := int(float64(total) * 0.6)
	if need < 1 {
		need = 1
	}
	if float64(need) < float64(total)*0.6 {
		need++
	}
	return need
}

func hasAnyTopicToken(topic []string, text string) bool {
	textTokens := tokenSet(tokenize(text))
	for _, token := range topic {
		if textTokens[token] {
			return true
		}
	}
	return false
}

func tokenInText(token string, text string) bool {
	return tokenSet(tokenize(text))[normalizeToken(token)]
}

func tokenSet(tokens []string) map[string]bool {
	out := make(map[string]bool, len(tokens))
	for _, token := range tokens {
		out[token] = true
	}
	return out
}

func firstTokenPosition(text string, tokens []string) int {
	lower := strings.ToLower(text)
	best := len(text)
	for _, token := range tokens {
		if token == "" {
			continue
		}
		for _, candidate := range []string{token, token + "s"} {
			if idx := strings.Index(lower, candidate); idx >= 0 && idx < best {
				best = idx
			}
		}
	}
	return best
}

func repeatedToken(tokens []string, threshold int) (string, int) {
	counts := map[string]int{}
	for _, token := range tokens {
		if genericURLTokens[token] || stopWords[token] {
			continue
		}
		counts[token]++
	}
	bestToken := ""
	bestCount := 0
	for token, count := range counts {
		if count > bestCount {
			bestToken, bestCount = token, count
		}
	}
	if bestCount >= threshold {
		return bestToken, bestCount
	}
	return "", 0
}

func containsPlaceholder(text string) bool {
	lower := strings.ToLower(text)
	needles := []string{"{{", "}}", "${", "%city%", "%model%", "[city]", "[model]", "{city}", "{model}", "undefined", "null", "nan"}
	for _, needle := range needles {
		if strings.Contains(lower, needle) {
			return true
		}
	}
	return false
}

var repeatedPunctuation = regexp.MustCompile(`[|!?.:_-]{4,}`)

func badSpecialChars(text string) bool {
	if strings.ContainsRune(text, '\uFFFD') || repeatedPunctuation.MatchString(text) || strings.Count(text, "|") > 3 {
		return true
	}
	for _, r := range text {
		if unicode.IsControl(r) && !unicode.IsSpace(r) {
			return true
		}
	}
	return false
}

func isHomePage(raw string) bool {
	parsed, err := url.Parse(raw)
	if err != nil {
		return false
	}
	path := strings.Trim(parsed.Path, "/")
	return path == ""
}

func isGenericTitle(title string) bool {
	tokens := tokenize(title)
	return len(tokens) <= 1 || (len(tokens) == 1 && (tokens[0] == "home" || tokens[0] == "homepage"))
}

func isBlogPage(raw string) bool {
	tokens := urlTopicTokens(raw)
	for _, token := range tokens {
		if blogTokens[token] {
			return true
		}
	}
	return false
}

func isTransactionalPage(t pageTopic) bool {
	for _, token := range append(append([]string{}, t.urlTokens...), t.titleTokens...) {
		if transactionalTokens[token] {
			return true
		}
	}
	return false
}

func hasCTA(text string) bool {
	for _, token := range tokenize(text) {
		if ctaTokens[token] {
			return true
		}
	}
	return false
}

func secondaryTokens(t pageTopic) []string {
	primary := tokenSet(t.primary)
	var secondary []string
	for _, token := range append(t.titleTokens, t.h1Tokens...) {
		if !primary[token] {
			secondary = append(secondary, token)
		}
	}
	if len(secondary) == 0 {
		secondary = t.primary
	}
	return compactTokens(secondary)
}

func firstWords(text string, n int) string {
	words := strings.Fields(text)
	if len(words) > n {
		words = words[:n]
	}
	return strings.Join(words, " ")
}

func keywordDensity(topic []string, text string) float64 {
	tokens := tokenize(text)
	if len(tokens) == 0 {
		return 0
	}
	topicSet := tokenSet(topic)
	matches := 0
	for _, token := range tokens {
		if topicSet[token] {
			matches++
		}
	}
	return float64(matches) / float64(len(tokens))
}

func brandFromURL(raw string) string {
	parsed, err := url.Parse(raw)
	if err != nil {
		return ""
	}
	host := strings.ToLower(parsed.Hostname())
	parts := strings.Split(host, ".")
	for len(parts) > 0 && (parts[0] == "www" || parts[0] == "m") {
		parts = parts[1:]
	}
	if len(parts) == 0 {
		return ""
	}
	return normalizeToken(parts[0])
}

func cityToken(tokens []string) string {
	for _, token := range tokens {
		if cityTokens[token] {
			return token
		}
	}
	return ""
}

func carModelTokens(tokens []string) []string {
	var out []string
	for _, token := range tokens {
		if carTokens[token] {
			out = append(out, token)
		}
	}
	sort.Strings(out)
	return compactTokens(out)
}

func firstYear(tokens []string) string {
	for _, token := range tokens {
		if _, ok := yearToken(token); ok {
			return token
		}
	}
	return ""
}

func yearToken(token string) (int, bool) {
	if len(token) != 4 {
		return 0, false
	}
	year := 0
	for _, r := range token {
		if r < '0' || r > '9' {
			return 0, false
		}
		year = year*10 + int(r-'0')
	}
	return year, year >= 1990 && year <= 2035
}

func looksLikeID(token string) bool {
	if _, ok := yearToken(token); ok {
		return false
	}
	if len(token) >= 8 {
		digits := 0
		for _, r := range token {
			if unicode.IsDigit(r) {
				digits++
			}
		}
		return digits >= len(token)/2
	}
	return false
}

func modelSeverity(modelTokens []string) models.Severity {
	if len(modelTokens) > 1 {
		return models.SeverityWarning
	}
	return models.SeverityNotice
}

var stopWords = map[string]bool{
	"about": true, "after": true, "also": true, "and": true, "are": true, "available": true, "best": true,
	"com": true, "for": true, "from": true, "have": true, "homepage": true, "https": true, "into": true,
	"near": true, "now": true, "online": true, "our": true, "page": true, "pages": true, "the": true,
	"this": true, "to": true, "with": true, "www": true, "you": true, "your": true,
}

var genericURLTokens = map[string]bool{
	"amp": true, "api": true, "category": true, "detail": true, "home": true, "html": true, "index": true,
	"listing": true, "listings": true, "page": true, "search": true, "tag": true,
}

var transactionalTokens = map[string]bool{
	"book": true, "buy": true, "car": true, "deal": true, "finance": true, "loan": true, "offer": true,
	"price": true, "quote": true, "sale": true, "sell": true, "used": true,
}

var ctaTokens = map[string]bool{
	"book": true, "buy": true, "call": true, "check": true, "compare": true, "discover": true, "explore": true,
	"find": true, "get": true, "learn": true, "read": true, "request": true, "schedule": true, "sell": true,
	"shop": true, "start": true, "view": true, "visit": true,
}

var blogTokens = map[string]bool{
	"article": true, "blog": true, "guide": true, "news": true, "review": true, "story": true, "tips": true,
}

var cityTokens = map[string]bool{
	"agra": true, "ahmedabad": true, "bangalore": true, "bengaluru": true, "bhopal": true, "chandigarh": true,
	"chennai": true, "delhi": true, "faridabad": true, "ghaziabad": true, "gurgaon": true, "gurugram": true,
	"hyderabad": true, "indore": true, "jaipur": true, "kolkata": true, "lucknow": true, "mumbai": true,
	"noida": true, "pune": true, "surat": true,
}

var carTokens = map[string]bool{
	"alto": true, "amaze": true, "audi": true, "baleno": true, "benz": true, "bmw": true, "city": true,
	"civic": true, "compass": true, "creta": true, "duster": true, "ecosport": true, "fortuner": true,
	"harrier": true, "hector": true, "honda": true, "hyundai": true, "i10": true, "i20": true, "innova": true,
	"jeep": true, "kia": true, "kwid": true, "mahindra": true, "maruti": true, "nexon": true, "nissan": true,
	"polo": true, "rapid": true, "renault": true, "scorpio": true, "seltos": true, "skoda": true, "sonet": true,
	"suzuki": true, "swift": true, "tata": true, "thar": true, "toyota": true, "venue": true, "vento": true,
	"volkswagen": true, "wagon": true, "wagonr": true, "xuv": true,
}
