package integrations

func globalInputs() []RequiredInput {
	return []RequiredInput{
		input("site.url", "Primary site URL", "Canonical production site URL used for property matching and reporting.", InputSourceUI, true, false, "https://www.cars24.com/"),
		input("site.root_domain", "Root domain", "Root domain for domain-level keyword, backlink, and brand checks.", InputSourceUI, true, false, "cars24.com"),
		input("site.brand_name", "Brand name", "Brand term used for branded search, mentions, reviews, and SERP checks.", InputSourceUI, true, false, "CARS24"),
		input("site.country", "Default country", "Primary market for GSC, SERP, keyword, and local visibility filters.", InputSourceUI, true, false, "IN"),
		input("site.locale", "Default locale", "Language or locale for GSC URL inspection, SERP, and AI evaluation prompts.", InputSourceUI, false, false, "en-IN"),
		input("site.target_keywords", "Target keywords", "Keyword set used by SERP, keyword-gap, cannibalization, and AI content checks.", InputSourceUI, false, false, "used cars in Delhi"),
		input("site.competitor_domains", "Competitor domains", "Competitor domains for keyword gap, backlink gap, and content comprehensiveness checks.", InputSourceUI, false, false, "spinny.com"),
		input("site.locations", "Business locations", "City or branch list for local SEO, GBP, and local pack checks.", InputSourceUI, false, false, "Delhi"),
	}
}
