package structured_data

import (
	"encoding/json"
	"fmt"
	"net/url"
	"regexp"
	"strings"

	"github.com/cars24/seo-automation/internal/models"
)

const structuredDataCategory = "Structured Data"

var (
	emptyLDJSONScriptPattern = regexp.MustCompile(`(?is)<script\b[^>]*type\s*=\s*["']application/ld\+json["'][^>]*(?:src|data-src|data-href)\s*=|<script\b[^>]*type\s*=\s*["']application/ld\+json["'][^>]*>\s*</script>`)
	whitespacePattern        = regexp.MustCompile(`\s+`)
)

// PageChecks returns per-page structured data checks.
func PageChecks() []models.PageCheck {
	return []models.PageCheck{
		&schemaJSONLDMissing{},
		&schemaJSONLDInvalidJSON{},
		&schemaJSONLDMissingContext{},
		&schemaJSONLDMissingType{},
		&schemaJSONLDDuplicateType{},
		&schemaArticleMissingFields{},
		&schemaProductMissingFields{},
		&schemaBreadcrumbInvalid{},
		&schemaFAQInvalid{},
		&schemaReviewRatingMissing{},
		&schemaLazyLoadedRisk{},
		&schemaHiddenContentMismatch{},
		&schemaHowToMissing{},
		&schemaProductListMissing{},
		&schemaEventMissing{},
		&schemaSpeakableMissing{},
	}
}

// SiteChecks returns site-wide structured data checks.
func SiteChecks() []models.SiteCheck {
	return []models.SiteCheck{
		&schemaOrganizationMissingHomepage{},
		&schemaWebSiteMissingHomepage{},
		&schemaSearchActionMissingHomepage{},
	}
}

type schemaJSONLDMissing struct{}

func (c *schemaJSONLDMissing) Run(p *models.PageData) []models.CheckResult {
	if len(p.SchemaJSONRaw) == 0 {
		return []models.CheckResult{{
			ID:       "schema.jsonld.missing",
			Category: "Structured Data",
			Severity: models.SeverityNotice,
			Message:  "No JSON-LD structured data found",
			URL:      p.URL,
		}}
	}
	return nil
}

type schemaJSONLDInvalidJSON struct{}

func (c *schemaJSONLDInvalidJSON) Run(p *models.PageData) []models.CheckResult {
	var results []models.CheckResult
	for _, raw := range p.SchemaJSONRaw {
		var obj interface{}
		if err := json.Unmarshal([]byte(raw), &obj); err != nil {
			results = append(results, models.CheckResult{
				ID:       "schema.jsonld.invalid_json",
				Category: "Structured Data",
				Severity: models.SeverityError,
				Message:  "Invalid JSON in JSON-LD structured data",
				URL:      p.URL,
				Details:  err.Error(),
			})
		}
	}
	return results
}

type schemaJSONLDMissingContext struct{}

func (c *schemaJSONLDMissingContext) Run(p *models.PageData) []models.CheckResult {
	var results []models.CheckResult
	for _, raw := range p.SchemaJSONRaw {
		var obj map[string]interface{}
		if err := json.Unmarshal([]byte(raw), &obj); err != nil {
			continue
		}
		ctx, _ := obj["@context"].(string)
		if ctx == "" {
			results = append(results, models.CheckResult{
				ID:       "schema.jsonld.missing_context",
				Category: "Structured Data",
				Severity: models.SeverityError,
				Message:  "JSON-LD schema missing @context",
				URL:      p.URL,
			})
		}
	}
	return results
}

type schemaJSONLDMissingType struct{}

func (c *schemaJSONLDMissingType) Run(p *models.PageData) []models.CheckResult {
	var results []models.CheckResult
	for _, raw := range p.SchemaJSONRaw {
		var obj map[string]interface{}
		if err := json.Unmarshal([]byte(raw), &obj); err != nil {
			continue
		}
		t, _ := obj["@type"].(string)
		if t == "" {
			results = append(results, models.CheckResult{
				ID:       "schema.jsonld.missing_type",
				Category: "Structured Data",
				Severity: models.SeverityError,
				Message:  "JSON-LD schema missing @type",
				URL:      p.URL,
			})
		}
	}
	return results
}

