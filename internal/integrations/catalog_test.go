package integrations

import "testing"

func TestGetCatalogHasConsistentProviderMappings(t *testing.T) {
	catalog := GetCatalog()
	if catalog.TotalChecks == 0 {
		t.Fatal("expected external check catalog to contain checks")
	}
	if catalog.TotalChecks != len(catalog.Checks) {
		t.Fatalf("total_checks=%d, len(checks)=%d", catalog.TotalChecks, len(catalog.Checks))
	}

	providers := map[string]ProviderDescriptor{}
	for _, provider := range catalog.Providers {
		providers[provider.ID] = provider
	}

	for _, check := range catalog.Checks {
		if check.ID == "" {
			t.Fatalf("check has empty ID: %#v", check)
		}
		if check.FeatureFlag == "" {
			t.Fatalf("%s has no feature flag", check.ID)
		}
		if _, ok := providers[check.PrimaryProviderID]; !ok {
			t.Fatalf("%s references unknown primary provider %q", check.ID, check.PrimaryProviderID)
		}
		for _, providerID := range check.ProviderIDs {
			if _, ok := providers[providerID]; !ok {
				t.Fatalf("%s references unknown provider %q", check.ID, providerID)
			}
		}
	}
}

func TestProviderCheckIDsReferenceCatalogChecks(t *testing.T) {
	catalog := GetCatalog()
	checkIDs := map[string]bool{}
	for _, check := range catalog.Checks {
		checkIDs[check.ID] = true
	}

	for _, provider := range catalog.Providers {
		for _, checkID := range provider.CheckIDs {
			if !checkIDs[checkID] {
				t.Fatalf("provider %s references unknown check %s", provider.ID, checkID)
			}
		}
	}
}
