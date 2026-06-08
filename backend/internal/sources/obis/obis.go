package obis

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/chetbackiewicz/water-data/backend/internal/config"
	"github.com/chetbackiewicz/water-data/backend/internal/sources"
)

const endpointURL = "https://api.obis.org/v3/occurrence?scientificname=Ophiodon+elongatus&geometry=POLYGON%28%28-123.5+47.0%2C-121.5+47.0%2C-121.5+48.5%2C-123.5+48.5%2C-123.5+47.0%29%29&size=5"

type source struct{}

type occurrenceResponse struct {
	Total   int                `json:"total"`
	Results []occurrenceRecord `json:"results"`
}

type occurrenceRecord struct {
	DecimalLatitude  *float64 `json:"decimalLatitude"`
	DecimalLongitude *float64 `json:"decimalLongitude"`
	EventDate        string   `json:"eventDate"`
	InstitutionCode  string   `json:"institutionCode"`
	Locality         string   `json:"locality"`
	License          string   `json:"license"`
	SST              *float64 `json:"sst"`
	SSS              *float64 `json:"sss"`
	Bathymetry       *float64 `json:"bathymetry"`
	ShoreDistance    *float64 `json:"shoredistance"`
}

type sampleRecord struct {
	Locality   string   `json:"locality"`
	Year       *int     `json:"year,omitempty"`
	SST        *float64 `json:"sst,omitempty"`
	Bathymetry *float64 `json:"bathymetry,omitempty"`
}

type summary struct {
	Total           int            `json:"total"`
	ResultsReturned int            `json:"results_returned"`
	Sample          []sampleRecord `json:"sample"`
}

func Register(reg *sources.Registry, cfg *config.Config) {
	_ = cfg
	reg.Register(source{})
}

func (source) Key() string { return "obis_lingcod_pugetsound" }

func (source) Description() string {
	return "OBIS lingcod occurrences in the Puget Sound bounding box."
}

func (source) AuthRequired() bool { return false }

func (s source) Probe(ctx context.Context, client *http.Client) sources.ProbeResult {
	var resp occurrenceResponse
	start := time.Now()
	body, status, err := sources.DoJSON(ctx, client, http.MethodGet, endpointURL, nil, &resp)

	sum := summary{Total: resp.Total, ResultsReturned: len(resp.Results)}
	limit := min(3, len(resp.Results))
	for _, record := range resp.Results[:limit] {
		sum.Sample = append(sum.Sample, sampleRecord{
			Locality:   record.Locality,
			Year:       yearFromEventDate(record.EventDate),
			SST:        record.SST,
			Bathymetry: record.Bathymetry,
		})
	}

	return sources.MakeResult(s.Key(), endpointURL, start, status, body, err, sum)
}

func yearFromEventDate(eventDate string) *int {
	if len(eventDate) < 4 {
		return nil
	}
	year, err := strconv.Atoi(eventDate[:4])
	if err != nil {
		return nil
	}
	return &year
}
