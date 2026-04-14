package server

import "time"

// AuditStatus represents the lifecycle state of an audit.
type AuditStatus string

const (
	StatusRunning   AuditStatus = "running"
	StatusComplete  AuditStatus = "complete"
	StatusFailed    AuditStatus = "failed"
	StatusCancelled AuditStatus = "cancelled"
)

// StartAuditRequest is the JSON body sent to POST /api/audits.
type StartAuditRequest struct {
	URL         string `json:"url"`
	MaxDepth    int    `json:"max_depth"`
	MaxPages    int    `json:"max_pages"`
	Concurrency int    `json:"concurrency"`
	Timeout     string `json:"timeout"`
	Platform    string `json:"platform"`
	OutputDir   string `json:"output_dir"` // optional — overrides default storage dir
}

// AuditRecord is the persistent metadata for one audit run, stored as meta.json.
type AuditRecord struct {
	ID           string             `json:"id"`
	URL          string             `json:"url"`
	Config       StartAuditRequest  `json:"config"`
	Status       AuditStatus        `json:"status"`
	CreatedAt    time.Time          `json:"created_at"`
	CompletedAt  *time.Time         `json:"completed_at,omitempty"`
	HealthScore  float64            `json:"health_score"`
	Grade        string             `json:"grade"`
	DesktopScore float64            `json:"desktop_score"`
	MobileScore  float64            `json:"mobile_score"`
	ErrorCount   int                `json:"error_count"`
	WarnCount    int                `json:"warn_count"`
	NoticeCount  int                `json:"notice_count"`
	PageCount    int                `json:"page_count"`
	ReportsDir   string             `json:"reports_dir"`
	ErrMsg       string             `json:"error,omitempty"`
}

// ProgressEvent is a single SSE message sent to the browser.
type ProgressEvent struct {
	Type         string  `json:"type"` // "progress" | "complete" | "error" | "cancelled"
	PagesCrawled int     `json:"pages_crawled,omitempty"`
	CurrentURL   string  `json:"current_url,omitempty"`
	Message      string  `json:"message,omitempty"`
	HealthScore  float64 `json:"health_score,omitempty"`
	Grade        string  `json:"grade,omitempty"`
	ErrorCount   int     `json:"error_count,omitempty"`
	WarnCount    int     `json:"warn_count,omitempty"`
	NoticeCount  int     `json:"notice_count,omitempty"`
	PageCount    int     `json:"page_count,omitempty"`
}

// DiffResponse is the response from GET /api/audits/diff.
type DiffResponse struct {
	AuditA      *AuditRecord `json:"audit_a"`
	AuditB      *AuditRecord `json:"audit_b"`
	ScoreDelta  float64      `json:"score_delta"`  // positive = improved
	ErrorDelta  int          `json:"error_delta"`  // negative = fewer errors (good)
	WarnDelta   int          `json:"warn_delta"`
	NoticeDelta int          `json:"notice_delta"`
	PageDelta   int          `json:"page_delta"`
}
