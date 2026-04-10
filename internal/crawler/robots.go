package crawler

import (
	"context"
	"net/url"
	"strings"
	"sync"

	"github.com/temoto/robotstxt"
	"golang.org/x/sync/singleflight"
)

// RobotsCache fetches and caches robots.txt per host.
type RobotsCache struct {
	cache   sync.Map // key: origin → *robotstxt.RobotsData
	rawCache sync.Map // key: origin → raw string
	group   singleflight.Group
	fetcher *Fetcher
}

// NewRobotsCache creates a new RobotsCache.
func NewRobotsCache(fetcher *Fetcher) *RobotsCache {
	return &RobotsCache{fetcher: fetcher}
}

// getRobots fetches and caches the robots.txt for the given origin.
func (rc *RobotsCache) getRobots(ctx context.Context, origin string) (*robotstxt.RobotsData, string) {
	if v, ok := rc.cache.Load(origin); ok {
		raw, _ := rc.rawCache.Load(origin)
		rawStr, _ := raw.(string)
		return v.(*robotstxt.RobotsData), rawStr
	}

	val, _, _ := rc.group.Do(origin, func() (interface{}, error) {
		robotsURL := origin + "/robots.txt"
		result := rc.fetcher.Fetch(ctx, robotsURL)
		var rawStr string
		var data *robotstxt.RobotsData
		if result.Error == "" && result.StatusCode == 200 && len(result.Body) > 0 {
			rawStr = string(result.Body)
			var err error
			data, err = robotstxt.FromString(rawStr)
			if err != nil {
				data = nil
			}
		}
		rc.rawCache.Store(origin, rawStr)
		rc.cache.Store(origin, data)
		return data, nil
	})

	raw, _ := rc.rawCache.Load(origin)
	rawStr, _ := raw.(string)
	if val == nil {
		return nil, rawStr
	}
	return val.(*robotstxt.RobotsData), rawStr
}

// IsAllowed returns true if the given URL is allowed for the user-agent.
func (rc *RobotsCache) IsAllowed(ctx context.Context, pageURL, ua string) bool {
	parsed, err := url.Parse(pageURL)
	if err != nil {
		return true
	}
	origin := strings.ToLower(parsed.Scheme) + "://" + strings.ToLower(parsed.Host)
	data, _ := rc.getRobots(ctx, origin)
	if data == nil {
		return true
	}
	return data.TestAgent(parsed.RequestURI(), ua)
}

// IsMissing returns true if the robots.txt could not be fetched (404 or error).
func (rc *RobotsCache) IsMissing(ctx context.Context, siteURL string) bool {
	parsed, err := url.Parse(siteURL)
	if err != nil {
		return false
	}
	origin := strings.ToLower(parsed.Scheme) + "://" + strings.ToLower(parsed.Host)
	_, raw := rc.getRobots(ctx, origin)
	return raw == ""
}

// BlocksAll returns true if robots.txt disallows all crawling for all agents.
func (rc *RobotsCache) BlocksAll(ctx context.Context, siteURL string) bool {
	parsed, err := url.Parse(siteURL)
	if err != nil {
		return false
	}
	origin := strings.ToLower(parsed.Scheme) + "://" + strings.ToLower(parsed.Host)
	data, _ := rc.getRobots(ctx, origin)
	if data == nil {
		return false
	}
	// Check if / is disallowed for *
	return !data.TestAgent("/", "*")
}

// HasSitemapDirective returns true if robots.txt contains a Sitemap: directive.
func (rc *RobotsCache) HasSitemapDirective(ctx context.Context, siteURL string) bool {
	parsed, err := url.Parse(siteURL)
	if err != nil {
		return false
	}
	origin := strings.ToLower(parsed.Scheme) + "://" + strings.ToLower(parsed.Host)
	_, raw := rc.getRobots(ctx, origin)
	if raw == "" {
		return false
	}
	return strings.Contains(strings.ToLower(raw), "sitemap:")
}
