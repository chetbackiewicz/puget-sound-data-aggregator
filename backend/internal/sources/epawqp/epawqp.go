package epawqp

import (
	"context"
	"encoding/csv"
	"net/http"
	"strings"
	"time"

	"github.com/chetbackiewicz/water-data/backend/internal/config"
	"github.com/chetbackiewicz/water-data/backend/internal/sources"
)

const waterTempKey = "epa_wqp_pugetsound_water_temp"

type waterTemp struct{}

func Register(reg *sources.Registry, cfg *config.Config) {
	_ = cfg
	reg.Register(waterTemp{})
}

func (s waterTemp) Key() string { return waterTempKey }

func (s waterTemp) Description() string {
	return "EPA Water Quality Portal Puget Sound water temperature results"
}

func (s waterTemp) AuthRequired() bool { return false }

func (s waterTemp) Probe(ctx context.Context, client *http.Client) sources.ProbeResult {
	start := time.Now()
	// WQP server-side processing scales with the unfiltered query, so we always
	// constrain by a narrow recent date window to keep the request under our
	// timeout. pageSize alone does NOT short-circuit server work.
	end := time.Now().UTC()
	begin := end.AddDate(0, -1, 0)
	endpoint := "https://www.waterqualitydata.us/data/Result/search?statecode=US%3A53&huc=17110019&characteristicName=Temperature%2C+water" +
		"&startDateLo=" + begin.Format("01-02-2006") +
		"&startDateHi=" + end.Format("01-02-2006") +
		"&mimeType=csv&zip=no&pageSize=20"

	body, status, err := sources.DoRaw(ctx, client, http.MethodGet, endpoint, nil)
	summary, parseErr := parseWaterTempCSV(body)
	if err == nil {
		err = parseErr
	}
	return sources.MakeResult(s.Key(), endpoint, start, status, body, err, summary)
}

func parseWaterTempCSV(body []byte) (map[string]any, error) {
	reader := csv.NewReader(strings.NewReader(string(body)))
	reader.FieldsPerRecord = -1
	records, err := reader.ReadAll()
	if err != nil {
		return map[string]any{
			"rows_returned":       0,
			"csv_header":          []string{},
			"sample_temperatures": []string{},
		}, err
	}
	if len(records) == 0 {
		return map[string]any{
			"rows_returned":       0,
			"csv_header":          []string{},
			"sample_temperatures": []string{},
		}, nil
	}
	header := records[0]
	valueIndex := indexOf(header, "ResultMeasureValue")
	temperatures := make([]string, 0, 5)
	dataRows := 0
	for _, row := range records[1:] {
		if len(row) == 0 {
			continue
		}
		dataRows++
		if valueIndex >= 0 && valueIndex < len(row) && row[valueIndex] != "" && len(temperatures) < 5 {
			temperatures = append(temperatures, row[valueIndex])
		}
		if dataRows == 10 {
			break
		}
	}
	return map[string]any{
		"rows_returned":       dataRows,
		"csv_header":          header,
		"sample_temperatures": temperatures,
	}, nil
}

func indexOf(values []string, target string) int {
	for i, value := range values {
		if value == target {
			return i
		}
	}
	return -1
}
