package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/cars24/seo-automation/internal/models"
)

var diffCmd = &cobra.Command{
	Use:   "diff <before.json> <after.json>",
	Short: "Compare two SEO audit reports and show changes",
	Args:  cobra.ExactArgs(2),
	RunE:  runDiff,
}

func runDiff(cmd *cobra.Command, args []string) error {
	before, err := loadAudit(args[0])
	if err != nil {
		return fmt.Errorf("load before report: %w", err)
	}
	after, err := loadAudit(args[1])
	if err != nil {
		return fmt.Errorf("load after report: %w", err)
	}

	// Build issue keys
	beforeIssues := issueMap(before)
	afterIssues := issueMap(after)

	var newIssues, resolved, persisting []string

	for key := range afterIssues {
		if _, exists := beforeIssues[key]; !exists {
			newIssues = append(newIssues, key)
		} else {
			persisting = append(persisting, key)
		}
	}
	for key := range beforeIssues {
		if _, exists := afterIssues[key]; !exists {
			resolved = append(resolved, key)
		}
	}

	fmt.Printf("╔══════════════════════════════════════════════════════════════\n")
	fmt.Printf("║  SEO Audit Diff Report\n")
	fmt.Printf("╠══════════════════════════════════════════════════════════════\n")
	fmt.Printf("║  Before: %s (score: %.1f %s)\n", before.SiteURL, before.HealthScore, before.Grade)
	fmt.Printf("║  After:  %s (score: %.1f %s)\n", after.SiteURL, after.HealthScore, after.Grade)
	scoreDelta := after.HealthScore - before.HealthScore
	if scoreDelta >= 0 {
		fmt.Printf("║  Score Delta: +%.1f\n", scoreDelta)
	} else {
		fmt.Printf("║  Score Delta: %.1f\n", scoreDelta)
	}
	fmt.Printf("╚══════════════════════════════════════════════════════════════\n\n")

	fmt.Printf("🔴 New Issues (%d)\n", len(newIssues))
	for _, k := range newIssues {
		r := afterIssues[k]
		fmt.Printf("  [%s] %s — %s\n  → %s\n", r.Severity, r.ID, r.Message, r.URL)
	}

	fmt.Printf("\n✅ Resolved Issues (%d)\n", len(resolved))
	for _, k := range resolved {
		r := beforeIssues[k]
		fmt.Printf("  [%s] %s — %s\n  → %s\n", r.Severity, r.ID, r.Message, r.URL)
	}

	fmt.Printf("\n⚪ Persisting Issues (%d)\n", len(persisting))

	fmt.Printf("\nSummary: %d new | %d resolved | %d persisting\n",
		len(newIssues), len(resolved), len(persisting))

	return nil
}

func loadAudit(path string) (*models.SiteAudit, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var audit models.SiteAudit
	if err := json.Unmarshal(data, &audit); err != nil {
		return nil, err
	}
	return &audit, nil
}

func issueMap(audit *models.SiteAudit) map[string]models.CheckResult {
	m := make(map[string]models.CheckResult)
	for _, page := range audit.Pages {
		for _, r := range page.CheckResults {
			key := r.ID + "|" + r.URL
			m[key] = r
		}
	}
	for _, r := range audit.SiteChecks {
		key := r.ID + "|" + r.URL
		m[key] = r
	}
	return m
}
