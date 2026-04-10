package mobile_desktop

import (
	"fmt"
	"math"
	"strings"

	"github.com/cars24/seo-automation/internal/models"
)

// PageChecks returns per-page mobile vs desktop checks.
func PageChecks() []models.PageCheck {
	return []models.PageCheck{
		&statusMismatch{},
		&titleMismatch{},
		&metaDescMismatch{},
		&h1Mismatch{},
		&canonicalMismatch{},
		&schemaMismatch{},
		&linksMismatch{},
		&separateMobileSite{},
		&ogImageMismatch{},
		&contentMismatch{},
	}
}

func hasMobile(p *models.PageData) bool {
	return p.MobileData != nil
}

type statusMismatch struct{}

func (c *statusMismatch) Run(p *models.PageData) []models.CheckResult {
	if !hasMobile(p) {
		return nil
	}
	if p.MobileData.StatusCode != 0 && p.StatusCode != p.MobileData.StatusCode {
		return []models.CheckResult{{
			ID:       "mob_desk.status_mismatch",
			Category: "Mobile vs Desktop",
			Severity: models.SeverityError,
			Message:  fmt.Sprintf("Different HTTP status: desktop=%d mobile=%d", p.StatusCode, p.MobileData.StatusCode),
			URL:      p.URL,
			Platform: models.PlatformDiff,
		}}
	}
	return nil
}

type titleMismatch struct{}

func (c *titleMismatch) Run(p *models.PageData) []models.CheckResult {
	if !hasMobile(p) {
		return nil
	}
	if p.MobileData.Title != "" && !strings.EqualFold(p.Title, p.MobileData.Title) {
		return []models.CheckResult{{
			ID:       "mob_desk.title_mismatch",
			Category: "Mobile vs Desktop",
			Severity: models.SeverityError,
			Message:  "Title differs between mobile and desktop",
			URL:      p.URL,
			Details:  fmt.Sprintf("desktop=%q mobile=%q", p.Title, p.MobileData.Title),
			Platform: models.PlatformDiff,
		}}
	}
	return nil
}

type metaDescMismatch struct{}

func (c *metaDescMismatch) Run(p *models.PageData) []models.CheckResult {
	if !hasMobile(p) {
		return nil
	}
	if p.MobileData.MetaDesc != "" && !strings.EqualFold(p.MetaDesc, p.MobileData.MetaDesc) {
		return []models.CheckResult{{
			ID:       "mob_desk.meta_desc_mismatch",
			Category: "Mobile vs Desktop",
			Severity: models.SeverityWarning,
			Message:  "Meta description differs between mobile and desktop",
			URL:      p.URL,
			Platform: models.PlatformDiff,
		}}
	}
	return nil
}

type h1Mismatch struct{}

func (c *h1Mismatch) Run(p *models.PageData) []models.CheckResult {
	if !hasMobile(p) {
		return nil
	}
	desktopH1 := ""
	if len(p.H1s) > 0 {
		desktopH1 = strings.TrimSpace(p.H1s[0])
	}
	mobileH1 := ""
	if len(p.MobileData.H1s) > 0 {
		mobileH1 = strings.TrimSpace(p.MobileData.H1s[0])
	}
	if mobileH1 != "" && !strings.EqualFold(desktopH1, mobileH1) {
		return []models.CheckResult{{
			ID:       "mob_desk.h1_mismatch",
			Category: "Mobile vs Desktop",
			Severity: models.SeverityWarning,
			Message:  "H1 differs between mobile and desktop",
			URL:      p.URL,
			Details:  fmt.Sprintf("desktop=%q mobile=%q", desktopH1, mobileH1),
			Platform: models.PlatformDiff,
		}}
	}
	return nil
}

type canonicalMismatch struct{}

func (c *canonicalMismatch) Run(p *models.PageData) []models.CheckResult {
	if !hasMobile(p) {
		return nil
	}
	if p.MobileData.Canonical != "" && !strings.EqualFold(
		strings.TrimRight(p.Canonical, "/"),
		strings.TrimRight(p.MobileData.Canonical, "/"),
	) {
		return []models.CheckResult{{
			ID:       "mob_desk.canonical_mismatch",
			Category: "Mobile vs Desktop",
			Severity: models.SeverityError,
			Message:  "Canonical URL differs between mobile and desktop",
			URL:      p.URL,
			Details:  fmt.Sprintf("desktop=%s mobile=%s", p.Canonical, p.MobileData.Canonical),
			Platform: models.PlatformDiff,
		}}
	}
	return nil
}

