package internal_linking

import (
	"testing"

	"github.com/cars24/seo-automation/internal/models"
)

func makePages() []*models.PageData {
	seed := &models.PageData{
		URL:   "https://example.com/",
		Depth: 0,
		Links: []models.Link{
			{URL: "https://example.com/about", IsInternal: true},
			{URL: "https://example.com/about", IsInternal: true},
			{URL: "https://example.com/about", IsInternal: true},
			{URL: "https://example.com/contact", IsInternal: true},
		},
	}
	about := &models.PageData{
		URL:   "https://example.com/about",
		Depth: 1,
		Links: []models.Link{
			{URL: "https://example.com/contact", IsInternal: true},
			{URL: "https://example.com/contact", IsInternal: true},
		},
	}
	contact := &models.PageData{
		URL:   "https://example.com/contact",
		Depth: 1,
	}
	orphan := &models.PageData{
		URL:   "https://example.com/orphan",
		Depth: 1,
	}
	lowLink := &models.PageData{
		URL:   "https://example.com/low",
		Depth: 1,
	}
	// Give lowLink exactly 1 inlink
	seed.Links = append(seed.Links, models.Link{URL: "https://example.com/low", IsInternal: true})

	return []*models.PageData{seed, about, contact, orphan, lowLink}
}

func TestLowInlinks(t *testing.T) {
	pages := makePages()
	check := &lowInlinks{}
	results := check.Run(pages)

	// contact has 3 inlinks (1 from seed + 2 from about) — should NOT fire
	// about has 3 inlinks from seed — should NOT fire
	// orphan has 0 inlinks — should NOT fire (that's orphanPage's job)
	// low has 1 inlink — should fire
	found := map[string]bool{}
	for _, r := range results {
		found[r.URL] = true
		if r.ID != "links.page.low_inlinks" {
			t.Errorf("wrong ID: %s", r.ID)
		}
	}

	if !found["https://example.com/low"] {
		t.Error("expected low_inlinks for /low (1 inlink)")
	}
	if found["https://example.com/about"] {
		t.Error("should not fire for /about (3 inlinks)")
	}
	if found["https://example.com/contact"] {
		t.Error("should not fire for /contact (3 inlinks)")
	}
	if found["https://example.com/orphan"] {
		t.Error("should not fire for /orphan (0 inlinks, orphan check handles this)")
	}
}

func TestOrphanPage(t *testing.T) {
	pages := makePages()
	check := &orphanPage{}
	results := check.Run(pages)

	found := map[string]bool{}
	for _, r := range results {
		found[r.URL] = true
	}

	if !found["https://example.com/orphan"] {
		t.Error("expected orphan for /orphan")
	}
	if found["https://example.com/about"] {
		t.Error("should not be orphan: /about")
	}
}

func TestInlinkCountPopulated(t *testing.T) {
	pages := makePages()
	buildInlinkMap(pages)

	for _, p := range pages {
		switch p.URL {
		case "https://example.com/about":
			if p.InlinkCount != 3 {
				t.Errorf("/about inlink count = %d, want 3", p.InlinkCount)
			}
		case "https://example.com/contact":
			if p.InlinkCount != 3 {
				t.Errorf("/contact inlink count = %d, want 3", p.InlinkCount)
			}
		case "https://example.com/orphan":
			if p.InlinkCount != 0 {
				t.Errorf("/orphan inlink count = %d, want 0", p.InlinkCount)
			}
		case "https://example.com/low":
			if p.InlinkCount != 1 {
				t.Errorf("/low inlink count = %d, want 1", p.InlinkCount)
			}
		}
	}
}
