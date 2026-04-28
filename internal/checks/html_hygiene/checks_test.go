package html_hygiene

import (
	"strings"
	"testing"

	"github.com/cars24/seo-automation/internal/models"
)

func TestMultipleHeadTags(t *testing.T) {
	page := testPage(`<!DOCTYPE html><html><head></head><head></head><body></body></html>`)

	results := (&multipleHeadTags{}).Run(page)

	assertSingleCheck(t, results, "html.multiple_head")
	if !strings.Contains(results[0].Details, "head_tags=2") {
		t.Fatalf("expected head tag count in details, got %q", results[0].Details)
	}
}

func TestDoctypeMissing(t *testing.T) {
	page := testPage(`<html><head></head><body></body></html>`)

	results := (&doctypeMissing{}).Run(page)

	assertSingleCheck(t, results, "html.doctype_missing")
}

func TestDoctypePresentPasses(t *testing.T) {
	page := testPage(`<!doctype html><html><head></head><body></body></html>`)

	if results := (&doctypeMissing{}).Run(page); len(results) != 0 {
		t.Fatalf("expected no doctype issue, got %#v", results)
	}
}

func TestDOMTooDeep(t *testing.T) {
	var nested strings.Builder
	nested.WriteString(`<!DOCTYPE html><html><head></head><body>`)
	for i := 0; i < 35; i++ {
		nested.WriteString("<div>")
	}
	for i := 0; i < 35; i++ {
		nested.WriteString("</div>")
	}
	nested.WriteString(`</body></html>`)
	page := testPage(nested.String())

	results := (&domTooDeep{}).Run(page)

	assertSingleCheck(t, results, "html.dom_too_deep")
	if !strings.Contains(results[0].Details, "max_depth=") {
		t.Fatalf("expected max depth in details, got %q", results[0].Details)
	}
}

func TestRobotsMetaOutsideHead(t *testing.T) {
	page := testPage(`<!DOCTYPE html><html><head></head><body><meta name="robots" content="noindex, follow"></body></html>`)

	results := (&robotsMetaOutsideHead{}).Run(page)

	assertSingleCheck(t, results, "html.robots_meta_in_body")
	if !strings.Contains(results[0].Details, "noindex") {
		t.Fatalf("expected robots content in details, got %q", results[0].Details)
	}
}

func TestRobotsMetaInHeadPasses(t *testing.T) {
	page := testPage(`<!DOCTYPE html><html><head><meta name="robots" content="index, follow"></head><body></body></html>`)

	if results := (&robotsMetaOutsideHead{}).Run(page); len(results) != 0 {
		t.Fatalf("expected no robots placement issue, got %#v", results)
	}
}

func TestPaginationMarkupInvalid(t *testing.T) {
	page := testPage(`<!DOCTYPE html><html><head><link rel="next" href=""></head><body><link rel="prev" href="?page=1"></body></html>`)
	page.URL = "https://example.com/list?page=1"
	page.FinalURL = page.URL

	results := (&paginationMarkupInvalid{}).Run(page)

	assertSingleCheck(t, results, "html.pagination_link_invalid")
	if !strings.Contains(results[0].Details, "missing href") || !strings.Contains(results[0].Details, "outside <head>") {
		t.Fatalf("expected pagination placement and href details, got %q", results[0].Details)
	}
}

func TestPaginationMarkupValidPasses(t *testing.T) {
	page := testPage(`<!DOCTYPE html><html><head><link rel="prev" href="?page=1"><link rel="next" href="?page=3"></head><body></body></html>`)
	page.URL = "https://example.com/list?page=2"
	page.FinalURL = page.URL

	if results := (&paginationMarkupInvalid{}).Run(page); len(results) != 0 {
		t.Fatalf("expected no pagination issue, got %#v", results)
	}
}

func testPage(rawHTML string) *models.PageData {
	return &models.PageData{
		URL:         "https://example.com/",
		FinalURL:    "https://example.com/",
		ContentType: "text/html; charset=utf-8",
		RawHTML:     rawHTML,
	}
}

func assertSingleCheck(t *testing.T, results []models.CheckResult, id string) {
	t.Helper()
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d: %#v", len(results), results)
	}
	if results[0].ID != id {
		t.Fatalf("expected check %q, got %q", id, results[0].ID)
	}
}
