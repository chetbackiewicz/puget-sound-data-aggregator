package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/chetbackiewicz/water-data/backend/internal/httpx"
	"github.com/chetbackiewicz/water-data/backend/internal/store"
	"github.com/go-chi/chi/v5"
)

type SpeciesHandler struct{ Store *store.SpeciesStore }

func (h *SpeciesHandler) List(w http.ResponseWriter, r *http.Request) {
	list, err := h.Store.List(r.Context())
	if err != nil {
		httpx.WriteError(w, 500, err.Error())
		return
	}
	httpx.WriteJSON(w, 200, list)
}

func (h *SpeciesHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))
	sp, err := h.Store.Get(r.Context(), id)
	if errors.Is(err, sql.ErrNoRows) {
		httpx.WriteError(w, 404, "not found")
		return
	} else if err != nil {
		httpx.WriteError(w, 500, err.Error())
		return
	}
	httpx.WriteJSON(w, 200, sp)
}

func (h *SpeciesHandler) Create(w http.ResponseWriter, r *http.Request) {
	var sp store.Species
	if err := json.NewDecoder(r.Body).Decode(&sp); err != nil {
		httpx.WriteError(w, 400, err.Error())
		return
	}
	if sp.ScientificName == "" || sp.CommonName == "" {
		httpx.WriteError(w, 400, "scientific_name and common_name are required")
		return
	}
	id, err := h.Store.Create(r.Context(), sp)
	if err != nil {
		httpx.WriteError(w, 500, err.Error())
		return
	}
	sp.ID = id
	httpx.WriteJSON(w, 201, sp)
}

func (h *SpeciesHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))
	var sp store.Species
	if err := json.NewDecoder(r.Body).Decode(&sp); err != nil {
		httpx.WriteError(w, 400, err.Error())
		return
	}
	sp.ID = id
	if err := h.Store.Update(r.Context(), sp); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			httpx.WriteError(w, 404, "not found")
			return
		}
		httpx.WriteError(w, 500, err.Error())
		return
	}
	httpx.WriteJSON(w, 200, sp)
}
