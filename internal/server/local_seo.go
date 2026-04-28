package server

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"
)

type LocalSEOCheck struct {
	ID             string   `json:"id"`
	Name           string   `json:"name"`
	Priority       string   `json:"priority"`
	Channel        string   `json:"channel"`
	PrimarySource  string   `json:"primary_source"`
	Automation     string   `json:"automation"`
	OperationTypes []string `json:"operation_types"`
	Notes          string   `json:"notes"`
}

type LocalSEOStatus struct {
	Provider            string   `json:"provider"`
	Mode                string   `json:"mode"`
	Configured          bool     `json:"configured"`
	OAuthScope          string   `json:"oauth_scope"`
	DocsURL             string   `json:"docs_url"`
	ConfiguredAccounts  []string `json:"configured_accounts"`
	ConfiguredLocations []string `json:"configured_locations"`
	Message             string   `json:"message"`
}

type LocalSEOSummary struct {
	TotalChecks          int `json:"total_checks"`
	GBPDirectChecks      int `json:"gbp_direct_checks"`
	WebsiteChecks        int `json:"website_checks"`
	VendorWorkflowChecks int `json:"vendor_workflow_checks"`
	ManualOnlyChecks     int `json:"manual_only_checks"`
}

type LocalSEOResponse struct {
	Status  LocalSEOStatus  `json:"status"`
	Summary LocalSEOSummary `json:"summary"`
	Checks  []LocalSEOCheck `json:"checks"`
	Actions []GBPActionSpec `json:"actions"`
}

type GBPActionSpec struct {
	Type        string   `json:"type"`
	Label       string   `json:"label"`
	Description string   `json:"description"`
	Fields      []string `json:"fields"`
}

type GBPActionRequest struct {
	Type       string         `json:"type"`
	LocationID string         `json:"location_id"`
	Title      string         `json:"title"`
	Summary    string         `json:"summary"`
	Payload    map[string]any `json:"payload"`
}

type GBPActionResponse struct {
	Status     string           `json:"status"`
	Message    string           `json:"message"`
	ReceivedAt time.Time        `json:"received_at"`
	Action     GBPActionRequest `json:"action"`
}

func (h *Handlers) getLocalSEO(w http.ResponseWriter, r *http.Request) {
	cfg := h.manager.GetSettings()
	checks := localSEOChecks()
	writeJSON(w, http.StatusOK, LocalSEOResponse{
		Status:  localSEOStatus(cfg),
		Summary: summarizeLocalSEO(checks),
		Checks:  checks,
		Actions: gbpActionSpecs(),
	})
}

func (h *Handlers) submitGBPAction(w http.ResponseWriter, r *http.Request) {
	var req GBPActionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	req.Type = strings.TrimSpace(strings.ToLower(req.Type))
	req.LocationID = strings.TrimSpace(req.LocationID)
	if req.Type != "update" && req.Type != "post" {
		writeErr(w, http.StatusBadRequest, "type must be update or post")
		return
	}
	if req.LocationID == "" {
		writeErr(w, http.StatusBadRequest, "location_id is required")
		return
	}
	if req.Type == "post" && strings.TrimSpace(req.Summary) == "" {
		writeErr(w, http.StatusBadRequest, "summary is required for post actions")
		return
	}
	if req.Payload == nil {
		req.Payload = map[string]any{}
	}

	writeJSON(w, http.StatusAccepted, GBPActionResponse{
		Status:     "placeholder_accepted",
		Message:    "GBP API integration is not connected yet. This action was validated and captured as a placeholder only.",
		ReceivedAt: time.Now(),
		Action:     req,
	})
}

func localSEOStatus(cfg AppSettings) LocalSEOStatus {
	gbp := cfg.Integrations.GBP
	configured := cfg.FeatureFlags.GoogleBusinessProfile && len(gbp.AccountIDs) > 0 && len(gbp.LocationIDs) > 0
	msg := "GBP API placeholder is ready; connect OAuth, approved API quota, account IDs, and location IDs to execute live operations."
	if configured {
		msg = "GBP settings are present, but this build still routes write operations to the placeholder handler."
	}
	return LocalSEOStatus{
		Provider:            "Google Business Profile",
		Mode:                "placeholder",
		Configured:          configured,
		OAuthScope:          "https://www.googleapis.com/auth/business.manage",
		DocsURL:             "https://developers.google.com/my-business",
		ConfiguredAccounts:  append([]string(nil), gbp.AccountIDs...),
		ConfiguredLocations: append([]string(nil), gbp.LocationIDs...),
		Message:             msg,
	}
}

