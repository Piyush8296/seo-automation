package report

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/cars24/seo-automation/internal/models"
)

// WriteMarkdown writes the audit as a Markdown report.
func WriteMarkdown(audit *models.SiteAudit, outputDir string) (string, error) {
	var sb strings.Builder

	sb.WriteString("# SEO Audit Report\n\n")
	sb.WriteString(fmt.Sprintf("**Site:** %s  \n", audit.SiteURL))
	sb.WriteString(fmt.Sprintf("**Crawled:** %s  \n", audit.CrawledAt.Format(time.RFC1123)))
	sb.WriteString(fmt.Sprintf("**Pages Crawled:** %d  \n", audit.PagesCrawled))
	sb.WriteString(fmt.Sprintf("**Health Score:** %.1f / 100 (%s)  \n\n", audit.HealthScore, audit.Grade))

	sb.WriteString("## Summary\n\n")
	sb.WriteString(fmt.Sprintf("| Severity | Count |\n|---|---|\n"))
	sb.WriteString(fmt.Sprintf("| Errors | %d |\n", audit.Stats.Errors))
	sb.WriteString(fmt.Sprintf("| Warnings | %d |\n", audit.Stats.Warnings))
	sb.WriteString(fmt.Sprintf("| Notices | %d |\n", audit.Stats.Notices))
	sb.WriteString(fmt.Sprintf("| Total Checks Run | %d |\n\n", audit.Stats.TotalChecksRun))

	sb.WriteString("## Crawl Setup\n\n")
	sb.WriteString("| Setting | Value |\n")
	sb.WriteString("|---|---|\n")
	sb.WriteString(fmt.Sprintf("| Seed URL | %s |\n", audit.CrawlConfig.SeedURL))
	sb.WriteString(fmt.Sprintf("| Scope | %s |\n", audit.CrawlConfig.Scope))
	scopePrefix := audit.CrawlConfig.ScopePrefix
	if scopePrefix == "" {
		scopePrefix = "—"
	}
	sb.WriteString(fmt.Sprintf("| Scope Prefix | %s |\n", scopePrefix))
	sb.WriteString(fmt.Sprintf("| Sitemap Mode | %s |\n", audit.CrawlConfig.SitemapMode))
	sitemapURL := audit.CrawlConfig.SitemapURL
	if sitemapURL == "" {
		sitemapURL = "auto / none"
	}
	sb.WriteString(fmt.Sprintf("| Sitemap URL | %s |\n", sitemapURL))
	sb.WriteString(fmt.Sprintf("| Max Depth | %d |\n", audit.CrawlConfig.MaxDepth))
	sb.WriteString(fmt.Sprintf("| Max Pages | %d |\n", audit.CrawlConfig.MaxPages))
	sb.WriteString(fmt.Sprintf("| Concurrency | %d |\n", audit.CrawlConfig.Concurrency))
	sb.WriteString(fmt.Sprintf("| Timeout | %s |\n", audit.CrawlConfig.Timeout))
	sb.WriteString(fmt.Sprintf("| Respect Robots | %t |\n", audit.CrawlConfig.RespectRobots))
	sb.WriteString(fmt.Sprintf("| Max Redirects | %d |\n", audit.CrawlConfig.MaxRedirects))
	sb.WriteString(fmt.Sprintf("| Max Page Size (KB) | %d |\n", audit.CrawlConfig.MaxPageSizeKB))
	sb.WriteString(fmt.Sprintf("| Max URL Length | %d |\n", audit.CrawlConfig.MaxURLLength))
	sb.WriteString(fmt.Sprintf("| Max Query Params | %d |\n", audit.CrawlConfig.MaxQueryParams))
	sb.WriteString(fmt.Sprintf("| Max Links Per Page | %d |\n", audit.CrawlConfig.MaxLinksPerPage))
	sb.WriteString(fmt.Sprintf("| Follow Nofollow Links | %t |\n", audit.CrawlConfig.FollowNofollowLinks))
	sb.WriteString(fmt.Sprintf("| Expand Noindex Pages | %t |\n", audit.CrawlConfig.ExpandNoindexPages))
	sb.WriteString(fmt.Sprintf("| Expand Canonicalized Pages | %t |\n", audit.CrawlConfig.ExpandCanonicalizedPages))
	sb.WriteString(fmt.Sprintf("| Render Mode | %s |\n\n", audit.CrawlConfig.RenderMode))

	// Collect all issues
	type issueEntry struct {
		result models.CheckResult
		url    string
	}
	var errors, warnings, notices []issueEntry
	for _, page := range audit.Pages {
		for _, r := range page.CheckResults {
			entry := issueEntry{result: r, url: page.URL}
			switch r.Severity {
			case models.SeverityError:
				errors = append(errors, entry)
			case models.SeverityWarning:
				warnings = append(warnings, entry)
			case models.SeverityNotice:
				notices = append(notices, entry)
			}
		}
	}
	for _, r := range audit.SiteChecks {
		entry := issueEntry{result: r, url: r.URL}
		switch r.Severity {
		case models.SeverityError:
			errors = append(errors, entry)
		case models.SeverityWarning:
			warnings = append(warnings, entry)
		case models.SeverityNotice:
			notices = append(notices, entry)
		}
	}

	writeIssueTable := func(title string, entries []issueEntry) {
		if len(entries) == 0 {
			return
		}
		sb.WriteString(fmt.Sprintf("## %s (%d)\n\n", title, len(entries)))
		sb.WriteString("| Check ID | URL | Message | Details |\n")
		sb.WriteString("|---|---|---|---|\n")
		for _, e := range entries {
			details := strings.ReplaceAll(e.result.Details, "|", "\\|")
			message := strings.ReplaceAll(e.result.Message, "|", "\\|")
			url := e.result.URL
			if url == "" {
				url = e.url
			}
			sb.WriteString(fmt.Sprintf("| `%s` | %s | %s | %s |\n",
				e.result.ID, url, message, details))
		}
		sb.WriteString("\n")
	}

	writeIssueTable("Errors", errors)
	writeIssueTable("Warnings", warnings)
	writeIssueTable("Notices", notices)

	sb.WriteString("## Pages Crawled\n\n")
	sb.WriteString("| URL | Status | Title | Words |\n")
	sb.WriteString("|---|---|---|---|\n")
	for _, page := range audit.Pages {
		sb.WriteString(fmt.Sprintf("| %s | %d | %s | %d |\n",
			page.URL, page.StatusCode,
			strings.ReplaceAll(page.Title, "|", "\\|"),
			page.WordCount))
	}

	path := filepath.Join(outputDir, "report.md")
	return path, os.WriteFile(path, []byte(sb.String()), 0644)
}
