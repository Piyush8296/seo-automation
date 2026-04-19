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

const imageValidationConcurrency = 5

// ValidateImages performs concurrent HEAD requests against each image to
// populate FileSize, ContentType, and StatusCode.
func ValidateImages(ctx context.Context, images []models.Image, ua string) {
	if len(images) == 0 {
		return
	}

	client := &http.Client{
		Timeout: 5 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 5 {
				return http.ErrUseLastResponse
			}
			return nil
		},
	}

	var wg sync.WaitGroup
	sem := make(chan struct{}, imageValidationConcurrency)

	for i := range images {
		img := &images[i]
		src := img.Src
		if src == "" || (!strings.HasPrefix(src, "http://") && !strings.HasPrefix(src, "https://")) {
			continue
		}

		wg.Add(1)
		sem <- struct{}{}
		go func(img *models.Image, src string) {
			defer wg.Done()
			defer func() { <-sem }()
			validateOneImage(ctx, client, img, src, ua)
		}(img, src)
	}

	wg.Wait()
}

func validateOneImage(ctx context.Context, client *http.Client, img *models.Image, src, ua string) {
	req, err := http.NewRequestWithContext(ctx, http.MethodHead, src, nil)
	if err != nil {
		return
	}
	req.Header.Set("User-Agent", ua)

	resp, err := client.Do(req)
	if err != nil {
		return
	}
	resp.Body.Close()

	img.StatusCode = resp.StatusCode

	// Content-Length for file size
	if cl := resp.Header.Get("Content-Length"); cl != "" {
		if size, err := strconv.ParseInt(cl, 10, 64); err == nil {
			img.FileSize = size
		}
	}

	// Content-Type → format detection
	ct := strings.ToLower(resp.Header.Get("Content-Type"))
	img.ContentType = ct
	if img.Format == "" {
		img.Format = formatFromContentType(ct)
	}
}

// formatFromContentType maps a Content-Type header to a short format string.
func formatFromContentType(ct string) string {
	switch {
	case strings.Contains(ct, "image/jpeg"):
		return "jpg"
	case strings.Contains(ct, "image/png"):
		return "png"
	case strings.Contains(ct, "image/gif"):
		return "gif"
	case strings.Contains(ct, "image/webp"):
		return "webp"
	case strings.Contains(ct, "image/avif"):
		return "avif"
	case strings.Contains(ct, "image/svg"):
		return "svg"
	case strings.Contains(ct, "image/bmp"):
		return "bmp"
	case strings.Contains(ct, "image/tiff"):
		return "tiff"
	default:
		return ""
	}
}
