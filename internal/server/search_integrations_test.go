package server

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSearchIntegrationChecksCoverPackItems(t *testing.T) {
	checks := searchIntegrationChecks()
	if got, want := len(checks), 9; got != want {
		t.Fatalf("searchIntegrationChecks() len=%d, want %d", got, want)
	}

	seen := map[string]SearchIntegrationCheck{}
	for _, check := range checks {
		seen[check.ID] = check
	}
	for _, id := range []string{
		"ANALYTICS-003", "ANALYTICS-005", "SITEMAP-019", "CRAWL-010", "CWV-004",
		"MOBILE-006", "KW-006", "CONTENT-006", "CONTENT-015",
	} {
		if _, ok := seen[id]; !ok {
			t.Fatalf("missing search integration check %s", id)
		}
	}

	if seen["SITEMAP-019"].Provider != searchProviderBing {
		t.Fatalf("SITEMAP-019 should be owned by Bing: %#v", seen["SITEMAP-019"])
	}
	if !seen["CRAWL-010"].NeedsEvidence {
		t.Fatalf("CRAWL-010 should remain manual evidence backed")
	}
}

func TestSearchOAuthConnectRequiresVerifiedProviderSettings(t *testing.T) {
	manager := NewManager(nil, nil)
	h := newHandlers(manager, nil, nil)

	req := httptestJSON(t, http.MethodPost, "/api/search-integrations/oauth/connect", `{"provider":"gsc"}`)
	rr := executeHandler(req, h.connectSearchOAuth)
	if rr.Code != http.StatusPreconditionRequired {
		t.Fatalf("status=%d, want %d", rr.Code, http.StatusPreconditionRequired)
	}

	cfg := manager.GetSettings()
	cfg.Integrations.GSC.PropertyURL = "sc-domain:cars24.com"
	manager.UpdateSettings(cfg)

	req = httptestJSON(t, http.MethodPost, "/api/search-integrations/oauth/connect", `{"provider":"gsc"}`)
	rr = executeHandler(req, h.connectSearchOAuth)
	if rr.Code != http.StatusAccepted {
		t.Fatalf("status=%d, want %d", rr.Code, http.StatusAccepted)
	}
	if !manager.GetSettings().Integrations.GSC.OAuthConnected {
		t.Fatal("GSC OAuthConnected was not marked")
	}
}

func httptestJSON(t *testing.T, method, path, body string) *http.Request {
	t.Helper()
	req := httptest.NewRequest(method, path, bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	return req
}

func executeHandler(req *http.Request, handler func(http.ResponseWriter, *http.Request)) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	handler(rr, req)
	return rr
}
