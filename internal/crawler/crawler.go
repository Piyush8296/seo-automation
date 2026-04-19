package crawler

import (
	"context"
	"sync"
	"time"

	"github.com/cars24/seo-automation/internal/models"
)

// Crawler orchestrates a BFS crawl with a worker pool.
type Crawler struct {
	config        *models.CrawlConfig
	fetcher       *Fetcher
	mobileFetcher *Fetcher
	robotsCache   *RobotsCache
}

// NewCrawler creates a Crawler with the given config.
func NewCrawler(config *models.CrawlConfig) *Crawler {
	fetcher := NewFetcher(config.Timeout, config.UserAgent, config.MaxRedirects, config.MaxPageSizeBytes)
	var mobileFetcher *Fetcher
	if !config.NoMobileCheck && config.MobileUA != "" {
		mobileFetcher = NewFetcher(config.Timeout, config.MobileUA, config.MaxRedirects, config.MaxPageSizeBytes)
	}
	return &Crawler{
		config:        config,
		fetcher:       fetcher,
		mobileFetcher: mobileFetcher,
		robotsCache:   NewRobotsCache(fetcher),
	}
}

type crawlItem struct {
	url   string
	depth int
}

// Crawl performs the full BFS crawl and returns a SiteAudit.
func (c *Crawler) Crawl(ctx context.Context) (*models.SiteAudit, error) {
	audit := &models.SiteAudit{
		SiteURL:     c.config.SeedURL,
		CrawledAt:   time.Now(),
		CrawlConfig: c.config.Snapshot(),
	}
	if audit.CrawlConfig.Scope == models.CrawlScopeSubfolder && audit.CrawlConfig.ScopePrefix == "" {
		audit.CrawlConfig.ScopePrefix = effectiveScopePrefix(c.config)
	}

	// Check robots.txt metadata
	audit.RobotsTxtMissing = c.robotsCache.IsMissing(ctx, c.config.SeedURL)
	audit.RobotsBlocksAll = c.robotsCache.BlocksAll(ctx, c.config.SeedURL)
	audit.RobotsSitemapDir = c.robotsCache.HasSitemapDirective(ctx, c.config.SeedURL)

	// Fetch sitemap URLs
	sitemapURL := c.config.SitemapURL
	if c.config.SitemapMode != models.SitemapModeOff && sitemapURL == "" {
		sitemapURL = DiscoverSitemapURL(ctx, c.fetcher, c.robotsCache, c.config.SeedURL)
	}
	sitemapSet := make(map[string]bool)
	if c.config.SitemapMode != models.SitemapModeOff && sitemapURL != "" {
		entries, _ := FetchSitemapURLs(ctx, c.fetcher, sitemapURL)
		for _, e := range entries {
			key := DedupeKey(e.URL)
			sitemapSet[key] = true
			audit.SitemapURLs = append(audit.SitemapURLs, e.URL)
		}
		audit.SitemapPageCount = len(entries)
	}

	// BFS queue
	queue := make(chan crawlItem, 10000)

	var (
		visited sync.Map
		pagesMu sync.Mutex
		pages   []*models.PageData
		wg      sync.WaitGroup
		pageCnt int
		sem     = make(chan struct{}, c.config.Concurrency)
	)

	enqueue := func(rawURL string, depth int) {
		key := DedupeKey(rawURL)
		if !shouldCrawlURL(key, c.config) {
			return
		}
		if _, loaded := visited.LoadOrStore(key, true); loaded {
			return
		}
		select {
		case queue <- crawlItem{url: key, depth: depth}:
		default:
		}
	}

	enqueue(c.config.SeedURL, 0)
	if c.config.SitemapMode == models.SitemapModeSeed {
		for _, sitemapURL := range audit.SitemapURLs {
			enqueue(sitemapURL, 0)
		}
	}

	// Drain queue until empty and all workers done
	for {
		select {
		case item, ok := <-queue:
			if !ok {
				goto done
			}
			// Check max pages
			pagesMu.Lock()
			if c.config.MaxPages > 0 && pageCnt >= c.config.MaxPages {
				pagesMu.Unlock()
				continue
			}
			pageCnt++
			pagesMu.Unlock()

			wg.Add(1)
			sem <- struct{}{}
			go func(ci crawlItem) {
				defer wg.Done()
				defer func() { <-sem }()

				result := ProcessURL(ctx, c.fetcher, c.mobileFetcher, c.robotsCache, ci.url, ci.depth, c.config)
				if result == nil || result.Page == nil {
					return
				}

				// Mark if in sitemap
				result.Page.InSitemap = sitemapSet[DedupeKey(ci.url)]

				pagesMu.Lock()
				pages = append(pages, result.Page)
				crawledCount := len(pages)
				pagesMu.Unlock()

				if c.config.OnProgress != nil {
					c.config.OnProgress(crawledCount, ci.url)
				}

				// Enqueue discovered URLs
				for _, discovered := range result.DiscoveredURLs {
					nextDepth := ci.depth + 1
					if c.config.MaxDepth >= 0 && nextDepth > c.config.MaxDepth {
						continue
					}
					pagesMu.Lock()
					exceeded := c.config.MaxPages > 0 && pageCnt >= c.config.MaxPages
					pagesMu.Unlock()
					if exceeded {
						continue
					}
					enqueue(discovered, nextDepth)
				}
			}(item)

		default:
			// No item immediately available; wait for workers then check again
			wg.Wait()
			// If queue is still empty after waiting, we're done
			if len(queue) == 0 {
				goto done
			}
		}
	}

done:
	wg.Wait()

	audit.Pages = pages
	audit.PagesCrawled = len(pages)
	audit.PagesTotal = len(pages)

	if c.config.ValidateExternalLinks {
		ValidateExternalLinks(ctx, pages, c.config.UserAgent)
	}

	if c.config.DiscoverResources {
		ValidateResources(ctx, pages, c.config.UserAgent)
	}

	return audit, nil
}
