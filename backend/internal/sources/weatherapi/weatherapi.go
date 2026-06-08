package weatherapi

import (
	"context"
	"net/http"
	"net/url"
	"time"

	"github.com/chetbackiewicz/water-data/backend/internal/config"
	"github.com/chetbackiewicz/water-data/backend/internal/sources"
)

type source struct {
	cfg *config.Config
}

func (s source) Key() string { return "weatherapi_marine_pugetsound" }
func (s source) Description() string {
	return "WeatherAPI.com marine forecast probe for Puget Sound near Seattle"
}
func (s source) AuthRequired() bool { return true }
func (s source) Probe(ctx context.Context, client *http.Client) sources.ProbeResult {
	start := time.Now()
	if s.cfg.WeatherAPIKey == "" {
		return sources.ProbeResult{
			SourceKey:    s.Key(),
			EndpointURL:  "http://api.weatherapi.com/v1/marine.json?q=47.6,-122.3&days=3&tides=no",
			HTTPStatus:   0,
			DurationMS:   time.Since(start).Milliseconds(),
			OK:           false,
			ErrorMessage: "WEATHERAPI_KEY not set",
			FetchedAt:    time.Now().UTC(),
		}
	}
	endpoint := "http://api.weatherapi.com/v1/marine.json?key=" + url.QueryEscape(s.cfg.WeatherAPIKey) + "&q=47.6,-122.3&days=3&tides=no"
	var resp marineResponse
	raw, status, err := sources.DoJSON(ctx, client, http.MethodGet, endpoint, nil, &resp)
	summary := map[string]any{
		"location_name": resp.Location.Name,
		"forecast_days": len(resp.Forecast.ForecastDay),
	}
	if len(resp.Forecast.ForecastDay) > 0 {
		day := resp.Forecast.ForecastDay[0]
		first := map[string]any{
			"date":              day.Date,
			"sig_wave_height_m": nil,
			"swell_period_sec":  nil,
			"max_wind_kn":       maxWindKn(day.Day),
		}
		if len(day.Hour) > 0 {
			first["sig_wave_height_m"] = day.Hour[0].SigWaveHeightM
			first["swell_period_sec"] = day.Hour[0].SwellPeriodSec
		}
		summary["first_day"] = first
	}
	return sources.MakeResult(s.Key(), endpoint, start, status, raw, err, summary)
}

// Register adds the WeatherAPI.com marine forecast probe.
func Register(reg *sources.Registry, cfg *config.Config) {
	reg.Register(source{cfg: cfg})
}

type marineResponse struct {
	Location struct {
		Name string `json:"name"`
	} `json:"location"`
	Forecast struct {
		ForecastDay []struct {
			Date string    `json:"date"`
			Day  dayFields `json:"day"`
			Hour []struct {
				SigWaveHeightM *float64 `json:"sig_ht_mt"`
				SwellPeriodSec *float64 `json:"swell_period_secs"`
			} `json:"hour"`
		} `json:"forecastday"`
	} `json:"forecast"`
}

type dayFields struct {
	MaxWindMPH *float64 `json:"maxwind_mph"`
	MaxWindKPH *float64 `json:"maxwind_kph"`
}

func maxWindKn(day dayFields) any {
	if day.MaxWindMPH != nil {
		return *day.MaxWindMPH * 0.868976
	}
	if day.MaxWindKPH != nil {
		return *day.MaxWindKPH * 0.539957
	}
	return nil
}
