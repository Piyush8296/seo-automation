// Package core_web_vitals implements HTML-based Core Web Vitals signal checks.
// While actual CWV measurements require browser rendering (Lighthouse/CrUX),
// we can detect HTML patterns that are known to cause poor CWV scores.
// These checks reflect Google's Page Experience ranking signals.
package core_web_vitals

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/cars24/seo-automation/internal/models"
)

// PageChecks returns per-page Core Web Vitals HTML signal checks.
func PageChecks() []models.PageCheck {
	return []models.PageCheck{
		&lcpImageNotPreloaded{},
		&clsLayoutShiftFonts{},
		&clsNoWidthHeightImages{},
		&fidBlockingScripts{},
		&largestContentfulElement{},
		&fontDisplayMissing{},
		&thirdPartyScriptsBlocking{},
	}
}

// lcpImageNotPreloaded: the LCP element (hero image) should be preloaded
// with <link rel="preload" as="image"> for fastest possible render.
type lcpImageNotPreloaded struct{}

func (c *lcpImageNotPreloaded) Run(p *models.PageData) []models.CheckResult {
	if len(p.Images) == 0 {
		return nil
	}
	// Check first above-fold non-lazy image
	for _, img := range p.Images {
		if !img.IsAboveFold || img.Loading == "lazy" {
			continue
		}
		// Check if this image is preloaded
		if !strings.Contains(p.RawHTML, "rel=\"preload\"") &&
			!strings.Contains(p.RawHTML, "rel='preload'") {
			return []models.CheckResult{{
				ID:       "cwv.lcp.image_not_preloaded",
				Category: "Core Web Vitals",
				Severity: models.SeverityWarning,
				Message:  "LCP candidate image not preloaded — add <link rel=\"preload\" as=\"image\"> in <head>",
				URL:      p.URL,
				Details:  fmt.Sprintf("Candidate: %s", img.Src),
			}}
		}
		return nil
	}
	return nil
}

// clsLayoutShiftFonts: web fonts without font-display:swap or optional cause
// layout shift as text re-renders when font loads (FOUT/FOIT → CLS).
type clsLayoutShiftFonts struct{}

var googleFontsRegex = regexp.MustCompile(`fonts\.googleapis\.com`)
var fontDisplayRegex = regexp.MustCompile(`font-display\s*:\s*(swap|optional|fallback)`)

func (c *clsLayoutShiftFonts) Run(p *models.PageData) []models.CheckResult {
	htmlLower := strings.ToLower(p.RawHTML)

	// Check for Google Fonts without display=swap
	if googleFontsRegex.MatchString(htmlLower) {
		if !strings.Contains(htmlLower, "display=swap") &&
			!strings.Contains(htmlLower, "display=optional") {
			return []models.CheckResult{{
				ID:       "cwv.cls.font_display_missing",
				Category: "Core Web Vitals",
				Severity: models.SeverityWarning,
				Message:  "Google Fonts loaded without font-display:swap — causes CLS from text reflow",
				URL:      p.URL,
				Details:  "Add &display=swap to Google Fonts URL or use font-display:swap in @font-face",
			}}
		}
	}

	// Check custom @font-face without font-display
	if strings.Contains(htmlLower, "@font-face") {
		if !fontDisplayRegex.MatchString(p.RawHTML) {
			return []models.CheckResult{{
				ID:       "cwv.cls.font_display_missing",
				Category: "Core Web Vitals",
				Severity: models.SeverityWarning,
				Message:  "@font-face declarations missing font-display property (causes CLS/FOUT)",
				URL:      p.URL,
				Details:  "Add font-display: swap; to all @font-face declarations",
			}}
		}
	}
	return nil
}

// clsNoWidthHeightImages: images without explicit dimensions cause layout
// shift as browser can't reserve space before loading.
type clsNoWidthHeightImages struct{}

func (c *clsNoWidthHeightImages) Run(p *models.PageData) []models.CheckResult {
	aboveFoldWithoutDims := 0
	for _, img := range p.Images {
		if img.IsAboveFold && (img.Width == 0 || img.Height == 0) {
			aboveFoldWithoutDims++
		}
	}
	if aboveFoldWithoutDims > 0 {
		return []models.CheckResult{{
			ID:       "cwv.cls.above_fold_images_no_dims",
			Category: "Core Web Vitals",
			Severity: models.SeverityError,
			Message: fmt.Sprintf("%d above-fold image(s) missing width/height — direct CLS cause",
				aboveFoldWithoutDims),
			URL:     p.URL,
			Details: "Add explicit width and height attributes to all above-fold images",
		}}
	}
	return nil
}

