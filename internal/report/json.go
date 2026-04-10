package report

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/cars24/seo-automation/internal/models"
)

// WriteJSON writes the audit as a pretty-printed JSON file.
func WriteJSON(audit *models.SiteAudit, outputDir string) (string, error) {
	data, err := json.MarshalIndent(audit, "", "  ")
	if err != nil {
		return "", err
	}
	path := filepath.Join(outputDir, "report.json")
	return path, os.WriteFile(path, data, 0644)
}
