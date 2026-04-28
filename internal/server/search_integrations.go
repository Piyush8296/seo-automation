package server

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"
)

const (
	searchProviderGSC  = "gsc"
	searchProviderBing = "bing"
)

type SearchIntegrationProviderStatus struct {
	ID                  string   `json:"id"`
	Name                string   `json:"name"`
	Configured          bool     `json:"configured"`
	Connected           bool     `json:"connected"`
	Mode                string   `json:"mode"`
	OAuthScope          string   `json:"oauth_scope"`
	DocsURL             string   `json:"docs_url"`
	VerifyFirstMessage  string   `json:"verify_first_message"`
	ConnectionMessage   string   `json:"connection_message"`
	ConfiguredProperty  string   `json:"configured_property,omitempty"`
	ConfiguredSite      string   `json:"configured_site,omitempty"`
	ConfiguredSitemaps  []string `json:"configured_sitemaps,omitempty"`
	ConfiguredCountry   string   `json:"configured_country,omitempty"`
	ConfiguredDevice    string   `json:"configured_device,omitempty"`
	ConfiguredDateRange string   `json:"configured_date_range,omitempty"`
}

type SearchIntegrationSummary struct {
	TotalChecks          int  `json:"total_checks"`
	GSCChecks            int  `json:"gsc_checks"`
	BingChecks           int  `json:"bing_checks"`
	ManualEvidenceChecks int  `json:"manual_evidence_checks"`
	WriteCapableChecks   int  `json:"write_capable_checks"`
	ReportReady          bool `json:"report_ready"`
}

type SearchIntegrationCheck struct {
	ID             string   `json:"id"`
	Name           string   `json:"name"`
	Category       string   `json:"category"`
	Provider       string   `json:"provider"`
	PrimaryAPI     string   `json:"primary_api"`
	Automation     string   `json:"automation"`
	OperationTypes []string `json:"operation_types"`
	NeedsEvidence  bool     `json:"needs_evidence"`
	Status         string   `json:"status"`
	Notes          string   `json:"notes"`
}

type SearchIntegrationReport struct {
	GeneratedAt time.Time                `json:"generated_at"`
	Mode        string                   `json:"mode"`
	Message     string                   `json:"message"`
	Items       []SearchIntegrationCheck `json:"items"`
}

type SearchIntegrationWorkspaceResponse struct {
	Summary             SearchIntegrationSummary          `json:"summary"`
	Providers           []SearchIntegrationProviderStatus `json:"providers"`
	Checks              []SearchIntegrationCheck          `json:"checks"`
	Report              *SearchIntegrationReport          `json:"report,omitempty"`
	Actions             []SearchIntegrationActionSpec     `json:"actions"`
	VerificationMessage string                            `json:"verification_message"`
}

type SearchIntegrationActionSpec struct {
	Type        string   `json:"type"`
	Label       string   `json:"label"`
	Description string   `json:"description"`
	Providers   []string `json:"providers"`
	Fields      []string `json:"fields"`
}

type SearchOAuthRequest struct {
	Provider string `json:"provider"`
}

type SearchOAuthResponse struct {
	Status           string    `json:"status"`
	Provider         string    `json:"provider"`
	Message          string    `json:"message"`
	OAuthScope       string    `json:"oauth_scope"`
	AuthorizationURL string    `json:"authorization_url"`
	ReceivedAt       time.Time `json:"received_at"`
}

type SearchIntegrationActionRequest struct {
	Type       string         `json:"type"`
	Provider   string         `json:"provider"`
	URL        string         `json:"url"`
	SitemapURL string         `json:"sitemap_url"`
	Payload    map[string]any `json:"payload"`
}

type SearchIntegrationActionResponse struct {
	Status     string                         `json:"status"`
	Message    string                         `json:"message"`
	ReceivedAt time.Time                      `json:"received_at"`
	Action     SearchIntegrationActionRequest `json:"action"`
}

