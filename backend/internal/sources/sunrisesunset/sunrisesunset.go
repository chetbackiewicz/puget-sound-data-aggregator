package sunrisesunset

import (
	"context"
	"net/http"
	"time"

	"github.com/chetbackiewicz/water-data/backend/internal/config"
	"github.com/chetbackiewicz/water-data/backend/internal/sources"
)

const seattleKey = "sunrise_sunset_seattle"

type seattle struct{}

type response struct {
	Results struct {
		Sunrise                   string `json:"sunrise"`
		Sunset                    string `json:"sunset"`
		SolarNoon                 string `json:"solar_noon"`
		DayLength                 int    `json:"day_length"`
		CivilTwilightBegin        string `json:"civil_twilight_begin"`
		CivilTwilightEnd          string `json:"civil_twilight_end"`
		NauticalTwilightBegin     string `json:"nautical_twilight_begin"`
		NauticalTwilightEnd       string `json:"nautical_twilight_end"`
		AstronomicalTwilightBegin string `json:"astronomical_twilight_begin"`
		AstronomicalTwilightEnd   string `json:"astronomical_twilight_end"`
	} `json:"results"`
	Status string `json:"status"`
	TZID   string `json:"tzid"`
}

func Register(reg *sources.Registry, cfg *config.Config) {
	_ = cfg
	reg.Register(seattle{})
}

func (s seattle) Key() string { return seattleKey }

func (s seattle) Description() string {
	return "Sunrise-Sunset.org solar and twilight times for Seattle"
}

func (s seattle) AuthRequired() bool { return false }

func (s seattle) Probe(ctx context.Context, client *http.Client) sources.ProbeResult {
	start := time.Now()
	endpoint := "https://api.sunrise-sunset.org/json?lat=47.6&lng=-122.3&date=today&formatted=0"

	var decoded response
	body, status, err := sources.DoJSON(ctx, client, http.MethodGet, endpoint, nil, &decoded)

	summary := map[string]any{
		"sunrise":                     decoded.Results.Sunrise,
		"sunset":                      decoded.Results.Sunset,
		"solar_noon":                  decoded.Results.SolarNoon,
		"day_length":                  decoded.Results.DayLength,
		"day_length_seconds":          decoded.Results.DayLength,
		"civil_twilight_begin":        decoded.Results.CivilTwilightBegin,
		"civil_twilight_end":          decoded.Results.CivilTwilightEnd,
		"nautical_twilight_begin":     decoded.Results.NauticalTwilightBegin,
		"nautical_twilight_end":       decoded.Results.NauticalTwilightEnd,
		"astronomical_twilight_begin": decoded.Results.AstronomicalTwilightBegin,
		"astronomical_twilight_end":   decoded.Results.AstronomicalTwilightEnd,
		"status":                      decoded.Status,
		"tzid":                        decoded.TZID,
	}
	return sources.MakeResult(s.Key(), endpoint, start, status, body, err, summary)
}
