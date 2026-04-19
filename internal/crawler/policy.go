package crawler

import (
	"net/url"
	"strings"

	"github.com/cars24/seo-automation/internal/models"
)

func effectiveScopePrefix(config *models.CrawlConfig) string {
	raw := strings.TrimSpace(config.ScopePrefix)
	if raw == "" {
		if parsed, err := url.Parse(config.SeedURL); err == nil {
			raw = parsed.Path
		}
	}
	if raw == "" || raw == "/" {
		return "/"
	}
	if !strings.HasPrefix(raw, "/") {
		raw = "/" + raw
	}
	raw = strings.TrimRight(raw, "/")
	if raw == "" {
		return "/"
	}
	return raw
}

func pathWithinScope(rawURL, prefix string) bool {
	if prefix == "" || prefix == "/" {
		return true
	}
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return false
	}
	path := parsed.Path
	if path == "" {
		path = "/"
	}
	if path != "/" {
		path = strings.TrimRight(path, "/")
	}
	return path == prefix || strings.HasPrefix(path, prefix+"/")
}

func queryParamCount(rawURL string) int {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return 0
	}
	return len(parsed.Query())
}

func shouldCrawlURL(rawURL string, config *models.CrawlConfig) bool {
	if !IsHTTPScheme(rawURL) || !SameHost(rawURL, config.SeedURL) {
		return false
	}
	if config.Scope == models.CrawlScopeSubfolder && !pathWithinScope(rawURL, effectiveScopePrefix(config)) {
		return false
	}
	if config.MaxURLLength > 0 && len(rawURL) > config.MaxURLLength {
		return false
	}
	if config.MaxQueryParams > 0 && queryParamCount(rawURL) > config.MaxQueryParams {
		return false
	}
	return true
}

func hasNoindexDirective(page *models.PageData) bool {
	for _, directive := range page.RobotsDirectives {
		if directive == "noindex" {
			return true
		}
	}
	return false
}

func canonicalPointsElsewhere(page *models.PageData) bool {
	canonical := strings.TrimSpace(page.Canonical)
	if canonical == "" {
		return false
	}
	resolved, err := NormalizeURL(canonical, page.FinalURL)
	if err != nil {
		return false
	}
	finalURL := page.FinalURL
	if finalURL == "" {
		finalURL = page.URL
	}
	return DedupeKey(resolved) != DedupeKey(finalURL)
}

func canExpandFromPage(page *models.PageData, config *models.CrawlConfig) bool {
	if !config.ExpandNoindexPages && hasNoindexDirective(page) {
		return false
	}
	if !config.ExpandCanonicalizedPages && canonicalPointsElsewhere(page) {
		return false
	}
	return true
}
