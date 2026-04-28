package crawler

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestFetchCapturesRedirectHopStatusCodes(t *testing.T) {
	fetcher := testFetcher(func(r *http.Request) (*http.Response, error) {
		header := http.Header{}
		switch r.URL.Path {
		case "/start":
			header.Set("Location", "/middle")
			return testResponse(r, http.StatusMovedPermanently, header, ""), nil
		case "/middle":
			header.Set("Location", "/final")
			return testResponse(r, http.StatusFound, header, ""), nil
		default:
			return testResponse(r, http.StatusOK, header, "<html><body>ok</body></html>"), nil
		}
	})

	result := fetcher.Fetch(context.Background(), "https://example.test/start")

	if result.Error != "" {
		t.Fatalf("unexpected fetch error: %s", result.Error)
	}
	if got, want := len(result.RedirectChain), 2; got != want {
		t.Fatalf("redirect chain len=%d, want %d: %#v", got, want, result.RedirectChain)
	}
	if result.RedirectChain[0].StatusCode != http.StatusMovedPermanently {
		t.Fatalf("first redirect status=%d, want 301", result.RedirectChain[0].StatusCode)
	}
	if result.RedirectChain[1].StatusCode != http.StatusFound {
		t.Fatalf("second redirect status=%d, want 302", result.RedirectChain[1].StatusCode)
	}
	if !strings.HasSuffix(result.FinalURL, "/final") {
		t.Fatalf("final URL=%q, want /final", result.FinalURL)
	}
}

func TestFetchDetectsRedirectLoop(t *testing.T) {
	fetcher := testFetcher(func(r *http.Request) (*http.Response, error) {
		header := http.Header{}
		if r.URL.Path == "/a" {
			header.Set("Location", "/b")
			return testResponse(r, http.StatusMovedPermanently, header, ""), nil
		}
		header.Set("Location", "/a")
		return testResponse(r, http.StatusMovedPermanently, header, ""), nil
	})

	result := fetcher.Fetch(context.Background(), "https://example.test/a")

	if !strings.Contains(result.Error, "redirect loop") {
		t.Fatalf("expected redirect loop error, got %q", result.Error)
	}
	if got, want := len(result.RedirectChain), 2; got != want {
		t.Fatalf("redirect chain len=%d, want %d: %#v", got, want, result.RedirectChain)
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (fn roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return fn(r)
}

func testFetcher(fn roundTripFunc) *Fetcher {
	return &Fetcher{
		client: &http.Client{
			Timeout:   2 * time.Second,
			Transport: fn,
		},
		UserAgent:    "test-agent",
		MaxRedirects: 10,
		MaxBodyBytes: 1024 * 1024,
	}
}

func testResponse(req *http.Request, status int, header http.Header, body string) *http.Response {
	return &http.Response{
		StatusCode: status,
		Header:     header,
		Body:       io.NopCloser(strings.NewReader(body)),
		Request:    req,
	}
}
