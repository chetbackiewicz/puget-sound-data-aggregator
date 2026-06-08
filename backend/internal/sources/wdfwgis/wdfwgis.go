package wdfwgis

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/chetbackiewicz/water-data/backend/internal/config"
	"github.com/chetbackiewicz/water-data/backend/internal/sources"
)

const baseURL = "https://geodataservices.wdfw.wa.gov/arcgis/rest/services/"

type arcGISSource struct {
	key       string
	desc      string
	path      string
	count     string
	summarize func(geoJSON) any
}

type geoJSON struct {
	Features []feature `json:"features"`
}

type feature struct {
	Properties map[string]any `json:"properties"`
}

func Register(reg *sources.Registry, cfg *config.Config) {
	_ = cfg
	reg.Register(arcGISSource{
		key:       "wdfw_marine_areas",
		desc:      "WDFW recreational marine area boundaries",
		path:      "FP_Projects/Recreational_Marine_Area/MapServer/2/query",
		count:     "200",
		summarize: marineAreasSummary,
	})
	reg.Register(arcGISSource{
		key:       "wdfw_marine_protected_areas",
		desc:      "WDFW marine protected area boundaries",
		path:      "FP_FishMaps/MarineProtectedAreas/MapServer/3/query",
		count:     "200",
		summarize: marineProtectedAreasSummary,
	})
	reg.Register(arcGISSource{
		key:       "wdfw_crab_emergency_areas",
		desc:      "WDFW recreational crabbing emergency regulation areas",
		path:      "FP_FishMaps/RecreationalCrabbingEmergencyRegulationAreas/MapServer/2/query",
		count:     "200",
		summarize: crabEmergencyAreasSummary,
	})
	reg.Register(arcGISSource{
		key:       "wdfw_shore_fishing_sites",
		desc:      "WDFW shore fishing sites",
		path:      "FP_FishMaps/ShoreFishingSites/MapServer/0/query",
		count:     "50",
		summarize: shoreFishingSitesSummary,
	})
}

func (s arcGISSource) Key() string { return s.key }

func (s arcGISSource) Description() string { return s.desc }

func (s arcGISSource) AuthRequired() bool { return false }

func (s arcGISSource) Probe(ctx context.Context, client *http.Client) sources.ProbeResult {
	start := time.Now()
	endpoint := baseURL + s.path + "?where=1%3D1&outFields=*&f=geojson&outSR=4326&resultRecordCount=" + s.count

	var decoded geoJSON
	body, status, err := sources.DoJSON(ctx, client, http.MethodGet, endpoint, nil, &decoded)
	return sources.MakeResult(s.Key(), endpoint, start, status, body, err, s.summarize(decoded))
}

func marineAreasSummary(g geoJSON) any {
	return map[string]any{
		"features_returned": len(g.Features),
		"sample_properties": firstProperties(g),
	}
}

func marineProtectedAreasSummary(g geoJSON) any {
	return map[string]any{
		"features_returned": len(g.Features),
		"mpa_names":         valuesForKey(g, "MPA_Name", 10),
	}
}

func crabEmergencyAreasSummary(g geoJSON) any {
	return map[string]any{
		"features_returned": len(g.Features),
		"sample_properties": firstProperties(g),
	}
}

func shoreFishingSitesSummary(g geoJSON) any {
	return map[string]any{
		"features_returned": len(g.Features),
		"sample_site_names": sampleSiteNames(g, 10),
	}
}

func firstProperties(g geoJSON) map[string]any {
	if len(g.Features) == 0 || g.Features[0].Properties == nil {
		return map[string]any{}
	}
	return g.Features[0].Properties
}

func valuesForKey(g geoJSON, key string, limit int) []any {
	values := make([]any, 0, limit)
	for _, f := range g.Features {
		if v, ok := f.Properties[key]; ok && v != nil && v != "" {
			values = append(values, v)
			if len(values) == limit {
				break
			}
		}
	}
	return values
}

func sampleSiteNames(g geoJSON, limit int) []any {
	preferred := []string{"Name", "NAME", "SiteName", "SITE_NAME", "Site_Name", "Site", "SITE"}
	values := make([]any, 0, limit)
	for _, f := range g.Features {
		var picked any
		for _, key := range preferred {
			if v, ok := f.Properties[key]; ok && v != nil && v != "" {
				picked = v
				break
			}
		}
		if picked == nil {
			for key, v := range f.Properties {
				if strings.Contains(strings.ToLower(key), "name") && v != nil && v != "" {
					picked = v
					break
				}
			}
		}
		if picked != nil {
			values = append(values, picked)
			if len(values) == limit {
				break
			}
		}
	}
	return values
}
