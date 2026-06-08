package coops

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/chetbackiewicz/water-data/backend/internal/config"
	"github.com/chetbackiewicz/water-data/backend/internal/sources"
)

type source struct {
	key         string
	description string
	url         string
	urlFn       func() string
	parse       func([]byte) any
}

func (s source) Key() string { return s.key }

func (s source) Description() string { return s.description }

func (s source) AuthRequired() bool { return false }

func (s source) buildURL() string {
	if s.urlFn != nil {
		return s.urlFn()
	}
	return s.url
}

func (s source) Probe(ctx context.Context, client *http.Client) sources.ProbeResult {
	start := time.Now()
	var summary any
	url := s.buildURL()
	body, status, err := sources.DoJSON(ctx, client, http.MethodGet, url, nil, &summary)
	if err == nil && status >= 200 && status < 300 && s.parse != nil {
		summary = s.parse(body)
	} else {
		summary = nil
	}
	return sources.MakeResult(s.key, url, start, status, body, err, summary)
}

func today() string { return time.Now().UTC().Format("20060102") }

func Register(reg *sources.Registry, cfg *config.Config) {
	reg.Register(source{
		key:         "coops_water_level_seattle",
		description: "NOAA CO-OPS latest observed water level at Seattle (9447130).",
		url:         "https://api.tidesandcurrents.noaa.gov/api/prod/datagetter?product=water_level&station=9447130&date=latest&datum=MLLW&time_zone=lst_ldt&units=english&format=json&application=psfish",
		parse:       parseWaterLevel,
	})
	reg.Register(source{
		key:         "coops_tide_predictions_seattle",
		description: "NOAA CO-OPS seven-day high/low tide predictions at Seattle (9447130).",
		urlFn: func() string {
			return "https://api.tidesandcurrents.noaa.gov/api/prod/datagetter?product=predictions&station=9447130&begin_date=" + today() + "&range=168&datum=MLLW&time_zone=lst_ldt&interval=hilo&units=english&format=json&application=psfish"
		},
		parse: parseTidePredictions,
	})
	reg.Register(source{
		key:         "coops_currents_predictions_pug1501",
		description: "NOAA CO-OPS current predictions for Puget Sound station PUG1501.",
		urlFn: func() string {
			return "https://api.tidesandcurrents.noaa.gov/api/prod/datagetter?product=currents_predictions&station=PUG1501&begin_date=" + today() + "&range=72&interval=max_slack&units=english&time_zone=lst_ldt&format=json&application=psfish"
		},
		parse: parseCurrentPredictions,
	})
	reg.Register(source{
		key:         "coops_water_temperature_seattle",
		description: "NOAA CO-OPS water temperature observations at Seattle (9447130).",
		url:         "https://api.tidesandcurrents.noaa.gov/api/prod/datagetter?product=water_temperature&station=9447130&date=today&time_zone=lst_ldt&units=english&format=json&application=psfish",
		parse:       parseWaterTemperature,
	})
	reg.Register(source{
		key:         "coops_metadata_seattle",
		description: "NOAA CO-OPS station metadata for Seattle (9447130).",
		url:         "https://api.tidesandcurrents.noaa.gov/mdapi/prod/webapi/stations/9447130.json?expand=details,datums",
		parse:       parseMetadata,
	})
}

type metadata struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type datum struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type observation struct {
	T string `json:"t"`
	V string `json:"v"`
	S string `json:"s"`
	F string `json:"f"`
	Q string `json:"q"`
}

type waterLevelResponse struct {
	Metadata metadata      `json:"metadata"`
	Data     []observation `json:"data"`
}

type waterLevelSummary struct {
	Station      string   `json:"station"`
	Name         string   `json:"name"`
	ObservedAt   string   `json:"observed_at"`
	WaterLevelFT *float64 `json:"water_level_ft,omitempty"`
	Quality      string   `json:"quality"`
}

func parseWaterLevel(body []byte) any {
	var resp waterLevelResponse
	if err := jsonUnmarshal(body, &resp); err != nil {
		return nil
	}
	var latest observation
	if len(resp.Data) > 0 {
		latest = resp.Data[len(resp.Data)-1]
	}
	return waterLevelSummary{
		Station:      resp.Metadata.ID,
		Name:         resp.Metadata.Name,
		ObservedAt:   latest.T,
		WaterLevelFT: parseFloat(latest.V),
		Quality:      latest.Q,
	}
}

type tidePrediction struct {
	T    string `json:"t"`
	V    string `json:"v"`
	Type string `json:"type"`
}

type tidePredictionsResponse struct {
	Predictions []tidePrediction `json:"predictions"`
}

type tideEvent struct {
	T string   `json:"t"`
	V *float64 `json:"v,omitempty"`
}

