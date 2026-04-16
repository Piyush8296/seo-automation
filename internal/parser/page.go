package parser

import (
	"bytes"
	"net/url"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/cars24/seo-automation/internal/models"
)

// ExtractPage parses raw HTML and populates a PageData struct.
func ExtractPage(body []byte, pageURL string, headers map[string]string) (*models.PageData, error) {
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	page := &models.PageData{}
	page.URL = pageURL
	page.Headers = headers
	page.HTMLSizeBytes = len(body)
	page.RawHTML = string(body)

	// Meta
	page.Title, page.MetaDesc, page.Canonical, page.RobotsTag, page.ViewportContent, page.HasViewport = ExtractMeta(doc)
	page.RobotsDirectives, page.XRobotsTag = ParseRobotsDirectives(page.RobotsTag, headers)

	// OG / Twitter / hreflang
	page.OGTags = ExtractOGTags(doc)
	page.TwitterTags = ExtractTwitterTags(doc)
	page.HreflangTags = ExtractHreflang(doc)

	// Headings
	page.H1s, page.H2s, page.H3s = ExtractHeadings(doc)

	// Links
	page.Links = ExtractLinks(doc, pageURL)

	// Schema
	page.SchemaJSONRaw = ExtractSchemaJSON(doc)

	// Images
	page.Images = extractImages(doc, pageURL)

	// Body text + word count
	page.BodyText, page.WordCount = extractBodyText(doc)

	// Performance metrics
	page.RenderBlockingScripts = CountRenderBlockingScripts(doc)
	page.RenderBlockingCSS = CountRenderBlockingCSS(doc)
	page.ExternalScriptCount = CountExternalScripts(doc, pageURL)
	page.InlineCSSBytes = GetInlineCSSBytes(doc)

	return page, nil
}

// ExtractMobileData extracts a lighter subset for mobile comparison.
func ExtractMobileData(body []byte, pageURL string) *models.MobilePageData {
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(body))
	if err != nil {
		return nil
	}
	title, metaDesc, canonical, _, _, _ := ExtractMeta(doc)
	h1s, _, _ := ExtractHeadings(doc)
	schema := ExtractSchemaJSON(doc)
	og := ExtractOGTags(doc)
	links := ExtractLinks(doc, pageURL)
	_, wordCount := extractBodyText(doc)

	return &models.MobilePageData{
		Title:         title,
		MetaDesc:      metaDesc,
		Canonical:     canonical,
		H1s:           h1s,
		SchemaJSONRaw: schema,
		OGTags:        og,
		Links:         links,
		WordCount:     wordCount,
	}
}

// extractBodyText removes script/style nodes and returns visible text and word count.
func extractBodyText(doc *goquery.Document) (string, int) {
	// Clone and remove noise elements
	doc.Find("script, style, noscript, head").Remove()
	text := strings.TrimSpace(doc.Find("body").Text())
	// Normalize whitespace
	text = strings.Join(strings.Fields(text), " ")
	words := strings.Fields(text)
	return text, len(words)
}

// extractImages returns all <img> elements with metadata.
func extractImages(doc *goquery.Document, baseURL string) []models.Image {
	var images []models.Image
	idx := 0
	doc.Find("img").Each(func(_ int, s *goquery.Selection) {
		src := strings.TrimSpace(s.AttrOr("src", ""))
		if src == "" {
			// Check data-src for lazy loaded images
			src = strings.TrimSpace(s.AttrOr("data-src", ""))
		}
		if src == "" {
			return
		}

		abs := src
		if parsed, err := url.Parse(src); err == nil && !parsed.IsAbs() {
			if resolvedURL, err2 := url.Parse(baseURL); err2 == nil {
				abs = resolvedURL.ResolveReference(parsed).String()
			}
		}

		_, altPresent := s.Attr("alt")
		alt := strings.TrimSpace(s.AttrOr("alt", ""))

		width := 0
		height := 0
		if w := s.AttrOr("width", ""); w != "" {
			width = parseIntAttr(w)
		}
		if h := s.AttrOr("height", ""); h != "" {
			height = parseIntAttr(h)
		}

		loading := strings.ToLower(strings.TrimSpace(s.AttrOr("loading", "")))
		_, hasSrcset := s.Attr("srcset")
		if !hasSrcset {
			_, hasSrcset = s.Attr("data-srcset")
		}

		img := models.Image{
			Src:        abs,
			Alt:        alt,
			AltPresent: altPresent,
			Width:      width,
			Height:     height,
			Loading:    loading,
			HasSrcset:  hasSrcset,
			Format:     imageFormatFromURL(abs),
			// Mark first 3 images as potential above-fold candidates
			IsAboveFold: idx < 3,
		}
		images = append(images, img)
		idx++
	})
	return images
}

// parseIntAttr safely parses an integer from an HTML attribute string.
func parseIntAttr(s string) int {
	// Remove "px" suffix if present
	s = strings.TrimSuffix(strings.TrimSpace(s), "px")
	n := 0
	for _, c := range s {
		if c >= '0' && c <= '9' {
			n = n*10 + int(c-'0')
		} else {
			break
		}
	}
	return n
}

// imageFormatFromURL extracts the image format from a URL's file extension.
func imageFormatFromURL(rawURL string) string {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return ""
	}
	path := strings.ToLower(parsed.Path)
	// Strip query/fragment already handled by url.Parse
	if idx := strings.LastIndex(path, "."); idx != -1 {
		ext := path[idx+1:]
		switch ext {
		case "jpg", "jpeg":
			return "jpg"
		case "png", "gif", "webp", "avif", "svg", "bmp", "ico", "tiff", "tif":
			return ext
		}
	}
	return ""
}
