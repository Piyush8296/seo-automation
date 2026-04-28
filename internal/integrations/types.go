package integrations

// InputSource describes where a required input should come from.
type InputSource string

const (
	InputSourceUI              InputSource = "ui"
	InputSourceOAuth           InputSource = "oauth"
	InputSourceSecretStore     InputSource = "secret_store"
	InputSourceEnvironment     InputSource = "environment"
	InputSourceProviderProject InputSource = "provider_project"
	InputSourceManual          InputSource = "manual_workflow"
)

const (
	ProviderGSC        = "gsc"
	ProviderGA4        = "ga4"
	ProviderBing       = "bing_webmaster"
	ProviderGBP        = "google_business_profile"
	ProviderSEOVendor  = "seo_vendor"
	ProviderSERP       = "serp_provider"
	ProviderPlagiarism = "plagiarism_vendor"
	ProviderNews       = "news_monitor"
	ProviderSocial     = "social_platforms"
	ProviderReviews    = "review_vendor"
	ProviderListings   = "listings_vendor"
	ProviderRUM        = "rum_vendor"
	ProviderManual     = "manual_workflow"
	ProviderAI         = "ai_content_evaluator"
)

// RequiredInput is a non-secret description of data needed to run a provider or check.
type RequiredInput struct {
	Key         string      `json:"key"`
	Label       string      `json:"label"`
	Description string      `json:"description"`
	Source      InputSource `json:"source"`
	Required    bool        `json:"required"`
	Secret      bool        `json:"secret"`
	Examples    []string    `json:"examples,omitempty"`
}

// ProviderDescriptor tells the UI what must be configured before a check pack can run.
type ProviderDescriptor struct {
	ID          string          `json:"id"`
	Name        string          `json:"name"`
	FeatureFlag string          `json:"feature_flag"`
	CostModel   string          `json:"cost_model"`
	AuthModel   string          `json:"auth_model"`
	DocsURL     string          `json:"docs_url,omitempty"`
	Inputs      []RequiredInput `json:"inputs"`
	CheckIDs    []string        `json:"check_ids"`
}

// CheckCapability maps a master-checklist item to the integrations and inputs needed.
type CheckCapability struct {
	ID                string   `json:"id"`
	Name              string   `json:"name"`
	Category          string   `json:"category"`
	PrimaryProviderID string   `json:"primary_provider_id"`
	ProviderIDs       []string `json:"provider_ids"`
	RequiredInputs    []string `json:"required_inputs"`
	FeatureFlag       string   `json:"feature_flag"`
	CostModel         string   `json:"cost_model"`
	RunMode           string   `json:"run_mode"`
	NeedsURLSource    bool     `json:"needs_url_source"`
	Notes             string   `json:"notes,omitempty"`
}

// Catalog is the public contract consumed by Settings and future execution workers.
type Catalog struct {
	TotalChecks  int                  `json:"total_checks"`
	GlobalInputs []RequiredInput      `json:"global_inputs"`
	Providers    []ProviderDescriptor `json:"providers"`
	Checks       []CheckCapability    `json:"checks"`
	FeatureFlags []string             `json:"feature_flags"`
}

func input(key, label, description string, source InputSource, required, secret bool, examples ...string) RequiredInput {
	return RequiredInput{
		Key:         key,
		Label:       label,
		Description: description,
		Source:      source,
		Required:    required,
		Secret:      secret,
		Examples:    examples,
	}
}

func capability(id, name, category, primaryProviderID string, providerIDs, requiredInputs []string, featureFlag, costModel, runMode string, needsURLSource bool, notes string) CheckCapability {
	return CheckCapability{
		ID:                id,
		Name:              name,
		Category:          category,
		PrimaryProviderID: primaryProviderID,
		ProviderIDs:       providerIDs,
		RequiredInputs:    requiredInputs,
		FeatureFlag:       featureFlag,
		CostModel:         costModel,
		RunMode:           runMode,
		NeedsURLSource:    needsURLSource,
		Notes:             notes,
	}
}
