package net

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/rs/zerolog"
)

type HTTPClient struct {
	client *http.Client
	logger zerolog.Logger
	debug  bool
}

func NewHTTPClient(logger zerolog.Logger, timeout time.Duration, debug bool) *HTTPClient {
	return &HTTPClient{
		client: &http.Client{Timeout: timeout},
		logger: logger,
		debug:  debug,
	}
}

func (c *HTTPClient) Do(ctx context.Context, method, url string, body []byte, headers map[string]string) ([]byte, int, error) {
	var bodyReader io.Reader
	if body != nil {
		bodyReader = bytes.NewReader(body)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		return nil, 0, fmt.Errorf("creating request: %w", err)
	}

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	if c.debug {
		c.logger.Debug().
			Str("method", method).
			Str("url", url).
			Msg("http client request")
	}

	start := time.Now()
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, fmt.Errorf("reading response body: %w", err)
	}

	if c.debug {
		c.logger.Debug().
			Str("method", method).
			Str("url", url).
			Int("status", resp.StatusCode).
			Dur("duration", time.Since(start)).
			Msg("http client response")
	}

	return respBody, resp.StatusCode, nil
}
