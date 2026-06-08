package dart

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/chetbackiewicz/water-data/backend/internal/config"
	"github.com/chetbackiewicz/water-data/backend/internal/sources"
)

const bonnevilleChinookKey = "dart_bonneville_chinook"

type bonnevilleChinook struct{}

func Register(reg *sources.Registry, cfg *config.Config) {
	_ = cfg
	reg.Register(bonnevilleChinook{})
}

func (s bonnevilleChinook) Key() string { return bonnevilleChinookKey }

func (s bonnevilleChinook) Description() string {
	return "Columbia River DART Bonneville Dam adult Chinook passage"
}

func (s bonnevilleChinook) AuthRequired() bool { return false }

func (s bonnevilleChinook) Probe(ctx context.Context, client *http.Client) sources.ProbeResult {
	start := time.Now()
	endpoint := "https://www.cbr.washington.edu/dart/query/adult_graph_text"
	year := time.Now().UTC().Year()
	attempted := []string{}

	body, status, err := postDART(ctx, client, endpoint, year)
	attempted = append(attempted, fmt.Sprintf("%s year=%d", endpoint, year))
	validErr := validateCSV(body)
	if err == nil && validErr != nil {
		previous := year - 1
		prevBody, prevStatus, prevErr := postDART(ctx, client, endpoint, previous)
		attempted = append(attempted, fmt.Sprintf("%s year=%d", endpoint, previous))
		if prevErr == nil && validateCSV(prevBody) == nil {
			year = previous
			body = prevBody
			status = prevStatus
			err = nil
			validErr = nil
		} else {
			body = prevBody
			status = prevStatus
			err = prevErr
			if err == nil {
				validErr = validateCSV(prevBody)
			}
		}
	}

	if err == nil && validErr != nil {
		err = fmt.Errorf("DART did not return valid CSV from attempted endpoints: %w", validErr)
	}

	summary := map[string]any{
		"dam":                        "BON",
		"species":                    "Chinook",
		"year":                       year,
		"attempted_endpoints":        attempted,
		"response_snippet_first_500": sources.Snippet(body, 500),
	}
	result := sources.MakeResult(s.Key(), endpoint, start, status, body, err, summary)
	if err == nil && status >= 200 && status < 300 {
		result.OK = true
	}
	if result.ParsedSummary == nil {
		if encoded, marshalErr := json.Marshal(summary); marshalErr == nil {
			result.ParsedSummary = encoded
		}
	}
	return result
}

func postDART(ctx context.Context, client *http.Client, endpoint string, year int) ([]byte, int, error) {
	form := url.Values{
		"outputFormat": {"csv"},
		"sc":           {"1"},
		"proj":         {"BON"},
		"species":      {"1"},
		"run":          {""},
		"startdate":    {"1/1"},
		"enddate":      {"12/31"},
		"year":         {fmt.Sprintf("%d", year)},
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewBufferString(form.Encode()))
	if err != nil {
		return nil, 0, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "text/csv, text/plain, */*")

	resp, err := client.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()
	body, readErr := io.ReadAll(io.LimitReader(resp.Body, 8*1024*1024))
	if readErr != nil {
		return body, resp.StatusCode, readErr
	}
	return body, resp.StatusCode, nil
}

func validateCSV(body []byte) error {
	trimmed := strings.TrimSpace(string(body))
	if trimmed == "" {
		return errors.New("empty response")
	}
	if strings.Contains(strings.ToLower(trimmed[:min(len(trimmed), 200)]), "<html") {
		return errors.New("HTML response")
	}
	reader := csv.NewReader(strings.NewReader(trimmed))
	reader.FieldsPerRecord = -1
	records, err := reader.ReadAll()
	if err != nil {
		return err
	}
	if len(records) < 2 {
		return errors.New("CSV contained no data rows")
	}
	return nil
}
