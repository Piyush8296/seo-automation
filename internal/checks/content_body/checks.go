package content_body

import (
	"fmt"
	"strings"

	"github.com/cars24/seo-automation/internal/models"
)

// PageChecks returns per-page content body checks.
func PageChecks() []models.PageCheck {
	return []models.PageCheck{
		&bodyVeryThin{},
		&bodyThin{},
		&bodyLoremIpsum{},
		&bodyTitleEqualsH1{},
		&bodyNoindexMeta{},
	}
}

// SiteChecks returns site-wide content body checks.
func SiteChecks() []models.SiteCheck {
	return []models.SiteCheck{
		&bodyNearDuplicate{},
	}
}

type bodyVeryThin struct{}

func (c *bodyVeryThin) Run(p *models.PageData) []models.CheckResult {
	if p.WordCount > 0 && p.WordCount < 100 {
		return []models.CheckResult{{
			ID:       "body.very_thin",
			Category: "Content",
			Severity: models.SeverityError,
			Message:  fmt.Sprintf("Very thin content (%d words, min 100)", p.WordCount),
			URL:      p.URL,
		}}
	}
	return nil
}

type bodyThin struct{}

func (c *bodyThin) Run(p *models.PageData) []models.CheckResult {
	if p.WordCount >= 100 && p.WordCount < 300 {
		return []models.CheckResult{{
			ID:       "body.thin",
			Category: "Content",
			Severity: models.SeverityWarning,
			Message:  fmt.Sprintf("Thin content (%d words, recommended 300+)", p.WordCount),
			URL:      p.URL,
		}}
	}
	return nil
}

type bodyLoremIpsum struct{}

func (c *bodyLoremIpsum) Run(p *models.PageData) []models.CheckResult {
	if strings.Contains(strings.ToLower(p.BodyText), "lorem ipsum") {
		return []models.CheckResult{{
			ID:       "body.lorem_ipsum",
			Category: "Content",
			Severity: models.SeverityError,
			Message:  "Page contains placeholder lorem ipsum text",
			URL:      p.URL,
		}}
	}
	return nil
}

type bodyTitleEqualsH1 struct{}

func (c *bodyTitleEqualsH1) Run(p *models.PageData) []models.CheckResult {
	if len(p.H1s) == 0 || p.Title == "" {
		return nil
	}
	if strings.EqualFold(strings.TrimSpace(p.Title), strings.TrimSpace(p.H1s[0])) {
		return []models.CheckResult{{
			ID:       "body.title_equals_h1",
			Category: "Content",
			Severity: models.SeverityNotice,
			Message:  "Page title is identical to H1 (missed opportunity for differentiation)",
			URL:      p.URL,
			Details:  p.Title,
		}}
	}
	return nil
}

type bodyNoindexMeta struct{}

func (c *bodyNoindexMeta) Run(p *models.PageData) []models.CheckResult {
	if strings.Contains(p.RobotsTag, "noindex") {
		return []models.CheckResult{{
			ID:       "body.noindex_meta",
			Category: "Content",
			Severity: models.SeverityNotice,
			Message:  "Page has meta robots noindex",
			URL:      p.URL,
			Details:  p.RobotsTag,
		}}
	}
	return nil
}

// Site-wide near-duplicate detection using Jaccard similarity on word sets.
type bodyNearDuplicate struct{}

func (c *bodyNearDuplicate) Run(pages []*models.PageData) []models.CheckResult {
	type pageWords struct {
		url   string
		words map[string]bool
	}
	var candidates []pageWords
	for _, p := range pages {
		if p.WordCount < 100 || p.BodyText == "" {
			continue
		}
		words := make(map[string]bool)
		for _, w := range strings.Fields(strings.ToLower(p.BodyText)) {
			if len(w) > 3 {
				words[w] = true
			}
		}
		if len(words) > 0 {
			candidates = append(candidates, pageWords{url: p.URL, words: words})
		}
	}

	reported := map[string]bool{}
	var results []models.CheckResult
	for i := 0; i < len(candidates); i++ {
		for j := i + 1; j < len(candidates); j++ {
			sim := jaccardSimilarity(candidates[i].words, candidates[j].words)
			if sim > 0.85 {
				if !reported[candidates[i].url] {
					results = append(results, models.CheckResult{
						ID:       "body.near_duplicate",
						Category: "Content",
						Severity: models.SeverityWarning,
						Message:  fmt.Sprintf("Near-duplicate content (%.0f%% similarity)", sim*100),
						URL:      candidates[i].url,
						Details:  candidates[j].url,
					})
					reported[candidates[i].url] = true
				}
				if !reported[candidates[j].url] {
					results = append(results, models.CheckResult{
						ID:       "body.near_duplicate",
						Category: "Content",
						Severity: models.SeverityWarning,
						Message:  fmt.Sprintf("Near-duplicate content (%.0f%% similarity)", sim*100),
						URL:      candidates[j].url,
						Details:  candidates[i].url,
					})
					reported[candidates[j].url] = true
				}
			}
		}
	}
	return results
}

func jaccardSimilarity(a, b map[string]bool) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	intersection := 0
	for w := range a {
		if b[w] {
			intersection++
		}
	}
	union := len(a) + len(b) - intersection
	if union == 0 {
		return 0
	}
	return float64(intersection) / float64(union)
}
