package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/cars24/seo-automation/internal/checks"
	"github.com/cars24/seo-automation/internal/crawler"
	"github.com/cars24/seo-automation/internal/models"
	"github.com/cars24/seo-automation/internal/report"
)

var (
	flagURL         string
	flagSitemap     string
	flagMaxDepth    int
	flagMaxPages    int
	flagConcurrency int
	flagTimeout     string
	flagNoMobile    bool
	flagFormats     string
	flagOutputDir   string
	flagExitCode    bool
	flagPlatform         string
	flagValidateExtLinks bool
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
	auditCmd.Flags().IntVar(&flagMaxDepth, "max-depth", -1, "Max crawl depth (-1 = unlimited)")
	auditCmd.Flags().IntVar(&flagMaxPages, "max-pages", 0, "Max pages to crawl (0 = unlimited)")
	auditCmd.Flags().IntVar(&flagConcurrency, "concurrency", 5, "Number of concurrent crawl workers")
	auditCmd.Flags().StringVar(&flagTimeout, "timeout", "30s", "Per-request timeout (e.g. 30s, 1m)")
	auditCmd.Flags().BoolVar(&flagNoMobile, "no-mobile-check", false, "Skip mobile vs desktop comparison")
	auditCmd.Flags().StringVar(&flagFormats, "format", "json,html,markdown", "Output formats (json,html,markdown)")
	auditCmd.Flags().StringVar(&flagOutputDir, "output-dir", "./reports", "Output directory for reports")
	auditCmd.Flags().BoolVar(&flagExitCode, "exit-code", false, "Exit 1 if any errors found")
	auditCmd.Flags().StringVar(&flagPlatform, "platform", "", "Focus platform: desktop, mobile, or all (default: show both, bifurcated)")
	auditCmd.Flags().BoolVar(&flagValidateExtLinks, "validate-external-links", false, "Validate external links via HEAD requests (slow, disabled by default)")
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

	// --platform desktop implies no mobile fetch (skip the overhead)
	noMobile := flagNoMobile || platform == models.PlatformDesktop

	config := &models.CrawlConfig{
		SeedURL:       flagURL,
		SitemapURL:    flagSitemap,
		MaxDepth:      flagMaxDepth,
		MaxPages:      flagMaxPages,
		Concurrency:   flagConcurrency,
		Timeout:       timeout,
		NoMobileCheck: noMobile,
		UserAgent:     "SEOAuditBot/1.0 (+https://github.com/cars24/seo-automation)",
		MobileUA:      "Mozilla/5.0 (Linux; Android 10; K) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Mobile Safari/537.36",
		Platform:              platform,
		ValidateExternalLinks: flagValidateExtLinks,
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

	c := crawler.NewCrawler(config)
	audit, err := c.Crawl(context.Background())
	if err != nil {
		return fmt.Errorf("crawl failed: %w", err)
	}

	fmt.Fprintf(os.Stderr, "Crawled %d pages, running checks...\n", audit.PagesCrawled)

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
