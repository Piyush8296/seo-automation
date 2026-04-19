package crawler

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/cars24/seo-automation/internal/models"
)

const maxBodyBytes = 5 * 1024 * 1024 // 5 MB

// FetchResult holds the result of fetching a single URL.
type FetchResult struct {
	URL           string
	FinalURL      string
	StatusCode    int
	Headers       http.Header
	Body          []byte
	ResponseTimeMs int64
	RedirectChain []models.RedirectHop
	TLSInfo       *models.TLSInfo
	Error         string
}

// Fetcher wraps an http.Client with redirect-chain tracking.
type Fetcher struct {
	client    *http.Client
	UserAgent string
}

// NewFetcher creates a Fetcher with the given timeout and user-agent.
func NewFetcher(timeout time.Duration, ua string) *Fetcher {
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
			if len(via) >= 10 {
				return fmt.Errorf("redirect loop: stopped after 10 redirects")
			}
			return nil
		},
	}

	_ = chain // chain is captured per-request below via a fresh closure each time

	return &Fetcher{
		client:    client,
		UserAgent: ua,
	}
}

// Fetch retrieves a URL and returns the full result including redirect chain.
func (f *Fetcher) Fetch(ctx context.Context, rawURL string) *FetchResult {
	var redirectChain []models.RedirectHop

	client := &http.Client{
		Timeout: f.client.Timeout,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) > 0 {
				// The previous request's response code is not directly available here,
				// so we store 0 and note: actual status captured from redirect response.
				redirectChain = append(redirectChain, models.RedirectHop{
					URL:        via[len(via)-1].URL.String(),
					StatusCode: 0,
				})
			}
			if len(via) >= 10 {
				return fmt.Errorf("redirect loop: stopped after 10 hops")
			}
			return nil
		},
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return &FetchResult{URL: rawURL, Error: err.Error()}
	}
	req.Header.Set("User-Agent", f.UserAgent)
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.5")

	start := time.Now()
	resp, err := client.Do(req)
	elapsed := time.Since(start).Milliseconds()

	if err != nil {
		return &FetchResult{URL: rawURL, ResponseTimeMs: elapsed, Error: err.Error(), RedirectChain: redirectChain}
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, maxBodyBytes))
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
