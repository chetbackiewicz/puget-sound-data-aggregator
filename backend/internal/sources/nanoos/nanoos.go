package nanoos

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/chetbackiewicz/water-data/backend/internal/config"
	"github.com/chetbackiewicz/water-data/backend/internal/sources"
)

type source struct {
	key         string
	description string
	dataset     string
	url         string
}

func (s source) Key() string { return s.key }

func (s source) Description() string { return s.description }

func (s source) AuthRequired() bool { return false }

func (s source) Probe(ctx context.Context, client *http.Client) sources.ProbeResult {
	start := time.Now()
	body, status, err := sources.DoRaw(ctx, client, http.MethodGet, s.url, nil)
	var summary any
	if err == nil && status >= 200 && status < 300 {
		summary, err = parseERDDAPTable(s.dataset, body)
	}
	return sources.MakeResult(s.key, s.url, start, status, body, err, summary)
}

func Register(reg *sources.Registry, cfg *config.Config) {
	reg.Register(source{
		key:         "nanoos_bellinghambay_hydro",
		description: "NANOOS Bellingham Bay deep hydrographic station (temperature, salinity, oxygen).",
		dataset:     "bellinghambay_deephydro",
		url:         "https://erddap.nanoos.org/erddap/tabledap/bellinghambay_deephydro.json?time,depth,sea_water_temperature,sea_water_practical_salinity,mass_concentration_of_oxygen_in_sea_water&orderByLimit(%2210%22)",
	})
	reg.Register(source{
		key:         "nanoos_orca_met_twanoh",
		description: "NANOOS ORCA meteorological observations at Twanoh (Hood Canal).",
		dataset:     "ORCA_Twanoh",
		url:         "https://erddap.nanoos.org/erddap/tabledap/ORCA_Twanoh.json?time,air_temperature,wind_speed,wind_from_direction,surface_air_pressure&orderByLimit(%2210%22)",
	})
}

type erddapTableResponse struct {
	Table struct {
		ColumnNames []string `json:"columnNames"`
		ColumnTypes []string `json:"columnTypes"`
		Rows        [][]any  `json:"rows"`
	} `json:"table"`
}

type tableSummary struct {
	Dataset      string         `json:"dataset"`
	RowsReturned int            `json:"rows_returned"`
	Columns      []string       `json:"columns"`
	LatestRow    map[string]any `json:"latest_row,omitempty"`
}

func parseERDDAPTable(dataset string, body []byte) (any, error) {
	var resp erddapTableResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, err
	}
	summary := tableSummary{Dataset: dataset, RowsReturned: len(resp.Table.Rows), Columns: resp.Table.ColumnNames}
	if len(resp.Table.Rows) > 0 {
		summary.LatestRow = rowToMap(resp.Table.ColumnNames, resp.Table.Rows[len(resp.Table.Rows)-1])
	}
	return summary, nil
}

func rowToMap(columns []string, row []any) map[string]any {
	out := make(map[string]any, len(columns))
	for i, column := range columns {
		if i < len(row) {
			out[column] = row[i]
		}
	}
	return out
}
