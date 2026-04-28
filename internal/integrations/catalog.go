package integrations

// GetCatalog returns the first production-ready scope for checks that do not need
// Screaming Frog as the data source. It intentionally stores no credentials.
func GetCatalog() Catalog {
	checks := checkCapabilities()

	return Catalog{
		TotalChecks:  len(checks),
		GlobalInputs: globalInputs(),
		Providers:    providerDescriptors(checks),
		Checks:       checks,
		FeatureFlags: featureFlags(),
	}
}

func featureFlags() []string {
	return []string{
		"third_party_checks.enabled",
		"integrations.gsc.enabled",
		"integrations.ga4.enabled",
		"integrations.bing_webmaster.enabled",
		"integrations.google_business_profile.enabled",
		"integrations.seo_vendor.enabled",
		"integrations.serp_provider.enabled",
		"integrations.plagiarism_vendor.enabled",
		"integrations.news_monitor.enabled",
		"integrations.social_platforms.enabled",
		"integrations.review_vendor.enabled",
		"integrations.listings_vendor.enabled",
		"integrations.rum_vendor.enabled",
		"manual_workflow.enabled",
		"ai_content_evaluator.enabled",
	}
}