type schemaJSONLDDuplicateType struct{}

func (c *schemaJSONLDDuplicateType) Run(p *models.PageData) []models.CheckResult {
	typeCounts := map[string]int{}
	for _, raw := range p.SchemaJSONRaw {
		var obj map[string]interface{}
		if err := json.Unmarshal([]byte(raw), &obj); err != nil {
			continue
		}
		t, _ := obj["@type"].(string)
		if t != "" {
			typeCounts[strings.ToLower(t)]++
		}
	}
	var results []models.CheckResult
	for t, count := range typeCounts {
		if count > 1 {
			results = append(results, models.CheckResult{
				ID:       "schema.jsonld.duplicate_type",
				Category: "Structured Data",
				Severity: models.SeverityWarning,
				Message:  fmt.Sprintf("Duplicate schema @type: %s (%d occurrences)", t, count),
				URL:      p.URL,
			})
		}
	}
	return results
}

type schemaArticleMissingFields struct{}

func (c *schemaArticleMissingFields) Run(p *models.PageData) []models.CheckResult {
	for _, raw := range p.SchemaJSONRaw {
		var obj map[string]interface{}
		if err := json.Unmarshal([]byte(raw), &obj); err != nil {
			continue
		}
		t := strings.ToLower(fmt.Sprintf("%v", obj["@type"]))
		if !strings.Contains(t, "article") && !strings.Contains(t, "newsarticle") && !strings.Contains(t, "blogposting") {
			continue
		}
		var missing []string
		if obj["headline"] == nil {
			missing = append(missing, "headline")
		}
		if obj["datePublished"] == nil {
			missing = append(missing, "datePublished")
		}
		if obj["author"] == nil {
			missing = append(missing, "author")
		}
		if len(missing) > 0 {
			return []models.CheckResult{{
				ID:       "schema.article.missing_fields",
				Category: "Structured Data",
				Severity: models.SeverityWarning,
				Message:  fmt.Sprintf("Article schema missing required fields: %s", strings.Join(missing, ", ")),
				URL:      p.URL,
			}}
		}
	}
	return nil
}

type schemaProductMissingFields struct{}

func (c *schemaProductMissingFields) Run(p *models.PageData) []models.CheckResult {
	for _, raw := range p.SchemaJSONRaw {
		var obj map[string]interface{}
		if err := json.Unmarshal([]byte(raw), &obj); err != nil {
			continue
		}
		if strings.ToLower(fmt.Sprintf("%v", obj["@type"])) != "product" {
			continue
		}
		var missing []string
		if obj["name"] == nil {
			missing = append(missing, "name")
		}
		if obj["offers"] == nil {
			missing = append(missing, "offers")
		}
		if len(missing) > 0 {
			return []models.CheckResult{{
				ID:       "schema.product.missing_fields",
				Category: "Structured Data",
				Severity: models.SeverityWarning,
				Message:  fmt.Sprintf("Product schema missing required fields: %s", strings.Join(missing, ", ")),
				URL:      p.URL,
			}}
		}
	}
	return nil
}

type schemaBreadcrumbInvalid struct{}

func (c *schemaBreadcrumbInvalid) Run(p *models.PageData) []models.CheckResult {
	for _, raw := range p.SchemaJSONRaw {
		var obj map[string]interface{}
		if err := json.Unmarshal([]byte(raw), &obj); err != nil {
			continue
		}
		if strings.ToLower(fmt.Sprintf("%v", obj["@type"])) != "breadcrumblist" {
			continue
		}
		items, ok := obj["itemListElement"].([]interface{})
		if !ok || len(items) == 0 {
			return []models.CheckResult{{
				ID:       "schema.breadcrumb.invalid",
				Category: "Structured Data",
				Severity: models.SeverityWarning,
				Message:  "BreadcrumbList schema missing itemListElement",
				URL:      p.URL,
			}}
		}
	}
	return nil
}