type tidePredictionsSummary struct {
	Station          string     `json:"station"`
	PredictionsCount int        `json:"predictions_count"`
	NextHigh         *tideEvent `json:"next_high,omitempty"`
	NextLow          *tideEvent `json:"next_low,omitempty"`
}

func parseTidePredictions(body []byte) any {
	var resp tidePredictionsResponse
	if err := jsonUnmarshal(body, &resp); err != nil {
		return nil
	}
	summary := tidePredictionsSummary{Station: "9447130", PredictionsCount: len(resp.Predictions)}
	for _, p := range resp.Predictions {
		e := &tideEvent{T: p.T, V: parseFloat(p.V)}
		switch p.Type {
		case "H":
			if summary.NextHigh == nil {
				summary.NextHigh = e
			}
		case "L":
			if summary.NextLow == nil {
				summary.NextLow = e
			}
		}
	}
	return summary
}

type currentPrediction struct {
	Type          string `json:"Type"`
	Time          string `json:"Time"`
	VelocityMajor string `json:"Velocity_Major"`
	MeanFloodDir  string `json:"meanFloodDir"`
	MeanEbbDir    string `json:"meanEbbDir"`
	Bin           string `json:"Bin"`
	Depth         string `json:"Depth"`
}

type currentPredictionsResponse struct {
	CurrentPredictions struct {
		CP []currentPrediction `json:"cp"`
	} `json:"current_predictions"`
}

type currentEvent struct {
	Type       string   `json:"type"`
	Time       string   `json:"time"`
	VelocityKN *float64 `json:"velocity_kn,omitempty"`
}

type currentPredictionsSummary struct {
	Station     string        `json:"station"`
	EventsCount int           `json:"events_count"`
	NextEvent   *currentEvent `json:"next_event,omitempty"`
}

func parseCurrentPredictions(body []byte) any {
	var resp currentPredictionsResponse
	if err := jsonUnmarshal(body, &resp); err != nil {
		return nil
	}
	summary := currentPredictionsSummary{Station: "PUG1501", EventsCount: len(resp.CurrentPredictions.CP)}
	if len(resp.CurrentPredictions.CP) > 0 {
		p := resp.CurrentPredictions.CP[0]
		summary.NextEvent = &currentEvent{Type: p.Type, Time: p.Time, VelocityKN: parseFloat(p.VelocityMajor)}
	}
	return summary
}

type waterTemperatureResponse struct {
	Data []observation `json:"data"`
}

type tempLatest struct {
	T  string   `json:"t"`
	VF *float64 `json:"v_f,omitempty"`
}

type waterTemperatureSummary struct {
	Station           string      `json:"station"`
	ObservationsCount int         `json:"observations_count"`
	Latest            *tempLatest `json:"latest,omitempty"`
}

func parseWaterTemperature(body []byte) any {
	var resp waterTemperatureResponse
	if err := jsonUnmarshal(body, &resp); err != nil {
		return nil
	}
	summary := waterTemperatureSummary{Station: "9447130", ObservationsCount: len(resp.Data)}
	if len(resp.Data) > 0 {
		latest := resp.Data[len(resp.Data)-1]
		summary.Latest = &tempLatest{T: latest.T, VF: parseFloat(latest.V)}
	}
	return summary
}

type metadataDatumSummary struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type metadataSummary struct {
	StationID   string                 `json:"station_id"`
	Name        string                 `json:"name"`
	Lat         string                 `json:"lat"`
	Lon         string                 `json:"lon"`
	Established string                 `json:"established"`
	Datums      []metadataDatumSummary `json:"datums"`
}

func parseMetadata(body []byte) any {
	var root map[string]any
	if err := jsonUnmarshal(body, &root); err != nil {
		return nil
	}
	station := root
	if stations, ok := root["stations"].([]any); ok && len(stations) > 0 {
		if first, ok := stations[0].(map[string]any); ok {
			station = first
		}
	}
	datums := make([]metadataDatumSummary, 0, 5)
	if rawDatums, ok := station["datums"].([]any); ok {
		for _, rawDatum := range rawDatums {
			if len(datums) == 5 {
				break
			}
			if d, ok := rawDatum.(map[string]any); ok {
				datums = append(datums, metadataDatumSummary{Name: stringify(d["name"]), Value: stringify(d["value"])})
			}
		}
	}
	return metadataSummary{
		StationID:   firstNonEmpty(stringify(station["id"]), stringify(station["station_id"])),
		Name:        stringify(station["name"]),
		Lat:         firstNonEmpty(stringify(station["lat"]), stringify(station["latitude"])),
		Lon:         firstNonEmpty(stringify(station["lng"]), stringify(station["lon"]), stringify(station["longitude"])),
		Established: stringify(station["established"]),
		Datums:      datums,
	}
}

func parseFloat(s string) *float64 {
	if s == "" {
		return nil
	}
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return nil
	}
	return &v
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}
