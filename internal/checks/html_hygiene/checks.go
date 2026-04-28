package html_hygiene

import (
	"fmt"
	"io"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/cars24/seo-automation/internal/models"
	"golang.org/x/net/html"
)

const (
	category          = "HTML Structure"
	domDepthThreshold = 32
	paginationRelNext = "next"
	paginationRelPrev = "prev"
)

var pageNumberPatterns = []*regexp.Regexp{
	regexp.MustCompile(`[?&](page|p|pg|paged)\s*=\s*(\d+)`),
	regexp.MustCompile(`/page/(\d+)/?`),
	regexp.MustCompile(`/p/(\d+)/?`),
}

// PageChecks returns fast parser-only HTML structure checks.
func PageChecks() []models.PageCheck {
	return []models.PageCheck{
		&multipleHeadTags{},
		&doctypeMissing{},
		&domTooDeep{},
		&robotsMetaOutsideHead{},
		&paginationMarkupInvalid{},
	}
}

type multipleHeadTags struct{}

func (c *multipleHeadTags) Run(p *models.PageData) []models.CheckResult {
	if !hasInspectableHTML(p) {
		return nil
	}
	scan := scanRawHTML(p.RawHTML)
	if scan.HeadStartCount <= 1 {
		return nil
	}
	return []models.CheckResult{{
		ID:       "html.multiple_head",
		Category: category,
		Severity: models.SeverityError,
		Message:  "Page has more than one <head> tag",
		URL:      p.URL,
		Details:  fmt.Sprintf("head_tags=%d", scan.HeadStartCount),
	}}
}

type doctypeMissing struct{}

func (c *doctypeMissing) Run(p *models.PageData) []models.CheckResult {
	if !hasInspectableHTML(p) {
		return nil
	}
	scan := scanRawHTML(p.RawHTML)
	if scan.HasHTMLDoctype {
		return nil
	}
	return []models.CheckResult{{
		ID:       "html.doctype_missing",
		Category: category,
		Severity: models.SeverityWarning,
		Message:  "Page is missing a <!DOCTYPE html> declaration",
		URL:      p.URL,
	}}
}

type domTooDeep struct{}

func (c *domTooDeep) Run(p *models.PageData) []models.CheckResult {
	if !hasInspectableHTML(p) {
		return nil
	}
	depth, err := maxElementDepth(p.RawHTML)
	if err != nil || depth < domDepthThreshold {
		return nil
	}
	return []models.CheckResult{{
		ID:       "html.dom_too_deep",
		Category: category,
		Severity: models.SeverityWarning,
		Message:  "Page DOM is nested too deeply",
		URL:      p.URL,
		Details:  fmt.Sprintf("max_depth=%d threshold=<%d", depth, domDepthThreshold),
	}}
}

type robotsMetaOutsideHead struct{}

func (c *robotsMetaOutsideHead) Run(p *models.PageData) []models.CheckResult {
	if !hasInspectableHTML(p) {
		return nil
	}
	scan := scanRawHTML(p.RawHTML)
	if len(scan.RobotsMetaOutsideHead) == 0 {
		return nil
	}
	return []models.CheckResult{{
		ID:       "html.robots_meta_in_body",
		Category: category,
		Severity: models.SeverityError,
		Message:  "Robots meta tag appears outside the document head",
		URL:      p.URL,
		Details:  strings.Join(scan.RobotsMetaOutsideHead, " | "),
	}}
}

type paginationMarkupInvalid struct{}

func (c *paginationMarkupInvalid) Run(p *models.PageData) []models.CheckResult {
	if !hasInspectableHTML(p) {
		return nil
	}
	scan := scanRawHTML(p.RawHTML)
	if len(scan.PaginationLinks) == 0 {
		return nil
	}
	issues := paginationIssues(p, scan.PaginationLinks)
	if len(issues) == 0 {
		return nil
	}
	return []models.CheckResult{{
		ID:       "html.pagination_link_invalid",
		Category: category,
		Severity: models.SeverityWarning,
		Message:  "Pagination rel=next/prev markup is invalid",
		URL:      p.URL,
		Details:  strings.Join(issues, " | "),
	}}
}

type rawHTMLScan struct {
	HasHTMLDoctype        bool
	HeadStartCount        int
	RobotsMetaOutsideHead []string
	PaginationLinks       []paginationLink
}

type paginationLink struct {
	Rel    string
	Href   string
	InHead bool
}

func hasInspectableHTML(p *models.PageData) bool {
	if p == nil || strings.TrimSpace(p.RawHTML) == "" {
		return false
	}
	contentType := strings.ToLower(p.ContentType)
	return contentType == "" || strings.Contains(contentType, "html")
}

func scanRawHTML(raw string) rawHTMLScan {
	var scan rawHTMLScan
	tokenizer := html.NewTokenizer(strings.NewReader(raw))
	var stack []string
	for {
		tokenType := tokenizer.Next()
		switch tokenType {
		case html.ErrorToken:
			if tokenizer.Err() == io.EOF {
				return scan
			}
			return scan
		case html.DoctypeToken:
			doctype := strings.ToLower(strings.TrimSpace(string(tokenizer.Text())))
			if doctype == "html" || strings.HasPrefix(doctype, "html ") {
				scan.HasHTMLDoctype = true
			}
		case html.StartTagToken, html.SelfClosingTagToken:
			token := tokenizer.Token()
			tag := strings.ToLower(token.Data)
			if tag == "head" {
				scan.HeadStartCount++
			}
			inHead := stackContains(stack, "head") && !stackContains(stack, "body")
			switch tag {
			case "meta":
				if name := attrValue(token.Attr, "name"); isRobotsMetaName(name) && !inHead {
					scan.RobotsMetaOutsideHead = append(scan.RobotsMetaOutsideHead, describeMetaRobots(token.Attr))
				}
			case "link":
				for _, rel := range paginationRelTokens(attrValue(token.Attr, "rel")) {
					scan.PaginationLinks = append(scan.PaginationLinks, paginationLink{
						Rel:    rel,
						Href:   strings.TrimSpace(attrValue(token.Attr, "href")),
						InHead: inHead,
					})
				}
			}
			if tokenType == html.StartTagToken && !isVoidElement(tag) {
				stack = append(stack, tag)
			}
		case html.EndTagToken:
			tag := strings.ToLower(tokenizer.Token().Data)
			stack = popStackToTag(stack, tag)
		}
	}
}