type schemaFAQInvalid struct{}

func (c *schemaFAQInvalid) Run(p *models.PageData) []models.CheckResult {
	for _, raw := range p.SchemaJSONRaw {
		var obj map[string]interface{}
		if err := json.Unmarshal([]byte(raw), &obj); err != nil {
			continue
		}
		if strings.ToLower(fmt.Sprintf("%v", obj["@type"])) != "faqpage" {
			continue
		}
		items, ok := obj["mainEntity"].([]interface{})
		if !ok || len(items) == 0 {
			return []models.CheckResult{{
				ID:       "schema.faq.invalid",
				Category: "Structured Data",
				Severity: models.SeverityWarning,
				Message:  "FAQPage schema missing mainEntity questions",
				URL:      p.URL,
			}}
		}
	}
	return nil
}

type schemaOrganizationMissingHomepage struct{}

func (c *schemaOrganizationMissingHomepage) Run(pages []*models.PageData) []models.CheckResult {
	for _, p := range pages {
		if p.Depth != 0 {
			continue
		}
		// Homepage found — check for org schema
		for _, raw := range p.SchemaJSONRaw {
			var obj map[string]interface{}
			if err := json.Unmarshal([]byte(raw), &obj); err != nil {
				continue
			}
			t := strings.ToLower(fmt.Sprintf("%v", obj["@type"]))
			if strings.Contains(t, "organization") || strings.Contains(t, "localbusiness") || strings.Contains(t, "corporation") {
				return nil
			}
		}
		return []models.CheckResult{{
			ID:       "schema.organization.missing_homepage",
			Category: "Structured Data",
			Severity: models.SeverityWarning,
			Message:  "Homepage has no Organization/LocalBusiness structured data",
			URL:      p.URL,
		}}
	}
	return nil
}

type schemaReviewRatingMissing struct{}

func (c *schemaReviewRatingMissing) Run(p *models.PageData) []models.CheckResult {
	nodes := schemaNodes(p.SchemaJSONRaw)
	if !looksLikeCarDetailPage(p, nodes) {
		return nil
	}
	if hasSchemaType(nodes, "review", "aggregaterating") || hasAnyProperty(nodes, "review", "aggregateRating") {
		return nil
	}
	return []models.CheckResult{{
		ID:       "schema.review_rating.missing",
		Category: structuredDataCategory,
		Severity: models.SeverityNotice,
		Message:  "Car detail page is missing Review or AggregateRating schema",
		URL:      p.URL,
	}}
}

type schemaLazyLoadedRisk struct{}

func (c *schemaLazyLoadedRisk) Run(p *models.PageData) []models.CheckResult {
	if strings.TrimSpace(p.RawHTML) == "" || !emptyLDJSONScriptPattern.MatchString(p.RawHTML) {
		return nil
	}
	return []models.CheckResult{{
		ID:       "schema.lazy_loaded_risk",
		Category: structuredDataCategory,
		Severity: models.SeverityWarning,
		Message:  "JSON-LD schema appears empty or lazy-loaded instead of present in raw HTML",
		URL:      p.URL,
		Details:  "Search crawlers should see JSON-LD in the initial HTML response.",
	}}
}

type schemaHiddenContentMismatch struct{}

func (c *schemaHiddenContentMismatch) Run(p *models.PageData) []models.CheckResult {
	nodes := schemaNodes(p.SchemaJSONRaw)
	if len(nodes) == 0 {
		return nil
	}
	visible := visiblePageText(p)
	if visible == "" {
		return nil
	}
	for _, candidate := range schemaVisibleTextCandidates(nodes) {
		if !textVisible(candidate, visible) {
			return []models.CheckResult{{
				ID:       "schema.hidden_content_mismatch",
				Category: structuredDataCategory,
				Severity: models.SeverityWarning,
				Message:  "Schema contains FAQ/HowTo content that is not visible on the page",
				URL:      p.URL,
				Details:  truncate(candidate, 120),
			}}
		}
	}
	return nil
}

