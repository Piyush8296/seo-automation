package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/cars24/seo-automation/internal/checks"
	"github.com/cars24/seo-automation/internal/checks/content_body"
	"github.com/cars24/seo-automation/internal/crawler"
	"github.com/cars24/seo-automation/internal/models"
	"github.com/cars24/seo-automation/internal/report"
)

var (
	flagURL               string
	flagSitemap           string
	flagScope             string
	flagScopePrefix       string
	flagMaxDepth          int
	flagMaxPages          int
	flagConcurrency       int
	flagTimeout           string
	flagNoMobile          bool
	flagFormats           string
	flagOutputDir         string
	flagExitCode          bool
	flagPlatform          string
	flagUserAgent         string
	flagMobileUserAgent   string
	flagRespectRobots     bool
	flagSitemapMode       string
	flagMaxRedirects      int
	flagMaxPageSizeKB     int64
	flagMaxURLLength      int
	flagMaxQueryParams    int
	flagMaxLinksPerPage   int
	flagFollowNofollow    bool
	flagExpandNoindex     bool
	flagExpandCanonical   bool
	flagValidateExtLinks  bool
	flagDiscoverResources bool
	flagSimHashDistance   int
)

var auditCmd = &cobra.Command{
	Use:   "audit",
	Short: "Crawl a website and run all SEO checks",
	Example: `  seo-audit audit --url https://www.cars24.com
  seo-audit audit --url https://example.com --max-depth 3 --max-pages 500 --format json,html
  seo-audit audit --url https://example.com --platform mobile
  seo-audit audit --url https://example.com --platform desktop --no-mobile-check`,
	RunE: runAudit,
}

func init() {
	auditCmd.Flags().StringVar(&flagURL, "url", "", "URL to audit (required)")
	auditCmd.Flags().StringVar(&flagSitemap, "sitemap", "", "Sitemap URL (auto-detected if not provided)")
	auditCmd.Flags().StringVar(&flagScope, "scope", string(models.CrawlScopeHost), "Crawl scope: host or subfolder")
	auditCmd.Flags().StringVar(&flagScopePrefix, "scope-prefix", "", "Optional path prefix restriction (defaults to seed path when --scope=subfolder)")
	auditCmd.Flags().IntVar(&flagMaxDepth, "max-depth", -1, "Max crawl depth (-1 = unlimited)")
	auditCmd.Flags().IntVar(&flagMaxPages, "max-pages", 0, "Max pages to crawl (0 = unlimited)")
	auditCmd.Flags().IntVar(&flagConcurrency, "concurrency", 5, "Number of concurrent crawl workers")
	auditCmd.Flags().StringVar(&flagTimeout, "timeout", "30s", "Per-request timeout (e.g. 30s, 1m)")
	auditCmd.Flags().BoolVar(&flagNoMobile, "no-mobile-check", false, "Skip mobile vs desktop comparison")
	auditCmd.Flags().StringVar(&flagFormats, "format", "json,html,markdown", "Output formats (json,html,markdown)")
	auditCmd.Flags().StringVar(&flagOutputDir, "output-dir", "./reports", "Output directory for reports")
	auditCmd.Flags().BoolVar(&flagExitCode, "exit-code", false, "Exit 1 if any errors found")
	auditCmd.Flags().StringVar(&flagPlatform, "platform", "", "Focus platform: desktop, mobile, or all (default: show both, bifurcated)")
	auditCmd.Flags().StringVar(&flagUserAgent, "user-agent", "SEOAuditBot/1.0 (+https://github.com/cars24/seo-automation)", "Desktop user-agent to send during crawl")
	auditCmd.Flags().StringVar(&flagMobileUserAgent, "mobile-user-agent", "Mozilla/5.0 (Linux; Android 10; K) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Mobile Safari/537.36", "Mobile user-agent for comparison fetches")
	auditCmd.Flags().BoolVar(&flagRespectRobots, "respect-robots", true, "Respect robots.txt allow/disallow rules")
	auditCmd.Flags().StringVar(&flagSitemapMode, "sitemap-mode", string(models.SitemapModeDiscover), "Sitemap behavior: discover, seed, or off")
	auditCmd.Flags().IntVar(&flagMaxRedirects, "max-redirects", 10, "Maximum redirects to follow per request")
	auditCmd.Flags().Int64Var(&flagMaxPageSizeKB, "max-page-size-kb", 5*1024, "Maximum response body size to read per page, in KB")
	auditCmd.Flags().IntVar(&flagMaxURLLength, "max-url-length", 0, "Skip discovered URLs longer than this (0 = unlimited)")
	auditCmd.Flags().IntVar(&flagMaxQueryParams, "max-query-params", 0, "Skip discovered URLs with more than this many query parameters (0 = unlimited)")
	auditCmd.Flags().IntVar(&flagMaxLinksPerPage, "max-links-per-page", 0, "Follow at most this many internal links from each page (0 = unlimited)")
	auditCmd.Flags().BoolVar(&flagFollowNofollow, "follow-nofollow-links", false, "Allow rel=nofollow internal links to expand the crawl")
	auditCmd.Flags().BoolVar(&flagExpandNoindex, "expand-noindex-pages", true, "Continue following links found on noindex pages")
	auditCmd.Flags().BoolVar(&flagExpandCanonical, "expand-canonicalized-pages", true, "Continue following links found on pages canonically pointing elsewhere")
	auditCmd.Flags().BoolVar(&flagValidateExtLinks, "validate-external-links", false, "Validate external links via HEAD requests (slow, disabled by default)")
	auditCmd.Flags().BoolVar(&flagDiscoverResources, "discover-resources", false, "Discover CSS/JS/font sub-resources and HEAD-validate them (slow, disabled by default)")
	auditCmd.Flags().IntVar(&flagSimHashDistance, "simhash-distance", 0, "Override Hamming-distance threshold for near-duplicate detection (0 = use default of 3). Higher = more pages flagged as near-duplicate.")
	_ = auditCmd.MarkFlagRequired("url")
}

