package handlers

import (
	"database/sql"
	"errors"
	"net/http"
	"strconv"

	"github.com/chetbackiewicz/water-data/backend/internal/httpx"
	"github.com/chetbackiewicz/water-data/backend/internal/store"
	"github.com/go-chi/chi/v5"
)

type MarineAreasHandler struct{ Store *store.MarineAreaStore }

func (h *MarineAreasHandler) List(w http.ResponseWriter, r *http.Request) {
	list, err := h.Store.List(r.Context())
	if err != nil {
		httpx.WriteError(w, 500, err.Error())
		return
	}
	httpx.WriteJSON(w, 200, list)
}

func (h *MarineAreasHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))
	ma, err := h.Store.Get(r.Context(), id)
	if errors.Is(err, sql.ErrNoRows) {
		httpx.WriteError(w, 404, "not found")
		return
	} else if err != nil {
		httpx.WriteError(w, 500, err.Error())
		return
	}
	httpx.WriteJSON(w, 200, ma)
}
