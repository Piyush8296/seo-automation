package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	flagCheckReport    string
	flagMaxErrors      int
	flagMaxWarnings    int
)

var checkExitCmd = &cobra.Command{
	Use:   "check-exit",
	Short: "Exit with non-zero code if audit thresholds are exceeded",
	Example: `  seo-audit check-exit --report reports/report.json --max-errors 0 --max-warnings 50`,
	RunE:  runCheckExit,
}

func init() {
	checkExitCmd.Flags().StringVar(&flagCheckReport, "report", "", "Path to JSON report (required)")
	checkExitCmd.Flags().IntVar(&flagMaxErrors, "max-errors", 0, "Maximum allowed errors (0 = fail on any)")
	checkExitCmd.Flags().IntVar(&flagMaxWarnings, "max-warnings", 50, "Maximum allowed warnings")
	_ = checkExitCmd.MarkFlagRequired("report")
}

func runCheckExit(cmd *cobra.Command, args []string) error {
	audit, err := loadAudit(flagCheckReport)
	if err != nil {
		return fmt.Errorf("load report: %w", err)
	}

	fmt.Printf("SEO Audit Gate Check\n")
	fmt.Printf("  Report: %s\n", flagCheckReport)
	fmt.Printf("  Site:   %s\n", audit.SiteURL)
	fmt.Printf("  Score:  %.1f (%s)\n", audit.HealthScore, audit.Grade)
	fmt.Printf("  Errors:   %d (max: %d)\n", audit.Stats.Errors, flagMaxErrors)
	fmt.Printf("  Warnings: %d (max: %d)\n", audit.Stats.Warnings, flagMaxWarnings)

	failed := false
	if audit.Stats.Errors > flagMaxErrors {
		fmt.Printf("\n❌ FAIL: errors (%d) > max-errors (%d)\n", audit.Stats.Errors, flagMaxErrors)
		failed = true
	}
	if audit.Stats.Warnings > flagMaxWarnings {
		fmt.Printf("\n❌ FAIL: warnings (%d) > max-warnings (%d)\n", audit.Stats.Warnings, flagMaxWarnings)
		failed = true
	}

	if failed {
		os.Exit(1)
	}

	fmt.Printf("\n✅ PASS: all thresholds within limits\n")
	return nil
}
