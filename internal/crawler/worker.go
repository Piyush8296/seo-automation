package crawler

import (
	"context"
	"strings"

	"github.com/cars24/seo-automation/internal/models"
	"github.com/cars24/seo-automation/internal/parser"
)

// WorkerResult is the output of processing a single URL.
type WorkerResult struct {
	Page          *models.PageData
	DiscoveredURLs []string
}

// ProcessURL fetches and parses a single URL, optionally with a mobile UA.
func ProcessURL(
	ctx context.Context,
	fetcher *Fetcher,
	mobileFetcher *Fetcher,
	robotsCache *RobotsCache,
	pageURL string,
	depth int,
	config *models.CrawlConfig,
) *WorkerResult {
	page := &models.PageData{
		URL:        pageURL,
		FinalURL:   pageURL,
		Depth:      depth,
		IsInternal: true,
	}

	// Check robots.txt
	if !robotsCache.IsAllowed(ctx, pageURL, config.UserAgent) {
		page.Error = "blocked by robots.txt"
		return &WorkerResult{Page: page}
	}

	// Fetch with desktop UA
	result := fetcher.Fetch(ctx, pageURL)
	page.FinalURL = result.FinalURL
	page.StatusCode = result.StatusCode
	page.ResponseTimeMs = result.ResponseTimeMs
	page.RedirectChain = result.RedirectChain
	page.HTMLSizeBytes = len(result.Body)
	page.TLSInfo = result.TLSInfo

	if result.Error != "" {
		page.Error = result.Error
		return &WorkerResult{Page: page}
	}

	// Extract headers
	page.Headers = parser.ExtractHeaders(result.Headers)
	page.ContentType = parser.GetHeader(page.Headers, "content-type")

	// X-Robots-Tag comes from HTTP headers, so extract it for all responses
	xrt := strings.TrimSpace(page.Headers["x-robots-tag"])
	if xrt != "" {
		page.XRobotsTag = xrt
		page.RobotsDirectives, _ = parser.ParseRobotsDirectives(page.RobotsTag, page.Headers)
	}

	// Only parse HTML pages
	if !isHTMLContent(page.ContentType) {
		return &WorkerResult{Page: page}
	}

	// Parse HTML
	extracted, err := parser.ExtractPage(result.Body, result.FinalURL, page.Headers)
	if err != nil {
		page.Error = "parse error: " + err.Error()
		return &WorkerResult{Page: page}
	}

	// Merge extracted data into page
	page.Title = extracted.Title
	page.MetaDesc = extracted.MetaDesc
	page.Canonical = extracted.Canonical
	page.RobotsTag = extracted.RobotsTag
	page.RobotsDirectives = extracted.RobotsDirectives
	page.XRobotsTag = extracted.XRobotsTag
	page.H1s = extracted.H1s
	page.H2s = extracted.H2s
	page.H3s = extracted.H3s
	page.Links = extracted.Links
	page.Images = extracted.Images
	page.SchemaJSONRaw = extracted.SchemaJSONRaw
	page.OGTags = extracted.OGTags
	page.TwitterTags = extracted.TwitterTags
	page.HreflangTags = extracted.HreflangTags
	page.BodyText = extracted.BodyText
	page.RawHTML = extracted.RawHTML
	page.WordCount = extracted.WordCount
	page.RenderBlockingScripts = extracted.RenderBlockingScripts
	page.RenderBlockingCSS = extracted.RenderBlockingCSS
	page.ExternalScriptCount = extracted.ExternalScriptCount
	page.InlineCSSBytes = extracted.InlineCSSBytes
	page.HasViewport = extracted.HasViewport
	page.ViewportContent = extracted.ViewportContent

	// Collect discovered internal URLs
	var discovered []string
	for _, link := range page.Links {
		if link.IsInternal && link.IsFollow {
			discovered = append(discovered, link.URL)
		}
	}

	// Mobile fetch
	if !config.NoMobileCheck && mobileFetcher != nil {
		mobileResult := mobileFetcher.Fetch(ctx, pageURL)
		if mobileResult.Error == "" && len(mobileResult.Body) > 0 {
			mobileData := parser.ExtractMobileData(mobileResult.Body, mobileResult.FinalURL)
			if mobileData != nil {
				mobileData.StatusCode = mobileResult.StatusCode
				mobileData.FinalURL = mobileResult.FinalURL
				page.MobileData = mobileData
			}
		}
	}

	return &WorkerResult{Page: page, DiscoveredURLs: discovered}
}

func isHTMLContent(ct string) bool {
	ct = strings.ToLower(ct)
	return strings.Contains(ct, "text/html") || ct == ""
}
