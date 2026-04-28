package structured_data

import (
	"testing"

	"github.com/cars24/seo-automation/internal/models"
)

func TestSchemaSearchActionMissingHomepage(t *testing.T) {
	pages := []*models.PageData{{
		URL:           "https://example.com/",
		FinalURL:      "https://example.com/",
		Depth:         0,
		SchemaJSONRaw: []string{`{"@context":"https://schema.org","@type":"WebSite","name":"Example"}`},
	}}

	results := (&schemaSearchActionMissingHomepage{}).Run(pages)

	assertSchemaCheck(t, results, "schema.website.searchaction_missing")
}

func TestSchemaSearchActionPresentPasses(t *testing.T) {
	pages := []*models.PageData{{
		URL:      "https://example.com/",
		FinalURL: "https://example.com/",
		Depth:    0,
		SchemaJSONRaw: []string{`{
			"@context":"https://schema.org",
			"@type":"WebSite",
			"potentialAction":{"@type":"SearchAction","target":"https://example.com/search?q={query}","query-input":"required name=query"}
		}`},
	}}

	if results := (&schemaSearchActionMissingHomepage{}).Run(pages); len(results) != 0 {
		t.Fatalf("expected no SearchAction issue, got %#v", results)
	}
}

func TestSchemaWebsiteMissingHomepage(t *testing.T) {
	pages := []*models.PageData{{
		URL:           "https://example.com/",
		FinalURL:      "https://example.com/",
		Depth:         0,
		SchemaJSONRaw: []string{`{"@context":"https://schema.org","@type":"Organization","name":"Example"}`},
	}}

	results := (&schemaWebSiteMissingHomepage{}).Run(pages)

	assertSchemaCheck(t, results, "schema.website.missing_homepage")
}

func TestReviewRatingMissingOnVehicleProduct(t *testing.T) {
	page := schemaPage(`{"@context":"https://schema.org","@type":"Product","name":"Honda City used car"}`)

	results := (&schemaReviewRatingMissing{}).Run(page)

	assertSchemaCheck(t, results, "schema.review_rating.missing")
}

func TestReviewRatingPresentPasses(t *testing.T) {
	page := schemaPage(`{"@context":"https://schema.org","@type":"Product","name":"Honda City used car","aggregateRating":{"@type":"AggregateRating","ratingValue":"4.6"}}`)

	if results := (&schemaReviewRatingMissing{}).Run(page); len(results) != 0 {
		t.Fatalf("expected no review rating issue, got %#v", results)
	}
}

func TestLazyLoadedSchemaRisk(t *testing.T) {
	page := &models.PageData{
		URL:     "https://example.com/",
		RawHTML: `<html><head><script type="application/ld+json" data-src="/schema.json"></script></head></html>`,
	}

	results := (&schemaLazyLoadedRisk{}).Run(page)

	assertSchemaCheck(t, results, "schema.lazy_loaded_risk")
}

func TestHiddenContentMismatchForFAQ(t *testing.T) {
	page := schemaPage(`{
		"@context":"https://schema.org",
		"@type":"FAQPage",
		"mainEntity":[{"@type":"Question","name":"What warranty is included?","acceptedAnswer":{"@type":"Answer","text":"A seven day return warranty is included."}}]
	}`)
	page.BodyText = "This page only talks about car prices and finance options."
	page.Title = "Used cars"

	results := (&schemaHiddenContentMismatch{}).Run(page)

	assertSchemaCheck(t, results, "schema.hidden_content_mismatch")
}

func TestHiddenContentVisiblePasses(t *testing.T) {
	page := schemaPage(`{
		"@context":"https://schema.org",
		"@type":"FAQPage",
		"mainEntity":[{"@type":"Question","name":"What warranty is included?","acceptedAnswer":{"@type":"Answer","text":"A seven day return warranty is included."}}]
	}`)
	page.BodyText = "What warranty is included? A seven day return warranty is included."
	page.Title = "Used cars"

	if results := (&schemaHiddenContentMismatch{}).Run(page); len(results) != 0 {
		t.Fatalf("expected no hidden content issue, got %#v", results)
	}
}

func TestHowToMissingForGuidePage(t *testing.T) {
	page := schemaPage(`{"@context":"https://schema.org","@type":"Article","headline":"How to sell a car"}`)
	page.URL = "https://example.com/guides/how-to-sell-car"
	page.Title = "How to sell a car"
	page.BodyText = "Step by step process to sell a car online."

	results := (&schemaHowToMissing{}).Run(page)

	assertSchemaCheck(t, results, "schema.howto.missing")
}

func TestProductListMissingForListingPage(t *testing.T) {
	page := schemaPage(`{"@context":"https://schema.org","@type":"WebPage","name":"Used cars"}`)
	page.URL = "https://example.com/buy-used-cars/"
	page.Title = "Used cars for sale"
	page.BodyText = "Filter and sort used cars by price, model, year and fuel."

	results := (&schemaProductListMissing{}).Run(page)

	assertSchemaCheck(t, results, "schema.product_list.missing")
}

func TestEventMissingForEventPage(t *testing.T) {
	page := schemaPage(`{"@context":"https://schema.org","@type":"WebPage","name":"Car launch event"}`)
	page.URL = "https://example.com/events/car-launch"
	page.Title = "Car launch event"
	page.BodyText = "Register for the launch event. Date, venue and schedule are available."

	results := (&schemaEventMissing{}).Run(page)

	assertSchemaCheck(t, results, "schema.event.missing")
}

func TestSpeakableMissingForNewsArticle(t *testing.T) {
	page := schemaPage(`{"@context":"https://schema.org","@type":"NewsArticle","headline":"Car market update"}`)
	page.URL = "https://example.com/news/car-market-update"
	page.Title = "Car market update"

	results := (&schemaSpeakableMissing{}).Run(page)

	assertSchemaCheck(t, results, "schema.speakable.missing")
}

func schemaPage(raw string) *models.PageData {
	return &models.PageData{
		URL:           "https://example.com/car/honda-city",
		FinalURL:      "https://example.com/car/honda-city",
		Title:         "Honda City used car",
		BodyText:      "Honda City used car price transmission fuel owner km",
		WordCount:     10,
		SchemaJSONRaw: []string{raw},
	}
}

func assertSchemaCheck(t *testing.T, results []models.CheckResult, id string) {
	t.Helper()
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d: %#v", len(results), results)
	}
	if results[0].ID != id {
		t.Fatalf("expected %q, got %q", id, results[0].ID)
	}
}