type schemaMismatch struct{}

func (c *schemaMismatch) Run(p *models.PageData) []models.CheckResult {
	if !hasMobile(p) {
		return nil
	}
	if len(p.SchemaJSONRaw) != len(p.MobileData.SchemaJSONRaw) {
		return []models.CheckResult{{
			ID:       "mob_desk.schema_mismatch",
			Category: "Mobile vs Desktop",
			Severity: models.SeverityWarning,
			Message: fmt.Sprintf("Different number of schema objects: desktop=%d mobile=%d",
				len(p.SchemaJSONRaw), len(p.MobileData.SchemaJSONRaw)),
			URL:      p.URL,
			Platform: models.PlatformDiff,
		}}
	}
	return nil
}

type linksMismatch struct{}

func (c *linksMismatch) Run(p *models.PageData) []models.CheckResult {
	if !hasMobile(p) {
		return nil
	}
	desktopLinks := 0
	for _, l := range p.Links {
		if l.IsInternal {
			desktopLinks++
		}
	}
	mobileLinks := 0
	for _, l := range p.MobileData.Links {
		if l.IsInternal {
			mobileLinks++
		}
	}
	if desktopLinks > 0 && mobileLinks > 0 {
		diff := math.Abs(float64(desktopLinks-mobileLinks)) / float64(desktopLinks)
		if diff > 0.3 {
			return []models.CheckResult{{
				ID:       "mob_desk.links_mismatch",
				Category: "Mobile vs Desktop",
				Severity: models.SeverityWarning,
				Message: fmt.Sprintf("Internal link count differs significantly: desktop=%d mobile=%d (%.0f%%)",
					desktopLinks, mobileLinks, diff*100),
				URL:      p.URL,
				Platform: models.PlatformDiff,
			}}
		}
	}
	return nil
}

type separateMobileSite struct{}

func (c *separateMobileSite) Run(p *models.PageData) []models.CheckResult {
	if !hasMobile(p) {
		return nil
	}
	if p.MobileData.FinalURL == "" {
		return nil
	}
	// Check if mobile redirects to m. subdomain
	mobileURL := strings.ToLower(p.MobileData.FinalURL)
	desktopURL := strings.ToLower(p.FinalURL)
	if mobileURL != desktopURL {
		if strings.Contains(mobileURL, "://m.") || strings.Contains(mobileURL, ".m.") {
			return []models.CheckResult{{
				ID:       "mob_desk.separate_mobile_site",
				Category: "Mobile vs Desktop",
				Severity: models.SeverityWarning,
				Message:  "Mobile requests redirect to a separate mobile site",
				URL:      p.URL,
				Details:  p.MobileData.FinalURL,
				Platform: models.PlatformDiff,
			}}
		}
	}
	return nil
}

type ogImageMismatch struct{}

func (c *ogImageMismatch) Run(p *models.PageData) []models.CheckResult {
	if !hasMobile(p) {
		return nil
	}
	desktopOG := p.OGTags["og:image"]
	mobileOG := p.MobileData.OGTags["og:image"]
	if desktopOG != "" && mobileOG != "" && !strings.EqualFold(desktopOG, mobileOG) {
		return []models.CheckResult{{
			ID:       "mob_desk.og_image_mismatch",
			Category: "Mobile vs Desktop",
			Severity: models.SeverityNotice,
			Message:  "og:image differs between mobile and desktop",
			URL:      p.URL,
			Platform: models.PlatformDiff,
		}}
	}
	return nil
}

type contentMismatch struct{}

func (c *contentMismatch) Run(p *models.PageData) []models.CheckResult {
	if !hasMobile(p) {
		return nil
	}
	if p.WordCount == 0 || p.MobileData.WordCount == 0 {
		return nil
	}
	diff := math.Abs(float64(p.WordCount-p.MobileData.WordCount)) / float64(p.WordCount)
	if diff > 0.5 {
		return []models.CheckResult{{
			ID:       "mob_desk.content_mismatch",
			Category: "Mobile vs Desktop",
			Severity: models.SeverityWarning,
			Message: fmt.Sprintf("Content word count differs significantly: desktop=%d mobile=%d (%.0f%%)",
				p.WordCount, p.MobileData.WordCount, diff*100),
			URL:      p.URL,
			Platform: models.PlatformDiff,
		}}
	}
	return nil
}
