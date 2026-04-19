package models

import "time"

// Severity levels for check results
type Severity string

const (
	SeverityError   Severity = "error"
	SeverityWarning Severity = "warning"
	SeverityNotice  Severity = "notice"
)

// Platform indicates which rendering context a check result applies to.
type Platform string

const (
	// PlatformBoth means the issue affects both desktop and mobile (default).
	PlatformBoth Platform = "both"
	// PlatformDesktop means the issue is specific to the desktop version.
	PlatformDesktop Platform = "desktop"
	// PlatformMobile means the issue is specific to the mobile version.
	PlatformMobile Platform = "mobile"
	// PlatformDiff means the issue is a discrepancy between mobile and desktop.
	PlatformDiff Platform = "diff"
)

// CrawlScope controls how far from the seed URL the crawler can roam.
type CrawlScope string

const (
	// CrawlScopeHost keeps crawling on the same hostname (default).
	CrawlScopeHost CrawlScope = "host"
	// CrawlScopeSubfolder restricts crawling to the seed path subtree.
	CrawlScopeSubfolder CrawlScope = "subfolder"
)

// SitemapMode controls how sitemap URLs are used during a crawl.
type SitemapMode string

const (
	// SitemapModeDiscover records sitemap URLs for coverage checks, but does not seed them.
	SitemapModeDiscover SitemapMode = "discover"
	// SitemapModeSeed adds sitemap URLs to the crawl frontier.
	SitemapModeSeed SitemapMode = "seed"
	// SitemapModeOff disables sitemap discovery entirely.
	SitemapModeOff SitemapMode = "off"
)

// CheckResult is a single SEO finding on a page or site-wide
type CheckResult struct {
	ID       string   `json:"id"`
	Category string   `json:"category"`
	Severity Severity `json:"severity"`
	Message  string   `json:"message"`
	URL      string   `json:"url"`
	Details  string   `json:"details,omitempty"`
	Platform Platform `json:"platform,omitempty"`
}

// PageCheck runs on a single page
type PageCheck interface {
	Run(page *PageData) []CheckResult
}

// SiteCheck runs across all crawled pages
type SiteCheck interface {
	Run(pages []*PageData) []CheckResult
}

// LinkPosition classifies where on a page a link appears.
type LinkPosition string

const (
	PositionHeader  LinkPosition = "header"
	PositionNav     LinkPosition = "nav"
	PositionContent LinkPosition = "content"
	PositionFooter  LinkPosition = "footer"
	PositionSidebar LinkPosition = "sidebar"
)

// Link represents a hyperlink found on a page
type Link struct {
	URL        string       `json:"url"`
	Text       string       `json:"text"`
	Rel        string       `json:"rel"`
	IsInternal bool         `json:"is_internal"`
	IsFollow   bool         `json:"is_follow"`
	Position   LinkPosition `json:"position,omitempty"`
	StatusCode int          `json:"status_code,omitempty"`
	Timeout    bool         `json:"timeout,omitempty"`
}

// ResourceType classifies a sub-resource discovered on a page.
type ResourceType string

const (
	ResourceScript ResourceType = "script"
	ResourceCSS    ResourceType = "css"
	ResourceFont   ResourceType = "font"
)

// Resource represents a sub-resource (CSS, JS, or font) referenced by a page.
type Resource struct {
	URL         string       `json:"url"`
	Type        ResourceType `json:"type"`
	StatusCode  int          `json:"status_code,omitempty"`
	FileSize    int64        `json:"file_size,omitempty"`
	ContentType string       `json:"content_type,omitempty"`
	IsInternal  bool         `json:"is_internal"`
}

// Image represents an <img> element
type Image struct {
	Src         string `json:"src"`
	Alt         string `json:"alt"`
	AltPresent  bool   `json:"alt_present"` // true if alt attr exists (even if empty)
	Width       int    `json:"width,omitempty"`
	Height      int    `json:"height,omitempty"`
	Loading     string `json:"loading,omitempty"`
	HasSrcset   bool   `json:"has_srcset"`
	IsAboveFold bool   `json:"is_above_fold"`
	StatusCode  int    `json:"status_code,omitempty"`
	FileSize    int64  `json:"file_size,omitempty"`    // bytes, from HEAD/GET response
	Format      string `json:"format,omitempty"`       // e.g. "jpg", "webp", "avif", "png"
	ContentType string `json:"content_type,omitempty"` // from Content-Type header
}

// Hreflang represents an alternate language link
type Hreflang struct {
	Lang string `json:"lang"`
	URL  string `json:"url"`
}

// RedirectHop is one step in a redirect chain
type RedirectHop struct {
	URL        string `json:"url"`
	StatusCode int    `json:"status_code"`
}

