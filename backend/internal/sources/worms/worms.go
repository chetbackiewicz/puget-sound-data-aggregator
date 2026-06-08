package worms

import (
	"context"
	"net/http"
	"time"

	"github.com/chetbackiewicz/water-data/backend/internal/config"
	"github.com/chetbackiewicz/water-data/backend/internal/sources"
)

const endpointURL = "https://www.marinespecies.org/rest/AphiaRecordsByMatchNames?scientificnames[]=Ophiodon+elongatus&marine_only=true"

type source struct{}

type aphiaRecord struct {
	AphiaID        int    `json:"AphiaID"`
	ScientificName string `json:"scientificname"`
	Authority      string `json:"authority"`
	Status         string `json:"status"`
	Citation       string `json:"citation"`
}

type firstMatchSummary struct {
	AphiaID        int    `json:"AphiaID"`
	ScientificName string `json:"scientificname"`
	Status         string `json:"status"`
	Authority      string `json:"authority"`
	Citation       string `json:"citation"`
}

type summary struct {
	MatchesCount int                `json:"matches_count"`
	FirstMatch   *firstMatchSummary `json:"first_match,omitempty"`
}

func Register(reg *sources.Registry, cfg *config.Config) {
	_ = cfg
	reg.Register(source{})
}

func (source) Key() string { return "worms_lingcod" }

func (source) Description() string {
	return "WoRMS Aphia match for lingcod (Ophiodon elongatus)."
}

func (source) AuthRequired() bool { return false }

func (s source) Probe(ctx context.Context, client *http.Client) sources.ProbeResult {
	var records [][]aphiaRecord
	start := time.Now()
	body, status, err := sources.DoJSON(ctx, client, http.MethodGet, endpointURL, nil, &records)

	sum := summary{}
	for _, group := range records {
		sum.MatchesCount += len(group)
		if sum.FirstMatch == nil && len(group) > 0 {
			first := group[0]
			sum.FirstMatch = &firstMatchSummary{
				AphiaID:        first.AphiaID,
				ScientificName: first.ScientificName,
				Status:         first.Status,
				Authority:      first.Authority,
				Citation:       first.Citation,
			}
		}
	}

	return sources.MakeResult(s.Key(), endpointURL, start, status, body, err, sum)
}
