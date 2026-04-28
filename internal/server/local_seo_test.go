package server

import "testing"

func TestLocalSEOChecksCoverSheetItems(t *testing.T) {
	checks := localSEOChecks()
	if got, want := len(checks), 15; got != want {
		t.Fatalf("localSEOChecks() len=%d, want %d", got, want)
	}

	seen := map[string]LocalSEOCheck{}
	for _, check := range checks {
		seen[check.ID] = check
	}
	for _, id := range []string{
		"LOCAL-001", "LOCAL-002", "LOCAL-003", "LOCAL-004", "LOCAL-005",
		"LOCAL-006", "LOCAL-007", "LOCAL-008", "LOCAL-009", "LOCAL-010",
		"LOCAL-011", "LOCAL-012", "LOCAL-013", "LOCAL-014", "LOCAL-015",
	} {
		if _, ok := seen[id]; !ok {
			t.Fatalf("missing local SEO check %s", id)
		}
	}

	qna := seen["LOCAL-006"]
	if qna.Channel != "Manual" || qna.Automation != "Manual only" {
		t.Fatalf("LOCAL-006 should remain manual after GBP Q&A API discontinuation: %#v", qna)
	}
}
