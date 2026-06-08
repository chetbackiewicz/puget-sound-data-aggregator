package gbif

import (
	"context"
	"net/http"
	"time"

	"github.com/chetbackiewicz/water-data/backend/internal/config"
	"github.com/chetbackiewicz/water-data/backend/internal/sources"
)

const endpointURL = "https://api.gbif.org/v1/occurrence/search?taxonKey=2336521&country=US&stateProvince=Washington&limit=5"

type source struct{}

type occurrenceResponse struct {
	Offset       int                `json:"offset"`
	Limit        int                `json:"limit"`
	EndOfRecords bool               `json:"endOfRecords"`
	Count        int                `json:"count"`
	Results      []occurrenceRecord `json:"results"`
}

type occurrenceRecord struct {
	Key              int      `json:"key"`
	ScientificName   string   `json:"scientificName"`
	DecimalLatitude  *float64 `json:"decimalLatitude"`
	DecimalLongitude *float64 `json:"decimalLongitude"`
	StateProvince    string   `json:"stateProvince"`
	Year             *int     `json:"year"`
	Month            *int     `json:"month"`
	Day              *int     `json:"day"`
	EventDate        string   `json:"eventDate"`
	BasisOfRecord    string   `json:"basisOfRecord"`
	License          string   `json:"license"`
	VerbatimLocality string   `json:"verbatimLocality"`
}

type sampleRecord struct {
	DecimalLatitude  *float64 `json:"decimalLatitude,omitempty"`
	DecimalLongitude *float64 `json:"decimalLongitude,omitempty"`
	Year             *int     `json:"year,omitempty"`
	BasisOfRecord    string   `json:"basisOfRecord"`
	License          string   `json:"license"`
}

type summary struct {
	TotalCount      int            `json:"total_count"`
	ResultsReturned int            `json:"results_returned"`
	Sample          []sampleRecord `json:"sample"`
}

func Register(reg *sources.Registry, cfg *config.Config) {
	_ = cfg
	reg.Register(source{})
}

func (source) Key() string { return "gbif_lingcod_wa" }

func (source) Description() string { return "GBIF lingcod occurrences in Washington State." }

func (source) AuthRequired() bool { return false }

func (s source) Probe(ctx context.Context, client *http.Client) sources.ProbeResult {
	var resp occurrenceResponse
	start := time.Now()
	body, status, err := sources.DoJSON(ctx, client, http.MethodGet, endpointURL, nil, &resp)

	sum := summary{TotalCount: resp.Count, ResultsReturned: len(resp.Results)}
	limit := min(3, len(resp.Results))
	for _, record := range resp.Results[:limit] {
		sum.Sample = append(sum.Sample, sampleRecord{
			DecimalLatitude:  record.DecimalLatitude,
			DecimalLongitude: record.DecimalLongitude,
			Year:             record.Year,
			BasisOfRecord:    record.BasisOfRecord,
			License:          record.License,
		})
	}

	return sources.MakeResult(s.Key(), endpointURL, start, status, body, err, sum)
}
