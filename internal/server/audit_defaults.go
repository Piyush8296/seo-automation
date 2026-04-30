package server

import (
	"fmt"
	"strings"
	"time"

	"github.com/cars24/seo-automation/internal/models"
)

const (
	maxAuditConcurrency             = 20
	defaultValidateExternalLinks    = true
	defaultDiscoverResources        = true
	defaultEnableCrawlerEvidence    = true
	defaultEnableRenderedSEO        = true
	defaultFollowNofollowLinks      = false
	defaultExpandNoindexPages       = true
	defaultExpandCanonicalizedPages = true
)

// AuditValidationError marks bad client-supplied audit config so handlers can
// return a 400 while storage/runtime failures still surface as 500s.
type AuditValidationError struct {
	Message string
}

func (e *AuditValidationError) Error() string {
	return e.Message
}

type IntOption struct {
	Value int    `json:"value"`
	Label string `json:"label"`
}

type StringOption struct {
	Value string `json:"value"`
	Label string `json:"label"`
}

type NumberControl struct {
	Min  int `json:"min"`
	Max  int `json:"max,omitempty"`
	Step int `json:"step"`
}

type AuditControlOptions struct {
	MaxPages                   NumberControl  `json:"max_pages"`
	Concurrency                NumberControl  `json:"concurrency"`
	MaxPagePresets             []IntOption    `json:"max_page_presets"`
	MaxDepthOptions            []IntOption    `json:"max_depth_options"`
	TimeoutOptions             []StringOption `json:"timeout_options"`
	PlatformOptions            []StringOption `json:"platform_options"`
	SitemapModeOptions         []StringOption `json:"sitemap_mode_options"`
	RenderedSampleLimitOptions []IntOption    `json:"rendered_sample_limit_options"`
	RenderedTimeoutOptions     []StringOption `json:"rendered_timeout_options"`
}

type AuditDefaultsResponse struct {
	DefaultConfig StartAuditRequest   `json:"default_config"`
	Controls      AuditControlOptions `json:"controls"`
}

func (m *Manager) GetAuditDefaults() AuditDefaultsResponse {
	return AuditDefaultsResponse{
		DefaultConfig: DefaultStartAuditRequest(),
		Controls:      DefaultAuditControlOptions(),
	}
}

func DefaultStartAuditRequest() StartAuditRequest {
	return StartAuditRequest{
		URL:                      "",
		Scope:                    string(defaultScope),
		ScopePrefix:              "",
		SitemapURL:               "",
		SitemapMode:              string(defaultSitemapMode),
		MaxDepth:                 -1,
		MaxPages:                 0,
		Concurrency:              defaultConcur,
		Timeout:                  defaultTimeout,
		Platform:                 "",
		UserAgent:                defaultUserAgent,
		MobileUserAgent:          defaultMobileUA,
		RespectRobots:            boolPtr(true),
		MaxRedirects:             defaultMaxRedirects,
		MaxPageSizeKB:            defaultMaxPageSizeKB,
		MaxURLLength:             0,
		MaxQueryParams:           0,
		MaxLinksPerPage:          0,
		FollowNofollowLinks:      defaultFollowNofollowLinks,
		ExpandNoindexPages:       boolPtr(defaultExpandNoindexPages),
		ExpandCanonicalizedPages: boolPtr(defaultExpandCanonicalizedPages),
		OutputDir:                "",
		ValidateExternalLinks:    boolPtr(defaultValidateExternalLinks),
		DiscoverResources:        boolPtr(defaultDiscoverResources),
		EnableCrawlerEvidence:    boolPtr(defaultEnableCrawlerEvidence),
		EnableRenderedSEO:        boolPtr(defaultEnableRenderedSEO),
		RenderedSampleLimit:      defaultRenderedSampleLimit,
		RenderedTimeout:          defaultRenderedTimeout.String(),
	}
}

