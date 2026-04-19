package server

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/cars24/seo-automation/internal/checks"
)

// Handlers holds references to all service-layer dependencies.
type Handlers struct {
	manager *Manager
	storage *Storage
	hub     *Hub
}

func newHandlers(manager *Manager, storage *Storage, hub *Hub) *Handlers {
	return &Handlers{manager: manager, storage: storage, hub: hub}
}

// ── helpers ────────────────────────────────────────────────────────────────────

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeErr(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

// ── POST /api/audits ──────────────────────────────────────────────────────────

func (h *Handlers) startAudit(w http.ResponseWriter, r *http.Request) {
	var req StartAuditRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if strings.TrimSpace(req.URL) == "" {
		writeErr(w, http.StatusBadRequest, "url is required")
		return
	}

	record, err := h.manager.StartAudit(req)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, record)
}

// ── GET /api/audits ───────────────────────────────────────────────────────────

func (h *Handlers) listAudits(w http.ResponseWriter, r *http.Request) {
	records, err := h.storage.List()
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	if records == nil {
		records = []*AuditRecord{} // always return an array, never null
	}
	writeJSON(w, http.StatusOK, records)
}

// ── GET /api/audits/{id} ──────────────────────────────────────────────────────

func (h *Handlers) getAudit(w http.ResponseWriter, r *http.Request, id string) {
	record, err := h.storage.Load(id)
	if err != nil {
		writeErr(w, http.StatusNotFound, "audit not found")
		return
	}
	writeJSON(w, http.StatusOK, record)
}

// ── DELETE /api/audits/{id} ───────────────────────────────────────────────────

func (h *Handlers) deleteAudit(w http.ResponseWriter, r *http.Request, id string) {
	h.manager.CancelAudit(id) // no-op if not running
	if err := h.storage.Delete(id); err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// ── POST /api/audits/{id}/cancel ──────────────────────────────────────────────

func (h *Handlers) cancelAudit(w http.ResponseWriter, r *http.Request, id string) {
	if !h.manager.CancelAudit(id) {
		writeErr(w, http.StatusConflict, "audit is not currently running")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "cancelling"})
}

// ── GET /api/audits/{id}/events  (SSE) ────────────────────────────────────────

func (h *Handlers) auditEvents(w http.ResponseWriter, r *http.Request, id string) {
	record, err := h.storage.Load(id)
	if err != nil {
		writeErr(w, http.StatusNotFound, "audit not found")
		return
	}

	// Already finished — send a single terminal event and close immediately.
	if record.Status != StatusRunning {
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")

		var evt ProgressEvent
		switch record.Status {
		case StatusComplete:
			evt = ProgressEvent{
				Type:        "complete",
				HealthScore: record.HealthScore,
				Grade:       record.Grade,
				ErrorCount:  record.ErrorCount,
				WarnCount:   record.WarnCount,
				NoticeCount: record.NoticeCount,
				PageCount:   record.PageCount,
			}
		case StatusFailed:
			evt = ProgressEvent{Type: "error", Message: record.ErrMsg}
		case StatusCancelled:
			evt = ProgressEvent{Type: "cancelled", Message: "audit was cancelled"}
		}

		data, _ := json.Marshal(evt)
		_, _ = w.Write([]byte("data: " + string(data) + "\n\n"))
		if f, ok := w.(http.Flusher); ok {
			f.Flush()
		}
		return
	}

	h.hub.ServeSSE(w, r, id)
}

// ── GET /api/audits/{id}/report.html ─────────────────────────────────────────

func (h *Handlers) serveReportHTML(w http.ResponseWriter, r *http.Request, id string) {
	http.ServeFile(w, r, h.storage.ReportPath(id, "html"))
}

// ── GET /api/audits/{id}/report.json ─────────────────────────────────────────

func (h *Handlers) serveReportJSON(w http.ResponseWriter, r *http.Request, id string) {
	http.ServeFile(w, r, h.storage.ReportPath(id, "json"))
}

// ── GET /api/settings ────────────────────────────────────────────────────────

func (h *Handlers) getSettings(w http.ResponseWriter, r *http.Request) {
	cfg := h.manager.GetSettings()
	writeJSON(w, http.StatusOK, cfg)
}

// ── PUT /api/settings ────────────────────────────────────────────────────────

func (h *Handlers) updateSettings(w http.ResponseWriter, r *http.Request) {
	var cfg AppSettings
	if err := json.NewDecoder(r.Body).Decode(&cfg); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	h.manager.UpdateSettings(cfg)
	writeJSON(w, http.StatusOK, cfg)
}

// ── GET /api/checks ──────────────────────────────────────────────────────────

func (h *Handlers) listChecks(w http.ResponseWriter, r *http.Request) {
	cat := checks.GetCatalog()
	writeJSON(w, http.StatusOK, map[string]any{
		"total":       cat.Total,
		"page_checks": cat.PageChecks,
		"site_checks": cat.SiteChecks,
		"check_ids":   cat.CheckIDs,
		"checks":      checks.GetCheckDescriptors(),
	})
}

// ── GET /api/audits/diff?a={id}&b={id} ───────────────────────────────────────

func (h *Handlers) diffAudits(w http.ResponseWriter, r *http.Request) {
	aID := r.URL.Query().Get("a")
	bID := r.URL.Query().Get("b")
	if aID == "" || bID == "" {
		writeErr(w, http.StatusBadRequest, "query params 'a' and 'b' are required")
		return
	}

	aRecord, err := h.storage.Load(aID)
	if err != nil {
		writeErr(w, http.StatusNotFound, "audit 'a' not found")
		return
	}
	bRecord, err := h.storage.Load(bID)
	if err != nil {
		writeErr(w, http.StatusNotFound, "audit 'b' not found")
		return
	}

	writeJSON(w, http.StatusOK, DiffResponse{
		AuditA:      aRecord,
		AuditB:      bRecord,
		ScoreDelta:  bRecord.HealthScore - aRecord.HealthScore,
		ErrorDelta:  bRecord.ErrorCount - aRecord.ErrorCount,
		WarnDelta:   bRecord.WarnCount - aRecord.WarnCount,
		NoticeDelta: bRecord.NoticeCount - aRecord.NoticeCount,
		PageDelta:   bRecord.PageCount - aRecord.PageCount,
	})
}