func (h *Handlers) getSearchIntegrations(w http.ResponseWriter, r *http.Request) {
	cfg := h.manager.GetSettings()
	checks := searchIntegrationChecks()
	providers := searchProviderStatuses(cfg)
	summary := summarizeSearchIntegrations(checks, providers)

	var report *SearchIntegrationReport
	if summary.ReportReady {
		reportItems := append([]SearchIntegrationCheck(nil), checks...)
		reportItems = annotateSearchCheckStatuses(reportItems, providers)
		report = &SearchIntegrationReport{
			GeneratedAt: time.Now(),
			Mode:        "placeholder",
			Message:     "OAuth is marked connected, so the pack is ready to execute live API calls once the real OAuth callback and token store are wired.",
			Items:       reportItems,
		}
	}

	writeJSON(w, http.StatusOK, SearchIntegrationWorkspaceResponse{
		Summary:             summary,
		Providers:           providers,
		Checks:              annotateSearchCheckStatuses(checks, providers),
		Report:              report,
		Actions:             searchIntegrationActionSpecs(),
		VerificationMessage: "Add a verified GSC property or Bing site, then connect OAuth before reports and POST operations are enabled.",
	})
}

func (h *Handlers) connectSearchOAuth(w http.ResponseWriter, r *http.Request) {
	var req SearchOAuthRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	provider := normalizeSearchProvider(req.Provider)
	if provider == "" {
		writeErr(w, http.StatusBadRequest, "provider must be gsc or bing")
		return
	}

	cfg := h.manager.GetSettings()
	switch provider {
	case searchProviderGSC:
		if strings.TrimSpace(cfg.Integrations.GSC.PropertyURL) == "" {
			writeErr(w, http.StatusPreconditionRequired, "configure and verify a GSC property URL before connecting OAuth")
			return
		}
		cfg.FeatureFlags.GSC = true
		cfg.Integrations.GSC.OAuthConnected = true
	case searchProviderBing:
		if strings.TrimSpace(cfg.Integrations.Bing.SiteURL) == "" {
			writeErr(w, http.StatusPreconditionRequired, "configure and verify a Bing site URL before connecting OAuth")
			return
		}
		cfg.FeatureFlags.BingWebmaster = true
		cfg.Integrations.Bing.OAuthConnected = true
	}
	h.manager.UpdateSettings(cfg)

	scope := searchOAuthScope(provider)
	writeJSON(w, http.StatusAccepted, SearchOAuthResponse{
		Status:           "oauth_placeholder_connected",
		Provider:         provider,
		Message:          "OAuth callback and token storage are placeholders in this build. The provider is marked connected so the report UI can be exercised without storing secrets.",
		OAuthScope:       scope,
		AuthorizationURL: "placeholder://oauth/" + provider,
		ReceivedAt:       time.Now(),
	})
}

func (h *Handlers) submitSearchIntegrationAction(w http.ResponseWriter, r *http.Request) {
	var req SearchIntegrationActionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	req.Type = strings.TrimSpace(strings.ToLower(req.Type))
	req.Provider = normalizeSearchProvider(req.Provider)
	req.URL = strings.TrimSpace(req.URL)
	req.SitemapURL = strings.TrimSpace(req.SitemapURL)
	if req.Payload == nil {
		req.Payload = map[string]any{}
	}
	if req.Provider == "" {
		writeErr(w, http.StatusBadRequest, "provider must be gsc or bing")
		return
	}
	if req.Type != "inspect_url" && req.Type != "submit_sitemap" {
		writeErr(w, http.StatusBadRequest, "type must be inspect_url or submit_sitemap")
		return
	}
	if req.Type == "inspect_url" && req.Provider != searchProviderGSC {
		writeErr(w, http.StatusBadRequest, "inspect_url is only available for GSC")
		return
	}
	if req.Type == "inspect_url" && req.URL == "" {
		writeErr(w, http.StatusBadRequest, "url is required for inspect_url actions")
		return
	}
	if req.Type == "submit_sitemap" && req.SitemapURL == "" {
		writeErr(w, http.StatusBadRequest, "sitemap_url is required for submit_sitemap actions")
		return
	}

	cfg := h.manager.GetSettings()
	if !searchProviderConnected(cfg, req.Provider) {
		writeErr(w, http.StatusPreconditionRequired, "connect OAuth and verify the provider first")
		return
	}

	writeJSON(w, http.StatusAccepted, SearchIntegrationActionResponse{
		Status:     "placeholder_accepted",
		Message:    "Search integration API is connected in placeholder mode. This action was validated but no live API call was made.",
		ReceivedAt: time.Now(),
		Action:     req,
	})
}

