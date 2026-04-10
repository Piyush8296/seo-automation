package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "seo-audit",
	Short: "Production-grade SEO site auditor (146 checks across 17 categories)",
	Long: `seo-audit crawls a website and runs 146 SEO checks modelled after
Screaming Frog and SEMrush Site Audit. Outputs JSON, HTML, and Markdown reports.`,
}

// Execute runs the root command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(auditCmd)
	rootCmd.AddCommand(diffCmd)
	rootCmd.AddCommand(checkExitCmd)
}
