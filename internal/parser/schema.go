package parser

import (
	"encoding/json"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// ExtractSchemaJSON returns raw JSON strings from <script type="application/ld+json"> tags.
func ExtractSchemaJSON(doc *goquery.Document) []string {
	var out []string
	doc.Find(`script[type="application/ld+json"]`).Each(func(_ int, s *goquery.Selection) {
		text := strings.TrimSpace(s.Text())
		if text != "" {
			out = append(out, text)
		}
	})
	return out
}

// ParseSchemaObjects attempts to parse each raw JSON string into a map.
// Returns the successfully parsed objects.
func ParseSchemaObjects(raw []string) []map[string]interface{} {
	var out []map[string]interface{}
	for _, r := range raw {
		var obj map[string]interface{}
		if err := json.Unmarshal([]byte(r), &obj); err == nil {
			out = append(out, obj)
			continue
		}
		// Try array of objects
		var arr []map[string]interface{}
		if err := json.Unmarshal([]byte(r), &arr); err == nil {
			out = append(out, arr...)
		}
	}
	return out
}

// SchemaType returns the @type value from a schema object as a lowercase string.
func SchemaType(obj map[string]interface{}) string {
	t, _ := obj["@type"].(string)
	return strings.ToLower(t)
}