func DefaultAuditControlOptions() AuditControlOptions {
	return AuditControlOptions{
		MaxPages:    NumberControl{Min: 0, Step: 50},
		Concurrency: NumberControl{Min: 1, Max: maxAuditConcurrency, Step: 1},
		MaxPagePresets: []IntOption{
			{Value: 0, Label: "Unlimited"},
			{Value: 100, Label: "100"},
			{Value: 500, Label: "500"},
			{Value: 1000, Label: "1k"},
			{Value: 5000, Label: "5k"},
		},
		MaxDepthOptions: []IntOption{
			{Value: -1, Label: "Unlimited"},
			{Value: 1, Label: "1 level"},
			{Value: 2, Label: "2 levels"},
			{Value: 3, Label: "3 levels"},
			{Value: 5, Label: "5 levels"},
		},
		TimeoutOptions: []StringOption{
			{Value: "10s", Label: "10 seconds"},
			{Value: defaultTimeout, Label: "30 seconds"},
			{Value: "1m", Label: "1 minute"},
			{Value: "2m", Label: "2 minutes"},
		},
		PlatformOptions: []StringOption{
			{Value: "", Label: "Both (bifurcated)"},
			{Value: string(models.PlatformDesktop), Label: "Desktop only"},
			{Value: string(models.PlatformMobile), Label: "Mobile focus"},
		},
		SitemapModeOptions: []StringOption{
			{Value: string(models.SitemapModeOff), Label: "Off (fastest)"},
			{Value: string(models.SitemapModeDiscover), Label: "Discover for coverage"},
			{Value: string(models.SitemapModeSeed), Label: "Seed crawl from sitemap"},
		},
		RenderedSampleLimitOptions: []IntOption{
			{Value: 3, Label: "3 pages"},
			{Value: defaultRenderedSampleLimit, Label: "5 pages"},
			{Value: 10, Label: "10 pages"},
			{Value: 20, Label: "20 pages"},
		},
		RenderedTimeoutOptions: []StringOption{
			{Value: "10s", Label: "10 seconds"},
			{Value: defaultRenderedTimeout.String(), Label: "20 seconds"},
			{Value: "30s", Label: "30 seconds"},
			{Value: "45s", Label: "45 seconds"},
		},
	}
}

