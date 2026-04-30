package onpage_keywords

import (
	"testing"

	"github.com/cars24/seo-automation/internal/models"
)

func TestTitleMetaKeywordAlignmentFindsExpectedIssues(t *testing.T) {
	page := &models.PageData{
		URL:       "https://www.cars24.com/buy-used-maruti-swift-2021-cars-bangalore/",
		Title:     "Best Deals | CARS24",
		MetaDesc:  "Browse inventory today.",
		H1s:       []string{"Used Cars"},
		H2s:       []string{"Cars Cars Cars Cars"},
		BodyText:  repeatedWords("Generic inventory copy without model or city terms.", 40),
		WordCount: 240,
	}

	got := runIDs(page)
	want := []string{
		"title.primary_keyword_missing",
		"title.model_missing",
		"title.year_missing",
		"title.city_model_missing",
		"meta_desc.primary_keyword_missing",
		"headings.h1.primary_keyword_missing",
		"headings.h2.keyword_stuffing",
		"keyword.first_100_missing",
		"keyword.density_out_of_range",
		"keyword.surface_mismatch",
	}
	for _, id := range want {
		if !got[id] {
			t.Fatalf("expected issue %q in %#v", id, got)
		}
	}
}

func TestTitleMetaKeywordAlignmentPassesAlignedPage(t *testing.T) {
	page := &models.PageData{
		URL:       "https://www.cars24.com/buy-used-maruti-swift-2021-cars-bangalore/",
		Title:     "Buy Used Maruti Swift 2021 Cars in Bangalore | CARS24",
		MetaDesc:  "Buy used Maruti Swift 2021 cars in Bangalore with verified listings and get the best CARS24 deal today.",
		H1s:       []string{"Used Maruti Swift 2021 Cars in Bangalore"},
		H2s:       []string{"Compare Maruti Swift prices", "Used car loan and inspection details"},
		BodyText:  repeatedWords("Buy used Maruti Swift 2021 cars in Bangalore with inspection reports and price details.", 30),
		WordCount: 330,
	}

	got := runIDs(page)
	unwanted := []string{
		"title.primary_keyword_missing",
		"title.keyword_not_near_start",
		"title.brand_missing",
		"title.model_missing",
		"title.year_missing",
		"title.city_model_missing",
		"meta_desc.primary_keyword_missing",
		"meta_desc.cta_missing",
		"headings.h1.primary_keyword_missing",
		"keyword.first_100_missing",
		"keyword.surface_mismatch",
	}
	for _, id := range unwanted {
		if got[id] {
			t.Fatalf("did not expect issue %q in %#v", id, got)
		}
	}
}

func TestDynamicPlaceholdersAndBadCharacters(t *testing.T) {
	page := &models.PageData{
		URL:       "https://www.cars24.com/sell-used-cars/",
		Title:     "Sell Used Cars in {{city}} |||| CARS24",
		MetaDesc:  "Sell your car in [city] for the best price.",
		H1s:       []string{"Sell Used Cars"},
		BodyText:  repeatedWords("Sell used cars online with CARS24.", 40),
		WordCount: 200,
	}

	got := runIDs(page)
	for _, id := range []string{"title.dynamic_placeholder", "title.special_chars", "meta_desc.dynamic_placeholder"} {
		if !got[id] {
			t.Fatalf("expected issue %q in %#v", id, got)
		}
	}
}

func runIDs(page *models.PageData) map[string]bool {
	ids := map[string]bool{}
	for _, check := range PageChecks() {
		for _, result := range check.Run(page) {
			ids[result.ID] = true
		}
	}
	return ids
}

func repeatedWords(text string, count int) string {
	out := ""
	for i := 0; i < count; i++ {
		if i > 0 {
			out += " "
		}
		out += text
	}
	return out
}
