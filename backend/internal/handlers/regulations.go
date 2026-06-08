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

type RegulationsHandler struct{ Store *store.RegulationStore }

func (h *RegulationsHandler) List(w http.ResponseWriter, r *http.Request) {
	maID, _ := strconv.Atoi(r.URL.Query().Get("marine_area_id"))
	spID, _ := strconv.Atoi(r.URL.Query().Get("species_id"))
	list, err := h.Store.List(r.Context(), maID, spID)
	if err != nil {
		httpx.WriteError(w, 500, err.Error())
		return
	}
	httpx.WriteJSON(w, 200, list)
}

func (h *RegulationsHandler) Create(w http.ResponseWriter, r *http.Request) {
	var reg store.Regulation
	if err := json.NewDecoder(r.Body).Decode(&reg); err != nil {
		httpx.WriteError(w, 400, err.Error())
		return
	}
	if reg.MarineAreaID == 0 || reg.SpeciesID == 0 {
		httpx.WriteError(w, 400, "marine_area_id and species_id are required")
		return
	}
	id, err := h.Store.Create(r.Context(), reg)
	if err != nil {
		httpx.WriteError(w, 500, err.Error())
		return
	}
	reg.ID = id
	httpx.WriteJSON(w, 201, reg)
}

func (h *RegulationsHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))
	var reg store.Regulation
	if err := json.NewDecoder(r.Body).Decode(&reg); err != nil {
		httpx.WriteError(w, 400, err.Error())
		return
	}
	reg.ID = id
	if err := h.Store.Update(r.Context(), reg); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			httpx.WriteError(w, 404, "not found")
			return
		}
		httpx.WriteError(w, 500, err.Error())
		return
	}
	httpx.WriteJSON(w, 200, reg)
}

func (h *RegulationsHandler) Delete(w http.ResponseWriter, r *http.Request) {
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