func searchProviderStatuses(cfg AppSettings) []SearchIntegrationProviderStatus {
	gsc := cfg.Integrations.GSC
	bing := cfg.Integrations.Bing
	gscConfigured := strings.TrimSpace(gsc.PropertyURL) != ""
	bingConfigured := strings.TrimSpace(bing.SiteURL) != ""

	return []SearchIntegrationProviderStatus{
		{
			ID:                  searchProviderGSC,
			Name:                "Google Search Console",
			Configured:          gscConfigured,
			Connected:           gscConfigured && gsc.OAuthConnected,
			Mode:                "oauth_placeholder",
			OAuthScope:          searchOAuthScope(searchProviderGSC),
			DocsURL:             "https://developers.google.com/webmaster-tools/about",
			VerifyFirstMessage:  "Verify the site in Search Console and save the exact property URL first.",
			ConnectionMessage:   searchConnectionMessage("GSC", gscConfigured, gsc.OAuthConnected),
			ConfiguredProperty:  gsc.PropertyURL,
			ConfiguredCountry:   gsc.Country,
			ConfiguredDevice:    gsc.Device,
			ConfiguredDateRange: gsc.DateRange,
		},
		{
			ID:                 searchProviderBing,
			Name:               "Bing Webmaster Tools",
			Configured:         bingConfigured,
			Connected:          bingConfigured && bing.OAuthConnected,
			Mode:               "oauth_placeholder",
			OAuthScope:         searchOAuthScope(searchProviderBing),
			DocsURL:            "https://learn.microsoft.com/en-us/bingwebmaster/",
			VerifyFirstMessage: "Add and verify the site in Bing Webmaster Tools, then save the site URL here.",
			ConnectionMessage:  searchConnectionMessage("Bing", bingConfigured, bing.OAuthConnected),
			ConfiguredSite:     bing.SiteURL,
			ConfiguredSitemaps: append([]string(nil), bing.SitemapURLs...),
		},
	}
}

func searchConnectionMessage(name string, configured, connected bool) string {
	switch {
	case connected:
		return name + " OAuth is connected in placeholder mode; reports are ready to render."
	case configured:
		return name + " property settings are saved. Connect OAuth to unlock report and POST actions."
	default:
		return name + " is not configured. Verify the property/site first, then connect OAuth."
	}
}

func searchIntegrationChecks() []SearchIntegrationCheck {
	return []SearchIntegrationCheck{
		searchCheck("ANALYTICS-003", "Google Search Console property verified", "Analytics & Monitoring", searchProviderGSC, "Sites API", "API verification", []string{}, false, "Requires a verified Search Console property URL."),
		searchCheck("ANALYTICS-005", "Bing Webmaster Tools set up", "Analytics & Monitoring", searchProviderBing, "Bing Webmaster API", "API verification", []string{}, false, "Requires a verified Bing Webmaster site."),
		searchCheck("SITEMAP-019", "Sitemap submitted to Bing Webmaster Tools", "Technical SEO", searchProviderBing, "Bing Webmaster API", "POST sitemap", []string{"submit_sitemap"}, false, "Submit and verify expected sitemap URLs."),
		searchCheck("CRAWL-010", "Crawl stats report checked in GSC", "Technical SEO", searchProviderGSC, "Search Console UI/export", "Manual/API evidence", []string{}, true, "No public Crawl Stats API; keep UI/export or log evidence."),
		searchCheck("CWV-004", "Core Web Vitals pass in Google Search Console", "Speed & Performance", searchProviderGSC, "GSC plus CrUX/PageSpeed", "API evidence", []string{}, true, "Wire CrUX or PageSpeed API for live field metrics."),
		searchCheck("MOBILE-006", "Mobile-first indexing confirmed in GSC", "Mobile SEO", searchProviderGSC, "Search Console evidence", "Manual/API evidence", []string{"inspect_url"}, true, "URL Inspection can help, but some property-level evidence stays manual."),
		searchCheck("KW-006", "Missing keyword opportunities identified (GSC)", "Content SEO", searchProviderGSC, "Search Analytics API", "API analysis", []string{}, false, "Use query/page dimensions to identify high-impression low-click gaps."),
		searchCheck("CONTENT-006", "Content is updated/freshness maintained", "Content SEO", searchProviderGSC, "Search Analytics API", "API trend analysis", []string{}, false, "Visibility declines can flag stale URLs; CMS dates improve confidence."),
		searchCheck("CONTENT-015", "Outdated content identified and updated", "Content SEO", searchProviderGSC, "Search Analytics API", "API workflow", []string{}, false, "Pair declining performance with an update workflow."),
	}
}

