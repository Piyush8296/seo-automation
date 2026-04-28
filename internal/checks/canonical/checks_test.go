package canonical

import (
	"testing"

	"github.com/cars24/seo-automation/internal/models"
)

func TestCanonicalMultipleDetectsMoreThanOneTag(t *testing.T) {
	page := &models.PageData{
		URL:      "https://example.com/a",
		FinalURL: "https://example.com/a",
		RawHTML:  `<html><head><link rel="canonical" href="https://example.com/a"><link rel="canonical" href="https://example.com/b"></head><body></body></html>`,
	}
	results := (&canonicalMultiple{}).Run(page)
	if len(results) != 1 || results[0].ID != "canonical.multiple" {
		t.Fatalf("results=%v, want canonical.multiple", results)
	}
}

func TestCanonicalInBodyDetectsBodyTag(t *testing.T) {
	page := &models.PageData{
		URL:      "https://example.com/a",
		FinalURL: "https://example.com/a",
		RawHTML:  `<html><head></head><body><link rel="canonical" href="https://example.com/a"></body></html>`,
	}
	results := (&canonicalInBody{}).Run(page)
	if len(results) != 1 || results[0].ID != "canonical.in_body" {
		t.Fatalf("results=%v, want canonical.in_body", results)
	}
}

func TestCanonicalHeaderMismatch(t *testing.T) {
	page := &models.PageData{
		URL:       "https://example.com/a",
		FinalURL:  "https://example.com/a",
		Canonical: "https://example.com/a",
		Headers:   map[string]string{"link": `<https://example.com/b>; rel="canonical"`},
	}
	results := (&canonicalHeaderMismatch{}).Run(page)
	if len(results) != 1 || results[0].ID != "canonical.header_mismatch" {
		t.Fatalf("results=%v, want canonical.header_mismatch", results)
	}
}

func TestCanonicalTargetNon200UsesCrawledTarget(t *testing.T) {
	pages := []*models.PageData{
		{URL: "https://example.com/a", FinalURL: "https://example.com/a", Canonical: "https://example.com/b", StatusCode: 200},
		{URL: "https://example.com/b", FinalURL: "https://example.com/b", StatusCode: 404},
	}
	results := (&canonicalTargetNon200{}).Run(pages)
	if len(results) != 1 || results[0].ID != "canonical.target_non_200" {
		t.Fatalf("results=%v, want canonical.target_non_200", results)
	}
}

func TestCanonicalLoopAndChain(t *testing.T) {
	loopPages := []*models.PageData{
		{URL: "https://example.com/a", FinalURL: "https://example.com/a", Canonical: "https://example.com/b", StatusCode: 200},
		{URL: "https://example.com/b", FinalURL: "https://example.com/b", Canonical: "https://example.com/a", StatusCode: 200},
	}
	loopResults := (&canonicalLoop{}).Run(loopPages)
	if len(loopResults) == 0 || loopResults[0].ID != "canonical.loop" {
		t.Fatalf("loop results=%v, want canonical.loop", loopResults)
	}

	chainPages := []*models.PageData{
		{URL: "https://example.com/a", FinalURL: "https://example.com/a", Canonical: "https://example.com/b", StatusCode: 200},
		{URL: "https://example.com/b", FinalURL: "https://example.com/b", Canonical: "https://example.com/c", StatusCode: 200},
		{URL: "https://example.com/c", FinalURL: "https://example.com/c", Canonical: "https://example.com/c", StatusCode: 200},
	}
	chainResults := (&canonicalChain{}).Run(chainPages)
	if len(chainResults) != 1 || chainResults[0].ID != "canonical.chain" {
		t.Fatalf("chain results=%v, want canonical.chain", chainResults)
	}
}

func TestCanonicalParamsSelfReference(t *testing.T) {
	page := &models.PageData{
		URL:       "https://example.com/cars?sort=price",
		FinalURL:  "https://example.com/cars?sort=price",
		Canonical: "https://example.com/cars?sort=price",
	}
	results := (&canonicalParamsSelfReference{}).Run(page)
	if len(results) != 1 || results[0].ID != "canonical.params_self_reference" {
		t.Fatalf("results=%v, want canonical.params_self_reference", results)
	}
}

func TestCanonicalCountryFolderMismatch(t *testing.T) {
	page := &models.PageData{
		URL:       "https://example.com/in/cars",
		FinalURL:  "https://example.com/in/cars",
		Canonical: "https://example.com/au/cars",
	}
	results := (&canonicalCountryFolderMismatch{}).Run(page)
	if len(results) != 1 || results[0].ID != "canonical.country_folder_mismatch" {
		t.Fatalf("results=%v, want canonical.country_folder_mismatch", results)
	}
}

func TestCanonicalWWWVariantDetectsMixedCanonicalHosts(t *testing.T) {
	pages := []*models.PageData{
		{URL: "https://www.example.com/a", FinalURL: "https://www.example.com/a", Canonical: "https://www.example.com/a", StatusCode: 200},
		{URL: "https://www.example.com/b", FinalURL: "https://www.example.com/b", Canonical: "https://example.com/b", StatusCode: 200},
	}
	results := (&canonicalWWWVariant{}).Run(pages)
	if len(results) != 1 || results[0].ID != "canonical.www_variant" {
		t.Fatalf("results=%v, want canonical.www_variant", results)
	}
}
