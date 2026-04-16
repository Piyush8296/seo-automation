package content_body

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"hash/fnv"
	"math/bits"
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
		&bodyExactDuplicate{},
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

// Site-wide exact duplicate detection using SHA-256 content hash.
type bodyExactDuplicate struct{}

func (c *bodyExactDuplicate) Run(pages []*models.PageData) []models.CheckResult {
	hashToURLs := make(map[string][]string)
	for _, p := range pages {
		if p.WordCount < 100 || p.BodyText == "" {
			continue
		}
		h := sha256.Sum256([]byte(p.BodyText))
		key := hex.EncodeToString(h[:])
		hashToURLs[key] = append(hashToURLs[key], p.URL)
	}

	var results []models.CheckResult
	for _, urls := range hashToURLs {
		if len(urls) < 2 {
			continue
		}
		for i, u := range urls {
			// Pick another URL from the group as the "details" reference.
			other := urls[0]
			if i == 0 {
				other = urls[1]
			}
			results = append(results, models.CheckResult{
				ID:       "body.exact_duplicate",
				Category: "Content",
				Severity: models.SeverityError,
				Message:  fmt.Sprintf("Exact duplicate content (%d pages share identical body)", len(urls)),
				URL:      u,
				Details:  other,
			})
		}
	}
	return results
}

// Site-wide near-duplicate detection using SimHash fingerprinting.
// SimHash produces a 64-bit fingerprint per document; documents with
// a Hamming distance ≤ 3 are flagged as near-duplicates.
type bodyNearDuplicate struct{}

const simhashMaxDistance = 3

func (c *bodyNearDuplicate) Run(pages []*models.PageData) []models.CheckResult {
	type candidate struct {
		url       string
		hash      string // exact content hash to skip exact duplicates
		fingerprint uint64
	}

	// Build fingerprints.
	hashToURL := make(map[string]string) // track exact hashes to skip them
	var candidates []candidate
	for _, p := range pages {
		if p.WordCount < 100 || p.BodyText == "" {
			continue
		}
		h := sha256.Sum256([]byte(p.BodyText))
		key := hex.EncodeToString(h[:])

		// Skip if we already have an exact duplicate — those are handled by bodyExactDuplicate.
		if _, seen := hashToURL[key]; seen {
			continue
		}
		hashToURL[key] = p.URL

		fp := simhash(p.BodyText)
		candidates = append(candidates, candidate{url: p.URL, hash: key, fingerprint: fp})
	}

	reported := make(map[string]bool)
	var results []models.CheckResult
	for i := 0; i < len(candidates); i++ {
		for j := i + 1; j < len(candidates); j++ {
			dist := hammingDistance(candidates[i].fingerprint, candidates[j].fingerprint)
			if dist <= simhashMaxDistance {
				sim := float64(64-dist) / 64.0 * 100
				if !reported[candidates[i].url] {
					results = append(results, models.CheckResult{
						ID:       "body.near_duplicate",
						Category: "Content",
						Severity: models.SeverityWarning,
						Message:  fmt.Sprintf("Near-duplicate content (~%.0f%% similarity, hamming=%d)", sim, dist),
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
						Message:  fmt.Sprintf("Near-duplicate content (~%.0f%% similarity, hamming=%d)", sim, dist),
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

// simhash computes a 64-bit SimHash fingerprint for the given text.
// Each word (len > 3) is hashed with FNV-64a; the bit-vector is accumulated
// and collapsed into the final fingerprint.
func simhash(text string) uint64 {
	var v [64]int
	for _, w := range strings.Fields(strings.ToLower(text)) {
		if len(w) <= 3 {
			continue
		}
		h := fnv.New64a()
		h.Write([]byte(w))
		hash := h.Sum64()
		for i := 0; i < 64; i++ {
			if hash&(1<<uint(i)) != 0 {
				v[i]++
			} else {
				v[i]--
			}
		}
	}
	var fp uint64
	for i := 0; i < 64; i++ {
		if v[i] > 0 {
			fp |= 1 << uint(i)
		}
	}
	return fp
}

// hammingDistance returns the number of differing bits between two uint64 values.
func hammingDistance(a, b uint64) int {
	return bits.OnesCount64(a ^ b)
}
