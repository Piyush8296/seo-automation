package server

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/cars24/seo-automation/internal/checks"
	"github.com/cars24/seo-automation/internal/crawler"
	"github.com/cars24/seo-automation/internal/models"
	"github.com/cars24/seo-automation/internal/report"
)

const (
	defaultUserAgent           = "SEOAuditBot/1.0 (+https://github.com/cars24/seo-automation)"
	defaultMobileUA            = "Mozilla/5.0 (Linux; Android 10; K) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Mobile Safari/537.36"
	defaultTimeout             = "30s"
	defaultConcur              = 5
	defaultScope               = models.CrawlScopeHost
	defaultSitemapMode         = models.SitemapModeDiscover
	defaultMaxRedirects        = 10
	defaultMaxPageSizeKB int64 = 5 * 1024
)

// Manager orchestrates the lifecycle of audit runs.
type Manager struct {
	storage    *Storage
	hub        *Hub
	mu         sync.Mutex
	cancels    map[string]context.CancelFunc
	settingsMu sync.RWMutex
	settings   AppSettings
}

// NewManager creates a Manager backed by the given Storage and Hub.
func NewManager(storage *Storage, hub *Hub) *Manager {
	return &Manager{
		storage:  storage,
		hub:      hub,
		cancels:  make(map[string]context.CancelFunc),
		settings: DefaultAppSettings(),
	}
}

// GetSettings returns a copy of the current settings.
func (m *Manager) GetSettings() AppSettings {
	m.settingsMu.RLock()
	defer m.settingsMu.RUnlock()
	return m.settings.Normalize()
}

// UpdateSettings replaces the current settings atomically.
func (m *Manager) UpdateSettings(cfg AppSettings) {
	m.settingsMu.Lock()
	defer m.settingsMu.Unlock()
	m.settings = cfg.Normalize()
}

