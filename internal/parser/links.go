package parser

import (
	"net/url"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/cars24/seo-automation/internal/models"
)

// resolveURL resolves rawURL relative to base and normalizes it.
func resolveURL(rawURL, base string) (string, error) {
	baseU, err := url.Parse(base)
	if err != nil {
		return "", err
	}
	ref, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}
	resolved := baseU.ResolveReference(ref)
	resolved.Fragment = ""
	resolved.Scheme = strings.ToLower(resolved.Scheme)
	resolved.Host = strings.ToLower(resolved.Host)
	return resolved.String(), nil
}

func isHTTPScheme(u string) bool {
	parsed, err := url.Parse(u)
	if err != nil {
		return false
	}
	s := strings.ToLower(parsed.Scheme)
	return s == "http" || s == "https"
}

func sameHost(a, b string) bool {
	au, err := url.Parse(a)
	if err != nil {
		return false
	}
	bu, err := url.Parse(b)
	if err != nil {
		return false
	}
	return strings.ToLower(au.Hostname()) == strings.ToLower(bu.Hostname())
}

// ExtractLinks returns all <a href> links from the page.
func ExtractLinks(doc *goquery.Document, baseURL string) []models.Link {
	var links []models.Link
	seen := make(map[string]bool)

	doc.Find("a[href]").Each(func(_ int, s *goquery.Selection) {
		href := strings.TrimSpace(s.AttrOr("href", ""))
		if href == "" || strings.HasPrefix(href, "#") ||
			strings.HasPrefix(strings.ToLower(href), "javascript:") ||
			strings.HasPrefix(strings.ToLower(href), "mailto:") ||
			strings.HasPrefix(strings.ToLower(href), "tel:") {
			return
		}

		abs, err := resolveURL(href, baseURL)
		if err != nil || !isHTTPScheme(abs) {
			return
		}
		if seen[abs] {
			return
		}
		seen[abs] = true

		rel := strings.ToLower(strings.TrimSpace(s.AttrOr("rel", "")))
		isFollow := !strings.Contains(rel, "nofollow")
		isInternal := sameHost(abs, baseURL)
		text := strings.TrimSpace(s.Text())

		links = append(links, models.Link{
			URL:        abs,
			Text:       text,
			Rel:        rel,
			IsInternal: isInternal,
			IsFollow:   isFollow,
		})
	})
	return links
}

// HasMixedContent returns true if an HTTPS page loads HTTP resources.
func HasMixedContent(doc *goquery.Document, pageURL string) bool {
	parsed, err := url.Parse(pageURL)
	if err != nil || strings.ToLower(parsed.Scheme) != "https" {
		return false
	}
	mixed := false
	doc.Find("script[src],link[href],img[src],iframe[src]").EachWithBreak(func(_ int, s *goquery.Selection) bool {
		for _, attr := range []string{"src", "href"} {
			val := s.AttrOr(attr, "")
			if strings.HasPrefix(val, "http://") {
				mixed = true
				return false
			}
		}
		return true
	})
	return mixed
}

// CountRenderBlockingScripts counts synchronous <script> tags in <head>.
func CountRenderBlockingScripts(doc *goquery.Document) int {
	count := 0
	doc.Find("head script[src]").Each(func(_ int, s *goquery.Selection) {
		_, hasAsync := s.Attr("async")
		_, hasDefer := s.Attr("defer")
		if !hasAsync && !hasDefer {
			count++
		}
	})
	return count
}

// CountRenderBlockingCSS counts <link rel="stylesheet"> tags in <head>.
func CountRenderBlockingCSS(doc *goquery.Document) int {
	count := 0
	doc.Find("head link[rel='stylesheet']").Each(func(_ int, s *goquery.Selection) {
		media := strings.ToLower(s.AttrOr("media", ""))
		if media == "" || media == "all" || media == "screen" {
			count++
		}
	})
	return count
}

// CountExternalScripts counts <script src> tags pointing to a different host.
func CountExternalScripts(doc *goquery.Document, baseURL string) int {
	count := 0
	doc.Find("script[src]").Each(func(_ int, s *goquery.Selection) {
		src := s.AttrOr("src", "")
		if src == "" {
			return
		}
		abs, err := resolveURL(src, baseURL)
		if err == nil && !sameHost(abs, baseURL) {
			count++
		}
	})
	return count
}

// GetInlineCSSBytes returns the total byte size of all <style> tag content.
func GetInlineCSSBytes(doc *goquery.Document) int {
	total := 0
	doc.Find("style").Each(func(_ int, s *goquery.Selection) {
		total += len(s.Text())
	})
	return total
}

// ExtractResources discovers CSS, JS, and font sub-resources referenced by the page.
// Deduplicates by URL across discovery sources.
func ExtractResources(doc *goquery.Document, baseURL string) []models.Resource {
	var resources []models.Resource
	seen := make(map[string]bool)

	add := func(raw string, rtype models.ResourceType) {
		raw = strings.TrimSpace(raw)
		if raw == "" {
			return
		}
		abs, err := resolveURL(raw, baseURL)
		if err != nil || !isHTTPScheme(abs) {
			return
		}
		key := string(rtype) + "|" + abs
		if seen[key] {
			return
		}
		seen[key] = true
		resources = append(resources, models.Resource{
			URL:        abs,
			Type:       rtype,
			IsInternal: sameHost(abs, baseURL),
		})
	}

	doc.Find("script[src]").Each(func(_ int, s *goquery.Selection) {
		add(s.AttrOr("src", ""), models.ResourceScript)
	})

	doc.Find("link[rel]").Each(func(_ int, s *goquery.Selection) {
		rel := strings.ToLower(s.AttrOr("rel", ""))
		href := s.AttrOr("href", "")
		if strings.Contains(rel, "stylesheet") {
			add(href, models.ResourceCSS)
			return
		}
		if strings.Contains(rel, "preload") {
			as := strings.ToLower(s.AttrOr("as", ""))
			switch as {
			case "font":
				add(href, models.ResourceFont)
			case "style":
				add(href, models.ResourceCSS)
			case "script":
				add(href, models.ResourceScript)
			}
		}
	})

	return resources
}

// CountFontFaceNoDisplay scans inline <style> tags for @font-face blocks missing font-display.
func CountFontFaceNoDisplay(doc *goquery.Document) int {
	count := 0
	doc.Find("style").Each(func(_ int, s *goquery.Selection) {
		css := strings.ToLower(s.Text())
		idx := 0
		for {
			found := strings.Index(css[idx:], "@font-face")
			if found == -1 {
				break
			}
			start := idx + found
			brace := strings.Index(css[start:], "{")
			if brace == -1 {
				break
			}
			end := strings.Index(css[start+brace:], "}")
			if end == -1 {
				break
			}
			block := css[start+brace : start+brace+end]
			if !strings.Contains(block, "font-display") {
				count++
			}
			idx = start + brace + end + 1
		}
	})
	return count
}

// HasPreconnect checks if the page has any <link rel="preconnect"> tags.
func HasPreconnect(doc *goquery.Document) bool {
	found := false
	doc.Find("link[rel]").EachWithBreak(func(_ int, s *goquery.Selection) bool {
		rel := strings.ToLower(s.AttrOr("rel", ""))
		if strings.Contains(rel, "preconnect") {
			found = true
			return false
		}
		return true
	})
	return found
}