type schemaHowToMissing struct{}

func (c *schemaHowToMissing) Run(p *models.PageData) []models.CheckResult {
	if !looksLikeHowToPage(p) || hasSchemaType(schemaNodes(p.SchemaJSONRaw), "howto") {
		return nil
	}
	return []models.CheckResult{{
		ID:       "schema.howto.missing",
		Category: structuredDataCategory,
		Severity: models.SeverityNotice,
		Message:  "Guide/process page appears eligible for HowTo schema",
		URL:      p.URL,
	}}
}

type schemaProductListMissing struct{}

func (c *schemaProductListMissing) Run(p *models.PageData) []models.CheckResult {
	if !looksLikeProductListPage(p) {
		return nil
	}
	nodes := schemaNodes(p.SchemaJSONRaw)
	if hasSchemaType(nodes, "itemlist", "productcollection", "collectionpage") || hasAnyProperty(nodes, "itemListElement") {
		return nil
	}
	return []models.CheckResult{{
		ID:       "schema.product_list.missing",
		Category: structuredDataCategory,
		Severity: models.SeverityNotice,
		Message:  "Listing/category page is missing ItemList/ProductList structured data",
		URL:      p.URL,
	}}
}

type schemaEventMissing struct{}

func (c *schemaEventMissing) Run(p *models.PageData) []models.CheckResult {
	if !looksLikeEventPage(p) || hasSchemaType(schemaNodes(p.SchemaJSONRaw), "event") {
		return nil
	}
	return []models.CheckResult{{
		ID:       "schema.event.missing",
		Category: structuredDataCategory,
		Severity: models.SeverityNotice,
		Message:  "Event-like page is missing Event schema",
		URL:      p.URL,
	}}
}

type schemaSpeakableMissing struct{}

func (c *schemaSpeakableMissing) Run(p *models.PageData) []models.CheckResult {
	nodes := schemaNodes(p.SchemaJSONRaw)
	if !looksLikeSpeakableCandidate(p, nodes) || hasAnyProperty(nodes, "speakable") || hasSchemaType(nodes, "speakablespecification") {
		return nil
	}
	return []models.CheckResult{{
		ID:       "schema.speakable.missing",
		Category: structuredDataCategory,
		Severity: models.SeverityNotice,
		Message:  "News/article page is missing SpeakableSpecification schema",
		URL:      p.URL,
	}}
}

type schemaWebSiteMissingHomepage struct{}

func (c *schemaWebSiteMissingHomepage) Run(pages []*models.PageData) []models.CheckResult {
	home := homepage(pages)
	if home == nil || hasSchemaType(schemaNodes(home.SchemaJSONRaw), "website") {
		return nil
	}
	return []models.CheckResult{{
		ID:       "schema.website.missing_homepage",
		Category: structuredDataCategory,
		Severity: models.SeverityWarning,
		Message:  "Homepage has no WebSite structured data",
		URL:      home.URL,
	}}
}

type schemaSearchActionMissingHomepage struct{}

func (c *schemaSearchActionMissingHomepage) Run(pages []*models.PageData) []models.CheckResult {
	home := homepage(pages)
	if home == nil {
		return nil
	}
	nodes := schemaNodes(home.SchemaJSONRaw)
	if !hasSchemaType(nodes, "website") || hasSchemaType(nodes, "searchaction") {
		return nil
	}
	return []models.CheckResult{{
		ID:       "schema.website.searchaction_missing",
		Category: structuredDataCategory,
		Severity: models.SeverityNotice,
		Message:  "WebSite schema is missing SearchAction for sitelinks searchbox",
		URL:      home.URL,
	}}
}

type schemaNode map[string]interface{}

func schemaNodes(raws []string) []schemaNode {
	var nodes []schemaNode
	for _, raw := range raws {
		var value interface{}
		if err := json.Unmarshal([]byte(raw), &value); err != nil {
			continue
		}
		flattenSchemaValue(value, &nodes)
	}
	return nodes
}

