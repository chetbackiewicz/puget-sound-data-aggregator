package datawagov

import (
	"context"
	"net/http"
	"sort"
	"time"

	"github.com/chetbackiewicz/water-data/backend/internal/config"
	"github.com/chetbackiewicz/water-data/backend/internal/sources"
)

const creelCatchKey = "datawa_creel_catch"

type creelCatch struct{}

func Register(reg *sources.Registry, cfg *config.Config) {
	_ = cfg
	reg.Register(creelCatch{})
}

func (s creelCatch) Key() string { return creelCatchKey }

func (s creelCatch) Description() string {
	return "data.wa.gov creel catch records"
}

func (s creelCatch) AuthRequired() bool { return false }

func (s creelCatch) Probe(ctx context.Context, client *http.Client) sources.ProbeResult {
	start := time.Now()
	endpoint := "https://data.wa.gov/resource/6y4e-8ftk.json?$limit=10"

	var decoded []map[string]any
	body, status, err := sources.DoJSON(ctx, client, http.MethodGet, endpoint, nil, &decoded)

	summary := map[string]any{
		"rows_returned":   len(decoded),
		"columns_present": columns(decoded),
		"sample_row":      sampleRow(decoded),
	}
	return sources.MakeResult(s.Key(), endpoint, start, status, body, err, summary)
}

func columns(rows []map[string]any) []string {
	if len(rows) == 0 {
		return []string{}
	}
	keys := make([]string, 0, len(rows[0]))
	for key := range rows[0] {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func sampleRow(rows []map[string]any) map[string]any {
	if len(rows) == 0 {
		return map[string]any{}
	}
	return rows[0]
}
