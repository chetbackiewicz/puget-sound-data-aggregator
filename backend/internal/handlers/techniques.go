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

type TechniquesHandler struct{ Store *store.TechniqueStore }

func (h *TechniquesHandler) List(w http.ResponseWriter, r *http.Request) {
	spID, _ := strconv.Atoi(r.URL.Query().Get("species_id"))
	method := r.URL.Query().Get("method")
	list, err := h.Store.List(r.Context(), spID, method)
	if err != nil {
		httpx.WriteError(w, 500, err.Error())
		return
	}
	httpx.WriteJSON(w, 200, list)
}

func (h *TechniquesHandler) Create(w http.ResponseWriter, r *http.Request) {
	var t store.Technique
	if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
		httpx.WriteError(w, 400, err.Error())
		return
	}
	if t.SpeciesID == 0 || t.Title == "" {
		httpx.WriteError(w, 400, "species_id and title are required")
		return
	}
	id, err := h.Store.Create(r.Context(), t)
	if err != nil {
		httpx.WriteError(w, 500, err.Error())
		return
	}
	t.ID = id
	httpx.WriteJSON(w, 201, t)
}

func (h *TechniquesHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))
	var t store.Technique
	if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
		httpx.WriteError(w, 400, err.Error())
		return
	}
	t.ID = id
	if err := h.Store.Update(r.Context(), t); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			httpx.WriteError(w, 404, "not found")
			return
		}
		httpx.WriteError(w, 500, err.Error())
		return
	}
	httpx.WriteJSON(w, 200, t)
}

func (h *TechniquesHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))
	if err := h.Store.Delete(r.Context(), id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			httpx.WriteError(w, 404, "not found")
			return
		}
		httpx.WriteError(w, 500, err.Error())
		return
	}
	w.WriteHeader(204)
}
