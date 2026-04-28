package server

import (
	"testing"

	"github.com/cars24/seo-automation/internal/models"
)

func TestRenderedSEOChecksCoverSheetPackItems(t *testing.T) {
	items := analyzeRenderedSEO([]renderedSEOPageInput{
		{
			URL:                  "https://example.com/",
			RawTitle:             "Example",
			RawH1s:               []string{"Example"},
			RawCanonical:         "https://example.com/",
			RawWordCount:         120,
			RawInternalLinkCount: 10,
			RawSchemaCount:       1,
		},
	}, renderedSEOProbeResponse{
		Renderer: "test",
		Pages: []renderedSEOProbePage{
			{
				URL:                       "https://example.com/",
				FinalURL:                  "https://example.com/",
				Title:                     "Example",
				Canonical:                 "https://example.com/",
				H1s:                       []string{"Example"},
				WordCount:                 130,
				InternalLinkCount:         11,
				SchemaCount:               1,
				BeforeScrollWordCount:     130,
				AfterScrollWordCount:      130,
				BeforeScrollInternalLinks: 11,
				AfterScrollInternalLinks:  11,
				RenderTimeMS:              850,
			},
		},
	})

	seen := map[string]models.EvidenceCheckResult{}
	for _, item := range items {
		seen[item.ID] = item
	}
	for _, id := range []string{"JS-001", "JS-002", "JS-003", "JS-004", "JS-005", "JS-006", "JS-007", "JS-008", "JS-009", "JS-010"} {
		if _, ok := seen[id]; !ok {
			t.Fatalf("missing rendered SEO check %s in %#v", id, seen)
		}
	}
	if seen["JS-004"].Status != "pass" {
		t.Fatalf("expected JS-004 to pass, got %#v", seen["JS-004"])
	}
}

func TestRenderedSEOFlagsRenderedOnlyCriticalContent(t *testing.T) {
	items := analyzeRenderedSEO([]renderedSEOPageInput{
		{URL: "https://example.com/", RawWordCount: 5},
	}, renderedSEOProbeResponse{
		Renderer: "test",
		Pages: []renderedSEOProbePage{
			{
				URL:               "https://example.com/",
				FinalURL:          "https://example.com/",
				Title:             "Rendered title",
				Canonical:         "https://example.com/",
				H1s:               []string{"Rendered H1"},
				WordCount:         180,
				InternalLinkCount: 20,
			},
		},
	})

	byID := map[string]models.EvidenceCheckResult{}
	for _, item := range items {
		byID[item.ID] = item
	}
	if byID["JS-001"].Status != "fail" {
		t.Fatalf("expected JS-001 to fail when title/h1/canonical are rendered-only, got %#v", byID["JS-001"])
	}
	if byID["JS-003"].Status != "warning" {
		t.Fatalf("expected JS-003 to warn when raw HTML is thin vs rendered DOM, got %#v", byID["JS-003"])
	}
}
