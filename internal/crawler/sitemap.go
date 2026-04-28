package crawler

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/xml"
	"io"
	"strings"
	"time"
)

// SitemapEntry is a single URL entry from a sitemap.
type SitemapEntry struct {
	URL     string
	LastMod time.Time
}

// xmlURL is used for sitemap urlset parsing.
type xmlURL struct {
	Loc     string `xml:"loc"`
	LastMod string `xml:"lastmod"`
}

// xmlURLSet is a standard sitemap.
type xmlURLSet struct {
	XMLName xml.Name `xml:"urlset"`
	URLs    []xmlURL `xml:"url"`
}

// xmlSitemapIndex is a sitemap index.
type xmlSitemapIndex struct {
	XMLName  xml.Name        `xml:"sitemapindex"`
	Sitemaps []xmlSitemapLoc `xml:"sitemap"`
}

type xmlSitemapLoc struct {
	Loc string `xml:"loc"`
}

// FetchSitemapURLs fetches and parses all URLs from a sitemap (index or urlset).
func FetchSitemapURLs(ctx context.Context, fetcher *Fetcher, sitemapURL string) ([]SitemapEntry, error) {
	return fetchSitemap(ctx, fetcher, sitemapURL, 0, 0)
}

// FetchSitemapURLsLimit fetches sitemap URLs up to limit. A limit <= 0 means no limit.
func FetchSitemapURLsLimit(ctx context.Context, fetcher *Fetcher, sitemapURL string, limit int) ([]SitemapEntry, error) {
	return fetchSitemap(ctx, fetcher, sitemapURL, 0, limit)
}

func fetchSitemap(ctx context.Context, fetcher *Fetcher, sitemapURL string, depth int, limit int) ([]SitemapEntry, error) {
	if depth > 3 {
		return nil, nil
	}
	result := fetcher.Fetch(ctx, sitemapURL)
	if result.Error != "" {
		return nil, nil
	}
	if result.StatusCode != 200 {
		return nil, nil
	}
	data := result.Body

	// Decompress gzip if needed
	if strings.HasSuffix(strings.ToLower(sitemapURL), ".gz") ||
		strings.Contains(result.Headers.Get("Content-Encoding"), "gzip") {
		gr, err := gzip.NewReader(bytes.NewReader(data))
		if err == nil {
			decompressed, err := io.ReadAll(gr)
			if err == nil {
				data = decompressed
			}
		}
	}

	// Try sitemap index first
	var index xmlSitemapIndex
	if err := xml.Unmarshal(data, &index); err == nil && len(index.Sitemaps) > 0 {
		var all []SitemapEntry
		for _, s := range index.Sitemaps {
			if s.Loc == "" {
				continue
			}
			if limit > 0 && len(all) >= limit {
				break
			}
			remaining := 0
			if limit > 0 {
				remaining = limit - len(all)
			}
			entries, _ := fetchSitemap(ctx, fetcher, strings.TrimSpace(s.Loc), depth+1, remaining)
			all = append(all, entries...)
		}
		return all, nil
	}

	// Try urlset
	var urlset xmlURLSet
	if err := xml.Unmarshal(data, &urlset); err == nil {
		entries := make([]SitemapEntry, 0, len(urlset.URLs))
		for _, u := range urlset.URLs {
			if u.Loc == "" {
				continue
			}
			if limit > 0 && len(entries) >= limit {
				break
			}
			entry := SitemapEntry{URL: strings.TrimSpace(u.Loc)}
			if u.LastMod != "" {
				if t, err := time.Parse("2006-01-02", u.LastMod); err == nil {
					entry.LastMod = t
				}
			}
			entries = append(entries, entry)
		}
		return entries, nil
	}

	return nil, nil
}

// DiscoverSitemapURL tries to find a sitemap URL from robots.txt or common paths.
func DiscoverSitemapURL(ctx context.Context, fetcher *Fetcher, robotsCache *RobotsCache, siteURL string) string {
	// Check robots.txt for Sitemap: directive
	result := robotsCache.fetcher.Fetch(ctx, OriginOf(siteURL)+"/robots.txt")
	if result.Error == "" && result.StatusCode == 200 {
		for _, line := range strings.Split(string(result.Body), "\n") {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(strings.ToLower(line), "sitemap:") {
				parts := strings.SplitN(line, ":", 2)
				if len(parts) == 2 {
					candidate := strings.TrimSpace(parts[1])
					// Handle full URL like https://...
					if strings.HasPrefix(candidate, "//") {
						candidate = "https:" + candidate
					}
					if strings.HasPrefix(candidate, "http") {
						return candidate
					}
				}
			}
		}
	}
	// Try common paths
	for _, path := range []string{"/sitemap.xml", "/sitemap_index.xml", "/sitemap.xml.gz"} {
		u := OriginOf(siteURL) + path
		r := fetcher.Fetch(ctx, u)
		if r.Error == "" && r.StatusCode == 200 {
			return u
		}
	}
	return ""
}
