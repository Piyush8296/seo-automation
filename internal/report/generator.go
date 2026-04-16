package report

import (
	"fmt"
	"math"
	"os"

	"github.com/cars24/seo-automation/internal/models"
)

func gradeFor(score float64) string {
	switch {
	case score >= 90:
		return "A"
	case score >= 80:
		return "B"
	case score >= 70:
		return "C"
	case score >= 50:
		return "D"
	default:
		return "F"
	}
}

// scoreFromStats returns a 0–100 health score from weighted issue density per page.
// Uses log-decay so issue-dense sites still gradate instead of flooring to 0.
// Calibration (weighted issues per page → score):
//   0 → 100 (A) · 2 → 90 (A) · 10 → 79 (C) · 30 → 70 (C) · 100 → 60 (D) · 1000 → 40 (F).
// Errors count 3×, warnings 1×, notices 0.3×.
func scoreFromStats(s models.AuditStats, pageCount int) float64 {
	if pageCount <= 0 {
		return 100
	}
	weighted := float64(s.Errors*3+s.Warnings) + float64(s.Notices)*0.3
	perPage := weighted / float64(pageCount)
	score := 100 - 20*math.Log10(1+perPage)
	return math.Round(math.Max(0, math.Min(100, score))*10) / 10
}

// ComputeHealthScore calculates the overall, desktop, and mobile health scores.
func ComputeHealthScore(audit *models.SiteAudit) {
	var all, desktop, mobile models.AuditStats

	countInto := func(r models.CheckResult, stats *models.AuditStats) {
		stats.TotalChecksRun++
		switch r.Severity {
		case models.SeverityError:
			stats.Errors++
		case models.SeverityWarning:
			stats.Warnings++
		case models.SeverityNotice:
			stats.Notices++
		}
	}

	tally := func(r models.CheckResult) {
		countInto(r, &all)
		switch r.Platform {
		case models.PlatformMobile, models.PlatformDiff:
			countInto(r, &mobile)
		case models.PlatformDesktop:
			countInto(r, &desktop)
		default:
			// PlatformBoth or empty — counts toward both desktop and mobile
			countInto(r, &desktop)
			countInto(r, &mobile)
		}
	}

	for _, page := range audit.Pages {
		for _, r := range page.CheckResults {
			tally(r)
		}
	}
	for _, r := range audit.SiteChecks {
		tally(r)
	}

	audit.Stats = all
	audit.DesktopStats = desktop
	audit.MobileStats = mobile

	pageCount := len(audit.Pages)

	audit.HealthScore = scoreFromStats(all, pageCount)
	audit.Grade = gradeFor(audit.HealthScore)

	audit.DesktopHealthScore = scoreFromStats(desktop, pageCount)
	audit.DesktopGrade = gradeFor(audit.DesktopHealthScore)

	audit.MobileHealthScore = scoreFromStats(mobile, pageCount)
	audit.MobileGrade = gradeFor(audit.MobileHealthScore)
}

// Generate creates all requested output formats and returns a map of format→filepath.
func Generate(audit *models.SiteAudit, formats []string, outputDir string) (map[string]string, error) {
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return nil, fmt.Errorf("create output directory: %w", err)
	}

	result := make(map[string]string)
	for _, format := range formats {
		switch format {
		case "json":
			path, err := WriteJSON(audit, outputDir)
			if err != nil {
				return result, fmt.Errorf("write JSON: %w", err)
			}
			result["json"] = path
		case "html":
			path, err := WriteHTML(audit, outputDir)
			if err != nil {
				return result, fmt.Errorf("write HTML: %w", err)
			}
			result["html"] = path
		case "markdown":
			path, err := WriteMarkdown(audit, outputDir)
			if err != nil {
				return result, fmt.Errorf("write Markdown: %w", err)
			}
			result["markdown"] = path
		}
	}
	return result, nil
}
