package handlers

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"time"

	"github.com/chetbackiewicz/water-data/backend/internal/httpx"
	"github.com/chetbackiewicz/water-data/backend/internal/sources"
	"github.com/chetbackiewicz/water-data/backend/internal/store"
	"github.com/go-chi/chi/v5"
)

type SourcesHandler struct {
	Registry *sources.Registry
	Store    *store.ProbeStore
	Client   *http.Client
}

func NewSourcesHandler(reg *sources.Registry, st *store.ProbeStore) *SourcesHandler {
	return &SourcesHandler{
		Registry: reg,
		Store:    st,
		Client:   &http.Client{Timeout: 60 * time.Second},
	}
}

type sourceInfo struct {
	Key          string                `json:"key"`
	Description  string                `json:"description"`
	AuthRequired bool                  `json:"auth_required"`
	LastProbe    *sources.ProbeResult  `json:"last_probe,omitempty"`
}

// GET /api/sources
func (h *SourcesHandler) List(w http.ResponseWriter, r *http.Request) {
	all := h.Registry.All()
	out := make([]sourceInfo, 0, len(all))
	for _, s := range all {
		info := sourceInfo{
			Key:          s.Key(),
			Description:  s.Description(),
			AuthRequired: s.AuthRequired(),
		}
		if last, err := h.Store.Latest(r.Context(), s.Key()); err == nil {
			info.LastProbe = &last
		}
		out = append(out, info)
	}
	httpx.WriteJSON(w, http.StatusOK, out)
}

// POST /api/probes/{key}
func (h *SourcesHandler) Run(w http.ResponseWriter, r *http.Request) {
	key := chi.URLParam(r, "key")
	src, ok := h.Registry.Get(key)
	if !ok {
		httpx.WriteError(w, http.StatusNotFound, "unknown source: "+key)
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 90*time.Second)
	defer cancel()
	result := src.Probe(ctx, h.Client)
	if _, err := h.Store.Save(r.Context(), result); err != nil {
		// Persisting is best-effort; log via header but still return result.
		w.Header().Set("X-Probe-Persist-Error", err.Error())
	}
	httpx.WriteJSON(w, http.StatusOK, result)
}

// GET /api/probes/{key}/latest
func (h *SourcesHandler) Latest(w http.ResponseWriter, r *http.Request) {
	key := chi.URLParam(r, "key")
	if _, ok := h.Registry.Get(key); !ok {
		httpx.WriteError(w, http.StatusNotFound, "unknown source: "+key)
		return
	}
	last, err := h.Store.Latest(r.Context(), key)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			httpx.WriteError(w, http.StatusNotFound, "no probe runs yet")
			return
		}
		httpx.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	httpx.WriteJSON(w, http.StatusOK, last)
}