func maxElementDepth(raw string) (int, error) {
	root, err := html.Parse(strings.NewReader(raw))
	if err != nil {
		return 0, err
	}
	maxDepth := 0
	var walk func(*html.Node, int)
	walk = func(node *html.Node, depth int) {
		if node.Type == html.ElementNode {
			depth++
			if depth > maxDepth {
				maxDepth = depth
			}
		}
		for child := node.FirstChild; child != nil; child = child.NextSibling {
			walk(child, depth)
		}
	}
	walk(root, 0)
	return maxDepth, nil
}

func paginationIssues(p *models.PageData, links []paginationLink) []string {
	currentURL := firstNonEmpty(p.FinalURL, p.URL)
	currentPageNum := pageNumber(currentURL)
	counts := map[string]int{}
	var issues []string
	for _, link := range links {
		counts[link.Rel]++
		label := "rel=" + link.Rel
		if !link.InHead {
			issues = append(issues, label+" link is outside <head>")
		}
		if link.Href == "" {
			issues = append(issues, label+" link is missing href")
			continue
		}
		resolved, err := resolveURL(currentURL, link.Href)
		if err != nil {
			issues = append(issues, label+" href is invalid: "+link.Href)
			continue
		}
		if resolved.Scheme != "http" && resolved.Scheme != "https" {
			issues = append(issues, label+" href is not crawlable: "+resolved.String())
			continue
		}
		if sameURLWithoutFragment(currentURL, resolved.String()) {
			issues = append(issues, label+" points to the current page")
		}
		if link.Rel == paginationRelPrev && currentPageNum == 1 {
			issues = append(issues, "rel=prev appears on page 1")
		}
	}
	for rel, count := range counts {
		if count > 1 {
			issues = append(issues, fmt.Sprintf("multiple rel=%s links found (%d)", rel, count))
		}
	}
	return uniqueStrings(issues)
}

func attrValue(attrs []html.Attribute, key string) string {
	key = strings.ToLower(key)
	for _, attr := range attrs {
		if strings.ToLower(attr.Key) == key {
			return attr.Val
		}
	}
	return ""
}

func isRobotsMetaName(name string) bool {
	name = strings.ToLower(strings.TrimSpace(name))
	switch name {
	case "robots", "googlebot", "bingbot", "slurp", "duckduckbot", "baiduspider", "yandexbot":
		return true
	default:
		return strings.HasPrefix(name, "googlebot-")
	}
}

func describeMetaRobots(attrs []html.Attribute) string {
	name := strings.TrimSpace(attrValue(attrs, "name"))
	content := strings.TrimSpace(attrValue(attrs, "content"))
	if content == "" {
		return "name=" + name
	}
	return "name=" + name + " content=" + content
}

func paginationRelTokens(rel string) []string {
	var tokens []string
	for _, token := range strings.Fields(strings.ToLower(rel)) {
		switch token {
		case paginationRelNext, paginationRelPrev:
			tokens = append(tokens, token)
		}
	}
	return tokens
}

func resolveURL(baseURL string, href string) (*url.URL, error) {
	linkURL, err := url.Parse(strings.TrimSpace(href))
	if err != nil {
		return nil, err
	}
	if linkURL.IsAbs() {
		return linkURL, nil
	}
	base, err := url.Parse(baseURL)
	if err != nil {
		return nil, err
	}
	return base.ResolveReference(linkURL), nil
}

func pageNumber(rawURL string) int {
	for _, pattern := range pageNumberPatterns {
		matches := pattern.FindStringSubmatch(rawURL)
		if len(matches) < 2 {
			continue
		}
		n, err := strconv.Atoi(matches[len(matches)-1])
		if err == nil {
			return n
		}
	}
	return 0
}

func sameURLWithoutFragment(a string, b string) bool {
	parsedA, errA := url.Parse(a)
	parsedB, errB := url.Parse(b)
	if errA != nil || errB != nil {
		return strings.TrimRight(a, "/") == strings.TrimRight(b, "/")
	}
	parsedA.Fragment = ""
	parsedB.Fragment = ""
	return strings.TrimRight(parsedA.String(), "/") == strings.TrimRight(parsedB.String(), "/")
}

func popStackToTag(stack []string, tag string) []string {
	for i := len(stack) - 1; i >= 0; i-- {
		if stack[i] == tag {
			return stack[:i]
		}
	}
	return stack
}

func stackContains(stack []string, tag string) bool {
	for i := len(stack) - 1; i >= 0; i-- {
		if stack[i] == tag {
			return true
		}
	}
	return false
}

func isVoidElement(tag string) bool {
	switch tag {
	case "area", "base", "br", "col", "embed", "hr", "img", "input", "link", "meta", "param", "source", "track", "wbr":
		return true
	default:
		return false
	}
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

func uniqueStrings(values []string) []string {
	seen := map[string]bool{}
	var unique []string
	for _, value := range values {
		if seen[value] {
			continue
		}
		seen[value] = true
		unique = append(unique, value)
	}
	return unique
}