func localSEOChecks() []LocalSEOCheck {
	return []LocalSEOCheck{
		localCheck("LOCAL-001", "Google Business Profile for each city branch", "Critical", "GBP", "Google Business Profile API", "API audit/create placeholder", []string{"update"}, "Verification can still require Google-controlled manual steps."),
		localCheck("LOCAL-002", "GBP name/address/phone (NAP) consistent", "Critical", "GBP", "Google Business Profile API", "API audit/update placeholder", []string{"update"}, "Compare live GBP location fields with branch master data."),
		localCheck("LOCAL-003", "GBP categories correctly set", "High", "GBP", "Google Business Profile API", "API audit/update placeholder", []string{"update"}, "Use GBP category IDs once API access is connected."),
		localCheck("LOCAL-004", "GBP photos added (interior, exterior, team)", "High", "GBP", "Google Business Profile API", "API audit/upload placeholder", []string{"update"}, "API can manage media metadata and uploads; visual QA remains branch-owned."),
		localCheck("LOCAL-005", "GBP posts published regularly", "Medium", "GBP", "Google Business Profile API", "API post placeholder", []string{"post"}, "Publish local posts after OAuth and quota are enabled."),
		localCheck("LOCAL-006", "GBP Q&A section populated", "Medium", "Manual", "Manual workflow", "Manual only", []string{}, "GBP Q&A API was discontinued; track website FAQ coverage instead."),
		localCheck("LOCAL-007", "Local schema (LocalBusiness) on location pages", "Critical", "Website", "Crawler / CMS", "Crawler audit", []string{}, "Validate LocalBusiness schema on branch pages."),
		localCheck("LOCAL-008", "City/location pages unique and optimized", "Critical", "Website", "Crawler / CMS", "Content audit", []string{}, "Use page crawl plus content similarity checks."),
		localCheck("LOCAL-009", "Consistent NAP in local directories", "High", "Listings", "Listings vendor", "Vendor/manual workflow", []string{}, "Needs citation/listings data beyond GBP."),
		localCheck("LOCAL-010", "Local keyword targeting ('used cars in Bangalore')", "Critical", "Website", "GSC / SEO vendor / CMS", "Content and keyword audit", []string{}, "GBP can provide keyword impressions; page targeting needs content checks."),
		localCheck("LOCAL-011", "Local reviews strategy in place", "High", "GBP", "Google Business Profile API", "API monitor plus workflow", []string{}, "API can monitor and reply to reviews; review generation is a CRM/workflow task."),
		localCheck("LOCAL-012", "Local pack (Map Pack) visibility monitored", "Critical", "Monitoring", "SERP provider", "Vendor monitoring", []string{}, "Official GBP APIs do not expose exact map-pack rankings."),
		localCheck("LOCAL-013", "Hyperlocal area landing pages created", "High", "Website", "CMS / inventory", "CMS workflow", []string{}, "Use owned city/neighborhood inventory and CMS state."),
		localCheck("LOCAL-014", "Local link building from city-based sites", "High", "Off-page", "SEO vendor / outreach workflow", "Vendor/manual workflow", []string{}, "Track prospects, outreach, and acquired links outside GBP."),
		localCheck("LOCAL-015", "Embedding Google Map on location pages", "Medium", "Website", "Crawler / CMS", "Crawler audit", []string{}, "Detect map embeds and match them to the branch location."),
	}
}

func localCheck(id, name, priority, channel, source, automation string, ops []string, notes string) LocalSEOCheck {
	return LocalSEOCheck{
		ID:             id,
		Name:           name,
		Priority:       priority,
		Channel:        channel,
		PrimarySource:  source,
		Automation:     automation,
		OperationTypes: append([]string(nil), ops...),
		Notes:          notes,
	}
}

func summarizeLocalSEO(checks []LocalSEOCheck) LocalSEOSummary {
	s := LocalSEOSummary{TotalChecks: len(checks)}
	for _, check := range checks {
		switch check.Channel {
		case "GBP":
			s.GBPDirectChecks++
		case "Website":
			s.WebsiteChecks++
		case "Manual":
			s.ManualOnlyChecks++
		default:
			s.VendorWorkflowChecks++
		}
	}
	return s
}

func gbpActionSpecs() []GBPActionSpec {
	return []GBPActionSpec{
		{
			Type:        "update",
			Label:       "Update GBP location",
			Description: "Placeholder for NAP, category, hours, services, and media updates.",
			Fields:      []string{"location_id", "payload"},
		},
		{
			Type:        "post",
			Label:       "Publish GBP post",
			Description: "Placeholder for local post publishing and recurring post workflows.",
			Fields:      []string{"location_id", "title", "summary", "payload"},
		},
	}
}