func runAudit(cmd *cobra.Command, args []string) error {
	timeout, err := time.ParseDuration(flagTimeout)
	if err != nil {
		return fmt.Errorf("invalid timeout %q: %w", flagTimeout, err)
	}

	platform := models.Platform(strings.ToLower(strings.TrimSpace(flagPlatform)))
	if platform == "all" {
		platform = ""
	}

	scope := models.CrawlScope(strings.ToLower(strings.TrimSpace(flagScope)))
	if scope == "" {
		scope = models.CrawlScopeHost
	}

	sitemapMode := models.SitemapMode(strings.ToLower(strings.TrimSpace(flagSitemapMode)))
	if sitemapMode == "" {
		sitemapMode = models.SitemapModeDiscover
	}

	// --platform desktop implies no mobile fetch (skip the overhead)
	noMobile := flagNoMobile || platform == models.PlatformDesktop

	config := &models.CrawlConfig{
		SeedURL:                  flagURL,
		SitemapURL:               flagSitemap,
		Scope:                    scope,
		ScopePrefix:              flagScopePrefix,
		SitemapMode:              sitemapMode,
		MaxDepth:                 flagMaxDepth,
		MaxPages:                 flagMaxPages,
		Concurrency:              flagConcurrency,
		Timeout:                  timeout,
		NoMobileCheck:            noMobile,
		UserAgent:                flagUserAgent,
		MobileUA:                 flagMobileUserAgent,
		RespectRobots:            flagRespectRobots,
		MaxRedirects:             flagMaxRedirects,
		MaxPageSizeBytes:         flagMaxPageSizeKB * 1024,
		MaxURLLength:             flagMaxURLLength,
		MaxQueryParams:           flagMaxQueryParams,
		MaxLinksPerPage:          flagMaxLinksPerPage,
		FollowNofollowLinks:      flagFollowNofollow,
		ExpandNoindexPages:       flagExpandNoindex,
		ExpandCanonicalizedPages: flagExpandCanonical,
		RenderMode:               "html-only",
		Platform:                 platform,
		ValidateExternalLinks:    flagValidateExtLinks,
		DiscoverResources:        flagDiscoverResources,
		OnProgress: func(crawled int, currentURL string) {
			fmt.Fprintf(os.Stderr, "  [%d] %s\n", crawled, currentURL)
		},
	}

	platformLabel := "desktop + mobile (bifurcated)"
	switch platform {
	case models.PlatformDesktop:
		platformLabel = "desktop only"
	case models.PlatformMobile:
		platformLabel = "mobile focus"
	}

	fmt.Fprintf(os.Stderr, "Starting audit of %s\n", flagURL)
	if flagMaxDepth >= 0 {
		fmt.Fprintf(os.Stderr, "  max-depth=%d", flagMaxDepth)
	} else {
		fmt.Fprintf(os.Stderr, "  max-depth=unlimited")
	}
	if flagMaxPages > 0 {
		fmt.Fprintf(os.Stderr, " max-pages=%d", flagMaxPages)
	} else {
		fmt.Fprintf(os.Stderr, " max-pages=unlimited")
	}
	fmt.Fprintf(os.Stderr, " concurrency=%d platform=%s\n", flagConcurrency, platformLabel)
	if flagValidateExtLinks {
		fmt.Fprintf(os.Stderr, "  external link validation: enabled\n")
	}
	if flagDiscoverResources {
		fmt.Fprintf(os.Stderr, "  sub-resource discovery: enabled\n")
	}
	fmt.Fprintf(os.Stderr, "  scope=%s sitemap-mode=%s respect-robots=%t max-redirects=%d max-page-size-kb=%d\n",
		config.Scope, config.SitemapMode, config.RespectRobots, config.MaxRedirects, config.MaxPageSizeBytes/1024)
	if config.ScopePrefix != "" {
		fmt.Fprintf(os.Stderr, "  scope-prefix=%s\n", config.ScopePrefix)
	}
	if config.MaxURLLength > 0 || config.MaxQueryParams > 0 || config.MaxLinksPerPage > 0 {
		fmt.Fprintf(os.Stderr, "  url-filters: max-url-length=%d max-query-params=%d max-links-per-page=%d\n",
			config.MaxURLLength, config.MaxQueryParams, config.MaxLinksPerPage)
	}
	if config.FollowNofollowLinks || !config.ExpandNoindexPages || !config.ExpandCanonicalizedPages {
		fmt.Fprintf(os.Stderr, "  crawl expansion: follow-nofollow=%t expand-noindex=%t expand-canonicalized=%t\n",
			config.FollowNofollowLinks, config.ExpandNoindexPages, config.ExpandCanonicalizedPages)
	}

	c := crawler.NewCrawler(config)
	audit, err := c.Crawl(context.Background())
	if err != nil {
		return fmt.Errorf("crawl failed: %w", err)
	}

	fmt.Fprintf(os.Stderr, "Crawled %d pages, running checks...\n", audit.PagesCrawled)

	// Apply near-duplicate threshold override if the CLI flag was set.
	if flagSimHashDistance > 0 {
		content_body.SetSimHashMaxDistanceOverride(flagSimHashDistance)
		fmt.Fprintf(os.Stderr, "  simhash-distance override: %d (default %d)\n", flagSimHashDistance, content_body.SimHashMaxDistance)
	}

	// Run per-page checks
	for _, page := range audit.Pages {
		page.CheckResults = checks.RunPageChecks(page)
	}

	// Run site-wide checks
	audit.SiteChecks = checks.RunSiteWideChecks(audit.Pages)

	// Filter issues by platform when a specific platform is requested
	if platform == models.PlatformDesktop || platform == models.PlatformMobile {
		filterByPlatform(audit, platform)
	}
	checks.AttachChecklistMappings(audit)

	// Compute health score and grade (overall + desktop + mobile split)
	report.ComputeHealthScore(audit)

	fmt.Fprintf(os.Stderr, "Health Score: %.1f (%s) | Desktop: %.1f (%s) | Mobile: %.1f (%s)\n",
		audit.HealthScore, audit.Grade,
		audit.DesktopHealthScore, audit.DesktopGrade,
		audit.MobileHealthScore, audit.MobileGrade)
	fmt.Fprintf(os.Stderr, "Errors: %d | Warnings: %d | Notices: %d\n",
		audit.Stats.Errors, audit.Stats.Warnings, audit.Stats.Notices)

	// Generate reports
	formats := strings.Split(flagFormats, ",")
	files, err := report.Generate(audit, formats, flagOutputDir)
	if err != nil {
		return fmt.Errorf("report generation failed: %w", err)
	}

	for format, path := range files {
		fmt.Fprintf(os.Stderr, "  [%s] → %s\n", format, path)
	}

	if flagExitCode && audit.Stats.Errors > 0 {
		os.Exit(1)
	}

	return nil
}

// filterByPlatform removes issues that don't apply to the requested platform.
func filterByPlatform(audit *models.SiteAudit, platform models.Platform) {
	keep := func(p models.Platform) bool {
		if p == models.PlatformBoth || p == "" {
			return true // "both" is always kept
		}
		switch platform {
		case models.PlatformDesktop:
			return p == models.PlatformDesktop
		case models.PlatformMobile:
			return p == models.PlatformMobile || p == models.PlatformDiff
		}
		return true
	}

	for _, page := range audit.Pages {
		filtered := page.CheckResults[:0]
		for _, r := range page.CheckResults {
			if keep(r.Platform) {
				filtered = append(filtered, r)
			}
		}
		page.CheckResults = filtered
	}

	filtered := audit.SiteChecks[:0]
	for _, r := range audit.SiteChecks {
		if keep(r.Platform) {
			filtered = append(filtered, r)
		}
	}
	audit.SiteChecks = filtered
}
