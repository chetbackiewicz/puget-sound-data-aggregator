package wikipedia

import (
	"context"
	"net/http"
	"sort"
	"time"

	"github.com/chetbackiewicz/water-data/backend/internal/config"
	"github.com/chetbackiewicz/water-data/backend/internal/sources"
)

const (
	endpointURL = "https://en.wikipedia.org/w/api.php?action=query&titles=Lingcod&prop=extracts&exintro=1&explaintext=1&format=json"
	userAgent   = "puget-sound-fishing-app/0.1 (contact@example.com)"
)

type source struct{}

type queryResponse struct {
	Query struct {
		Pages map[string]page `json:"pages"`
	} `json:"query"`
}

type page struct {
	PageID  int    `json:"pageid"`
	NS      int    `json:"ns"`
	Title   string `json:"title"`
	Extract string `json:"extract"`
}

type summary struct {
	PageID               int    `json:"page_id"`
	Title                string `json:"title"`
	ExtractLength        int    `json:"extract_length"`
	ExtractFirst500Chars string `json:"extract_first_500_chars"`
}

func Register(reg *sources.Registry, cfg *config.Config) {
	_ = cfg
	reg.Register(source{})
}

func (source) Key() string { return "wikipedia_lingcod" }

func (source) Description() string { return "Wikipedia introductory extract for Lingcod." }

func (source) AuthRequired() bool { return false }

func (s source) Probe(ctx context.Context, client *http.Client) sources.ProbeResult {
	var resp queryResponse
	start := time.Now()
	headers := http.Header{"User-Agent": []string{userAgent}}
	body, status, err := sources.DoJSON(ctx, client, http.MethodGet, endpointURL, headers, &resp)

	sum := summary{}
	if p, ok := firstPage(resp.Query.Pages); ok {
		sum = summary{
			PageID:               p.PageID,
			Title:                p.Title,
			ExtractLength:        len(p.Extract),
			ExtractFirst500Chars: firstNRunes(p.Extract, 500),
		}
	}

	return sources.MakeResult(s.Key(), endpointURL, start, status, body, err, sum)
}

func firstPage(pages map[string]page) (page, bool) {
	if len(pages) == 0 {
		return page{}, false
	}
	keys := make([]string, 0, len(pages))
	for key := range pages {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return pages[keys[0]], true
}

func firstNRunes(s string, n int) string {
	runes := []rune(s)
	if len(runes) <= n {
		return s
	}
	return string(runes[:n])
}
