package crawler

import (
	"context"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/cars24/seo-automation/internal/models"
)

const resourceValidationConcurrency = 10

// resourceResult caches the HEAD outcome for a single resource URL.
type resourceResult struct {
	StatusCode  int
	FileSize    int64
	ContentType string
}

// ValidateResources performs concurrent HEAD requests against every unique
// sub-resource URL across all pages. Results are fanned back onto each Resource
// struct so checks can inspect StatusCode/FileSize/ContentType.
func ValidateResources(ctx context.Context, pages []*models.PageData, ua string) {
	urlResources := make(map[string][]*models.Resource)
	for _, p := range pages {
		for i := range p.Resources {
			r := &p.Resources[i]
			src := r.URL
			if src == "" || (!strings.HasPrefix(src, "http://") && !strings.HasPrefix(src, "https://")) {
				continue
			}
			urlResources[src] = append(urlResources[src], r)
		}
	}

	if len(urlResources) == 0 {
		return
	}

	client := &http.Client{
		Timeout: 10 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 5 {
				return http.ErrUseLastResponse
			}
			return nil
		},
	}

	results := make(map[string]*resourceResult, len(urlResources))
	var mu sync.Mutex
	var wg sync.WaitGroup
	sem := make(chan struct{}, resourceValidationConcurrency)

	for url := range urlResources {
		wg.Add(1)
		sem <- struct{}{}
		go func(target string) {
			defer wg.Done()
			defer func() { <-sem }()

			res := validateOneResource(ctx, client, target, ua)

			mu.Lock()
			results[target] = res
			mu.Unlock()
		}(url)
	}

	wg.Wait()

	for url, rs := range urlResources {
		res := results[url]
		if res == nil {
			continue
		}
		for _, r := range rs {
			r.StatusCode = res.StatusCode
			r.FileSize = res.FileSize
			r.ContentType = res.ContentType
		}
	}
}

func validateOneResource(ctx context.Context, client *http.Client, target, ua string) *resourceResult {
	req, err := http.NewRequestWithContext(ctx, http.MethodHead, target, nil)
	if err != nil {
		return &resourceResult{}
	}
	req.Header.Set("User-Agent", ua)

	resp, err := client.Do(req)
	if err != nil {
		return &resourceResult{}
	}
	resp.Body.Close()

	// Fall back to GET on 405.
	if resp.StatusCode == http.StatusMethodNotAllowed {
		req2, err := http.NewRequestWithContext(ctx, http.MethodGet, target, nil)
		if err != nil {
			return &resourceResult{StatusCode: resp.StatusCode}
		}
		req2.Header.Set("User-Agent", ua)
		resp2, err := client.Do(req2)
		if err != nil {
			return &resourceResult{}
		}
		resp2.Body.Close()
		return resultFromResponse(resp2)
	}

	return resultFromResponse(resp)
}

func resultFromResponse(resp *http.Response) *resourceResult {
	out := &resourceResult{
		StatusCode:  resp.StatusCode,
		ContentType: strings.ToLower(resp.Header.Get("Content-Type")),
	}
	if cl := resp.Header.Get("Content-Length"); cl != "" {
		if size, err := strconv.ParseInt(cl, 10, 64); err == nil {
			out.FileSize = size
		}
	}
	return out
}