// PageData holds all extracted data and check results for a single URL
type PageData struct {
	URL               string            `json:"url"`
	FinalURL          string            `json:"final_url"`
	StatusCode        int               `json:"status_code"`
	ResponseTimeMs    int64             `json:"response_time_ms"`
	ContentType       string            `json:"content_type"`
	Headers           map[string]string `json:"headers"`
	Title             string            `json:"title"`
	MetaDesc          string            `json:"meta_description"`
	H1s               []string          `json:"h1s"`
	H2s               []string          `json:"h2s"`
	H3s               []string          `json:"h3s"`
	Canonical         string            `json:"canonical"`
	RobotsTag         string            `json:"robots_tag"`
	RobotsDirectives  []string          `json:"robots_directives,omitempty"`
	XRobotsTag        string            `json:"x_robots_tag,omitempty"`
	RedirectChain     []RedirectHop     `json:"redirect_chain,omitempty"`
	Links             []Link            `json:"links"`
	Images            []Image           `json:"images"`
	Resources         []Resource        `json:"resources,omitempty"`
	FontFaceNoDisplay int               `json:"font_face_no_display,omitempty"`
	SchemaJSONRaw     []string          `json:"schema_json_raw"`
	OGTags            map[string]string `json:"og_tags"`
	TwitterTags       map[string]string `json:"twitter_tags"`
	HreflangTags      []Hreflang        `json:"hreflang_tags"`
	Depth             int               `json:"depth"`
	InlinkCount       int               `json:"inlink_count"`
	IsInternal        bool              `json:"is_internal"`
	InSitemap         bool              `json:"in_sitemap"`
	CheckResults      []CheckResult     `json:"check_results"`
	Error             string            `json:"error,omitempty"`
	// Body analysis (not serialized to reduce JSON size)
	BodyText              string `json:"-"`
	RawHTML               string `json:"-"`
	WordCount             int    `json:"word_count"`
	HTMLSizeBytes         int    `json:"html_size_bytes"`
	RenderBlockingScripts int    `json:"render_blocking_scripts"`
	RenderBlockingCSS     int    `json:"render_blocking_css"`
	ExternalScriptCount   int    `json:"external_script_count"`
	InlineCSSBytes        int    `json:"inline_css_bytes"`
	HasViewport           bool   `json:"has_viewport"`
	ViewportContent       string `json:"viewport_content"`
	// Mobile comparison data
	MobileData *MobilePageData `json:"mobile_data,omitempty"`
	// TLS / SSL certificate info
	TLSInfo *TLSInfo `json:"tls_info,omitempty"`
}

// TLSInfo holds TLS connection and certificate details captured during fetch.
type TLSInfo struct {
	Version       string    `json:"version"` // e.g. "TLS 1.3"
	CipherSuite   string    `json:"cipher_suite"`
	CertSubject   string    `json:"cert_subject,omitempty"`
	CertIssuer    string    `json:"cert_issuer,omitempty"`
	CertNotBefore time.Time `json:"cert_not_before,omitempty"`
	CertNotAfter  time.Time `json:"cert_not_after,omitempty"`
	CertDNSNames  []string  `json:"cert_dns_names,omitempty"`
	ChainLength   int       `json:"chain_length"`
}

// MobilePageData holds data fetched with a mobile user-agent for comparison
type MobilePageData struct {
	StatusCode    int               `json:"status_code"`
	FinalURL      string            `json:"final_url"`
	Title         string            `json:"title"`
	MetaDesc      string            `json:"meta_description"`
	H1s           []string          `json:"h1s"`
	Canonical     string            `json:"canonical"`
	SchemaJSONRaw []string          `json:"schema_json_raw"`
	OGTags        map[string]string `json:"og_tags"`
	Links         []Link            `json:"links"`
	WordCount     int               `json:"word_count"`
}

// CrawlConfigSnapshot stores the effective crawl setup used for an audit.
type CrawlConfigSnapshot struct {
	SeedURL                  string      `json:"seed_url"`
	Scope                    CrawlScope  `json:"scope"`
	ScopePrefix              string      `json:"scope_prefix,omitempty"`
	SitemapMode              SitemapMode `json:"sitemap_mode"`
	SitemapURL               string      `json:"sitemap_url,omitempty"`
	MaxDepth                 int         `json:"max_depth"`
	MaxPages                 int         `json:"max_pages"`
	Concurrency              int         `json:"concurrency"`
	Timeout                  string      `json:"timeout"`
	Platform                 Platform    `json:"platform,omitempty"`
	UserAgent                string      `json:"user_agent,omitempty"`
	MobileUserAgent          string      `json:"mobile_user_agent,omitempty"`
	RespectRobots            bool        `json:"respect_robots"`
	MaxRedirects             int         `json:"max_redirects"`
	MaxPageSizeKB            int64       `json:"max_page_size_kb"`
	MaxURLLength             int         `json:"max_url_length"`
	MaxQueryParams           int         `json:"max_query_params"`
	MaxLinksPerPage          int         `json:"max_links_per_page"`
	FollowNofollowLinks      bool        `json:"follow_nofollow_links"`
	ExpandNoindexPages       bool        `json:"expand_noindex_pages"`
	ExpandCanonicalizedPages bool        `json:"expand_canonicalized_pages"`
	RenderMode               string      `json:"render_mode"`
}

