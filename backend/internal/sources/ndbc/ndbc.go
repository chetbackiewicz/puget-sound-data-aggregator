package ndbc

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/chetbackiewicz/water-data/backend/internal/config"
	"github.com/chetbackiewicz/water-data/backend/internal/sources"
)

type stationSource struct {
	key         string
	description string
	station     string
}

func (s stationSource) Key() string         { return s.key }
func (s stationSource) Description() string { return s.description }
func (s stationSource) AuthRequired() bool  { return false }
func (s stationSource) Probe(ctx context.Context, client *http.Client) sources.ProbeResult {
	start := time.Now()
	url := fmt.Sprintf("https://www.ndbc.noaa.gov/data/realtime2/%s.txt", s.station)
	raw, status, err := sources.DoRaw(ctx, client, http.MethodGet, url, nil)
	var summary any
	if err == nil && status >= 200 && status < 300 {
		summary, err = parseObservation(s.station, string(raw))
	}
	return sources.MakeResult(s.key, url, start, status, raw, err, summary)
}

// Register adds NOAA NDBC real-time station probes for Puget Sound waters.
func Register(reg *sources.Registry, cfg *config.Config) {
	_ = cfg
	reg.Register(stationSource{key: "ndbc_wpow1", station: "WPOW1", description: "NDBC real-time wind and water observations for West Point, Seattle"})
	reg.Register(stationSource{key: "ndbc_sisw1", station: "SISW1", description: "NDBC real-time wind and water observations for Smith Island"})
	reg.Register(stationSource{key: "ndbc_46087", station: "46087", description: "NDBC real-time wind and wave observations for Neah Bay offshore buoy 46087"})
	reg.Register(stationSource{key: "ndbc_ptaw1", station: "PTAW1", description: "NDBC real-time wind and water observations for Port Angeles harbor"})
}

type observationSummary struct {
	Station            string   `json:"station"`
	ObservationTimeUTC string   `json:"observation_time_utc"`
	WindDirDeg         *int     `json:"wind_dir_deg"`
	WindSpeedMS        *float64 `json:"wind_speed_ms"`
	GustMS             *float64 `json:"gust_ms"`
	WaveHeightM        *float64 `json:"wave_height_m"`
	WavePeriodSec      *float64 `json:"wave_period_sec"`
	WaveDirDeg         *int     `json:"wave_dir_deg"`
	PressureHPA        *float64 `json:"pressure_hpa"`
	AirTempC           *float64 `json:"air_temp_c"`
	WaterTempC         *float64 `json:"water_temp_c"`
}

func parseObservation(station, body string) (observationSummary, error) {
	for _, line := range strings.Split(body, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 16 {
			return observationSummary{}, fmt.Errorf("parse ndbc row: got %d fields", len(fields))
		}
		obsTime, err := parseTime(fields[0:5])
		if err != nil {
			return observationSummary{}, err
		}
		return observationSummary{
			Station:            station,
			ObservationTimeUTC: obsTime.Format(time.RFC3339),
			WindDirDeg:         parseIntPtr(fields[5]),
			WindSpeedMS:        parseFloatPtr(fields[6]),
			GustMS:             parseFloatPtr(fields[7]),
			WaveHeightM:        parseFloatPtr(fields[8]),
			WavePeriodSec:      parseFloatPtr(fields[9]),
			WaveDirDeg:         parseIntPtr(fields[11]),
			PressureHPA:        parseFloatPtr(fields[12]),
			AirTempC:           parseFloatPtr(fields[13]),
			WaterTempC:         parseFloatPtr(fields[14]),
		}, nil
	}
	return observationSummary{}, fmt.Errorf("parse ndbc row: no observation rows")
}

func parseTime(parts []string) (time.Time, error) {
	vals := make([]int, len(parts))
	for i, p := range parts {
		v, err := strconv.Atoi(p)
		if err != nil {
			return time.Time{}, fmt.Errorf("parse observation time: %w", err)
		}
		vals[i] = v
	}
	return time.Date(vals[0], time.Month(vals[1]), vals[2], vals[3], vals[4], 0, 0, time.UTC), nil
}

func parseFloatPtr(s string) *float64 {
	if s == "MM" {
		return nil
	}
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return nil
	}
	return &v
}

func parseIntPtr(s string) *int {
	if s == "MM" {
		return nil
	}
	v, err := strconv.Atoi(s)
	if err != nil {
		return nil
	}
	return &v
}
