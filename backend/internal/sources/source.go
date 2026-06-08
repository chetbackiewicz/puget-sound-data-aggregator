// Package sources defines the core Source interface for upstream API probes.
//
// Each upstream data provider (NWS, NDBC, CO-OPS, etc.) implements one or
// more Source values and registers them with a Registry. The probe framework
// in internal/handlers consumes the Registry to enumerate available probes
// and to execute them on demand.
package sources

import (
	"context"
	"encoding/json"
	"net/http"
	"sync"
	"time"
)

// ProbeResult is the canonical output of a single Source.Probe call.
// It is both returned over the wire and persisted in api_probes.
type ProbeResult struct {
	SourceKey          string          `json:"source_key"`
	EndpointURL        string          `json:"endpoint_url"`
	HTTPStatus         int             `json:"http_status"`
	DurationMS         int64           `json:"duration_ms"`
	OK                 bool            `json:"ok"`
	ErrorMessage       string          `json:"error_message,omitempty"`
	RawResponseSnippet string          `json:"raw_response_snippet,omitempty"`
	ParsedSummary      json.RawMessage `json:"parsed_summary,omitempty"`
	FetchedAt          time.Time       `json:"fetched_at"`
}

// Source represents a single upstream API endpoint that can be probed.
type Source interface {
	Key() string
	Description() string
	AuthRequired() bool
	Probe(ctx context.Context, client *http.Client) ProbeResult
}

// Registry holds all registered Sources and is safe for concurrent use.
type Registry struct {
	mu      sync.RWMutex
	sources map[string]Source
}

func NewRegistry() *Registry {
	return &Registry{sources: make(map[string]Source)}
}

func (r *Registry) Register(s Source) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.sources[s.Key()] = s
}

func (r *Registry) Get(key string) (Source, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	s, ok := r.sources[key]
	return s, ok
}

func (r *Registry) All() []Source {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]Source, 0, len(r.sources))
	for _, s := range r.sources {
		out = append(out, s)
	}
	return out
}

// Snippet trims a byte slice to at most n bytes for raw_response_snippet storage.
func Snippet(b []byte, n int) string {
	if len(b) <= n {
		return string(b)
	}
	return string(b[:n]) + "...[truncated]"
}
