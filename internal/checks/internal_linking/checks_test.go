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

func TestFooterHeavy(t *testing.T) {
	// Build a page with 12 internal links, 10 in footer (83%), 2 in content → should fire.
	var links []models.Link
	for i := 0; i < 10; i++ {
		links = append(links, models.Link{URL: "https://example.com/f", IsInternal: true, Position: models.PositionFooter})
	}
	links = append(links,
		models.Link{URL: "https://example.com/c1", IsInternal: true, Position: models.PositionContent},
		models.Link{URL: "https://example.com/c2", IsInternal: true, Position: models.PositionContent},
	)
	p := &models.PageData{URL: "https://example.com/heavy", Links: links}
	results := (&footerHeavy{}).Run(p)
	if len(results) != 1 {
		t.Fatalf("want 1 result, got %d", len(results))
	}
	if results[0].ID != "links.footer_heavy" {
		t.Errorf("wrong ID: %s", results[0].ID)
	}

	// Page with 6 internal links, 5 in footer (83%) but below the 10-link threshold → should NOT fire.
	var thin []models.Link
	for i := 0; i < 5; i++ {
		thin = append(thin, models.Link{URL: "https://example.com/f", IsInternal: true, Position: models.PositionFooter})
	}
	thin = append(thin, models.Link{URL: "https://example.com/c", IsInternal: true, Position: models.PositionContent})
	p2 := &models.PageData{URL: "https://example.com/thin", Links: thin}
	if r := (&footerHeavy{}).Run(p2); len(r) != 0 {
		t.Errorf("thin page should not fire, got %d results", len(r))
	}
}

func TestNoContentLinks(t *testing.T) {
	// Page with 6 internal links, all in nav/footer, none in content → fire.
	links := []models.Link{
		{URL: "https://example.com/a", IsInternal: true, Position: models.PositionNav},
		{URL: "https://example.com/b", IsInternal: true, Position: models.PositionNav},
		{URL: "https://example.com/c", IsInternal: true, Position: models.PositionFooter},
		{URL: "https://example.com/d", IsInternal: true, Position: models.PositionFooter},
		{URL: "https://example.com/e", IsInternal: true, Position: models.PositionHeader},
		{URL: "https://example.com/f", IsInternal: true, Position: models.PositionSidebar},
	}
	p := &models.PageData{URL: "https://example.com/bare", Links: links}
	results := (&noContentLinks{}).Run(p)
	if len(results) != 1 || results[0].ID != "links.no_content_links" {
		t.Fatalf("want 1 no_content_links result, got %+v", results)
	}

	// With one content link → should NOT fire.
	links = append(links, models.Link{URL: "https://example.com/g", IsInternal: true, Position: models.PositionContent})
	p2 := &models.PageData{URL: "https://example.com/ok", Links: links}
	if r := (&noContentLinks{}).Run(p2); len(r) != 0 {
		t.Errorf("page with 1 content link should not fire, got %d", len(r))
	}
}

func TestNavOrphan(t *testing.T) {
	// /in-nav is linked from nav; /not-in-nav is linked 3x from content only; /thin has 1 inlink.
	p1 := &models.PageData{
		URL:   "https://example.com/",
		Depth: 0,
		Links: []models.Link{
			{URL: "https://example.com/in-nav", IsInternal: true, Position: models.PositionNav},
			{URL: "https://example.com/not-in-nav", IsInternal: true, Position: models.PositionContent},
			{URL: "https://example.com/thin", IsInternal: true, Position: models.PositionContent},
		},
	}
	p2 := &models.PageData{
		URL:   "https://example.com/other",
		Depth: 1,
		Links: []models.Link{
			{URL: "https://example.com/not-in-nav", IsInternal: true, Position: models.PositionContent},
			{URL: "https://example.com/not-in-nav", IsInternal: true, Position: models.PositionFooter},
		},
	}
	inNav := &models.PageData{URL: "https://example.com/in-nav", Depth: 1}
	notInNav := &models.PageData{URL: "https://example.com/not-in-nav", Depth: 1}
	thin := &models.PageData{URL: "https://example.com/thin", Depth: 1}

	// Note: second link to /not-in-nav from p2 is counted twice (we tally all occurrences).
	pages := []*models.PageData{p1, p2, inNav, notInNav, thin}
	results := (&navOrphan{}).Run(pages)

	found := map[string]bool{}
	for _, r := range results {
		if r.ID != "links.nav_orphan" {
			t.Errorf("wrong ID: %s", r.ID)
		}
		found[r.URL] = true
	}
	if !found["https://example.com/not-in-nav"] {
		t.Error("expected nav_orphan for /not-in-nav (3 inlinks, no nav inlinks)")
	}
	if found["https://example.com/in-nav"] {
		t.Error("should not fire for /in-nav (has nav inlink)")
	}
	if found["https://example.com/thin"] {
		t.Error("should not fire for /thin (only 1 inlink, below threshold)")
	}
}

func TestNavOrphan_SkipWhenNoNavLinksSiteWide(t *testing.T) {
	// If no page uses nav-positioned links at all, the check should be silent.
	pages := []*models.PageData{
		{URL: "https://example.com/", Depth: 0, Links: []models.Link{
			{URL: "https://example.com/a", IsInternal: true, Position: models.PositionContent},
			{URL: "https://example.com/a", IsInternal: true, Position: models.PositionContent},
			{URL: "https://example.com/a", IsInternal: true, Position: models.PositionContent},
		}},
		{URL: "https://example.com/a", Depth: 1},
	}
	if r := (&navOrphan{}).Run(pages); len(r) != 0 {
		t.Errorf("no-nav-site should yield 0 results, got %d", len(r))
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