func flattenSchemaValue(value interface{}, nodes *[]schemaNode) {
	switch v := value.(type) {
	case map[string]interface{}:
		if _, ok := v["@type"]; ok {
			*nodes = append(*nodes, schemaNode(v))
		}
		for _, child := range v {
			flattenSchemaValue(child, nodes)
		}
	case []interface{}:
		for _, child := range v {
			flattenSchemaValue(child, nodes)
		}
	}
}

func hasSchemaType(nodes []schemaNode, types ...string) bool {
	for _, node := range nodes {
		if schemaTypeMatches(node, types...) {
			return true
		}
	}
	return false
}

func schemaTypeMatches(node schemaNode, types ...string) bool {
	actual := schemaTypeNames(node["@type"])
	for _, expected := range types {
		expected = normalizeSchemaType(expected)
		for _, t := range actual {
			if t == expected {
				return true
			}
		}
	}
	return false
}

func schemaTypeNames(raw interface{}) []string {
	switch v := raw.(type) {
	case string:
		return []string{normalizeSchemaType(v)}
	case []interface{}:
		var out []string
		for _, item := range v {
			if s, ok := item.(string); ok {
				out = append(out, normalizeSchemaType(s))
			}
		}
		return out
	default:
		return nil
	}
}

func normalizeSchemaType(value string) string {
	value = strings.TrimSpace(strings.ToLower(value))
	value = strings.TrimPrefix(value, "https://schema.org/")
	value = strings.TrimPrefix(value, "http://schema.org/")
	return strings.TrimPrefix(value, "schema:")
}

func hasAnyProperty(nodes []schemaNode, keys ...string) bool {
	for _, node := range nodes {
		for _, key := range keys {
			if value, ok := findProperty(node, key); ok && value != nil {
				return true
			}
		}
	}
	return false
}

func findProperty(node map[string]interface{}, key string) (interface{}, bool) {
	for k, v := range node {
		if strings.EqualFold(k, key) {
			return v, true
		}
	}
	return nil, false
}

func homepage(pages []*models.PageData) *models.PageData {
	for _, p := range pages {
		if p != nil && p.Depth == 0 {
			return p
		}
	}
	for _, p := range pages {
		if p != nil && isHomepageURL(p.FinalURL) {
			return p
		}
	}
	return nil
}

func isHomepageURL(raw string) bool {
	u, err := url.Parse(raw)
	return err == nil && (u.Path == "" || u.Path == "/")
}

func pageSignal(p *models.PageData) string {
	parts := []string{p.URL, p.FinalURL, p.Title, p.BodyText}
	parts = append(parts, p.H1s...)
	parts = append(parts, p.H2s...)
	return strings.ToLower(strings.Join(parts, " "))
}

func looksLikeCarDetailPage(p *models.PageData, nodes []schemaNode) bool {
	if hasSchemaType(nodes, "vehicle", "car", "product") {
		return true
	}
	signal := pageSignal(p)
	return containsAny(signal, "/buy-used-car/", "/used-car/", "used car details", "vehicle details") &&
		containsAny(signal, "price", "km", "transmission", "fuel", "owner")
}

func looksLikeHowToPage(p *models.PageData) bool {
	signal := pageSignal(p)
	return containsAny(signal, "how to", "how-to", "/guide", "/guides", "step by step", "steps to", "process")
}

func looksLikeProductListPage(p *models.PageData) bool {
	signal := pageSignal(p)
	if containsAny(signal, "used cars", "cars for sale", "buy used car", "/buy-used-cars", "/used-cars") {
		return containsAtLeast(signal, 2, "filter", "sort", "price", "model", "year", "fuel", "km")
	}
	return containsAny(signal, "category page", "listing page", "search results") &&
		containsAtLeast(signal, 2, "filter", "sort", "price", "cars")
}

