package usgs

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
	url         string
}

func (s source) Key() string { return s.key }

func (s source) Description() string { return s.description }

func (s source) AuthRequired() bool { return false }

func (s source) Probe(ctx context.Context, client *http.Client) sources.ProbeResult {
	start := time.Now()
	var resp waterMLResponse
	body, status, err := sources.DoJSON(ctx, client, http.MethodGet, s.url, nil, &resp)
	var summary any
	if err == nil && status >= 200 && status < 300 {
		summary = summarize(resp)
	}
	return sources.MakeResult(s.key, s.url, start, status, body, err, summary)
}

func Register(reg *sources.Registry, cfg *config.Config) {
	reg.Register(source{
		key:         "usgs_nwis_puget_sound_rivers",
		description: "USGS NWIS instantaneous flow and water temperature readings for Puget Sound rivers.",
		url:         "https://waterservices.usgs.gov/nwis/iv/?format=json&sites=12101500,12089500,12061500,12150800,12181000,12200500,12048000&parameterCd=00060,00010&siteStatus=active",
	})
}

type waterMLResponse struct {
	Value struct {
		TimeSeries []timeSeries `json:"timeSeries"`
	} `json:"value"`
}

type timeSeries struct {
	SourceInfo struct {
		SiteName string `json:"siteName"`
		SiteCode []struct {
			Value string `json:"value"`
		} `json:"siteCode"`
	} `json:"sourceInfo"`
	Variable struct {
		VariableCode []struct {
			Value string `json:"value"`
		} `json:"variableCode"`
	} `json:"variable"`
	Values []struct {
		Value []struct {
			Value    string `json:"value"`
			DateTime string `json:"dateTime"`
		} `json:"value"`
	} `json:"values"`
}

type readingSummary struct {
	SiteCode       string `json:"site_code"`
	SiteName       string `json:"site_name"`
	ParameterCode  string `json:"parameter_code"`
	LatestValue    string `json:"latest_value"`
	LatestDateTime string `json:"latest_dateTime"`
}

type summary struct {
	SitesReturned int              `json:"sites_returned"`
	Readings      []readingSummary `json:"readings"`
}

func summarize(resp waterMLResponse) summary {
	readings := make([]readingSummary, 0, 14)
	seenSites := make(map[string]struct{})
	for _, ts := range resp.Value.TimeSeries {
		if len(readings) == 14 {
			break
		}
		siteCode := ""
		if len(ts.SourceInfo.SiteCode) > 0 {
			siteCode = ts.SourceInfo.SiteCode[0].Value
			seenSites[siteCode] = struct{}{}
		}
		parameterCode := ""
		if len(ts.Variable.VariableCode) > 0 {
			parameterCode = ts.Variable.VariableCode[0].Value
		}
		latestValue := ""
		latestDateTime := ""
		if len(ts.Values) > 0 && len(ts.Values[0].Value) > 0 {
			values := ts.Values[0].Value
			latest := values[len(values)-1]
			latestValue = latest.Value
			latestDateTime = latest.DateTime
		}
		readings = append(readings, readingSummary{
			SiteCode:       siteCode,
			SiteName:       ts.SourceInfo.SiteName,
			ParameterCode:  parameterCode,
			LatestValue:    latestValue,
			LatestDateTime: latestDateTime,
		})
	}
	return summary{SitesReturned: len(seenSites), Readings: readings}
}
