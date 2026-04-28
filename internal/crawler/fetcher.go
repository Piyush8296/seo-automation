package crawler

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/cars24/seo-automation/internal/models"
)

const defaultMaxBodyBytes int64 = 5 * 1024 * 1024 // 5 MB

// FetchResult holds the result of fetching a single URL.
type FetchResult struct {
	URL            string
	FinalURL       string
	StatusCode     int
	Headers        http.Header
	Body           []byte
	ResponseTimeMs int64
	RedirectChain  []models.RedirectHop
	TLSInfo        *models.TLSInfo
	Error          string
}

// Fetcher wraps an http.Client with redirect-chain tracking.
type Fetcher struct {
	client       *http.Client
	UserAgent    string
	MaxRedirects int
	MaxBodyBytes int64
}

// NewFetcher creates a Fetcher with the given timeout and user-agent.
func NewFetcher(timeout time.Duration, ua string, maxRedirects int, maxBodyBytes int64) *Fetcher {
	if maxRedirects <= 0 {
		maxRedirects = 10
	}
	if maxBodyBytes <= 0 {
		maxBodyBytes = defaultMaxBodyBytes
	}
	var chain []models.RedirectHop

	client := &http.Client{
		Timeout: timeout,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) > 0 {
				chain = append(chain, models.RedirectHop{
					URL:        via[len(via)-1].URL.String(),
					StatusCode: 0, // filled after response
				})
			}
			if len(via) >= maxRedirects {
				return fmt.Errorf("redirect loop: stopped after %d redirects", maxRedirects)
			}
			return nil
		},
	}

	_ = chain // chain is captured per-request below via a fresh closure each time

	return &Fetcher{
		client:       client,
		UserAgent:    ua,
		MaxRedirects: maxRedirects,
		MaxBodyBytes: maxBodyBytes,
	}
}

// Fetch retrieves a URL and returns the full result including redirect chain.
func (f *Fetcher) Fetch(ctx context.Context, rawURL string) *FetchResult {
	var redirectChain []models.RedirectHop
	currentURL := rawURL
	seen := map[string]bool{rawURL: true}
	start := time.Now()

	client := &http.Client{
		Timeout:   f.client.Timeout,
		Transport: f.client.Transport,
		CheckRedirect: func(_ *http.Request, _ []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	for {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, currentURL, nil)
		if err != nil {
			return &FetchResult{URL: rawURL, FinalURL: currentURL, ResponseTimeMs: time.Since(start).Milliseconds(), Error: err.Error(), RedirectChain: redirectChain}
		}
		req.Header.Set("User-Agent", f.UserAgent)
		req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
		req.Header.Set("Accept-Language", "en-US,en;q=0.5")

		resp, err := client.Do(req)
		elapsed := time.Since(start).Milliseconds()
		if err != nil {
			return &FetchResult{URL: rawURL, FinalURL: currentURL, ResponseTimeMs: elapsed, Error: err.Error(), RedirectChain: redirectChain}
		}

		if isRedirectStatus(resp.StatusCode) {
			location := resp.Header.Get("Location")
			if location != "" {
				redirectChain = append(redirectChain, models.RedirectHop{
					URL:        resp.Request.URL.String(),
					StatusCode: resp.StatusCode,
				})
				closeRedirectBody(resp.Body)
				if len(redirectChain) > f.MaxRedirects {
					return &FetchResult{URL: rawURL, FinalURL: currentURL, ResponseTimeMs: elapsed, Error: fmt.Sprintf("redirect loop: stopped after %d hops", f.MaxRedirects), RedirectChain: redirectChain}
				}
				nextURL, err := resolveRedirectLocation(resp.Request.URL, location)
				if err != nil {
					return &FetchResult{URL: rawURL, FinalURL: currentURL, ResponseTimeMs: elapsed, Error: err.Error(), RedirectChain: redirectChain}
				}
				if seen[nextURL] {
					return &FetchResult{URL: rawURL, FinalURL: nextURL, ResponseTimeMs: elapsed, Error: "redirect loop: repeated URL " + nextURL, RedirectChain: redirectChain}
				}
				seen[nextURL] = true
				currentURL = nextURL
				continue
			}
		}

		defer resp.Body.Close()
		body, err := io.ReadAll(io.LimitReader(resp.Body, f.MaxBodyBytes))
		if err != nil {
			return &FetchResult{
				URL:            rawURL,
				FinalURL:       resp.Request.URL.String(),
				StatusCode:     resp.StatusCode,
				Headers:        resp.Header,
				ResponseTimeMs: elapsed,
				RedirectChain:  redirectChain,
				Error:          err.Error(),
			}
		}

		return &FetchResult{
			URL:            rawURL,
			FinalURL:       resp.Request.URL.String(),
			StatusCode:     resp.StatusCode,
			Headers:        resp.Header,
			Body:           body,
			ResponseTimeMs: elapsed,
			RedirectChain:  redirectChain,
			TLSInfo:        extractTLSInfo(resp.TLS),
		}
	}
}

func isRedirectStatus(status int) bool {
	switch status {
	case http.StatusMovedPermanently, http.StatusFound, http.StatusSeeOther, http.StatusTemporaryRedirect, http.StatusPermanentRedirect:
		return true
	default:
		return false
	}
}

func closeRedirectBody(body io.ReadCloser) {
	if body == nil {
		return
	}
	_, _ = io.Copy(io.Discard, io.LimitReader(body, 1024))
	_ = body.Close()
}

func resolveRedirectLocation(base *url.URL, location string) (string, error) {
	next, err := url.Parse(location)
	if err != nil {
		return "", err
	}
	return base.ResolveReference(next).String(), nil
}

// extractTLSInfo converts a tls.ConnectionState into our TLSInfo model.
func extractTLSInfo(state *tls.ConnectionState) *models.TLSInfo {
	if state == nil {
		return nil
	}
	info := &models.TLSInfo{
		Version:     tlsVersionName(state.Version),
		CipherSuite: tls.CipherSuiteName(state.CipherSuite),
		ChainLength: len(state.PeerCertificates),
	}
	if len(state.PeerCertificates) > 0 {
		leaf := state.PeerCertificates[0]
		info.CertSubject = leaf.Subject.CommonName
		info.CertIssuer = leaf.Issuer.CommonName
		info.CertNotBefore = leaf.NotBefore
		info.CertNotAfter = leaf.NotAfter
		info.CertDNSNames = leaf.DNSNames
	}
	return info
}

func tlsVersionName(v uint16) string {
	switch v {
	case tls.VersionTLS10:
		return "TLS 1.0"
	case tls.VersionTLS11:
		return "TLS 1.1"
	case tls.VersionTLS12:
		return "TLS 1.2"
	case tls.VersionTLS13:
		return "TLS 1.3"
	default:
		return fmt.Sprintf("unknown (0x%04x)", v)
	}
}

// FetchStatus fetches only the status code (HEAD request, falls back to GET).
func (f *Fetcher) FetchStatus(ctx context.Context, rawURL string) int {
	req, err := http.NewRequestWithContext(ctx, http.MethodHead, rawURL, nil)
	if err != nil {
		return 0
	}
	req.Header.Set("User-Agent", f.UserAgent)
	resp, err := f.client.Do(req)
	if err != nil {
		// Try GET
		req2, err2 := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
		if err2 != nil {
			return 0
		}
		req2.Header.Set("User-Agent", f.UserAgent)
		resp2, err2 := f.client.Do(req2)
		if err2 != nil {
			return 0
		}
		resp2.Body.Close()
		return resp2.StatusCode
	}
	resp.Body.Close()
	return resp.StatusCode
}
