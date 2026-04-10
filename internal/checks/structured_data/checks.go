package structured_data

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/cars24/seo-automation/internal/models"
)

// PageChecks returns per-page structured data checks.
func PageChecks() []models.PageCheck {
	return []models.PageCheck{
		&schemaJSONLDMissing{},
		&schemaJSONLDInvalidJSON{},
		&schemaJSONLDMissingContext{},
		&schemaJSONLDMissingType{},
		&schemaJSONLDDuplicateType{},
		&schemaArticleMissingFields{},
		&schemaProductMissingFields{},
		&schemaBreadcrumbInvalid{},
		&schemaFAQInvalid{},
	}
}

// SiteChecks returns site-wide structured data checks.
func SiteChecks() []models.SiteCheck {
	return []models.SiteCheck{
		&schemaOrganizationMissingHomepage{},
	}
}

type schemaJSONLDMissing struct{}

func (c *schemaJSONLDMissing) Run(p *models.PageData) []models.CheckResult {
	if len(p.SchemaJSONRaw) == 0 {
		return []models.CheckResult{{
			ID:       "schema.jsonld.missing",
			Category: "Structured Data",
			Severity: models.SeverityNotice,
			Message:  "No JSON-LD structured data found",
			URL:      p.URL,
		}}
	}
	return nil
}

type schemaJSONLDInvalidJSON struct{}

func (c *schemaJSONLDInvalidJSON) Run(p *models.PageData) []models.CheckResult {
	var results []models.CheckResult
	for _, raw := range p.SchemaJSONRaw {
		var obj interface{}
		if err := json.Unmarshal([]byte(raw), &obj); err != nil {
			results = append(results, models.CheckResult{
				ID:       "schema.jsonld.invalid_json",
				Category: "Structured Data",
				Severity: models.SeverityError,
				Message:  "Invalid JSON in JSON-LD structured data",
				URL:      p.URL,
				Details:  err.Error(),
			})
		}
	}
	return results
}

type schemaJSONLDMissingContext struct{}

func (c *schemaJSONLDMissingContext) Run(p *models.PageData) []models.CheckResult {
	var results []models.CheckResult
	for _, raw := range p.SchemaJSONRaw {
		var obj map[string]interface{}
		if err := json.Unmarshal([]byte(raw), &obj); err != nil {
			continue
		}
		ctx, _ := obj["@context"].(string)
		if ctx == "" {
			results = append(results, models.CheckResult{
				ID:       "schema.jsonld.missing_context",
				Category: "Structured Data",
				Severity: models.SeverityError,
				Message:  "JSON-LD schema missing @context",
				URL:      p.URL,
			})
		}
	}
	return results
}

type schemaJSONLDMissingType struct{}

func (c *schemaJSONLDMissingType) Run(p *models.PageData) []models.CheckResult {
	var results []models.CheckResult
	for _, raw := range p.SchemaJSONRaw {
		var obj map[string]interface{}
		if err := json.Unmarshal([]byte(raw), &obj); err != nil {
			continue
		}
		t, _ := obj["@type"].(string)
		if t == "" {
			results = append(results, models.CheckResult{
				ID:       "schema.jsonld.missing_type",
				Category: "Structured Data",
				Severity: models.SeverityError,
				Message:  "JSON-LD schema missing @type",
				URL:      p.URL,
			})
		}
	}
	return results
}

type schemaJSONLDDuplicateType struct{}

func (c *schemaJSONLDDuplicateType) Run(p *models.PageData) []models.CheckResult {
	typeCounts := map[string]int{}
	for _, raw := range p.SchemaJSONRaw {
		var obj map[string]interface{}
		if err := json.Unmarshal([]byte(raw), &obj); err != nil {
			continue
		}
		t, _ := obj["@type"].(string)
		if t != "" {
			typeCounts[strings.ToLower(t)]++
		}
	}
	var results []models.CheckResult
	for t, count := range typeCounts {
		if count > 1 {
			results = append(results, models.CheckResult{
				ID:       "schema.jsonld.duplicate_type",
				Category: "Structured Data",
				Severity: models.SeverityWarning,
				Message:  fmt.Sprintf("Duplicate schema @type: %s (%d occurrences)", t, count),
				URL:      p.URL,
			})
		}
	}
	return results
}

type schemaArticleMissingFields struct{}

