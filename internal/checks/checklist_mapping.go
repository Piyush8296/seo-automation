package checks

import (
	"strings"

	"github.com/cars24/seo-automation/internal/models"
)

// checklistIDMap links the internal check IDs to the checklist registry IDs in
// /Users/user/cars24/screamingfrog/rules/checklist_registry.json.
//
// Some internal checks are narrower than the registry row and some are broader.
// In those cases the mapping points to the closest checklist item that owns the
// same remediation/evidence workflow.
var checklistIDMap = map[string][]string{
	// AMP
	"amp.canonical.missing":        {"CANONICAL-017"},
	"amp.canonical.points_to_amp":  {"CANONICAL-017"},
	"amp.canonical.self_reference": {"CANONICAL-017"},
	"amp.regular.missing_amphtml":  {"MOBILE-011"},

	// Canonical
	"canonical.chain":                   {"CANONICAL-011"},
	"canonical.conflict_og_url":         {"HTML-004"},
	"canonical.country_folder_mismatch": {"CANONICAL-022"},
	"canonical.has_fragment":            {"CANONICAL-021"},
	"canonical.header_mismatch":         {"CANONICAL-009"},
	"canonical.insecure":                {"CANONICAL-013"},
	"canonical.in_body":                 {"CANONICAL-020"},
	"canonical.loop":                    {"CANONICAL-010"},
	"canonical.missing":                 {"CANONICAL-001"},
	"canonical.multiple":                {"CANONICAL-007"},
	"canonical.not_absolute":            {"CANONICAL-006"},
	"canonical.params_self_reference":   {"CANONICAL-019"},
	"canonical.points_elsewhere":        {"CANONICAL-002", "CANONICAL-003"},
	"canonical.target_non_200":          {"CANONICAL-008"},
	"canonical.www_variant":             {"CANONICAL-014"},

	// Content
	"body.exact_duplicate": {"CONTENT-002"},
	"body.lorem_ipsum":     {"CONTENT-005"},
	"body.near_duplicate":  {"CONTENT-002"},
	"body.noindex_meta":    {"INDEX-006"},
	"body.thin":            {"CONTENT-001"},
	"body.title_equals_h1": {"TITLE-011", "H1-005"},
	"body.very_thin":       {"CONTENT-001"},

	// Core Web Vitals
	"cwv.cls.above_fold_images_no_dims":        {"CWV-003", "CWV-006", "IMG-015"},
	"cwv.cls.font_display_missing":             {"SPEED-019"},
	"cwv.fid.blocking_scripts":                 {"CWV-002", "CWV-007", "SPEED-027"},
	"cwv.inp.heavy_third_party_scripts":        {"CWV-002", "CWV-007", "SPEED-020"},
	"cwv.lcp.background_image":                 {"CWV-005", "SPEED-023"},
	"cwv.lcp.font_preload_missing_crossorigin": {"SPEED-019", "SPEED-026"},
	"cwv.lcp.image_not_preloaded":              {"CWV-005", "IMG-009"},

	// Crawl Budget
	"crawl_budget.faceted_navigation":       {"CRAWL-001", "URL-013"},
	"crawl_budget.high_waste_ratio":         {"CRAWL-012"},
	"crawl_budget.low_value_archive":        {"DUPE-010"},
	"crawl_budget.low_value_page":           {"CRAWL-012", "CONTENT-001"},
	"crawl_budget.moderate_waste_ratio":     {"CRAWL-012"},
	"crawl_budget.search_page_indexable":    {"CRAWL-002", "ROBOTS-012"},
	"crawl_budget.sitemap_noindex_conflict": {"SITEMAP-010"},
	"crawl_budget.tracking_params":          {"DUPE-001", "URL-005"},

	// Crawlability
	"crawl.noindex.in_sitemap":                   {"SITEMAP-010", "INDEX-006"},
	"crawl.page_depth.orphan_external_only":      {"CRAWL-008", "INTLINK-003"},
	"crawl.page_depth.too_deep":                  {"CRAWL-005", "INTLINK-010"},
	"crawl.redirect.302_permanent":               {"REDIRECT-016"},
	"crawl.redirect.chain":                       {"REDIRECT-002", "REDIRECT-008"},
	"crawl.redirect.destination_not_indexable":   {"REDIRECT-013"},
	"crawl.redirect.http_to_https_not_301":       {"REDIRECT-004"},
	"crawl.redirect.javascript":                  {"REDIRECT-011"},
	"crawl.redirect.loop":                        {"REDIRECT-003"},
	"crawl.redirect.meta_refresh":                {"REDIRECT-012"},
	"crawl.redirect.trailing_slash_inconsistent": {"REDIRECT-006"},
	"crawl.redirect.www_variant_not_301":         {"REDIRECT-005"},
	"crawl.response.4xx":                         {"STATUS-002", "INDEX-STATE-OTHER4XX"},
	"crawl.response.5xx":                         {"STATUS-003", "INDEX-STATE-5XX", "ALERT-001"},
	"crawl.response.content_type_mismatch":       {"STATUS-009"},
	"crawl.response.image_non_200":               {"STATUS-006"},
	"crawl.response.soft_404":                    {"STATUS-004"},
	"crawl.response.timeout":                     {"ALERT-001", "SPEED-015"},
	"crawl.response.vary_header_invalid":         {"STATUS-010"},
	"crawl.robots.directive_conflict":            {"ROBOTS-008", "HTML-007"},
	"crawl.robots.noarchive":                     {"HTML-007", "INDEX-006"},
	"crawl.robots.nofollow_page":                 {"INTLINK-009"},
	"crawl.robots.nosnippet":                     {"HTML-007", "INDEX-006"},
	"crawl.robots.page_blocked_but_linked":       {"ROBOTS-005"},
	"crawl.robots.x_robots_tag":                  {"HTML-007"},

	// E-E-A-T
	"eeat.about_page.missing":        {"CONTENT-003"},
	"eeat.about_page.thin":           {"CONTENT-003", "CONTENT-001"},
	"eeat.author_info.missing":       {"CONTENT-004"},
	"eeat.breadcrumbs.missing":       {"INTLINK-007", "SCHEMA-003"},
	"eeat.contact_page.insufficient": {"CONTENT-003", "SCHEMA-001"},
	"eeat.contact_page.missing":      {"CONTENT-003", "SCHEMA-001"},
	"eeat.dates.missing":             {"CONTENT-006"},
	"eeat.privacy_policy.thin":       {"CONTENT-003"},

	// Headings
	"headings.h1.duplicate":                  {"H1-004"},
	"headings.h1.empty":                      {"H1-001", "H1-014"},
	"headings.h1.missing":                    {"H1-001"},
	"headings.h1.multiple":                   {"H1-002"},
	"headings.h1.primary_keyword_missing":    {"H1-003"},
	"headings.h1.too_long":                   {"H1-014"},
	"headings.h1.too_short":                  {"H1-014"},
	"headings.h2.keyword_stuffing":           {"H1-015"},
	"headings.h2.missing":                    {"H1-006"},
	"headings.h2.secondary_keywords_missing": {"H1-009"},
	"headings.hierarchy.skipped_level":       {"H1-008", "H1-007", "H1-011"},

	// HTML Structure
	"html.doctype_missing":         {"HTML-002"},
	"html.dom_too_deep":            {"HTML-003"},
	"html.multiple_head":           {"HTML-001"},
	"html.pagination_link_invalid": {"HTML-006"},
	"html.robots_meta_in_body":     {"HTML-005"},

	// HTTPS & Security
	"https.mixed_content":                 {"SSL-006"},
	"https.page_insecure":                 {"SSL-001"},
	"security.csp.missing":                {"CRAWL-SEC-CSP"},
	"security.hsts.max_age_too_short":     {"SSL-005", "CRAWL-SEC-HSTS"},
	"security.hsts.missing":               {"SSL-005", "CRAWL-SEC-HSTS"},
	"security.permissions_policy.missing": {"CRAWL-SEC-CSP"},
	"security.referrer_policy.missing":    {"CRAWL-SEC-REFPOL"},
	"security.x_frame.missing":            {"CRAWL-SEC-XFO"},
	"security.xcto.missing":               {"CRAWL-SEC-XCTO"},
	"ssl.cert_expired":                    {"SSL-002"},
	"ssl.cert_expiring_soon":              {"SSL-002"},
	"ssl.cert_mismatch":                   {"SSL-003"},
	"ssl.chain_incomplete":                {"SSL-008"},
	"ssl.hsts_preload_missing":            {"SSL-005"},
	"ssl.tls_version_old":                 {"SSL-009"},

	// Images
	"images.alt.empty_non_decorative": {"IMG-004", "IMG-002"},
	"images.alt.is_filename":          {"IMG-002", "IMG-005"},
	"images.alt.missing":              {"IMG-001"},
	"images.alt.too_long":             {"IMG-002", "IMG-003"},
	"images.broken":                   {"IMG-014"},
	"images.dimensions.missing":       {"IMG-015"},
	"images.format.not_modern":        {"IMG-007"},
	"images.lazy.above_fold":          {"IMG-009", "IMG-008"},
	"images.missing_srcset":           {"IMG-006"},
	"images.no_width_height_cls":      {"IMG-015", "CWV-006"},
	"images.size.too_large":           {"IMG-006", "IMG-007", "SPEED-012"},

	// Internal Linking
	"links.anchor.empty":              {"INTLINK-002"},
	"links.anchor.generic":            {"INTLINK-002"},
	"links.external.broken_4xx":       {"STATUS-008"},
	"links.external.broken_5xx":       {"STATUS-008"},
	"links.external.missing_noopener": {"CRAWL-SEC-XORIGIN"},
	"links.external.redirect":         {"STATUS-008"},
	"links.external.timeout":          {"STATUS-008"},
	"links.footer_heavy":              {"INTLINK-008", "INTLINK-014"},
	"links.internal.broken_4xx":       {"INTLINK-004", "STATUS-001"},
	"links.internal.broken_5xx":       {"INTLINK-004", "STATUS-001"},
	"links.internal.nofollow":         {"INTLINK-009"},
	"links.internal.to_redirect":      {"REDIRECT-002", "STATUS-001"},
	"links.nav_orphan":                {"INTLINK-003", "INTLINK-015"},
	"links.no_content_links":          {"INTLINK-005"},
	"links.page.low_inlinks":          {"CRAWL-005", "INTLINK-010"},
	"links.page.orphan":               {"INTLINK-003", "CRAWL-008"},
	"links.page.too_many_outlinks":    {"INTLINK-014"},

	// International
	"i18n.hreflang.invalid_lang_code":   {"HREFLANG-002"},
	"i18n.hreflang.missing_return_link": {"HREFLANG-003"},
	"i18n.hreflang.missing_self_ref":    {"HREFLANG-001"},
	"i18n.hreflang.missing_x_default":   {"HREFLANG-004"},
	"i18n.hreflang.non_canonical_url":   {"CANONICAL-018", "HREFLANG-005"},
	"i18n.hreflang.url_not_absolute":    {"HREFLANG-005"},

	// Meta Descriptions
	"meta_desc.cta_missing":             {"META-005"},
	"meta_desc.duplicate":               {"META-002"},
	"meta_desc.dynamic_placeholder":     {"META-009"},
	"meta_desc.missing":                 {"META-001"},
	"meta_desc.primary_keyword_missing": {"META-004"},
	"meta_desc.too_long":                {"META-001"},
	"meta_desc.too_short":               {"META-001"},

	// Mobile
	"mobile.font_size.too_small":    {"MOBILE-004"},
	"mobile.user_scalable.disabled": {"MOBILE-001"},
	"mobile.viewport.invalid":       {"MOBILE-002"},
	"mobile.viewport.missing":       {"MOBILE-002"},

	// Mobile vs Desktop
	"mob_desk.canonical_mismatch":   {"PARITY-002", "CANONICAL-018"},
	"mob_desk.content_mismatch":     {"MOBILE-007", "PARITY-002"},
	"mob_desk.h1_mismatch":          {"PARITY-003"},
	"mob_desk.links_mismatch":       {"MOBILE-007", "PARITY-002", "INTLINK-015"},
	"mob_desk.meta_desc_mismatch":   {"MOBILE-007", "PARITY-002"},
	"mob_desk.og_image_mismatch":    {"MOBILE-007", "PARITY-002"},
	"mob_desk.schema_mismatch":      {"MOBILE-007", "PARITY-002"},
	"mob_desk.separate_mobile_site": {"PARITY-001"},
	"mob_desk.status_mismatch":      {"PARITY-001", "STATUS-001"},
	"mob_desk.title_mismatch":       {"MOBILE-007", "PARITY-002"},

	// Pagination
	"pagination.canonical_first_page":            {"CANONICAL-023"},
	"pagination.inconsistent_canonical_strategy": {"CRAWL-003", "CANONICAL-005"},
	"pagination.missing_canonical":               {"CANONICAL-004"},
	"pagination.noindex":                         {"INDEX-006", "CANONICAL-016"},
	"pagination.thin_content":                    {"DUPE-009"},
	"pagination.title_not_unique":                {"TITLE-002"},

	// Performance
	"perf.cache_control.missing":         {"SPEED-013"},
	"perf.cls_risk.images_no_dimensions": {"CWV-003", "IMG-015"},
	"perf.compression.missing":           {"SPEED-016"},
	"perf.external_scripts.count":        {"SPEED-006", "SPEED-020"},
	"perf.html_size.large":               {"CRAWL-VAL-2MB", "SPEED-005"},
	"perf.inline_css.large":              {"SPEED-009"},
	"perf.lcp_candidate.lazy_loaded":     {"CWV-005", "IMG-009"},
	"perf.missing_preconnect":            {"SPEED-018"},
	"perf.render_blocking.css_count":     {"SPEED-007"},
	"perf.render_blocking.scripts":       {"SPEED-008"},
	"perf.response_time.critical":        {"SPEED-015", "CWV-010"},
	"perf.response_time.slow":            {"SPEED-015", "CWV-010"},

	// Resources
	"resources.css.broken":           {"STATUS-007"},
	"resources.font.broken":          {"STATUS-007", "SPEED-019"},
	"resources.font.no_display_swap": {"SPEED-019"},
	"resources.script.broken":        {"STATUS-007"},
	"resources.too_many_requests":    {"SPEED-006"},
	"resources.total_size_too_large": {"SPEED-005"},

	// Sitemap
	"sitemap.coverage_low":   {"SITEMAP-001"},
	"sitemap.missing":        {"SITEMAP-001"},
	"sitemap.too_large":      {"SITEMAP-005", "SITEMAP-006"},
	"sitemap.url_4xx":        {"SITEMAP-007"},
	"sitemap.url_blocked":    {"SITEMAP-011", "SITEMAP-021"},
	"sitemap.url_noindex":    {"SITEMAP-010"},
	"sitemap.url_redirected": {"SITEMAP-007"},

	// Social metadata
	"og.description.missing":    {"META-001"},
	"og.image.missing":          {"CONTENT-014"},
	"og.title.missing":          {"TITLE-014"},
	"og.url.mismatch_canonical": {"HTML-004"},
	"og.url.missing":            {"HTML-004"},
	"twitter.card.missing":      {"CONTENT-014"},
	"twitter.image.missing":     {"CONTENT-014"},
	"twitter.title.missing":     {"TITLE-014"},

	// Structured Data
	"schema.article.missing_fields":        {"SCHEMA-013"},
	"schema.breadcrumb.invalid":            {"SCHEMA-003"},
	"schema.event.missing":                 {"SCHEMA-019"},
	"schema.faq.invalid":                   {"SCHEMA-004"},
	"schema.hidden_content_mismatch":       {"SCHEMA-010"},
	"schema.howto.missing":                 {"SCHEMA-012"},
	"schema.jsonld.duplicate_type":         {"SCHEMA-015"},
	"schema.jsonld.invalid_json":           {"SCHEMA-008", "SCHEMA-017"},
	"schema.jsonld.missing":                {"SCHEMA-016"},
	"schema.jsonld.missing_context":        {"SCHEMA-015"},
	"schema.jsonld.missing_type":           {"SCHEMA-015"},
	"schema.lazy_loaded_risk":              {"SCHEMA-009"},
	"schema.organization.missing_homepage": {"SCHEMA-001"},
	"schema.product_list.missing":          {"SCHEMA-018"},
	"schema.product.missing_fields":        {"SCHEMA-002", "SCHEMA-011"},
	"schema.review_rating.missing":         {"SCHEMA-005"},
	"schema.speakable.missing":             {"SCHEMA-020"},
	"schema.website.missing_homepage":      {"SCHEMA-014"},
	"schema.website.searchaction_missing":  {"SCHEMA-007"},

	// Titles
	"keyword.density_out_of_range":     {"KW-002"},
	"keyword.first_100_missing":        {"KW-001"},
	"keyword.surface_mismatch":         {"KW-009"},
	"title.blog_keyword_missing":       {"TITLE-020"},
	"title.brand_missing":              {"TITLE-007"},
	"title.city_model_missing":         {"TITLE-022"},
	"title.duplicate":                  {"TITLE-002"},
	"title.dynamic_placeholder":        {"TITLE-010"},
	"title.homepage_not_brand_focused": {"TITLE-015"},
	"title.keyword_not_near_start":     {"TITLE-005"},
	"title.keyword_stuffing":           {"TITLE-008"},
	"title.missing":                    {"TITLE-001"},
	"title.model_missing":              {"TITLE-017"},
	"title.primary_keyword_missing":    {"TITLE-004"},
	"title.special_chars":              {"TITLE-013"},
	"title.too_long":                   {"TITLE-003", "TITLE-006"},
	"title.too_short":                  {"TITLE-003"},
	"title.topic_mismatch":             {"TITLE-009"},
	"title.year_missing":               {"TITLE-019"},

	// URL Structure
	"url.breadcrumb_mismatch":         {"URL-012"},
	"url.consistent_structure":        {"URL-010"},
	"url.contains_stop_words":         {"URL-011"},
	"url.double_slash":                {"URL-002"},
	"url.has_session_params":          {"URL-014"},
	"url.has_spaces":                  {"URL-002"},
	"url.has_underscores":             {"URL-002"},
	"url.has_uppercase":               {"URL-001"},
	"url.keyword_topic_mismatch":      {"URL-007"},
	"url.non_ascii":                   {"URL-016"},
	"url.non_descriptive":             {"URL-003"},
	"url.path_depth_too_deep":         {"URL-004"},
	"url.print_canonical_conflict":    {"URL-015"},
	"url.too_long":                    {"URL-009"},
	"url.too_many_params":             {"URL-005", "URL-013"},
	"url.trailing_slash_inconsistent": {"URL-017"},
}

