package crawler

import (
	"context"
	"errors"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/cars24/seo-automation/internal/models"
)

const linkValidationConcurrency = 10

// linkResult caches the validation outcome for a single external URL.
type linkResult struct {
	StatusCode int
	Timeout    bool
}

// ValidateExternalLinks performs concurrent HEAD requests against every unique
// external URL found across all pages. Results are fanned back onto each Link
// struct so the check layer can inspect StatusCode and Timeout.
func ValidateExternalLinks(ctx context.Context, pages []*models.PageData, ua string) {
	// Collect unique external URLs and map them to their Link pointers.
	urlLinks := make(map[string][]*models.Link)
	for _, p := range pages {
		for i := range p.Links {
			link := &p.Links[i]
			if link.IsInternal {
				continue
			}
			src := link.URL
			if src == "" || (!strings.HasPrefix(src, "http://") && !strings.HasPrefix(src, "https://")) {
				continue
			}
			urlLinks[src] = append(urlLinks[src], link)
		}
	}

	if len(urlLinks) == 0 {
		return
	}

	client := &http.Client{
		Timeout: 10 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			// Don't follow redirects — capture the 3xx status code.
			return http.ErrUseLastResponse
		},
	}

	results := make(map[string]*linkResult, len(urlLinks))
	var mu sync.Mutex
	var wg sync.WaitGroup
	sem := make(chan struct{}, linkValidationConcurrency)

	for url := range urlLinks {
		wg.Add(1)
		sem <- struct{}{}
		go func(target string) {
			defer wg.Done()
			defer func() { <-sem }()

			res := validateOneLink(ctx, client, target, ua)

			mu.Lock()
			results[target] = res
			mu.Unlock()
		}(url)
	}

	wg.Wait()

	// Fan results back onto every Link struct.
	for url, links := range urlLinks {
		res := results[url]
		if res == nil {
			continue
		}
		for _, link := range links {
			link.StatusCode = res.StatusCode
			link.Timeout = res.Timeout
		}
	}
}

// validateOneLink sends a HEAD request (falling back to GET on 405) and returns
// the result. Timeout vs other network errors are distinguished.
func validateOneLink(ctx context.Context, client *http.Client, target, ua string) *linkResult {
	req, err := http.NewRequestWithContext(ctx, http.MethodHead, target, nil)
	if err != nil {
		return &linkResult{}
	}
	req.Header.Set("User-Agent", ua)

	resp, err := client.Do(req)
	if err != nil {
		if isTimeoutErr(err) {
			return &linkResult{Timeout: true}
		}
		return &linkResult{}
	}
	resp.Body.Close()

	// Some servers reject HEAD — fall back to GET.
	if resp.StatusCode == http.StatusMethodNotAllowed {
		req2, err := http.NewRequestWithContext(ctx, http.MethodGet, target, nil)
		if err != nil {
			return &linkResult{StatusCode: resp.StatusCode}
		}
		req2.Header.Set("User-Agent", ua)

		resp2, err := client.Do(req2)
		if err != nil {
			if isTimeoutErr(err) {
				return &linkResult{Timeout: true}
			}
			return &linkResult{}
		}
		resp2.Body.Close()
		return &linkResult{StatusCode: resp2.StatusCode}
	}

	return &linkResult{StatusCode: resp.StatusCode}
}

// isTimeoutErr checks whether an error is a network timeout or context deadline.
func isTimeoutErr(err error) bool {
	if errors.Is(err, context.DeadlineExceeded) {
		return true
	}
	var netErr net.Error
	if errors.As(err, &netErr) && netErr.Timeout() {
		return true
	}
	return false
}
