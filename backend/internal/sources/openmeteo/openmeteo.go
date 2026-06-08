package openmeteo

import (
	"context"
	"net/http"
	"time"

	"github.com/chetbackiewicz/water-data/backend/internal/config"
	"github.com/chetbackiewicz/water-data/backend/internal/sources"
)

type source struct {
	key         string
	description string
	probe       func(context.Context, *http.Client, string) sources.ProbeResult
}

func (s source) Key() string         { return s.key }
func (s source) Description() string { return s.description }
func (s source) AuthRequired() bool  { return false }
func (s source) Probe(ctx context.Context, client *http.Client) sources.ProbeResult {
	return s.probe(ctx, client, s.key)
}

// Register adds Open-Meteo marine and wind forecast probes for Seattle waters.
func Register(reg *sources.Registry, cfg *config.Config) {
	_ = cfg
	reg.Register(source{key: "openmeteo_marine_pugetsound", description: "Open-Meteo marine wave forecast probe for Puget Sound near Seattle", probe: probeMarine})
	reg.Register(source{key: "openmeteo_forecast_wind_seattle", description: "Open-Meteo 10 m wind forecast probe for Seattle", probe: probeWind})
}

type marineResponse struct {
	Hourly struct {
		Time          []string   `json:"time"`
		WaveHeight    []*float64 `json:"wave_height"`
		WaveDirection []*float64 `json:"wave_direction"`
		WavePeriod    []*float64 `json:"wave_period"`
	} `json:"hourly"`
}

func probeMarine(ctx context.Context, client *http.Client, key string) sources.ProbeResult {
	start := time.Now()
	query := "latitude=47.6062&longitude=-122.3321&hourly=wave_height,wave_direction,wave_period&forecast_days=3"
	url := "https://marine-api.open-meteo.com/v1/marine?" + query
	subdomain := "marine-api.open-meteo.com"
	var resp marineResponse
	raw, status, err := sources.DoJSON(ctx, client, http.MethodGet, url, nil, &resp)
	if err == nil && status == http.StatusNotFound {
		url = "https://api.open-meteo.com/v1/marine?" + query
		subdomain = "api.open-meteo.com"
		resp = marineResponse{}
		raw, status, err = sources.DoJSON(ctx, client, http.MethodGet, url, nil, &resp)
	}
	summary := map[string]any{"subdomain_used": subdomain, "hours_returned": len(resp.Hourly.Time)}
	if len(resp.Hourly.Time) > 0 {
		summary["first_hour"] = map[string]any{
			"time":               resp.Hourly.Time[0],
			"wave_height_m":      floatAt(resp.Hourly.WaveHeight, 0),
			"wave_direction_deg": floatAt(resp.Hourly.WaveDirection, 0),
			"wave_period_sec":    floatAt(resp.Hourly.WavePeriod, 0),
		}
	}
	return sources.MakeResult(key, url, start, status, raw, err, summary)
}

type windResponse struct {
	Hourly struct {
		Time          []string   `json:"time"`
		WindSpeed     []*float64 `json:"wind_speed_10m"`
		WindDirection []*float64 `json:"wind_direction_10m"`
		WindGusts     []*float64 `json:"wind_gusts_10m"`
	} `json:"hourly"`
}

func probeWind(ctx context.Context, client *http.Client, key string) sources.ProbeResult {
	start := time.Now()
	url := "https://api.open-meteo.com/v1/forecast?latitude=47.6062&longitude=-122.3321&hourly=wind_speed_10m,wind_direction_10m,wind_gusts_10m&wind_speed_unit=kn&timezone=America/Los_Angeles&forecast_days=3"
	var resp windResponse
	raw, status, err := sources.DoJSON(ctx, client, http.MethodGet, url, nil, &resp)
	summary := map[string]any{"hours_returned": len(resp.Hourly.Time)}
	if len(resp.Hourly.Time) > 0 {
		summary["first_hour"] = map[string]any{
			"time":               resp.Hourly.Time[0],
			"wind_speed_kn":      floatAt(resp.Hourly.WindSpeed, 0),
			"wind_direction_deg": floatAt(resp.Hourly.WindDirection, 0),
			"wind_gusts_kn":      floatAt(resp.Hourly.WindGusts, 0),
		}
	}
	return sources.MakeResult(key, url, start, status, raw, err, summary)
}

func floatAt(values []*float64, i int) any {
	if i < 0 || i >= len(values) {
		return nil
	}
	return values[i]
}
