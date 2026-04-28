package server

// AppSettings holds runtime configuration managed via the UI.
type AppSettings struct {
	SkipLinkHosts []string            `json:"skip_link_hosts"`
	SiteProfile   SiteProfileSettings `json:"site_profile"`
	FeatureFlags  FeatureFlagSettings `json:"feature_flags"`
	Integrations  IntegrationSettings `json:"integrations"`
}

// SiteProfileSettings is shared by external API checks. It intentionally stores
// only non-secret configuration that can safely round-trip through the UI.
type SiteProfileSettings struct {
	SiteURL           string   `json:"site_url"`
	RootDomain        string   `json:"root_domain"`
	BrandName         string   `json:"brand_name"`
	Country           string   `json:"country"`
	Locale            string   `json:"locale"`
	TargetKeywords    []string `json:"target_keywords"`
	CompetitorDomains []string `json:"competitor_domains"`
	Locations         []string `json:"locations"`
}

// FeatureFlagSettings keeps external checks dark until credentials, quota, and
// reporting expectations are configured.
type FeatureFlagSettings struct {
	ThirdPartyChecks      bool `json:"third_party_checks"`
	GSC                   bool `json:"gsc"`
	GA4                   bool `json:"ga4"`
	BingWebmaster         bool `json:"bing_webmaster"`
	GoogleBusinessProfile bool `json:"google_business_profile"`
	SEOVendor             bool `json:"seo_vendor"`
	SERPProvider          bool `json:"serp_provider"`
	PlagiarismVendor      bool `json:"plagiarism_vendor"`
	NewsMonitor           bool `json:"news_monitor"`
	SocialPlatforms       bool `json:"social_platforms"`
	ReviewVendor          bool `json:"review_vendor"`
	ListingsVendor        bool `json:"listings_vendor"`
	RUMVendor             bool `json:"rum_vendor"`
	ManualWorkflow        bool `json:"manual_workflow"`
	AIContentEvaluator    bool `json:"ai_content_evaluator"`
}

// IntegrationSettings stores external provider IDs and property references. API
// keys, OAuth tokens, and refresh tokens must live in a secret store or env vars.
type IntegrationSettings struct {
	GSC        GSCIntegrationSettings       `json:"gsc"`
	GA4        GA4IntegrationSettings       `json:"ga4"`
	Bing       BingIntegrationSettings      `json:"bing_webmaster"`
	GBP        GBPIntegrationSettings       `json:"google_business_profile"`
	SEOVendor  SEOVendorIntegrationSettings `json:"seo_vendor"`
	SERP       SERPIntegrationSettings      `json:"serp_provider"`
	AI         AIIntegrationSettings        `json:"ai_content_evaluator"`
	Plagiarism GenericVendorSettings        `json:"plagiarism_vendor"`
	News       GenericVendorSettings        `json:"news_monitor"`
	Social     SocialIntegrationSettings    `json:"social_platforms"`
	Reviews    GenericVendorSettings        `json:"review_vendor"`
	Listings   GenericVendorSettings        `json:"listings_vendor"`
	RUM        GenericVendorSettings        `json:"rum_vendor"`
}

type GSCIntegrationSettings struct {
	PropertyURL    string `json:"property_url"`
	Country        string `json:"country"`
	Device         string `json:"device"`
	DateRange      string `json:"date_range"`
	OAuthConnected bool   `json:"oauth_connected"`
}

type GA4IntegrationSettings struct {
	PropertyID        string   `json:"property_id"`
	WebStreamID       string   `json:"web_stream_id"`
	ExpectedKeyEvents []string `json:"expected_key_events"`
	ExpectedUTMRules  []string `json:"expected_utm_rules"`
	DateRange         string   `json:"date_range"`
}

type BingIntegrationSettings struct {
	SiteURL        string   `json:"site_url"`
	SitemapURLs    []string `json:"sitemap_urls"`
	OAuthConnected bool     `json:"oauth_connected"`
}

type GBPIntegrationSettings struct {
	AccountIDs         []string `json:"account_ids"`
	LocationIDs        []string `json:"location_ids"`
	NAPBaseline        string   `json:"nap_baseline"`
	RequiredCategories []string `json:"required_categories"`
	PostsCadenceDays   int      `json:"posts_cadence_days"`
}

