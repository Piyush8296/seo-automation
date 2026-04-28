package checks

import (
	"reflect"
	"testing"

	"github.com/cars24/seo-automation/internal/models"
)

func TestChecklistIDsForCoversCatalog(t *testing.T) {
	for _, descriptor := range GetCheckDescriptors() {
		if len(descriptor.ChecklistIDs) == 0 {
			t.Fatalf("check %q has no checklist mapping", descriptor.ID)
		}
	}
}

func TestChecklistIDsForReturnsDefensiveCopy(t *testing.T) {
	ids := ChecklistIDsFor("canonical.missing")
	if !reflect.DeepEqual(ids, []string{"CANONICAL-001"}) {
		t.Fatalf("canonical.missing mapped to %v", ids)
	}

	ids[0] = "MUTATED"
	if got := ChecklistIDsFor("canonical.missing"); !reflect.DeepEqual(got, []string{"CANONICAL-001"}) {
		t.Fatalf("mapping was mutated: %v", got)
	}
}

func TestChecklistIDsForRegistryIDMapsToSelf(t *testing.T) {
	if got := ChecklistIDsFor("JS-001"); !reflect.DeepEqual(got, []string{"JS-001"}) {
		t.Fatalf("registry ID should map to itself, got %v", got)
	}
}

func TestAttachChecklistMappings(t *testing.T) {
	audit := &models.SiteAudit{
		Pages: []*models.PageData{
			{
				URL: "https://example.com/",
				CheckResults: []models.CheckResult{
					{ID: "title.missing"},
				},
			},
		},
		SiteChecks: []models.CheckResult{
			{ID: "sitemap.missing"},
		},
		RenderedSEO: []models.EvidenceCheckResult{
			{ID: "JS-004"},
		},
	}

	AttachChecklistMappings(audit)

	if got := audit.Pages[0].CheckResults[0].ChecklistIDs; !reflect.DeepEqual(got, []string{"TITLE-001"}) {
		t.Fatalf("page check mapping = %v", got)
	}
	if got := audit.SiteChecks[0].ChecklistIDs; !reflect.DeepEqual(got, []string{"SITEMAP-001"}) {
		t.Fatalf("site check mapping = %v", got)
	}
	if got := audit.RenderedSEO[0].ChecklistIDs; !reflect.DeepEqual(got, []string{"JS-004"}) {
		t.Fatalf("evidence mapping = %v", got)
	}
}
