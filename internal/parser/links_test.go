package parser

import (
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
	"github.com/cars24/seo-automation/internal/models"
)

func extract(t *testing.T, html string) []models.Link {
	t.Helper()
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	return ExtractLinks(doc, "https://example.com/")
}

func positionOf(links []models.Link, url string) models.LinkPosition {
	for _, l := range links {
		if l.URL == url {
			return l.Position
		}
	}
	return ""
}

func TestClassifyLinkPosition_SemanticTags(t *testing.T) {
	html := `<html><body>
		<header><a href="/h">H</a></header>
		<nav><a href="/n">N</a></nav>
		<aside><a href="/s">S</a></aside>
		<main><a href="/c">C</a></main>
		<footer><a href="/f">F</a></footer>
		<a href="/orphan">O</a>
	</body></html>`
	links := extract(t, html)

	tests := map[string]models.LinkPosition{
		"https://example.com/h":      models.PositionHeader,
		"https://example.com/n":      models.PositionNav,
		"https://example.com/s":      models.PositionSidebar,
		"https://example.com/c":      models.PositionContent,
		"https://example.com/f":      models.PositionFooter,
		"https://example.com/orphan": models.PositionContent, // default
	}
	for url, want := range tests {
		if got := positionOf(links, url); got != want {
			t.Errorf("%s: got %q, want %q", url, got, want)
		}
	}
}

func TestClassifyLinkPosition_NavInsideHeader(t *testing.T) {
	// <nav> inside <header> should classify as nav (more specific wins — nearest ancestor).
	html := `<html><body>
		<header><nav><a href="/x">X</a></nav></header>
	</body></html>`
	links := extract(t, html)
	if got := positionOf(links, "https://example.com/x"); got != models.PositionNav {
		t.Errorf("nav inside header: got %q, want nav", got)
	}
}

func TestClassifyLinkPosition_ARIARoles(t *testing.T) {
	html := `<html><body>
		<div role="navigation"><a href="/n">N</a></div>
		<div role="contentinfo"><a href="/f">F</a></div>
		<div role="banner"><a href="/h">H</a></div>
		<div role="complementary"><a href="/s">S</a></div>
		<div role="main"><a href="/c">C</a></div>
	</body></html>`
	links := extract(t, html)

	tests := map[string]models.LinkPosition{
		"https://example.com/n": models.PositionNav,
		"https://example.com/f": models.PositionFooter,
		"https://example.com/h": models.PositionHeader,
		"https://example.com/s": models.PositionSidebar,
		"https://example.com/c": models.PositionContent,
	}
	for url, want := range tests {
		if got := positionOf(links, url); got != want {
			t.Errorf("%s: got %q, want %q", url, got, want)
		}
	}
}

func TestClassifyLinkPosition_IDClassHeuristics(t *testing.T) {
	html := `<html><body>
		<div id="footer"><a href="/f">F</a></div>
		<div class="site-wrapper header"><a href="/h">H</a></div>
		<div class="sidebar"><a href="/s">S</a></div>
		<div class="subheader"><a href="/sub">Sub</a></div>
	</body></html>`
	links := extract(t, html)

	if got := positionOf(links, "https://example.com/f"); got != models.PositionFooter {
		t.Errorf("id=footer: got %q, want footer", got)
	}
	if got := positionOf(links, "https://example.com/h"); got != models.PositionHeader {
		t.Errorf("class~=header: got %q, want header", got)
	}
	if got := positionOf(links, "https://example.com/s"); got != models.PositionSidebar {
		t.Errorf("class~=sidebar: got %q, want sidebar", got)
	}
	// "subheader" token should NOT match "header" (exact token match).
	if got := positionOf(links, "https://example.com/sub"); got != models.PositionContent {
		t.Errorf("class~=subheader: got %q, want content (no false-positive)", got)
	}
}
