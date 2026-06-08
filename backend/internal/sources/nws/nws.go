package nws

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/chetbackiewicz/water-data/backend/internal/config"
	"github.com/chetbackiewicz/water-data/backend/internal/sources"
)

const (
	forecastKey = "nws_forecast_seattle"
	marineKey   = "nws_marine_cwf_sew"
	alertsKey   = "nws_alerts_pz135"
)

type source struct {
	key         string
	description string
	cfg         *config.Config
	probe       func(context.Context, *http.Client, *config.Config, string) sources.ProbeResult
}

func (s source) Key() string         { return s.key }
func (s source) Description() string { return s.description }
func (s source) AuthRequired() bool  { return false }
func (s source) Probe(ctx context.Context, client *http.Client) sources.ProbeResult {
	return s.probe(ctx, client, s.cfg, s.key)
}

// Register adds National Weather Service weather and marine forecast probes.
func Register(reg *sources.Registry, cfg *config.Config) {
	reg.Register(source{key: forecastKey, description: "NWS hourly forecast probe for Seattle via points grid metadata", cfg: cfg, probe: probeForecast})
	reg.Register(source{key: marineKey, description: "NWS Seattle Coastal Waters Forecast text product probe", cfg: cfg, probe: probeMarineCWF})
	reg.Register(source{key: alertsKey, description: "NWS active marine alerts probe for Puget Sound zone PZZ135", cfg: cfg, probe: probeAlerts})
}

func nwsHeaders(cfg *config.Config) http.Header {
	h := http.Header{}
	h.Set("User-Agent", cfg.NWSUserAgent)
	h.Set("Accept", "application/geo+json")
	return h
}

type pointsResponse struct {
	Properties struct {
		GridID         string `json:"gridId"`
		GridX          int    `json:"gridX"`
		GridY          int    `json:"gridY"`
		ForecastHourly string `json:"forecastHourly"`
	} `json:"properties"`
}

type forecastResponse struct {
	Properties struct {
		Periods []struct {
			StartTime     string `json:"startTime"`
			Temperature   int    `json:"temperature"`
			WindSpeed     string `json:"windSpeed"`
			WindDirection string `json:"windDirection"`
			ShortForecast string `json:"shortForecast"`
		} `json:"periods"`
	} `json:"properties"`
}

func probeForecast(ctx context.Context, client *http.Client, cfg *config.Config, key string) sources.ProbeResult {
	start := time.Now()
	pointsURL := "https://api.weather.gov/points/47.6062,-122.3321"
	var points pointsResponse
	raw, status, err := sources.DoJSON(ctx, client, http.MethodGet, pointsURL, nwsHeaders(cfg), &points)
	if err != nil || status < 200 || status >= 300 || points.Properties.ForecastHourly == "" {
		if err == nil && status >= 200 && status < 300 {
			err = fmt.Errorf("missing forecastHourly URL")
		}
		return sources.MakeResult(key, pointsURL, start, status, raw, err, nil)
	}

	var forecast forecastResponse
	raw, status, err = sources.DoJSON(ctx, client, http.MethodGet, points.Properties.ForecastHourly, nwsHeaders(cfg), &forecast)
	summary := map[string]any{
		"grid_id":              points.Properties.GridID,
		"grid_x":               points.Properties.GridX,
		"grid_y":               points.Properties.GridY,
		"hourly_periods_count": len(forecast.Properties.Periods),
	}
	if len(forecast.Properties.Periods) > 0 {
		p := forecast.Properties.Periods[0]
		summary["first_period"] = map[string]any{
			"start_time":     p.StartTime,
			"temperature":    p.Temperature,
			"wind_speed":     p.WindSpeed,
			"wind_direction": p.WindDirection,
			"short_forecast": p.ShortForecast,
		}
	}
	return sources.MakeResult(key, points.Properties.ForecastHourly, start, status, raw, err, summary)
}

type productListResponse struct {
	Graph []struct {
		AtID         string `json:"@id"`
		ID           string `json:"id"`
		IssuanceTime string `json:"issuanceTime"`
	} `json:"@graph"`
}

type productResponse struct {
	AtID         string `json:"@id"`
	ID           string `json:"id"`
	IssuanceTime string `json:"issuanceTime"`
	ProductText  string `json:"productText"`
}

func probeMarineCWF(ctx context.Context, client *http.Client, cfg *config.Config, key string) sources.ProbeResult {
	start := time.Now()
	listURL := "https://api.weather.gov/products/types/CWF/locations/SEW"
	var list productListResponse
	raw, status, err := sources.DoJSON(ctx, client, http.MethodGet, listURL, nwsHeaders(cfg), &list)
	if err != nil || status < 200 || status >= 300 || len(list.Graph) == 0 || list.Graph[0].AtID == "" {
		if err == nil && status >= 200 && status < 300 {
			err = fmt.Errorf("missing latest CWF product URL")
		}
		return sources.MakeResult(key, listURL, start, status, raw, err, nil)
	}

	var product productResponse
	productURL := list.Graph[0].AtID
	raw, status, err = sources.DoJSON(ctx, client, http.MethodGet, productURL, nwsHeaders(cfg), &product)
	latestID := product.ID
	if latestID == "" {
		latestID = list.Graph[0].ID
	}
	if latestID == "" {
		latestID = product.AtID
	}
	if latestID == "" {
		latestID = productURL
	}
	issuance := product.IssuanceTime
	if issuance == "" {
		issuance = list.Graph[0].IssuanceTime
	}
	summary := map[string]any{
		"latest_product_id": latestID,
		"issuance_time":     issuance,
		"text_length":       len(product.ProductText),
		"snippet_first_500": sources.Snippet([]byte(product.ProductText), 500),
	}
	return sources.MakeResult(key, productURL, start, status, raw, err, summary)
}

type alertsResponse struct {
	Features []struct {
		Properties struct {
			Event    string `json:"event"`
			Severity string `json:"severity"`
			Headline string `json:"headline"`
			Expires  string `json:"expires"`
		} `json:"properties"`
	} `json:"features"`
}

func probeAlerts(ctx context.Context, client *http.Client, cfg *config.Config, key string) sources.ProbeResult {
	start := time.Now()
	url := "https://api.weather.gov/alerts/active/zone/PZZ135"
	var alerts alertsResponse
	raw, status, err := sources.DoJSON(ctx, client, http.MethodGet, url, nwsHeaders(cfg), &alerts)
	items := make([]map[string]string, 0, min(len(alerts.Features), 5))
	for i, f := range alerts.Features {
		if i == 5 {
			break
		}
		items = append(items, map[string]string{
			"event":    f.Properties.Event,
			"severity": f.Properties.Severity,
			"headline": f.Properties.Headline,
			"expires":  f.Properties.Expires,
		})
	}
	summary := map[string]any{"active_alert_count": len(alerts.Features), "alerts": items}
	return sources.MakeResult(key, url, start, status, raw, err, summary)
}
