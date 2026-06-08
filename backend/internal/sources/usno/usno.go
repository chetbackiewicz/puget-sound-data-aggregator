package usno

import (
	"context"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/chetbackiewicz/water-data/backend/internal/config"
	"github.com/chetbackiewicz/water-data/backend/internal/sources"
)

const sunMoonSeattleKey = "usno_sun_moon_seattle"

type sunMoonSeattle struct {
	appID string
}

type rsttResponse struct {
	Properties struct {
		Data struct {
			CurPhase  string `json:"curphase"`
			FracIllum string `json:"fracillum"`
			SunData   []phen `json:"sundata"`
			MoonData  []phen `json:"moondata"`
		} `json:"data"`
	} `json:"properties"`
}

type phen struct {
	Phen string `json:"phen"`
	Time string `json:"time"`
}

func Register(reg *sources.Registry, cfg *config.Config) {
	appID := ""
	if cfg != nil {
		appID = cfg.USNOAppID
	}
	reg.Register(sunMoonSeattle{appID: appID})
}

func (s sunMoonSeattle) Key() string { return sunMoonSeattleKey }

func (s sunMoonSeattle) Description() string {
	return "USNO one-day sun and moon rise, set, and transit data for Seattle"
}

func (s sunMoonSeattle) AuthRequired() bool { return false }

func (s sunMoonSeattle) Probe(ctx context.Context, client *http.Client) sources.ProbeResult {
	start := time.Now()
	date := time.Now().UTC().Format("2006-01-02")
	endpoint := "https://aa.usno.navy.mil/api/rstt/oneday?date=" + url.QueryEscape(date) +
		"&coords=47.6,-122.3&tz=-7&ID=" + url.QueryEscape(s.appID)

	var decoded rsttResponse
	body, status, err := sources.DoJSON(ctx, client, http.MethodGet, endpoint, nil, &decoded)

	summary := map[string]any{
		"date":               date,
		"moon_phase":         decoded.Properties.Data.CurPhase,
		"illumination_pct":   decoded.Properties.Data.FracIllum,
		"sunrise":            findPhen(decoded.Properties.Data.SunData, "rise"),
		"sunset":             findPhen(decoded.Properties.Data.SunData, "set"),
		"moonrise":           findPhen(decoded.Properties.Data.MoonData, "rise"),
		"moonset":            findPhen(decoded.Properties.Data.MoonData, "set"),
		"moon_upper_transit": findPhen(decoded.Properties.Data.MoonData, "upper"),
	}
	return sources.MakeResult(s.Key(), endpoint, start, status, body, err, summary)
}

func findPhen(items []phen, needle string) string {
	needle = strings.ToLower(needle)
	for _, item := range items {
		if strings.Contains(strings.ToLower(item.Phen), needle) {
			return item.Time
		}
	}
	return ""
}
