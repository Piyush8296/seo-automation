package checks

import (
	"sort"

	"github.com/cars24/seo-automation/internal/models"
)

// checkDescriptors is the authoritative list of all registered checks.
// Add new check IDs here alongside their implementation in the relevant package.
var checkDescriptors = []models.CheckDescriptor{
	// ── AMP ─────────────────────────────────────────────────────────────────
	{ID: "amp.canonical.missing", Category: "AMP", Description: "AMP page is missing a canonical link pointing to the non-AMP version."},
	{ID: "amp.canonical.points_to_amp", Category: "AMP", Description: "AMP page's canonical points to another AMP page instead of the HTML version."},
	{ID: "amp.canonical.self_reference", Category: "AMP", Description: "AMP page canonicalises to itself instead of the desktop URL."},
	{ID: "amp.regular.missing_amphtml", Category: "AMP", Description: "Regular page is missing a <link rel=\"amphtml\"> tag pointing to its AMP counterpart."},

	// ── Canonical ───────────────────────────────────────────────────────────
	{ID: "canonical.conflict_og_url", Category: "Canonical", Description: "Canonical URL does not match the og:url property, sending mixed signals."},
	{ID: "canonical.chain", Category: "Canonical", Description: "Canonical target points to another canonical target, creating a canonical chain."},
	{ID: "canonical.country_folder_mismatch", Category: "Canonical", Description: "Canonical URL points to a different country or locale folder than the current page."},
	{ID: "canonical.has_fragment", Category: "Canonical", Description: "Canonical URL contains a fragment identifier, which search engines ignore."},
	{ID: "canonical.header_mismatch", Category: "Canonical", Description: "HTTP Link header canonical does not match the HTML canonical tag."},
	{ID: "canonical.insecure", Category: "Canonical", Description: "Canonical URL uses http:// but the page is served over https://."},
	{ID: "canonical.in_body", Category: "Canonical", Description: "Canonical tag appears outside the document head."},
	{ID: "canonical.loop", Category: "Canonical", Description: "Canonical URLs form a loop between crawled pages."},
	{ID: "canonical.missing", Category: "Canonical", Description: "Page has no canonical tag, leaving search engines to choose the preferred URL."},
	{ID: "canonical.multiple", Category: "Canonical", Description: "Page has more than one canonical tag."},
	{ID: "canonical.not_absolute", Category: "Canonical", Description: "Canonical URL is relative; search engines require an absolute URL."},
	{ID: "canonical.params_self_reference", Category: "Canonical", Description: "Parameterized URL is self-canonical instead of consolidating to a clean URL."},
	{ID: "canonical.points_elsewhere", Category: "Canonical", Description: "Canonical points to a different URL, consolidating link signals away from this page."},
	{ID: "canonical.target_non_200", Category: "Canonical", Description: "Canonical target does not return HTTP 200 within the crawl."},
	{ID: "canonical.www_variant", Category: "Canonical", Description: "Canonical host alternates between www and non-www variants."},

	// ── Content ─────────────────────────────────────────────────────────────
	{ID: "body.exact_duplicate", Category: "Content", Description: "Page body is an exact duplicate of another crawled page."},
	{ID: "body.lorem_ipsum", Category: "Content", Description: "Page contains placeholder lorem ipsum text that was never replaced."},
	{ID: "body.near_duplicate", Category: "Content", Description: "Page body is nearly identical to another crawled page (high simhash similarity)."},
	{ID: "body.noindex_meta", Category: "Content", Description: "Page has a noindex meta tag but is still linked from internal pages."},
	{ID: "body.thin", Category: "Content", Description: "Page has very few words; thin content may be treated as low quality by search engines."},
	{ID: "body.title_equals_h1", Category: "Content", Description: "The page title and H1 are identical, a missed opportunity for keyword variation."},
	{ID: "body.very_thin", Category: "Content", Description: "Page has extremely few words; likely insufficient content for indexing."},

	// ── Core Web Vitals ─────────────────────────────────────────────────────
	{ID: "cwv.cls.above_fold_images_no_dims", Category: "Core Web Vitals", Description: "Above-the-fold images have no width/height attributes, causing layout shift (CLS)."},
	{ID: "cwv.cls.font_display_missing", Category: "Core Web Vitals", Description: "Font lacks font-display property, risking invisible text during load (FOIT → CLS)."},
	{ID: "cwv.fid.blocking_scripts", Category: "Core Web Vitals", Description: "Synchronous scripts block the main thread, delaying first input response (FID)."},
	{ID: "cwv.inp.heavy_third_party_scripts", Category: "Core Web Vitals", Description: "Heavy third-party scripts increase Interaction to Next Paint (INP)."},
	{ID: "cwv.lcp.background_image", Category: "Core Web Vitals", Description: "LCP element is a CSS background image; not eligible for preload priority hints."},
	{ID: "cwv.lcp.font_preload_missing_crossorigin", Category: "Core Web Vitals", Description: "Font preload link is missing crossorigin attribute, causing a double fetch."},
	{ID: "cwv.lcp.image_not_preloaded", Category: "Core Web Vitals", Description: "Largest Contentful Paint image is not preloaded, delaying the LCP metric."},

	// ── Crawl Budget ────────────────────────────────────────────────────────
	{ID: "crawl_budget.faceted_navigation", Category: "Crawl Budget", Description: "Faceted navigation URLs are indexable, wasting crawl budget on near-duplicate pages."},
	{ID: "crawl_budget.high_waste_ratio", Category: "Crawl Budget", Description: "Large proportion of crawled pages are non-canonical or low-value."},
	{ID: "crawl_budget.low_value_archive", Category: "Crawl Budget", Description: "Date-based archive URLs are indexable and consuming crawl budget unnecessarily."},
	{ID: "crawl_budget.low_value_page", Category: "Crawl Budget", Description: "Page shows signals of low value: thin content, no inlinks, or non-canonical."},
	{ID: "crawl_budget.moderate_waste_ratio", Category: "Crawl Budget", Description: "Moderate share of crawled pages are duplicates or low-value."},
	{ID: "crawl_budget.search_page_indexable", Category: "Crawl Budget", Description: "Internal search result pages are indexable, wasting crawl budget."},
	{ID: "crawl_budget.sitemap_noindex_conflict", Category: "Crawl Budget", Description: "Sitemap lists URLs that are noindexed, sending conflicting signals to crawlers."},
	{ID: "crawl_budget.tracking_params", Category: "Crawl Budget", Description: "URLs with tracking parameters are indexable, creating unnecessary duplicate pages."},

	// ── Crawlability ────────────────────────────────────────────────────────
	{ID: "crawl.noindex.in_sitemap", Category: "Crawlability", Description: "Noindexed page is listed in the sitemap, sending conflicting directives to Googlebot."},
	{ID: "crawl.page_depth.orphan_external_only", Category: "Crawlability", Description: "Page has no internal links; only reachable via external sources."},
	{ID: "crawl.page_depth.too_deep", Category: "Crawlability", Description: "Page is buried too many clicks from the homepage, reducing crawl priority."},
	{ID: "crawl.redirect.302_permanent", Category: "Crawlability", Description: "Permanent redirect is implemented as a temporary 302 instead of 301."},
	{ID: "crawl.redirect.chain", Category: "Crawlability", Description: "Page is part of a redirect chain, wasting crawl budget and losing link equity."},
	{ID: "crawl.redirect.destination_not_indexable", Category: "Crawlability", Description: "Redirect destination is not indexable because it is non-200 or marked noindex."},
	{ID: "crawl.redirect.http_to_https_not_301", Category: "Crawlability", Description: "HTTP URL does not permanently redirect to the HTTPS version."},
	{ID: "crawl.redirect.javascript", Category: "Crawlability", Description: "Page uses a JavaScript-based redirect instead of an HTTP redirect."},
	{ID: "crawl.redirect.loop", Category: "Crawlability", Description: "Page is caught in a redirect loop that never resolves to a final destination."},
	{ID: "crawl.redirect.meta_refresh", Category: "Crawlability", Description: "Page uses a meta refresh redirect instead of an HTTP redirect."},
	{ID: "crawl.redirect.trailing_slash_inconsistent", Category: "Crawlability", Description: "Trailing slash variants redirect non-permanently or both return 200."},
	{ID: "crawl.redirect.www_variant_not_301", Category: "Crawlability", Description: "WWW/non-WWW variant redirect is not permanent."},
	{ID: "crawl.response.4xx", Category: "Crawlability", Description: "Page returns a 4xx client error (e.g. 404 Not Found), indicating a broken URL."},
	{ID: "crawl.response.5xx", Category: "Crawlability", Description: "Page returns a 5xx server error, meaning the server failed to respond."},
	{ID: "crawl.response.content_type_mismatch", Category: "Crawlability", Description: "URL returns HTTP 200 with a Content-Type that does not match its file extension."},
	{ID: "crawl.response.image_non_200", Category: "Crawlability", Description: "Image URL does not return HTTP 200."},
	{ID: "crawl.response.soft_404", Category: "Crawlability", Description: "Page returns HTTP 200 but appears to be a not-found or unavailable page."},
	{ID: "crawl.response.timeout", Category: "Crawlability", Description: "Page timed out during crawl; server may be slow or unresponsive."},
	{ID: "crawl.response.vary_header_invalid", Category: "Crawlability", Description: "Cacheable response has an invalid Vary header or omits Vary: Accept-Encoding when compressed."},
	{ID: "crawl.robots.directive_conflict", Category: "Crawlability", Description: "Page has conflicting robots directives (e.g. both index and noindex set)."},
	{ID: "crawl.robots.noarchive", Category: "Crawlability", Description: "Page has noarchive directive, preventing a cached copy from appearing in search results."},
	{ID: "crawl.robots.nofollow_page", Category: "Crawlability", Description: "Page-level nofollow prevents link equity flowing to any outgoing links."},
	{ID: "crawl.robots.nosnippet", Category: "Crawlability", Description: "Page has nosnippet directive, suppressing preview text in search results."},
	{ID: "crawl.robots.page_blocked_but_linked", Category: "Crawlability", Description: "Page is blocked by robots.txt but still linked internally, wasting crawl budget."},
	{ID: "crawl.robots.x_robots_tag", Category: "Crawlability", Description: "Page uses an X-Robots-Tag HTTP header to control crawling or indexing behaviour."},

	// ── E-E-A-T ─────────────────────────────────────────────────────────────
	{ID: "eeat.about_page.missing", Category: "E-E-A-T", Description: "Site has no About page, a key trust signal for Experience, Expertise, Authoritativeness, and Trust."},
	{ID: "eeat.about_page.thin", Category: "E-E-A-T", Description: "About page exists but has too little content to establish expertise or credibility."},
	{ID: "eeat.author_info.missing", Category: "E-E-A-T", Description: "Article pages lack author attribution, a key Experience and Authoritativeness signal."},
	{ID: "eeat.breadcrumbs.missing", Category: "E-E-A-T", Description: "Page has no breadcrumb navigation, reducing site structure clarity and user trust."},
	{ID: "eeat.contact_page.insufficient", Category: "E-E-A-T", Description: "Contact page lacks sufficient detail (address, phone, or email) to build user trust."},
	{ID: "eeat.contact_page.missing", Category: "E-E-A-T", Description: "Site has no Contact page, reducing user trust and E-E-A-T signals."},
	{ID: "eeat.dates.missing", Category: "E-E-A-T", Description: "Article lacks a publication or last-modified date, weakening Freshness signals."},
	{ID: "eeat.privacy_policy.thin", Category: "E-E-A-T", Description: "Privacy policy page exists but is too short to be credible or legally sufficient."},

	// ── Headings ────────────────────────────────────────────────────────────
	{ID: "headings.h1.duplicate", Category: "Headings", Description: "Multiple pages share the same H1 text, reducing keyword uniqueness across the site."},
	{ID: "headings.h1.empty", Category: "Headings", Description: "Page has an H1 tag present but with no visible text content."},
	{ID: "headings.h1.missing", Category: "Headings", Description: "Page is missing an H1 heading, a critical on-page SEO element."},
	{ID: "headings.h1.multiple", Category: "Headings", Description: "Page has more than one H1 tag, which can confuse crawlers about the primary topic."},
	{ID: "headings.h1.too_long", Category: "Headings", Description: "H1 text is excessively long; concise headings carry more weight."},
	{ID: "headings.h1.too_short", Category: "Headings", Description: "H1 text is too short to convey meaningful context to search engines."},
	{ID: "headings.h2.missing", Category: "Headings", Description: "Page has no H2 subheadings; subheadings help structure content for users and crawlers."},
	{ID: "headings.hierarchy.skipped_level", Category: "Headings", Description: "Heading levels are skipped (e.g. H1 → H3), breaking the logical document outline."},

	// ── HTML Structure ──────────────────────────────────────────────────────
	{ID: "html.doctype_missing", Category: "HTML Structure", Description: "Page is missing a <!DOCTYPE html> declaration, which can trigger browser quirks mode."},
	{ID: "html.dom_too_deep", Category: "HTML Structure", Description: "Page DOM is nested 32 or more element levels deep, which can hurt render performance and crawl efficiency."},
	{ID: "html.multiple_head", Category: "HTML Structure", Description: "Page has more than one <head> tag, which can break parsing of metadata, canonical tags, and schema."},
	{ID: "html.pagination_link_invalid", Category: "HTML Structure", Description: "Page uses rel=next/prev pagination markup with invalid placement, hrefs, duplicates, or self-references."},
	{ID: "html.robots_meta_in_body", Category: "HTML Structure", Description: "Robots meta tag appears outside <head>, where search engines may ignore it."},

	// ── HTTPS & Security ────────────────────────────────────────────────────
	{ID: "https.mixed_content", Category: "HTTPS & Security", Description: "Page is served over HTTPS but loads resources (images, scripts, CSS) over HTTP."},
	{ID: "https.page_insecure", Category: "HTTPS & Security", Description: "Page is served over HTTP instead of HTTPS, risking user data and browser warnings."},
	{ID: "security.csp.missing", Category: "HTTPS & Security", Description: "Content Security Policy header is missing, increasing exposure to XSS attacks."},
	{ID: "security.hsts.max_age_too_short", Category: "HTTPS & Security", Description: "HSTS max-age is too short to provide reliable protection against protocol downgrade."},
	{ID: "security.hsts.missing", Category: "HTTPS & Security", Description: "HTTP Strict Transport Security header is absent; browsers may allow HTTP connections."},
	{ID: "security.permissions_policy.missing", Category: "HTTPS & Security", Description: "Permissions-Policy header is missing, leaving browser features uncontrolled."},
	{ID: "security.referrer_policy.missing", Category: "HTTPS & Security", Description: "Referrer-Policy header is absent; sensitive referrer data may leak to third parties."},
	{ID: "security.x_frame.missing", Category: "HTTPS & Security", Description: "X-Frame-Options header is absent, leaving the page vulnerable to clickjacking."},
	{ID: "security.xcto.missing", Category: "HTTPS & Security", Description: "X-Content-Type-Options header is missing; browsers may MIME-sniff responses insecurely."},
	{ID: "ssl.cert_expired", Category: "HTTPS & Security", Description: "SSL certificate has expired; visitors will see browser security warnings."},
	{ID: "ssl.cert_expiring_soon", Category: "HTTPS & Security", Description: "SSL certificate is expiring within 30 days; renew promptly to avoid downtime."},
	{ID: "ssl.cert_mismatch", Category: "HTTPS & Security", Description: "SSL certificate domain does not match the URL being served."},
	{ID: "ssl.chain_incomplete", Category: "HTTPS & Security", Description: "SSL certificate chain is incomplete; some clients may reject the connection."},
	{ID: "ssl.hsts_preload_missing", Category: "HTTPS & Security", Description: "HSTS preload directive is absent; site cannot be submitted to browser preload lists."},
	{ID: "ssl.tls_version_old", Category: "HTTPS & Security", Description: "Server supports deprecated TLS versions (1.0/1.1) that are no longer secure."},

	// ── Images ──────────────────────────────────────────────────────────────
	{ID: "images.alt.empty_non_decorative", Category: "Images", Description: "Non-decorative image has an empty alt attribute, losing keyword and accessibility value."},
	{ID: "images.alt.is_filename", Category: "Images", Description: "Image alt text is the raw filename rather than a meaningful description."},
	{ID: "images.alt.missing", Category: "Images", Description: "Image is missing an alt attribute entirely, failing accessibility and SEO."},
	{ID: "images.alt.too_long", Category: "Images", Description: "Image alt text is excessively long; keep descriptions to 125 characters or fewer."},
	{ID: "images.broken", Category: "Images", Description: "Image URL returns an error, causing broken images for users and crawlers."},
	{ID: "images.dimensions.missing", Category: "Images", Description: "Image lacks explicit width/height attributes, risking layout shift during load."},
	{ID: "images.format.not_modern", Category: "Images", Description: "Image uses an older format (JPEG/PNG/GIF) instead of WebP or AVIF."},
	{ID: "images.lazy.above_fold", Category: "Images", Description: "Above-the-fold image uses lazy loading, delaying LCP and harming perceived performance."},
	{ID: "images.missing_srcset", Category: "Images", Description: "Responsive image is missing a srcset attribute for serving different resolutions."},
	{ID: "images.no_width_height_cls", Category: "Images", Description: "Image has no width/height set, directly contributing to Cumulative Layout Shift (CLS)."},
	{ID: "images.size.too_large", Category: "Images", Description: "Image file size is too large; compress or resize to improve page load speed."},

	// ── Internal Linking ────────────────────────────────────────────────────
	{ID: "links.anchor.empty", Category: "Internal Linking", Description: "Anchor tag has no visible text, making it inaccessible and meaningless to crawlers."},
	{ID: "links.anchor.generic", Category: "Internal Linking", Description: "Anchor uses generic text like \"click here\" or \"read more\" with no descriptive context."},
	{ID: "links.external.broken_4xx", Category: "Internal Linking", Description: "External link points to a page returning a 4xx client error."},
	{ID: "links.external.broken_5xx", Category: "Internal Linking", Description: "External link points to a page returning a 5xx server error."},
	{ID: "links.external.missing_noopener", Category: "Internal Linking", Description: "External link opens in a new tab without rel=\"noopener\", a security risk."},
	{ID: "links.external.redirect", Category: "Internal Linking", Description: "External link targets a URL that redirects, adding unnecessary latency."},
	{ID: "links.external.timeout", Category: "Internal Linking", Description: "External link timed out during validation; the target may be slow or unreachable."},
	{ID: "links.footer_heavy", Category: "Internal Linking", Description: "Page has a disproportionate number of links concentrated in the footer."},
	{ID: "links.internal.broken_4xx", Category: "Internal Linking", Description: "Internal link points to a page returning a 4xx error (broken internal link)."},
	{ID: "links.internal.broken_5xx", Category: "Internal Linking", Description: "Internal link points to a page returning a 5xx server error."},
	{ID: "links.internal.nofollow", Category: "Internal Linking", Description: "Internal link uses rel=\"nofollow\", unnecessarily blocking link equity flow."},
	{ID: "links.internal.to_redirect", Category: "Internal Linking", Description: "Internal link targets a redirect; update it to point directly to the final URL."},
	{ID: "links.nav_orphan", Category: "Internal Linking", Description: "Page appears in navigation but has no other internal links pointing to it."},
	{ID: "links.no_content_links", Category: "Internal Linking", Description: "Page body has no outgoing links to other internal pages."},
	{ID: "links.page.low_inlinks", Category: "Internal Linking", Description: "Page has very few internal links pointing to it, reducing crawlability and authority."},
	{ID: "links.page.orphan", Category: "Internal Linking", Description: "Page has no internal links pointing to it; effectively unreachable through navigation."},
	{ID: "links.page.too_many_outlinks", Category: "Internal Linking", Description: "Page has an unusually high number of outgoing links, diluting link equity."},

	// ── International ───────────────────────────────────────────────────────
	{ID: "i18n.hreflang.invalid_lang_code", Category: "International", Description: "hreflang attribute uses an invalid or unrecognised BCP 47 language code."},
	{ID: "i18n.hreflang.missing_return_link", Category: "International", Description: "hreflang alternate page does not link back to this page, breaking the bidirectional requirement."},
	{ID: "i18n.hreflang.missing_self_ref", Category: "International", Description: "Page is missing a self-referencing hreflang tag for its own locale."},
	{ID: "i18n.hreflang.missing_x_default", Category: "International", Description: "hreflang implementation lacks an x-default fallback for unmatched locales."},
	{ID: "i18n.hreflang.non_canonical_url", Category: "International", Description: "hreflang URL does not match the canonical URL for that locale's page."},
	{ID: "i18n.hreflang.url_not_absolute", Category: "International", Description: "hreflang URL is relative; the spec requires absolute URLs."},

	// ── Meta Descriptions ───────────────────────────────────────────────────
	{ID: "meta_desc.duplicate", Category: "Meta Descriptions", Description: "Multiple pages share the same meta description, reducing click-through differentiation."},
	{ID: "meta_desc.missing", Category: "Meta Descriptions", Description: "Page has no meta description tag; search engines will auto-generate a snippet."},
	{ID: "meta_desc.too_long", Category: "Meta Descriptions", Description: "Meta description exceeds ~160 characters and will likely be truncated in SERPs."},
	{ID: "meta_desc.too_short", Category: "Meta Descriptions", Description: "Meta description is too short to effectively summarise the page content."},

	// ── Mobile ──────────────────────────────────────────────────────────────
	{ID: "mobile.font_size.too_small", Category: "Mobile", Description: "Text font size is below 12px, making it difficult to read on mobile without zooming."},
	{ID: "mobile.user_scalable.disabled", Category: "Mobile", Description: "Viewport meta disables user zoom (user-scalable=no), failing accessibility requirements."},
	{ID: "mobile.viewport.invalid", Category: "Mobile", Description: "Viewport meta tag has an invalid or unrecognised value."},
	{ID: "mobile.viewport.missing", Category: "Mobile", Description: "Page is missing a viewport meta tag required for correct mobile rendering."},

	// ── Mobile vs Desktop ───────────────────────────────────────────────────
	{ID: "mob_desk.canonical_mismatch", Category: "Mobile vs Desktop", Description: "Canonical URL differs between the mobile and desktop versions of this page."},
	{ID: "mob_desk.content_mismatch", Category: "Mobile vs Desktop", Description: "Main body content differs significantly between mobile and desktop versions."},
	{ID: "mob_desk.h1_mismatch", Category: "Mobile vs Desktop", Description: "H1 heading text differs between mobile and desktop versions of this page."},
	{ID: "mob_desk.links_mismatch", Category: "Mobile vs Desktop", Description: "Internal link sets differ between mobile and desktop versions, risking cloaking signals."},
	{ID: "mob_desk.meta_desc_mismatch", Category: "Mobile vs Desktop", Description: "Meta description differs between mobile and desktop versions."},
	{ID: "mob_desk.og_image_mismatch", Category: "Mobile vs Desktop", Description: "OG image differs between mobile and desktop versions of this page."},
	{ID: "mob_desk.schema_mismatch", Category: "Mobile vs Desktop", Description: "Structured data markup differs between mobile and desktop versions."},
	{ID: "mob_desk.separate_mobile_site", Category: "Mobile vs Desktop", Description: "Site uses a separate m-dot domain for mobile instead of a responsive design approach."},
	{ID: "mob_desk.status_mismatch", Category: "Mobile vs Desktop", Description: "HTTP status code differs between mobile and desktop versions; may indicate cloaking."},
	{ID: "mob_desk.title_mismatch", Category: "Mobile vs Desktop", Description: "Page title differs between mobile and desktop versions of this page."},

	// ── Pagination ──────────────────────────────────────────────────────────
	{ID: "pagination.canonical_first_page", Category: "Pagination", Description: "Paginated page canonicalises to page 1, ignoring its own unique content."},
	{ID: "pagination.inconsistent_canonical_strategy", Category: "Pagination", Description: "Pagination uses mixed canonicalization strategies across the series."},
	{ID: "pagination.missing_canonical", Category: "Pagination", Description: "Paginated page lacks a canonical tag, risking duplicate content issues."},
	{ID: "pagination.noindex", Category: "Pagination", Description: "Paginated page is marked noindex; verify whether all paginated pages should be excluded."},
	{ID: "pagination.thin_content", Category: "Pagination", Description: "Paginated page has very little unique content compared to other pages in the series."},
	{ID: "pagination.title_not_unique", Category: "Pagination", Description: "Paginated pages share an identical title instead of differentiating by page number."},

	// ── Performance ─────────────────────────────────────────────────────────
	{ID: "perf.cache_control.missing", Category: "Performance", Description: "Page response lacks Cache-Control headers, preventing efficient browser caching."},
	{ID: "perf.cls_risk.images_no_dimensions", Category: "Performance", Description: "Images without explicit dimensions increase Cumulative Layout Shift (CLS) risk."},
	{ID: "perf.compression.missing", Category: "Performance", Description: "Server is not compressing responses with gzip or Brotli, inflating transfer size."},
	{ID: "perf.external_scripts.count", Category: "Performance", Description: "Too many external JavaScript files are loaded, increasing HTTP request overhead."},
	{ID: "perf.html_size.large", Category: "Performance", Description: "HTML document is excessively large; trim unused markup to improve Time to First Byte."},
	{ID: "perf.inline_css.large", Category: "Performance", Description: "Page contains excessive inline CSS; move styles to external cached stylesheets."},
	{ID: "perf.lcp_candidate.lazy_loaded", Category: "Performance", Description: "Likely LCP image is lazily loaded, delaying when the Largest Contentful Paint fires."},
	{ID: "perf.missing_preconnect", Category: "Performance", Description: "Page does not preconnect to critical third-party origins, adding DNS/handshake latency."},
	{ID: "perf.render_blocking.css_count", Category: "Performance", Description: "Too many render-blocking CSS files delay first paint; inline critical CSS or defer."},
	{ID: "perf.render_blocking.scripts", Category: "Performance", Description: "Synchronous <script> tags in <head> block rendering; use defer or async."},
	{ID: "perf.response_time.critical", Category: "Performance", Description: "Page response time exceeds 3 seconds; critically slow for both users and crawlers."},
	{ID: "perf.response_time.slow", Category: "Performance", Description: "Page response time exceeds 1 second; investigate server-side or infrastructure bottlenecks."},

	// ── Resources ───────────────────────────────────────────────────────────
	{ID: "resources.css.broken", Category: "Resources", Description: "A CSS file referenced by this page returns an error, potentially breaking styles."},
	{ID: "resources.font.broken", Category: "Resources", Description: "A font file referenced by this page returns an error, causing fallback font rendering."},
	{ID: "resources.font.no_display_swap", Category: "Resources", Description: "Font is loaded without font-display: swap, causing invisible text during load (FOIT)."},
	{ID: "resources.script.broken", Category: "Resources", Description: "A JavaScript file referenced by this page returns an error, potentially breaking functionality."},
	{ID: "resources.too_many_requests", Category: "Resources", Description: "Page makes an excessive number of sub-resource requests, increasing load time."},
	{ID: "resources.total_size_too_large", Category: "Resources", Description: "Total weight of all page resources exceeds recommended limits for fast loading."},

	// ── Sitemap ─────────────────────────────────────────────────────────────
	{ID: "sitemap.coverage_low", Category: "Sitemap", Description: "Sitemap covers a low proportion of crawlable pages; many URLs may go undiscovered."},
	{ID: "sitemap.missing", Category: "Sitemap", Description: "No XML sitemap was found at standard locations or referenced in robots.txt."},
	{ID: "sitemap.too_large", Category: "Sitemap", Description: "Sitemap exceeds the 50,000 URL or 50MB limit; split into a sitemap index."},
	{ID: "sitemap.url_4xx", Category: "Sitemap", Description: "Sitemap references a URL that returns a 4xx error; remove or fix the dead entry."},
	{ID: "sitemap.url_blocked", Category: "Sitemap", Description: "Sitemap references a URL blocked by robots.txt, sending conflicting signals."},
	{ID: "sitemap.url_noindex", Category: "Sitemap", Description: "Sitemap references a URL marked noindex; remove it from the sitemap."},
	{ID: "sitemap.url_redirected", Category: "Sitemap", Description: "Sitemap references a redirect URL; update to the final canonical destination."},

	// ── Social ──────────────────────────────────────────────────────────────
	{ID: "og.description.missing", Category: "Social", Description: "Page is missing an og:description tag; social previews will lack a description."},
	{ID: "og.image.missing", Category: "Social", Description: "Page is missing an og:image tag; social shares will display no preview image."},
	{ID: "og.title.missing", Category: "Social", Description: "Page is missing an og:title tag; social previews will fall back to the page title."},
	{ID: "og.url.mismatch_canonical", Category: "Social", Description: "og:url does not match the canonical URL, causing inconsistent link attribution."},
	{ID: "og.url.missing", Category: "Social", Description: "Page is missing an og:url tag used to specify the canonical URL for social sharing."},
	{ID: "twitter.card.missing", Category: "Social", Description: "Page is missing a twitter:card meta tag required for Twitter/X rich previews."},
	{ID: "twitter.image.missing", Category: "Social", Description: "Page is missing a twitter:image tag; Twitter/X shares will show no image."},
	{ID: "twitter.title.missing", Category: "Social", Description: "Page is missing a twitter:title tag for Twitter/X card previews."},

	// ── Structured Data ─────────────────────────────────────────────────────
	{ID: "schema.article.missing_fields", Category: "Structured Data", Description: "Article schema is missing required fields (headline, author, or datePublished)."},
	{ID: "schema.breadcrumb.invalid", Category: "Structured Data", Description: "BreadcrumbList schema has invalid or missing item properties."},
	{ID: "schema.faq.invalid", Category: "Structured Data", Description: "FAQPage schema has invalid structure or missing question/answer pairs."},
	{ID: "schema.jsonld.duplicate_type", Category: "Structured Data", Description: "Multiple JSON-LD blocks on this page define the same schema type."},
	{ID: "schema.jsonld.invalid_json", Category: "Structured Data", Description: "JSON-LD block contains malformed JSON that cannot be parsed by search engines."},
	{ID: "schema.jsonld.missing", Category: "Structured Data", Description: "Page has no structured data markup (JSON-LD, Microdata, or RDFa)."},
	{ID: "schema.jsonld.missing_context", Category: "Structured Data", Description: "JSON-LD block is missing the required @context property (should be schema.org)."},
	{ID: "schema.jsonld.missing_type", Category: "Structured Data", Description: "JSON-LD block is missing the required @type property."},
	{ID: "schema.organization.missing_homepage", Category: "Structured Data", Description: "Organization schema on the homepage is missing or has an invalid URL field."},
	{ID: "schema.product.missing_fields", Category: "Structured Data", Description: "Product schema is missing recommended fields (price, availability, or rating)."},

	// ── Titles ──────────────────────────────────────────────────────────────
	{ID: "title.duplicate", Category: "Titles", Description: "Multiple pages share the same title tag, reducing uniqueness in search results."},
	{ID: "title.missing", Category: "Titles", Description: "Page has no <title> tag, a critical ranking and click-through rate factor."},
	{ID: "title.too_long", Category: "Titles", Description: "Title tag exceeds ~60 characters and will likely be truncated in SERPs."},
	{ID: "title.too_short", Category: "Titles", Description: "Title tag is too short to effectively describe the page content."},

	// ── URL Structure ───────────────────────────────────────────────────────
	{ID: "url.contains_stop_words", Category: "URL Structure", Description: "URL contains common stop words (a, the, and…) that add length without SEO value."},
	{ID: "url.double_slash", Category: "URL Structure", Description: "URL contains double slashes (//), which may cause duplicate content or crawl issues."},
	{ID: "url.has_session_params", Category: "URL Structure", Description: "URL contains session ID parameters, creating unique duplicate pages for each session."},
	{ID: "url.has_spaces", Category: "URL Structure", Description: "URL contains spaces encoded as %20; use hyphens to separate words instead."},
	{ID: "url.has_underscores", Category: "URL Structure", Description: "URL uses underscores instead of hyphens; Google treats underscores as word joiners."},
	{ID: "url.has_uppercase", Category: "URL Structure", Description: "URL contains uppercase letters, risking duplicate content at different case variants."},
	{ID: "url.non_descriptive", Category: "URL Structure", Description: "URL slug is non-descriptive (numeric ID or random string) with no keyword value."},
	{ID: "url.path_depth_too_deep", Category: "URL Structure", Description: "URL has too many path segments, signalling low priority to crawlers."},
	{ID: "url.too_long", Category: "URL Structure", Description: "URL exceeds 2048 characters, which may cause issues with browsers and crawlers."},
	{ID: "url.too_many_params", Category: "URL Structure", Description: "URL has too many query parameters, risking crawl budget waste and duplicate content."},
}

// GetCheckDescriptors returns the full list of registered checks, sorted by category then ID.
func GetCheckDescriptors() []models.CheckDescriptor {
	result := make([]models.CheckDescriptor, len(checkDescriptors))
	copy(result, checkDescriptors)
	for i := range result {
		result[i].ChecklistIDs = ChecklistIDsFor(result[i].ID)
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].Category != result[j].Category {
			return result[i].Category < result[j].Category
		}
		return result[i].ID < result[j].ID
	})
	return result
}