func looksLikeEventPage(p *models.PageData) bool {
	signal := pageSignal(p)
	return containsAny(signal, "event", "launch", "expo", "webinar", "auction", "fair") &&
		containsAny(signal, "date", "venue", "register", "schedule", "location")
}

func looksLikeSpeakableCandidate(p *models.PageData, nodes []schemaNode) bool {
	if hasSchemaType(nodes, "newsarticle") {
		return true
	}
	signal := pageSignal(p)
	return containsAny(signal, "/news/", "news article") && hasSchemaType(nodes, "article", "blogposting")
}

func containsAny(value string, needles ...string) bool {
	for _, needle := range needles {
		if strings.Contains(value, needle) {
			return true
		}
	}
	return false
}

func containsAtLeast(value string, count int, needles ...string) bool {
	found := 0
	for _, needle := range needles {
		if strings.Contains(value, needle) {
			found++
		}
	}
	return found >= count
}

func visiblePageText(p *models.PageData) string {
	parts := []string{p.Title, p.BodyText}
	parts = append(parts, p.H1s...)
	parts = append(parts, p.H2s...)
	parts = append(parts, p.H3s...)
	return normalizeVisibleText(strings.Join(parts, " "))
}

func schemaVisibleTextCandidates(nodes []schemaNode) []string {
	var out []string
	for _, node := range nodes {
		switch {
		case schemaTypeMatches(node, "faqpage", "question"):
			collectSchemaStrings(node, &out, "name", "text")
			if answer, ok := findProperty(node, "acceptedAnswer"); ok {
				collectSchemaValueStrings(answer, &out, "text", "name")
			}
		case schemaTypeMatches(node, "howto", "howtostep", "howtosection"):
			collectSchemaStrings(node, &out, "name", "text")
			if steps, ok := findProperty(node, "step"); ok {
				collectSchemaValueStrings(steps, &out, "name", "text")
			}
		}
	}
	return uniqueMeaningfulStrings(out)
}

func collectSchemaStrings(node map[string]interface{}, out *[]string, keys ...string) {
	for _, key := range keys {
		if raw, ok := findProperty(node, key); ok {
			collectSchemaValueStrings(raw, out)
		}
	}
}

func collectSchemaValueStrings(value interface{}, out *[]string, keys ...string) {
	switch v := value.(type) {
	case string:
		*out = append(*out, v)
	case map[string]interface{}:
		if len(keys) == 0 {
			for _, child := range v {
				collectSchemaValueStrings(child, out)
			}
			return
		}
		for _, key := range keys {
			if child, ok := findProperty(v, key); ok {
				collectSchemaValueStrings(child, out)
			}
		}
	case []interface{}:
		for _, child := range v {
			collectSchemaValueStrings(child, out, keys...)
		}
	}
}

func uniqueMeaningfulStrings(values []string) []string {
	seen := map[string]bool{}
	var out []string
	for _, value := range values {
		value = strings.TrimSpace(value)
		if len(value) < 12 {
			continue
		}
		key := normalizeVisibleText(value)
		if key == "" || seen[key] {
			continue
		}
		seen[key] = true
		out = append(out, value)
	}
	return out
}

func textVisible(candidate string, visible string) bool {
	candidate = normalizeVisibleText(candidate)
	if len(candidate) < 12 {
		return true
	}
	if strings.Contains(visible, candidate) {
		return true
	}
	words := strings.Fields(candidate)
	if len(words) < 4 {
		return false
	}
	matched := 0
	for _, word := range words {
		if len(word) >= 4 && strings.Contains(visible, word) {
			matched++
		}
	}
	return float64(matched)/float64(len(words)) >= 0.7
}

func normalizeVisibleText(value string) string {
	value = strings.ToLower(value)
	value = whitespacePattern.ReplaceAllString(value, " ")
	return strings.TrimSpace(value)
}

func truncate(value string, max int) string {
	value = strings.TrimSpace(value)
	if len(value) <= max {
		return value
	}
	return value[:max] + "..."
}