// SiteAudit is the top-level result of a complete site crawl + analysis
type SiteAudit struct {
	SiteURL            string              `json:"site_url"`
	CrawledAt          time.Time           `json:"crawled_at"`
	CrawlConfig        CrawlConfigSnapshot `json:"crawl_config"`
	PagesTotal         int                 `json:"pages_total"`
	PagesCrawled       int                 `json:"pages_crawled"`
	Pages              []*PageData         `json:"pages"`
	SiteChecks         []CheckResult       `json:"site_checks"`
	HealthScore        float64             `json:"health_score"`
	Grade              string              `json:"grade"`
	DesktopHealthScore float64             `json:"desktop_health_score"`
	DesktopGrade       string              `json:"desktop_grade"`
	MobileHealthScore  float64             `json:"mobile_health_score"`
	MobileGrade        string              `json:"mobile_grade"`
	Stats              AuditStats          `json:"stats"`
	DesktopStats       AuditStats          `json:"desktop_stats"`
	MobileStats        AuditStats          `json:"mobile_stats"`
	RobotsTxtMissing   bool                `json:"robots_txt_missing"`
	RobotsBlocksAll    bool                `json:"robots_blocks_all"`
	RobotsSitemapDir   bool                `json:"robots_has_sitemap_directive"`
	SitemapURLs        []string            `json:"sitemap_urls"`
	SitemapPageCount   int                 `json:"sitemap_page_count"`
}

// AuditStats aggregates counts of issues across severity levels
type AuditStats struct {
	Errors         int `json:"errors"`
	Warnings       int `json:"warnings"`
	Notices        int `json:"notices"`
	TotalChecksRun int `json:"total_checks_run"`
}

// CrawlConfig holds all configuration for a crawl run
type CrawlConfig struct {
	SeedURL                  string
	SitemapURL               string
	Scope                    CrawlScope
	ScopePrefix              string
	SitemapMode              SitemapMode
	MaxDepth                 int
	MaxPages                 int
	Concurrency              int
	Timeout                  time.Duration
	NoMobileCheck            bool
	UserAgent                string
	MobileUA                 string
	RespectRobots            bool
	MaxRedirects             int
	MaxPageSizeBytes         int64
	MaxURLLength             int
	MaxQueryParams           int
	MaxLinksPerPage          int
	FollowNofollowLinks      bool
	ExpandNoindexPages       bool
	ExpandCanonicalizedPages bool
	RenderMode               string
	// Platform filters the audit to a specific rendering context.
	// "" or "all" = run both and show bifurcated report (default).
	// "desktop" = skip mobile fetch, only surface desktop issues.
	// "mobile"  = only surface mobile + diff issues.
	Platform              Platform
	ValidateExternalLinks bool
	DiscoverResources     bool
	// OnProgress is called after each page is successfully crawled.
	// crawled = total pages done so far; currentURL = the URL just processed.
	// Safe to leave nil.
	OnProgress func(crawled int, currentURL string)
}

// Snapshot returns the effective crawl settings in a report-friendly form.
func (c *CrawlConfig) Snapshot() CrawlConfigSnapshot {
	timeout := ""
	if c.Timeout > 0 {
		timeout = c.Timeout.String()
	}
	maxPageSizeKB := int64(0)
	if c.MaxPageSizeBytes > 0 {
		maxPageSizeKB = c.MaxPageSizeBytes / 1024
	}
	return CrawlConfigSnapshot{
		SeedURL:                  c.SeedURL,
		Scope:                    c.Scope,
		ScopePrefix:              c.ScopePrefix,
		SitemapMode:              c.SitemapMode,
		SitemapURL:               c.SitemapURL,
		MaxDepth:                 c.MaxDepth,
		MaxPages:                 c.MaxPages,
		Concurrency:              c.Concurrency,
		Timeout:                  timeout,
		Platform:                 c.Platform,
		UserAgent:                c.UserAgent,
		MobileUserAgent:          c.MobileUA,
		RespectRobots:            c.RespectRobots,
		MaxRedirects:             c.MaxRedirects,
		MaxPageSizeKB:            maxPageSizeKB,
		MaxURLLength:             c.MaxURLLength,
		MaxQueryParams:           c.MaxQueryParams,
		MaxLinksPerPage:          c.MaxLinksPerPage,
		FollowNofollowLinks:      c.FollowNofollowLinks,
		ExpandNoindexPages:       c.ExpandNoindexPages,
		ExpandCanonicalizedPages: c.ExpandCanonicalizedPages,
		RenderMode:               c.RenderMode,
	}
}