func searchCheck(id, name, category, provider, api, automation string, ops []string, needsEvidence bool, notes string) SearchIntegrationCheck {
	return SearchIntegrationCheck{
		ID:             id,
		Name:           name,
		Category:       category,
		Provider:       provider,
		PrimaryAPI:     api,
		Automation:     automation,
		OperationTypes: append([]string(nil), ops...),
		NeedsEvidence:  needsEvidence,
		Notes:          notes,
	}
}

func summarizeSearchIntegrations(checks []SearchIntegrationCheck, providers []SearchIntegrationProviderStatus) SearchIntegrationSummary {
	s := SearchIntegrationSummary{TotalChecks: len(checks)}
	for _, check := range checks {
		switch check.Provider {
		case searchProviderGSC:
			s.GSCChecks++
		case searchProviderBing:
			s.BingChecks++
		}
		if check.NeedsEvidence {
			s.ManualEvidenceChecks++
		}
		if len(check.OperationTypes) > 0 {
			s.WriteCapableChecks++
		}
	}
	for _, provider := range providers {
		if provider.Connected {
			s.ReportReady = true
			break
		}
	}
	return s
}

func annotateSearchCheckStatuses(checks []SearchIntegrationCheck, providers []SearchIntegrationProviderStatus) []SearchIntegrationCheck {
	connected := map[string]bool{}
	for _, provider := range providers {
		connected[provider.ID] = provider.Connected
	}
	out := make([]SearchIntegrationCheck, 0, len(checks))
	for _, check := range checks {
		next := check
		if !connected[check.Provider] {
			next.Status = "verify_first"
		} else if check.NeedsEvidence {
			next.Status = "needs_evidence"
		} else {
			next.Status = "ready"
		}
		out = append(out, next)
	}
	return out
}

func searchIntegrationActionSpecs() []SearchIntegrationActionSpec {
	return []SearchIntegrationActionSpec{
		{
			Type:        "inspect_url",
			Label:       "Inspect URL in GSC",
			Description: "Placeholder for Search Console URL Inspection API calls.",
			Providers:   []string{searchProviderGSC},
			Fields:      []string{"provider", "url"},
		},
		{
			Type:        "submit_sitemap",
			Label:       "Submit sitemap",
			Description: "Placeholder for GSC or Bing sitemap submission.",
			Providers:   []string{searchProviderGSC, searchProviderBing},
			Fields:      []string{"provider", "sitemap_url"},
		},
	}
}

func normalizeSearchProvider(provider string) string {
	switch strings.TrimSpace(strings.ToLower(provider)) {
	case "gsc", "google", "google_search_console":
		return searchProviderGSC
	case "bing", "bing_webmaster", "bing_webmaster_tools":
		return searchProviderBing
	default:
		return ""
	}
}

func searchOAuthScope(provider string) string {
	switch provider {
	case searchProviderGSC:
		return "https://www.googleapis.com/auth/webmasters.readonly https://www.googleapis.com/auth/webmasters"
	case searchProviderBing:
		return "Webmaster.read Webmaster.manage"
	default:
		return ""
	}
}

func searchProviderConnected(cfg AppSettings, provider string) bool {
	switch provider {
	case searchProviderGSC:
		return strings.TrimSpace(cfg.Integrations.GSC.PropertyURL) != "" && cfg.Integrations.GSC.OAuthConnected
	case searchProviderBing:
		return strings.TrimSpace(cfg.Integrations.Bing.SiteURL) != "" && cfg.Integrations.Bing.OAuthConnected
	default:
		return false
	}
}