// StartAudit validates the request, persists an initial record, and launches the
// crawl goroutine. It returns immediately with the new AuditRecord.
func (m *Manager) StartAudit(req StartAuditRequest) (*AuditRecord, error) {
	if strings.TrimSpace(req.URL) == "" {
		return nil, fmt.Errorf("url is required")
	}

	// Apply defaults
	if req.Concurrency <= 0 {
		req.Concurrency = defaultConcur
	}
	if req.MaxDepth == 0 {
		req.MaxDepth = -1 // unlimited
	}
	if req.Timeout == "" {
		req.Timeout = defaultTimeout
	}
	if strings.TrimSpace(req.Scope) == "" {
		req.Scope = string(defaultScope)
	}
	if strings.TrimSpace(req.SitemapMode) == "" {
		req.SitemapMode = string(defaultSitemapMode)
	}
	if strings.TrimSpace(req.UserAgent) == "" {
		req.UserAgent = defaultUserAgent
	}
	if strings.TrimSpace(req.MobileUserAgent) == "" {
		req.MobileUserAgent = defaultMobileUA
	}
	if req.MaxRedirects <= 0 {
		req.MaxRedirects = defaultMaxRedirects
	}
	if req.MaxPageSizeKB <= 0 {
		req.MaxPageSizeKB = defaultMaxPageSizeKB
	}

	timeout, err := time.ParseDuration(req.Timeout)
	if err != nil {
		return nil, fmt.Errorf("invalid timeout %q: %w", req.Timeout, err)
	}

	id := newID()
	reportsDir := m.storage.AuditDir(id)
	if req.OutputDir != "" {
		reportsDir = req.OutputDir + "/" + id
	}

	record := &AuditRecord{
		ID:         id,
		URL:        req.URL,
		Config:     req,
		Status:     StatusRunning,
		CreatedAt:  time.Now(),
		ReportsDir: reportsDir,
	}
	if err := m.storage.Save(record); err != nil {
		return nil, fmt.Errorf("persist record: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	m.mu.Lock()
	m.cancels[id] = cancel
	m.mu.Unlock()

	go m.runAudit(ctx, id, req, timeout, reportsDir)
	return record, nil
}

// CancelAudit cancels a running audit. Returns false if the audit is not running.
func (m *Manager) CancelAudit(id string) bool {
	m.mu.Lock()
	cancel, ok := m.cancels[id]
	m.mu.Unlock()
	if ok {
		cancel()
	}
	return ok
}

// runAudit is the background goroutine that drives the full crawl → check → report pipeline.
func (m *Manager) runAudit(ctx context.Context, id string, req StartAuditRequest, timeout time.Duration, reportsDir string) {
	defer func() {
		m.mu.Lock()
		delete(m.cancels, id)
		m.mu.Unlock()
	}()

	platform := models.Platform(strings.ToLower(strings.TrimSpace(req.Platform)))
	if platform == "all" {
		platform = ""
	}
	scope := models.CrawlScope(strings.ToLower(strings.TrimSpace(req.Scope)))
	if scope == "" {
		scope = defaultScope
	}
	sitemapMode := models.SitemapMode(strings.ToLower(strings.TrimSpace(req.SitemapMode)))
	if sitemapMode == "" {
		sitemapMode = defaultSitemapMode
	}
	noMobile := platform == models.PlatformDesktop
	respectRobots := true
	if req.RespectRobots != nil {
		respectRobots = *req.RespectRobots
	}
	expandNoindex := true
	if req.ExpandNoindexPages != nil {
		expandNoindex = *req.ExpandNoindexPages
	}
	expandCanonical := true
	if req.ExpandCanonicalizedPages != nil {
		expandCanonical = *req.ExpandCanonicalizedPages
	}

	cfg := m.GetSettings()

	config := &models.CrawlConfig{
		SeedURL:                  req.URL,
		SitemapURL:               req.SitemapURL,
		Scope:                    scope,
		ScopePrefix:              req.ScopePrefix,
		SitemapMode:              sitemapMode,
		MaxDepth:                 req.MaxDepth,
		MaxPages:                 req.MaxPages,
		Concurrency:              req.Concurrency,
		Timeout:                  timeout,
		NoMobileCheck:            noMobile,
		UserAgent:                req.UserAgent,
		MobileUA:                 req.MobileUserAgent,
		RespectRobots:            respectRobots,
		MaxRedirects:             req.MaxRedirects,
		MaxPageSizeBytes:         req.MaxPageSizeKB * 1024,
		MaxURLLength:             req.MaxURLLength,
		MaxQueryParams:           req.MaxQueryParams,
		MaxLinksPerPage:          req.MaxLinksPerPage,
		FollowNofollowLinks:      req.FollowNofollowLinks,
		ExpandNoindexPages:       expandNoindex,
		ExpandCanonicalizedPages: expandCanonical,
		RenderMode:               "html-only",
		Platform:                 platform,
		ValidateExternalLinks:    req.ValidateExternalLinks,
		DiscoverResources:        req.DiscoverResources,
		SkipLinkHosts:            cfg.SkipLinkHosts,
		OnProgress: func(crawled int, currentURL string) {
			m.hub.Broadcast(id, ProgressEvent{
				Type:         "progress",
				PagesCrawled: crawled,
				CurrentURL:   currentURL,
			})
		},
	}

	// ── Crawl ──────────────────────────────────────────────────────────────
	c := crawler.NewCrawler(config)
	audit, err := c.Crawl(ctx)
	if err != nil {
		if ctx.Err() != nil {
			m.finishRecord(id, func(r *AuditRecord) {
				r.Status = StatusCancelled
			})
			m.hub.Broadcast(id, ProgressEvent{Type: "cancelled", Message: "audit cancelled by user"})
			return
		}
		m.finishRecord(id, func(r *AuditRecord) {
			r.Status = StatusFailed
			r.ErrMsg = err.Error()
		})
		m.hub.Broadcast(id, ProgressEvent{Type: "error", Message: err.Error()})
		return
	}

	// ── Checks ─────────────────────────────────────────────────────────────
	for _, page := range audit.Pages {
		page.CheckResults = checks.RunPageChecks(page)
	}
	audit.SiteChecks = checks.RunSiteWideChecks(audit.Pages)
	if req.EnableCrawlerEvidence == nil || *req.EnableCrawlerEvidence {
		audit.CrawlerEvidence = crawlerEvidenceForAudit(ctx, audit, req, timeout)
		audit.SiteChecks = append(audit.SiteChecks, crawlerEvidenceFindings(audit.CrawlerEvidence)...)
	}
	if req.EnableRenderedSEO == nil || *req.EnableRenderedSEO {
		m.hub.Broadcast(id, ProgressEvent{Type: "progress", PagesCrawled: audit.PagesCrawled, CurrentURL: "Running rendered SEO checks"})
		audit.RenderedSEO = renderedSEOForAudit(ctx, audit, req, timeout)
		audit.SiteChecks = append(audit.SiteChecks, renderedSEOFindings(audit.RenderedSEO)...)
	}

	if platform == models.PlatformDesktop || platform == models.PlatformMobile {
		filterByPlatform(audit, platform)
	}
	checks.AttachChecklistMappings(audit)

	// ── Score ──────────────────────────────────────────────────────────────
	report.ComputeHealthScore(audit)

	// ── Reports ────────────────────────────────────────────────────────────
	if _, err := report.Generate(audit, []string{"json", "html", "markdown"}, reportsDir); err != nil {
		m.finishRecord(id, func(r *AuditRecord) {
			r.Status = StatusFailed
			r.ErrMsg = "report generation failed: " + err.Error()
		})
		m.hub.Broadcast(id, ProgressEvent{Type: "error", Message: "report generation failed"})
		return
	}

	// ── Persist final record ────────────────────────────────────────────────
	m.finishRecord(id, func(r *AuditRecord) {
		r.Status = StatusComplete
		r.HealthScore = audit.HealthScore
		r.Grade = audit.Grade
		r.DesktopScore = audit.DesktopHealthScore
		r.MobileScore = audit.MobileHealthScore
		r.ErrorCount = audit.Stats.Errors
		r.WarnCount = audit.Stats.Warnings
		r.NoticeCount = audit.Stats.Notices
		r.PageCount = audit.PagesCrawled
	})

	record, _ := m.storage.Load(id)
	m.hub.Broadcast(id, ProgressEvent{
		Type:        "complete",
		PageCount:   record.PageCount,
		HealthScore: record.HealthScore,
		Grade:       record.Grade,
		ErrorCount:  record.ErrorCount,
		WarnCount:   record.WarnCount,
		NoticeCount: record.NoticeCount,
	})
}

// finishRecord loads, mutates, and saves an AuditRecord atomically from the
// caller's perspective (storage.Save is internally write-locked).
func (m *Manager) finishRecord(id string, fn func(*AuditRecord)) {
	record, err := m.storage.Load(id)
	if err != nil {
		return
	}
	fn(record)
	now := time.Now()
	record.CompletedAt = &now
	_ = m.storage.Save(record)
}

// filterByPlatform removes check results that don't apply to the requested platform.
// Mirrors the identical logic in cmd/audit.go.
func filterByPlatform(audit *models.SiteAudit, platform models.Platform) {
	keep := func(p models.Platform) bool {
		if p == models.PlatformBoth || p == "" {
			return true
		}
		switch platform {
		case models.PlatformDesktop:
			return p == models.PlatformDesktop
		case models.PlatformMobile:
			return p == models.PlatformMobile || p == models.PlatformDiff
		}
		return true
	}
	for _, page := range audit.Pages {
		out := page.CheckResults[:0]
		for _, r := range page.CheckResults {
			if keep(r.Platform) {
				out = append(out, r)
			}
		}
		page.CheckResults = out
	}
	out := audit.SiteChecks[:0]
	for _, r := range audit.SiteChecks {
		if keep(r.Platform) {
			out = append(out, r)
		}
	}
	audit.SiteChecks = out
}

// newID generates a cryptographically random 8-byte hex string.
func newID() string {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		// fallback: timestamp-based (practically impossible to hit)
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(b)
}