type SEOVendorIntegrationSettings struct {
	Vendor     string `json:"vendor"`
	ProjectID  string `json:"project_id"`
	CampaignID string `json:"campaign_id"`
	Market     string `json:"market"`
}

type SERPIntegrationSettings struct {
	Vendor   string `json:"vendor"`
	Location string `json:"location"`
	Device   string `json:"device"`
	Language string `json:"language"`
}

type AIIntegrationSettings struct {
	Model   string `json:"model"`
	Gateway string `json:"gateway"`
}

type GenericVendorSettings struct {
	Vendor    string `json:"vendor"`
	ProjectID string `json:"project_id"`
}

type SocialIntegrationSettings struct {
	OfficialHandles []string `json:"official_handles"`
	Vendor          string   `json:"vendor"`
}

// DefaultAppSettings returns defaults for a fresh manager.
func DefaultAppSettings() AppSettings {
	return AppSettings{
		SkipLinkHosts: append([]string(nil), DefaultSkipLinkHosts...),
		SiteProfile: SiteProfileSettings{
			Country: "IN",
			Locale:  "en-IN",
		},
		Integrations: IntegrationSettings{
			GSC: GSCIntegrationSettings{
				DateRange: "last_28_days",
			},
			GA4: GA4IntegrationSettings{
				DateRange: "last_28_days",
			},
			SERP: SERPIntegrationSettings{
				Device:   "mobile",
				Language: "en",
			},
		},
	}
}

// Normalize fills defaults and copies slice fields so settings are safe to store.
func (s AppSettings) Normalize() AppSettings {
	if s.SkipLinkHosts == nil {
		s.SkipLinkHosts = append([]string(nil), DefaultSkipLinkHosts...)
	} else {
		s.SkipLinkHosts = append([]string(nil), s.SkipLinkHosts...)
	}
	s.SiteProfile.TargetKeywords = append([]string(nil), s.SiteProfile.TargetKeywords...)
	s.SiteProfile.CompetitorDomains = append([]string(nil), s.SiteProfile.CompetitorDomains...)
	s.SiteProfile.Locations = append([]string(nil), s.SiteProfile.Locations...)
	s.Integrations.GA4.ExpectedKeyEvents = append([]string(nil), s.Integrations.GA4.ExpectedKeyEvents...)
	s.Integrations.GA4.ExpectedUTMRules = append([]string(nil), s.Integrations.GA4.ExpectedUTMRules...)
	s.Integrations.Bing.SitemapURLs = append([]string(nil), s.Integrations.Bing.SitemapURLs...)
	s.Integrations.GBP.AccountIDs = append([]string(nil), s.Integrations.GBP.AccountIDs...)
	s.Integrations.GBP.LocationIDs = append([]string(nil), s.Integrations.GBP.LocationIDs...)
	s.Integrations.GBP.RequiredCategories = append([]string(nil), s.Integrations.GBP.RequiredCategories...)
	s.Integrations.Social.OfficialHandles = append([]string(nil), s.Integrations.Social.OfficialHandles...)
	if s.SiteProfile.Country == "" {
		s.SiteProfile.Country = "IN"
	}
	if s.SiteProfile.Locale == "" {
		s.SiteProfile.Locale = "en-IN"
	}
	if s.Integrations.GSC.DateRange == "" {
		s.Integrations.GSC.DateRange = "last_28_days"
	}
	if s.Integrations.GSC.Device == "" {
		s.Integrations.GSC.Device = "mobile"
	}
	if s.Integrations.GA4.DateRange == "" {
		s.Integrations.GA4.DateRange = "last_28_days"
	}
	if s.Integrations.SERP.Device == "" {
		s.Integrations.SERP.Device = "mobile"
	}
	if s.Integrations.SERP.Language == "" {
		s.Integrations.SERP.Language = "en"
	}
	return s
}

// DefaultSkipLinkHosts are platforms known to block automated requests.
var DefaultSkipLinkHosts = []string{
	"linkedin.com",
	"www.linkedin.com",
	"twitter.com",
	"www.twitter.com",
	"x.com",
	"www.x.com",
	"instagram.com",
	"www.instagram.com",
	"facebook.com",
	"www.facebook.com",
	"tiktok.com",
	"www.tiktok.com",
}