// fidBlockingScripts: First Input Delay (FID) / Interaction to Next Paint (INP)
// is caused by long tasks blocking the main thread. Render-blocking scripts are
// the primary cause.
type fidBlockingScripts struct{}

func (c *fidBlockingScripts) Run(p *models.PageData) []models.CheckResult {
	if p.RenderBlockingScripts > 2 {
		return []models.CheckResult{{
			ID:       "cwv.fid.blocking_scripts",
			Category: "Core Web Vitals",
			Severity: models.SeverityError,
			Message: fmt.Sprintf("%d render-blocking scripts block main thread (FID/INP impact)",
				p.RenderBlockingScripts),
			URL:     p.URL,
			Details: "Add async or defer to non-critical scripts. Move to bottom of body or use dynamic import.",
		}}
	}
	return nil
}

// largestContentfulElement: LCP should be an image or heading text, not a
// background image (which can't be detected/preloaded by preload scanner).
type largestContentfulElement struct{}

func (c *largestContentfulElement) Run(p *models.PageData) []models.CheckResult {
	// Check if hero section uses background-image (bad for LCP)
	cssBackgroundPattern := regexp.MustCompile(`background-image\s*:\s*url\(`)
	if cssBackgroundPattern.MatchString(p.RawHTML) {
		// Only flag if there are no above-fold img elements
		hasAboveFoldImg := false
		for _, img := range p.Images {
			if img.IsAboveFold {
				hasAboveFoldImg = true
				break
			}
		}
		if !hasAboveFoldImg {
			return []models.CheckResult{{
				ID:       "cwv.lcp.background_image",
				Category: "Core Web Vitals",
				Severity: models.SeverityWarning,
				Message:  "Hero content appears to use CSS background-image — not preloadable, harms LCP",
				URL:      p.URL,
				Details:  "Convert hero background images to <img> elements for better LCP and preload support",
			}}
		}
	}
	return nil
}

// fontDisplayMissing: broader font-display check already covered but also
// check for self-hosted fonts via link tags.
type fontDisplayMissing struct{}

func (c *fontDisplayMissing) Run(p *models.PageData) []models.CheckResult {
	// Check for font preloads without crossorigin (causes double download)
	if strings.Contains(p.RawHTML, `as="font"`) || strings.Contains(p.RawHTML, `as='font'`) {
		if !strings.Contains(p.RawHTML, "crossorigin") {
			return []models.CheckResult{{
				ID:       "cwv.lcp.font_preload_missing_crossorigin",
				Category: "Core Web Vitals",
				Severity: models.SeverityWarning,
				Message:  "Font preload missing crossorigin attribute — causes double font download",
				URL:      p.URL,
				Details:  "Add crossorigin=\"anonymous\" to <link rel=\"preload\" as=\"font\"> tags",
			}}
		}
	}
	return nil
}

// thirdPartyScriptsBlocking: third-party scripts (analytics, ads, chat) are a
// major source of poor INP and FCP scores.
type thirdPartyScriptsBlocking struct{}

var knownHeavyThirdParties = []string{
	"googletagmanager.com", "google-analytics.com",
	"facebook.net", "connect.facebook.net",
	"platform.twitter.com", "snap.licdn.com",
	"hotjar.com", "fullstory.com", "mouseflow.com",
	"intercom.io", "crisp.chat", "tawk.to",
	"zopim.com", "freshchat.com",
}

func (c *thirdPartyScriptsBlocking) Run(p *models.PageData) []models.CheckResult {
	if p.ExternalScriptCount < 3 {
		return nil
	}
	blockingThirdParty := 0
	for _, domain := range knownHeavyThirdParties {
		if strings.Contains(p.RawHTML, domain) {
			blockingThirdParty++
		}
	}
	if blockingThirdParty >= 3 {
		return []models.CheckResult{{
			ID:       "cwv.inp.heavy_third_party_scripts",
			Category: "Core Web Vitals",
			Severity: models.SeverityWarning,
			Message: fmt.Sprintf("%d heavy third-party scripts detected — significant INP/TBT impact",
				blockingThirdParty),
			URL:     p.URL,
			Details: "Audit third-party scripts with WebPageTest. Consider: load after user interaction, use Partytown, or remove non-essential scripts.",
		}}
	}
	return nil
}
