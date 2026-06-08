package inaturalist

import (
	"context"
	"net/http"
	"time"

	"github.com/chetbackiewicz/water-data/backend/internal/config"
	"github.com/chetbackiewicz/water-data/backend/internal/sources"
)

const endpointURL = "https://api.inaturalist.org/v1/taxa?q=lingcod&per_page=5"

type source struct{}

type taxaResponse struct {
	TotalResults int        `json:"total_results"`
	Results      []taxonHit `json:"results"`
}

type taxonHit struct {
	ID                  int          `json:"id"`
	Name                string       `json:"name"`
	Rank                string       `json:"rank"`
	PreferredCommonName string       `json:"preferred_common_name"`
	ObservationsCount   int          `json:"observations_count"`
	DefaultPhoto        defaultPhoto `json:"default_photo"`
	WikipediaURL        string       `json:"wikipedia_url"`
}

type defaultPhoto struct {
	SquareURL   string `json:"square_url"`
	LicenseCode string `json:"license_code"`
	Attribution string `json:"attribution"`
}

type topMatchSummary struct {
	ID                  int    `json:"id"`
	Name                string `json:"name"`
	PreferredCommonName string `json:"preferred_common_name"`
	ObservationsCount   int    `json:"observations_count"`
	DefaultPhotoLicense string `json:"default_photo_license"`
}

type summary struct {
	TotalResults int              `json:"total_results"`
	TopMatch     *topMatchSummary `json:"top_match,omitempty"`
}

func Register(reg *sources.Registry, cfg *config.Config) {
	_ = cfg
	reg.Register(source{})
}

func (source) Key() string { return "inat_lingcod" }

func (source) Description() string { return "iNaturalist taxa search for lingcod." }

func (source) AuthRequired() bool { return false }

func (s source) Probe(ctx context.Context, client *http.Client) sources.ProbeResult {
	var resp taxaResponse
	start := time.Now()
	body, status, err := sources.DoJSON(ctx, client, http.MethodGet, endpointURL, nil, &resp)

	sum := summary{TotalResults: resp.TotalResults}
	if len(resp.Results) > 0 {
		top := resp.Results[0]
		sum.TopMatch = &topMatchSummary{
			ID:                  top.ID,
			Name:                top.Name,
			PreferredCommonName: top.PreferredCommonName,
			ObservationsCount:   top.ObservationsCount,
			DefaultPhotoLicense: top.DefaultPhoto.LicenseCode,
		}
	}

	return sources.MakeResult(s.Key(), endpointURL, start, status, body, err, sum)
}
