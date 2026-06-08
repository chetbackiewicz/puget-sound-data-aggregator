package sources

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
)

// MaxSnippetBytes caps stored raw response per probe.
const MaxSnippetBytes = 64 * 1024

// MaxBodyBytes caps how much we read from an upstream response.
// Large enough for WDFW ArcGIS GeoJSON exports of full state layers.
const MaxBodyBytes = 32 * 1024 * 1024

// DoJSON performs a GET and unmarshals JSON into out. Returns raw body, status, and error.
// Caller decides what counts as success; this helper is non-opinionated.
func DoJSON(ctx context.Context, client *http.Client, method, url string, headers http.Header, out any) (raw []byte, status int, err error) {
	req, err := http.NewRequestWithContext(ctx, method, url, nil)
	if err != nil {
		return nil, 0, err
	}
	for k, vs := range headers {
		for _, v := range vs {
			req.Header.Add(k, v)
		}
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, MaxBodyBytes))
	if err != nil {
		return body, resp.StatusCode, err
	}
	if out != nil && len(body) > 0 {
		if err := json.Unmarshal(body, out); err != nil {
			return body, resp.StatusCode, fmt.Errorf("decode json: %w", err)
		}
	}
	return body, resp.StatusCode, nil
}

// DoRaw performs a GET and returns the raw body. Caller parses (e.g., NDBC text).
func DoRaw(ctx context.Context, client *http.Client, method, url string, headers http.Header) (raw []byte, status int, err error) {
	req, err := http.NewRequestWithContext(ctx, method, url, nil)
	if err != nil {
		return nil, 0, err
	}
	for k, vs := range headers {
		for _, v := range vs {
			req.Header.Add(k, v)
		}
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, MaxBodyBytes))
	return body, resp.StatusCode, err
}

// MakeResult is a convenience for building a ProbeResult from a fetch.
// summary is optional; if non-nil it is marshaled into ParsedSummary.
func MakeResult(key, url string, start time.Time, status int, body []byte, err error, summary any) ProbeResult {
	r := ProbeResult{
		SourceKey:   key,
		EndpointURL: url,
		HTTPStatus:  status,
		DurationMS:  time.Since(start).Milliseconds(),
		OK:          err == nil && status >= 200 && status < 300,
		FetchedAt:   time.Now().UTC(),
	}
	if err != nil {
		r.ErrorMessage = err.Error()
	} else if status >= 400 {
		r.ErrorMessage = fmt.Sprintf("HTTP %d", status)
	}
	if len(body) > 0 {
		r.RawResponseSnippet = Snippet(body, MaxSnippetBytes)
	}
	if summary != nil {
		if b, e := json.Marshal(summary); e == nil {
			r.ParsedSummary = b
		}
	}
	return r
}

// ErrSkipped indicates a probe was deliberately skipped (e.g., missing API key).
var ErrSkipped = errors.New("skipped")