func (c *schemaArticleMissingFields) Run(p *models.PageData) []models.CheckResult {
	for _, raw := range p.SchemaJSONRaw {
		var obj map[string]interface{}
		if err := json.Unmarshal([]byte(raw), &obj); err != nil {
			continue
		}
		t := strings.ToLower(fmt.Sprintf("%v", obj["@type"]))
		if !strings.Contains(t, "article") && !strings.Contains(t, "newsarticle") && !strings.Contains(t, "blogposting") {
			continue
		}
		var missing []string
		if obj["headline"] == nil {
			missing = append(missing, "headline")
		}
		if obj["datePublished"] == nil {
			missing = append(missing, "datePublished")
		}
		if obj["author"] == nil {
			missing = append(missing, "author")
		}
		if len(missing) > 0 {
			return []models.CheckResult{{
				ID:       "schema.article.missing_fields",
				Category: "Structured Data",
				Severity: models.SeverityWarning,
				Message:  fmt.Sprintf("Article schema missing required fields: %s", strings.Join(missing, ", ")),
				URL:      p.URL,
			}}
		}
	}
	return nil
}

type schemaProductMissingFields struct{}

func (c *schemaProductMissingFields) Run(p *models.PageData) []models.CheckResult {
	for _, raw := range p.SchemaJSONRaw {
		var obj map[string]interface{}
		if err := json.Unmarshal([]byte(raw), &obj); err != nil {
			continue
		}
		if strings.ToLower(fmt.Sprintf("%v", obj["@type"])) != "product" {
			continue
		}
		var missing []string
		if obj["name"] == nil {
			missing = append(missing, "name")
		}
		if obj["offers"] == nil {
			missing = append(missing, "offers")
		}
		if len(missing) > 0 {
			return []models.CheckResult{{
				ID:       "schema.product.missing_fields",
				Category: "Structured Data",
				Severity: models.SeverityWarning,
				Message:  fmt.Sprintf("Product schema missing required fields: %s", strings.Join(missing, ", ")),
				URL:      p.URL,
			}}
		}
	}
	return nil
}

type schemaBreadcrumbInvalid struct{}

func (c *schemaBreadcrumbInvalid) Run(p *models.PageData) []models.CheckResult {
	for _, raw := range p.SchemaJSONRaw {
		var obj map[string]interface{}
		if err := json.Unmarshal([]byte(raw), &obj); err != nil {
			continue
		}
		if strings.ToLower(fmt.Sprintf("%v", obj["@type"])) != "breadcrumblist" {
			continue
		}
		items, ok := obj["itemListElement"].([]interface{})
		if !ok || len(items) == 0 {
			return []models.CheckResult{{
				ID:       "schema.breadcrumb.invalid",
				Category: "Structured Data",
				Severity: models.SeverityWarning,
				Message:  "BreadcrumbList schema missing itemListElement",
				URL:      p.URL,
			}}
		}
	}
	return nil
}

type schemaFAQInvalid struct{}

func (c *schemaFAQInvalid) Run(p *models.PageData) []models.CheckResult {
	for _, raw := range p.SchemaJSONRaw {
		var obj map[string]interface{}
		if err := json.Unmarshal([]byte(raw), &obj); err != nil {
			continue
		}
		if strings.ToLower(fmt.Sprintf("%v", obj["@type"])) != "faqpage" {
			continue
		}
		items, ok := obj["mainEntity"].([]interface{})
		if !ok || len(items) == 0 {
			return []models.CheckResult{{
				ID:       "schema.faq.invalid",
				Category: "Structured Data",
				Severity: models.SeverityWarning,
				Message:  "FAQPage schema missing mainEntity questions",
				URL:      p.URL,
			}}
		}
	}
	return nil
}

type schemaOrganizationMissingHomepage struct{}

func (c *schemaOrganizationMissingHomepage) Run(pages []*models.PageData) []models.CheckResult {
	for _, p := range pages {
		if p.Depth != 0 {
			continue
		}
		// Homepage found — check for org schema
		for _, raw := range p.SchemaJSONRaw {
			var obj map[string]interface{}
			if err := json.Unmarshal([]byte(raw), &obj); err != nil {
				continue
			}
			t := strings.ToLower(fmt.Sprintf("%v", obj["@type"]))
			if strings.Contains(t, "organization") || strings.Contains(t, "localbusiness") || strings.Contains(t, "corporation") {
				return nil
			}
		}
		return []models.CheckResult{{
			ID:       "schema.organization.missing_homepage",
			Category: "Structured Data",
			Severity: models.SeverityWarning,
			Message:  "Homepage has no Organization/LocalBusiness structured data",
			URL:      p.URL,
		}}
	}
	return nil
}
