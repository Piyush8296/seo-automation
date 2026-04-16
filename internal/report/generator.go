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

func scoreFromStats(s models.AuditStats) float64 {
	if s.TotalChecksRun == 0 {
		return 100
	}
	penalty := float64(s.Errors*10+s.Warnings*3+s.Notices*1) * 10.0 / float64(s.TotalChecksRun)
	return math.Round(math.Max(0, math.Min(100, 100-penalty))*10) / 10
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

	audit.HealthScore = scoreFromStats(all)
	audit.Grade = gradeFor(audit.HealthScore)

	audit.DesktopHealthScore = scoreFromStats(desktop)
	audit.DesktopGrade = gradeFor(audit.DesktopHealthScore)

	audit.MobileHealthScore = scoreFromStats(mobile)
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