func NormalizeStartAuditRequest(req StartAuditRequest) (StartAuditRequest, error) {
	defaults := DefaultStartAuditRequest()

	req.URL = strings.TrimSpace(req.URL)
	req.Scope = strings.ToLower(strings.TrimSpace(req.Scope))
	req.ScopePrefix = strings.TrimSpace(req.ScopePrefix)
	req.SitemapURL = strings.TrimSpace(req.SitemapURL)
	req.SitemapMode = strings.ToLower(strings.TrimSpace(req.SitemapMode))
	req.Platform = strings.ToLower(strings.TrimSpace(req.Platform))
	req.OutputDir = strings.TrimSpace(req.OutputDir)
	req.UserAgent = strings.TrimSpace(req.UserAgent)
	req.MobileUserAgent = strings.TrimSpace(req.MobileUserAgent)
	req.RenderedTimeout = strings.TrimSpace(req.RenderedTimeout)

	if req.URL == "" {
		return req, invalidAuditRequest("url is required")
	}
	if req.Scope == "" {
		req.Scope = defaults.Scope
	}
	if req.Scope != string(models.CrawlScopeHost) && req.Scope != string(models.CrawlScopeSubfolder) {
		return req, invalidAuditRequest("invalid scope %q", req.Scope)
	}
	if req.SitemapMode == "" {
		req.SitemapMode = defaults.SitemapMode
	}
	switch models.SitemapMode(req.SitemapMode) {
	case models.SitemapModeOff, models.SitemapModeDiscover, models.SitemapModeSeed:
	default:
		return req, invalidAuditRequest("invalid sitemap_mode %q", req.SitemapMode)
	}
	if req.Platform == "all" {
		req.Platform = ""
	}
	switch models.Platform(req.Platform) {
	case "", models.PlatformDesktop, models.PlatformMobile:
	default:
		return req, invalidAuditRequest("invalid platform %q", req.Platform)
	}
	if req.MaxDepth == 0 {
		req.MaxDepth = defaults.MaxDepth
	}
	if req.MaxDepth < -1 {
		return req, invalidAuditRequest("max_depth must be -1 or greater")
	}
	if req.MaxPages < 0 {
		return req, invalidAuditRequest("max_pages must be 0 or greater")
	}
	if req.Concurrency <= 0 {
		req.Concurrency = defaults.Concurrency
	}
	if req.Concurrency > maxAuditConcurrency {
		return req, invalidAuditRequest("concurrency must be %d or less", maxAuditConcurrency)
	}
	if req.Timeout == "" {
		req.Timeout = defaults.Timeout
	}
	if _, err := parsePositiveDuration("timeout", req.Timeout); err != nil {
		return req, err
	}
	if req.UserAgent == "" {
		req.UserAgent = defaults.UserAgent
	}
	if req.MobileUserAgent == "" {
		req.MobileUserAgent = defaults.MobileUserAgent
	}
	if req.RespectRobots == nil {
		req.RespectRobots = defaults.RespectRobots
	}
	if req.MaxRedirects <= 0 {
		req.MaxRedirects = defaults.MaxRedirects
	}
	if req.MaxPageSizeKB <= 0 {
		req.MaxPageSizeKB = defaults.MaxPageSizeKB
	}
	if req.MaxURLLength < 0 {
		return req, invalidAuditRequest("max_url_length must be 0 or greater")
	}
	if req.MaxQueryParams < 0 {
		return req, invalidAuditRequest("max_query_params must be 0 or greater")
	}
	if req.MaxLinksPerPage < 0 {
		return req, invalidAuditRequest("max_links_per_page must be 0 or greater")
	}
	if req.ExpandNoindexPages == nil {
		req.ExpandNoindexPages = defaults.ExpandNoindexPages
	}
	if req.ExpandCanonicalizedPages == nil {
		req.ExpandCanonicalizedPages = defaults.ExpandCanonicalizedPages
	}
	if req.ValidateExternalLinks == nil {
		req.ValidateExternalLinks = defaults.ValidateExternalLinks
	}
	if req.DiscoverResources == nil {
		req.DiscoverResources = defaults.DiscoverResources
	}
	if req.EnableCrawlerEvidence == nil {
		req.EnableCrawlerEvidence = defaults.EnableCrawlerEvidence
	}
	if req.EnableRenderedSEO == nil {
		req.EnableRenderedSEO = defaults.EnableRenderedSEO
	}
	if req.RenderedSampleLimit <= 0 {
		req.RenderedSampleLimit = defaults.RenderedSampleLimit
	}
	if req.RenderedSampleLimit > maxRenderedSampleLimit {
		return req, invalidAuditRequest("rendered_sample_limit must be %d or less", maxRenderedSampleLimit)
	}
	if req.RenderedTimeout == "" {
		req.RenderedTimeout = defaults.RenderedTimeout
	}
	if _, err := parsePositiveDuration("rendered_timeout", req.RenderedTimeout); err != nil {
		return req, err
	}
	req.ExpectedInventoryURLs = cleanStringSlice(req.ExpectedInventoryURLs)
	req.ExpectedParameterNames = cleanStringSlice(req.ExpectedParameterNames)
	req.AllowedImageCDNHosts = cleanStringSlice(req.AllowedImageCDNHosts)
	req.RequiredLiveText = cleanStringSlice(req.RequiredLiveText)

	return req, nil
}

func boolValue(v *bool, fallback bool) bool {
	if v == nil {
		return fallback
	}
	return *v
}

func boolPtr(v bool) *bool {
	return &v
}

func invalidAuditRequest(format string, args ...any) error {
	return &AuditValidationError{Message: fmt.Sprintf(format, args...)}
}

func parsePositiveDuration(field string, raw string) (time.Duration, error) {
	d, err := time.ParseDuration(raw)
	if err != nil {
		return 0, invalidAuditRequest("invalid %s %q", field, raw)
	}
	if d <= 0 {
		return 0, invalidAuditRequest("%s must be positive", field)
	}
	return d, nil
}

func cleanStringSlice(in []string) []string {
	if len(in) == 0 {
		return nil
	}
	out := make([]string, 0, len(in))
	for _, value := range in {
		value = strings.TrimSpace(value)
		if value != "" {
			out = append(out, value)
		}
	}
	return out
}
