package httputil

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"sync"
)

// FetchAsDataURL downloads a single image and returns a data: URL.
func FetchAsDataURL(
	ctx context.Context,
	client *http.Client,
	rawURL string,
	maxBytes int64,
) (string, error) {
	if err := ValidateHTTPURL(rawURL); err != nil {
		return "", err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return "", err
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("http %d", resp.StatusCode)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, maxBytes+1))
	if err != nil {
		return "", err
	}
	if int64(len(body)) > maxBytes {
		return "", fmt.Errorf("response exceeds %d bytes", maxBytes)
	}
	return BytesToDataURL(body, resp.Header.Get("Content-Type")), nil
}

// FetchMapAsDataURLs downloads URLs concurrently; failed URLs are omitted from the result.
func FetchMapAsDataURLs(
	ctx context.Context,
	client *http.Client,
	urls []string,
	maxBytes int64,
	maxParallel int,
	onError func(url string, err error),
) map[string]string {
	if len(urls) == 0 {
		return nil
	}
	if maxParallel <= 0 {
		maxParallel = 8
	}

	sem := make(chan struct{}, maxParallel)
	var mu sync.Mutex
	out := make(map[string]string, len(urls))
	var wg sync.WaitGroup

	for _, rawURL := range urls {
		wg.Add(1)
		go func(u string) {
			defer wg.Done()
			if ctx.Err() != nil {
				return
			}
			sem <- struct{}{}
			defer func() { <-sem }()

			data, err := FetchAsDataURL(ctx, client, u, maxBytes)
			if err != nil {
				if onError != nil {
					onError(u, err)
				}
				return
			}
			mu.Lock()
			out[u] = data
			mu.Unlock()
		}(rawURL)
	}
	wg.Wait()
	return out
}
