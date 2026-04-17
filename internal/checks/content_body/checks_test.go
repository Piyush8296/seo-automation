package content_body

import (
	"fmt"
	"strings"
	"testing"

	"github.com/cars24/seo-automation/internal/models"
)

// makePageWithBody creates a PageData with the given URL and body text.
func makePageWithBody(url, body string) *models.PageData {
	return &models.PageData{
		URL:       url,
		BodyText:  body,
		WordCount: len(strings.Fields(body)),
	}
}

// longText returns a string with n repeated words to exceed the 100-word minimum.
func longText(base string, n int) string {
	words := make([]string, n)
	for i := range words {
		words[i] = base
	}
	return strings.Join(words, " ")
}

func TestBodyExactDuplicate_IdenticalPages(t *testing.T) {
	body := longText("duplicate content words here today", 50)
	pages := []*models.PageData{
		makePageWithBody("https://example.com/a", body),
		makePageWithBody("https://example.com/b", body),
		makePageWithBody("https://example.com/c", body),
	}

	check := &bodyExactDuplicate{}
	results := check.Run(pages)

	if len(results) != 3 {
		t.Fatalf("expected 3 results (one per duplicate page), got %d", len(results))
	}
	for _, r := range results {
		if r.ID != "body.exact_duplicate" {
			t.Errorf("expected ID body.exact_duplicate, got %s", r.ID)
		}
		if r.Severity != models.SeverityError {
			t.Errorf("expected severity error, got %s", r.Severity)
		}
	}
}

func TestBodyExactDuplicate_UniquePages(t *testing.T) {
	pages := []*models.PageData{
		makePageWithBody("https://example.com/a", longText("alpha bravo charlie delta", 50)),
		makePageWithBody("https://example.com/b", longText("echo foxtrot golf hotel", 50)),
	}

	check := &bodyExactDuplicate{}
	results := check.Run(pages)

	if len(results) != 0 {
		t.Fatalf("expected 0 results for unique pages, got %d", len(results))
	}
}

func TestBodyExactDuplicate_SkipsThinContent(t *testing.T) {
	body := "short text only"
	pages := []*models.PageData{
		makePageWithBody("https://example.com/a", body),
		makePageWithBody("https://example.com/b", body),
	}

	check := &bodyExactDuplicate{}
	results := check.Run(pages)

	if len(results) != 0 {
		t.Fatalf("expected 0 results for thin pages, got %d", len(results))
	}
}

func TestSimhash_IdenticalTexts(t *testing.T) {
	text := longText("hello world testing simhash algorithm", 50)
	fp1 := simhash(text)
	fp2 := simhash(text)
	if fp1 != fp2 {
		t.Errorf("identical texts should produce identical fingerprints: %x != %x", fp1, fp2)
	}
}

func TestSimhash_SimilarTexts(t *testing.T) {
	// Build two texts that share 95%+ of their words (unique words matter for SimHash).
	shared := make([]string, 200)
	for i := range shared {
		shared[i] = fmt.Sprintf("word%04d", i)
	}
	base := strings.Join(shared, " ")
	// Replace only a few words to create a similar document.
	tweaked := make([]string, 200)
	copy(tweaked, shared)
	tweaked[0] = "changed0001"
	tweaked[1] = "changed0002"
	similar := strings.Join(tweaked, " ")

	fp1 := simhash(base)
	fp2 := simhash(similar)
	dist := hammingDistance(fp1, fp2)
	if dist > SimHashMaxDistance {
		t.Errorf("similar texts should have hamming distance <= %d, got %d", SimHashMaxDistance, dist)
	}
}

func TestSimhash_DifferentTexts(t *testing.T) {
	text1 := longText("javascript react frontend development webpack bundler", 50)
	text2 := longText("python machine learning neural network tensorflow", 50)

	fp1 := simhash(text1)
	fp2 := simhash(text2)
	dist := hammingDistance(fp1, fp2)
	if dist <= SimHashMaxDistance {
		t.Errorf("very different texts should have hamming distance > %d, got %d", SimHashMaxDistance, dist)
	}
}

func TestHammingDistance(t *testing.T) {
	tests := []struct {
		a, b uint64
		want int
	}{
		{0, 0, 0},
		{0xFF, 0xFF, 0},
		{0, 1, 1},
		{0b1010, 0b0101, 4},
		{^uint64(0), 0, 64},
	}
	for _, tt := range tests {
		got := hammingDistance(tt.a, tt.b)
		if got != tt.want {
			t.Errorf("hammingDistance(%x, %x) = %d, want %d", tt.a, tt.b, got, tt.want)
		}
	}
}

func TestBodyNearDuplicate_SkipsExactDuplicates(t *testing.T) {
	// Exact duplicates should NOT appear in near-duplicate results
	// because bodyExactDuplicate handles them.
	body := longText("identical content across multiple pages here", 50)
	pages := []*models.PageData{
		makePageWithBody("https://example.com/a", body),
		makePageWithBody("https://example.com/b", body),
	}

	check := &bodyNearDuplicate{}
	results := check.Run(pages)

	for _, r := range results {
		if r.ID == "body.near_duplicate" {
			t.Errorf("exact duplicates should not be flagged as near-duplicates")
		}
	}
}

func TestBodyNearDuplicate_FlagsSimilarContent(t *testing.T) {
	shared := make([]string, 200)
	for i := range shared {
		shared[i] = fmt.Sprintf("word%04d", i)
	}
	base := strings.Join(shared, " ")
	tweaked := make([]string, 200)
	copy(tweaked, shared)
	tweaked[0] = "changed0001"
	tweaked[1] = "changed0002"
	similar := strings.Join(tweaked, " ")

	pages := []*models.PageData{
		makePageWithBody("https://example.com/a", base),
		makePageWithBody("https://example.com/b", similar),
	}

	check := &bodyNearDuplicate{}
	results := check.Run(pages)

	if len(results) != 2 {
		t.Fatalf("expected 2 near-duplicate results, got %d", len(results))
	}
	for _, r := range results {
		if r.ID != "body.near_duplicate" {
			t.Errorf("expected ID body.near_duplicate, got %s", r.ID)
		}
		if r.Severity != models.SeverityWarning {
			t.Errorf("expected severity warning, got %s", r.Severity)
		}
	}
}
