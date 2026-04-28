package parser

import (
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/cars24/seo-automation/internal/models"
)

// ExtractMeta returns common meta SEO fields from the document.
func ExtractMeta(doc *goquery.Document) (title, metaDesc, canonical, robotsTag, viewportContent string, hasViewport bool) {
	title = strings.TrimSpace(doc.Find("title").First().Text())

	doc.Find("meta").Each(func(_ int, s *goquery.Selection) {
		name := strings.ToLower(strings.TrimSpace(s.AttrOr("name", "")))
		prop := strings.ToLower(strings.TrimSpace(s.AttrOr("property", "")))
		content := strings.TrimSpace(s.AttrOr("content", ""))

		switch name {
		case "description":
			metaDesc = content
		case "robots":
			robotsTag = strings.ToLower(content)
		case "viewport":
			hasViewport = true
			viewportContent = content
		}
		_ = prop
	})

	doc.Find("link[rel]").Each(func(_ int, s *goquery.Selection) {
		rel := strings.ToLower(strings.TrimSpace(s.AttrOr("rel", "")))
		if relTokenContains(rel, "canonical") {
			canonical = strings.TrimSpace(s.AttrOr("href", ""))
		}
	})
	return
}

func relTokenContains(rel string, token string) bool {
	token = strings.ToLower(strings.TrimSpace(token))
	for _, part := range strings.Fields(strings.ToLower(rel)) {
		if part == token {
			return true
		}
	}
	return false
}

// ParseRobotsDirectives merges directives from <meta name="robots"> and
// the X-Robots-Tag HTTP header into a deduplicated, lowercased slice.
func ParseRobotsDirectives(metaContent string, headers map[string]string) (directives []string, xRobotsRaw string) {
	seen := map[string]bool{}
	for _, d := range strings.Split(metaContent, ",") {
		d = strings.TrimSpace(d)
		if d != "" && !seen[d] {
			seen[d] = true
			directives = append(directives, d)
		}
	}
	xRobotsRaw = strings.TrimSpace(headers["x-robots-tag"])
	if xRobotsRaw != "" {
		for _, d := range strings.Split(strings.ToLower(xRobotsRaw), ",") {
			d = strings.TrimSpace(d)
			if idx := strings.Index(d, ":"); idx != -1 {
				d = strings.TrimSpace(d[idx+1:])
			}
			if d != "" && !seen[d] {
				seen[d] = true
				directives = append(directives, d)
			}
		}
	}
	return
}

// ExtractOGTags returns all og: Open Graph meta properties.
func ExtractOGTags(doc *goquery.Document) map[string]string {
	tags := make(map[string]string)
	doc.Find("meta[property]").Each(func(_ int, s *goquery.Selection) {
		prop := strings.ToLower(strings.TrimSpace(s.AttrOr("property", "")))
		if strings.HasPrefix(prop, "og:") {
			tags[prop] = strings.TrimSpace(s.AttrOr("content", ""))
		}
	})
	return tags
}

// ExtractTwitterTags returns all twitter: meta name tags.
func ExtractTwitterTags(doc *goquery.Document) map[string]string {
	tags := make(map[string]string)
	doc.Find("meta[name]").Each(func(_ int, s *goquery.Selection) {
		name := strings.ToLower(strings.TrimSpace(s.AttrOr("name", "")))
		if strings.HasPrefix(name, "twitter:") {
			tags[name] = strings.TrimSpace(s.AttrOr("content", ""))
		}
	})
	return tags
}

// ExtractHreflang returns all hreflang alternate link tags.
func ExtractHreflang(doc *goquery.Document) []models.Hreflang {
	var out []models.Hreflang
	doc.Find("link[rel='alternate']").Each(func(_ int, s *goquery.Selection) {
		lang := strings.TrimSpace(s.AttrOr("hreflang", ""))
		href := strings.TrimSpace(s.AttrOr("href", ""))
		if lang != "" && href != "" {
			out = append(out, models.Hreflang{Lang: lang, URL: href})
		}
	})
	return out
}

// ExtractHeadings returns text content of h1, h2, h3 elements.
func ExtractHeadings(doc *goquery.Document) (h1s, h2s, h3s []string) {
	doc.Find("h1").Each(func(_ int, s *goquery.Selection) {
		h1s = append(h1s, strings.TrimSpace(s.Text()))
	})
	doc.Find("h2").Each(func(_ int, s *goquery.Selection) {
		h2s = append(h2s, strings.TrimSpace(s.Text()))
	})
	doc.Find("h3").Each(func(_ int, s *goquery.Selection) {
		h3s = append(h3s, strings.TrimSpace(s.Text()))
	})
	return
}
