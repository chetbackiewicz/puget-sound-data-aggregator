package coastwatch

import (
	"context"
	"encoding/csv"
	"errors"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/chetbackiewicz/water-data/backend/internal/config"
	"github.com/chetbackiewicz/water-data/backend/internal/sources"
)

type source struct {
	key         string
	description string
}

func (s source) Key() string { return s.key }

func (s source) Description() string { return s.description }

func (s source) AuthRequired() bool { return false }

func buildURL(ts string) string {
	return "https://coastwatch.noaa.gov/erddap/griddap/noaacwSNPPACSPOSSTL3GCDaily.csv?sea_surface_temperature[(" + ts + ")][0][(47.0):(49.0)][(-124.0):(-122.0)]"
}

var axisMaxRe = regexp.MustCompile(`even ([0-9T:\-Z]+)`)

func (s source) Probe(ctx context.Context, client *http.Client) sources.ProbeResult {
	start := time.Now()
	queryDate := time.Now().UTC().AddDate(0, 0, -2)
	queryDate = time.Date(queryDate.Year(), queryDate.Month(), queryDate.Day(), 12, 0, 0, 0, time.UTC)
	ts := queryDate.Format("2006-01-02T15:04:05Z")
	url := buildURL(ts)
	body, status, err := sources.DoRaw(ctx, client, http.MethodGet, url, nil)

	// If dataset is stale, the error message includes the real axis maximum. Retry once.
	if status == 404 && err == nil {
		if m := axisMaxRe.FindStringSubmatch(string(body)); len(m) == 2 {
			ts = m[1]
			url = buildURL(ts)
			body, status, err = sources.DoRaw(ctx, client, http.MethodGet, url, nil)
		}
	}

	var summary any
	if err == nil && status >= 200 && status < 300 {
		summary, err = summarizeCSV(ts, body)
	}
	return sources.MakeResult(s.key, url, start, status, body, err, summary)
}

func Register(reg *sources.Registry, cfg *config.Config) {
	reg.Register(source{key: "coastwatch_sst_pugetsound", description: "NOAA CoastWatch daily sea surface temperature grid sample for Puget Sound."})
}

type csvSummary struct {
	DateQueried     string   `json:"date_queried"`
	CSVRowsReturned int      `json:"csv_rows_returned"`
	CSVHeader       []string `json:"csv_header"`
	SampleSSTValues []string `json:"sample_sst_values"`
}

func summarizeCSV(dateString string, body []byte) (any, error) {
	text := string(body)
	if strings.Contains(strings.ToLower(text), "no matching results") {
		return csvSummary{DateQueried: dateString}, nil
	}
	records, err := csv.NewReader(strings.NewReader(text)).ReadAll()
	if err != nil {
		return nil, err
	}
	if len(records) == 0 {
		return nil, errors.New("empty csv response")
	}
	header := records[0]
	dataStart := 1
	if len(records) > 1 && looksLikeUnits(records[1]) {
		dataStart = 2
	}
	rows := records[dataStart:]
	if len(rows) > 10 {
		rows = rows[:10]
	}
	sstIndex := len(header) - 1
	for i, column := range header {
		if column == "sea_surface_temperature" {
			sstIndex = i
			break
		}
	}
	samples := make([]string, 0, 5)
	for _, row := range rows {
		if len(samples) == 5 {
			break
		}
		if sstIndex >= len(row) {
			continue
		}
		value := strings.TrimSpace(row[sstIndex])
		if value == "" || strings.EqualFold(value, "NaN") {
			continue
		}
		samples = append(samples, value)
	}
	return csvSummary{DateQueried: dateString, CSVRowsReturned: len(rows), CSVHeader: header, SampleSSTValues: samples}, nil
}

func looksLikeUnits(record []string) bool {
	for _, value := range record {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if strings.Contains(value, "degrees") || strings.Contains(value, "UTC") || strings.Contains(value, "m") {
			return true
		}
	}
	return false
}