// ChecklistIDsFor returns the registry checklist IDs associated with an
// internal check ID. Evidence checks that already use registry IDs are mapped to
// themselves.
func ChecklistIDsFor(checkID string) []string {
	if ids, ok := checklistIDMap[checkID]; ok {
		return cloneChecklistIDs(ids)
	}
	if looksLikeChecklistRegistryID(checkID) {
		return []string{checkID}
	}
	return nil
}

// AttachChecklistMappings decorates audit results with checklist registry IDs.
func AttachChecklistMappings(audit *models.SiteAudit) {
	if audit == nil {
		return
	}
	for _, page := range audit.Pages {
		for i := range page.CheckResults {
			attachChecklistIDsToResult(&page.CheckResults[i])
		}
	}
	for i := range audit.SiteChecks {
		attachChecklistIDsToResult(&audit.SiteChecks[i])
	}
	for i := range audit.CrawlerEvidence {
		attachChecklistIDsToEvidence(&audit.CrawlerEvidence[i])
	}
	for i := range audit.RenderedSEO {
		attachChecklistIDsToEvidence(&audit.RenderedSEO[i])
	}
}

func attachChecklistIDsToResult(result *models.CheckResult) {
	if result == nil || len(result.ChecklistIDs) > 0 {
		return
	}
	result.ChecklistIDs = ChecklistIDsFor(result.ID)
}

func attachChecklistIDsToEvidence(result *models.EvidenceCheckResult) {
	if result == nil || len(result.ChecklistIDs) > 0 {
		return
	}
	result.ChecklistIDs = ChecklistIDsFor(result.ID)
}

func looksLikeChecklistRegistryID(id string) bool {
	return id != "" && strings.Contains(id, "-") && !strings.Contains(id, ".") && id == strings.ToUpper(id)
}

func cloneChecklistIDs(ids []string) []string {
	if len(ids) == 0 {
		return nil
	}
	out := make([]string, len(ids))
	copy(out, ids)
	return out
}
