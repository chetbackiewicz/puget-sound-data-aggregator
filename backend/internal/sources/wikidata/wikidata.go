package wikidata

import (
	"context"
	"net/http"
	"sort"
	"time"

	"github.com/chetbackiewicz/water-data/backend/internal/config"
	"github.com/chetbackiewicz/water-data/backend/internal/sources"
)

const (
	endpointURL = "https://www.wikidata.org/wiki/Special:EntityData/Q81964.json"
	userAgent   = "puget-sound-fishing-app/0.1 (contact@example.com)"
)

type source struct{}

type entityResponse struct {
	Entities map[string]entity `json:"entities"`
}

type entity struct {
	ID           string                   `json:"id"`
	Labels       map[string]languageValue `json:"labels"`
	Descriptions map[string]languageValue `json:"descriptions"`
	Claims       map[string]any           `json:"claims"`
}

type languageValue struct {
	Language string `json:"language"`
	Value    string `json:"value"`
}

type summary struct {
	QID           string   `json:"qid"`
	LabelEN       string   `json:"label_en"`
	DescriptionEN string   `json:"description_en"`
	ClaimKeys     []string `json:"claim_keys"`
}

func Register(reg *sources.Registry, cfg *config.Config) {
	_ = cfg
	reg.Register(source{})
}

func (source) Key() string { return "wikidata_lingcod" }

func (source) Description() string { return "Wikidata entity data for lingcod (Q81964)." }

func (source) AuthRequired() bool { return false }

func (s source) Probe(ctx context.Context, client *http.Client) sources.ProbeResult {
	var resp entityResponse
	start := time.Now()
	headers := http.Header{"User-Agent": []string{userAgent}}
	body, status, err := sources.DoJSON(ctx, client, http.MethodGet, endpointURL, headers, &resp)

	sum := summary{}
	if ent, ok := resp.Entities["Q81964"]; ok {
		sum = summary{
			QID:           ent.ID,
			LabelEN:       ent.Labels["en"].Value,
			DescriptionEN: ent.Descriptions["en"].Value,
			ClaimKeys:     firstClaimKeys(ent.Claims, 10),
		}
	}

	return sources.MakeResult(s.Key(), endpointURL, start, status, body, err, sum)
}

func firstClaimKeys(claims map[string]any, limit int) []string {
	keys := make([]string, 0, len(claims))
	for key := range claims {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	if len(keys) > limit {
		keys = keys[:limit]
	}
	return keys
}
