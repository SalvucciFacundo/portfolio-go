// api_experience.go — handlers JSON de /api/v1/experience.
package handler

import (
	"net/http"
	"strings"

	"github.com/SalvucciFacundo/portfolio-go/internal/domain"
)

// ListExperience devuelve todas las experiencias ordenadas por position.
func (a *API) ListExperience(w http.ResponseWriter, r *http.Request) {
	items, err := a.store.Experience.List(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	if items == nil {
		items = []domain.Experience{}
	}
	writeJSON(w, http.StatusOK, items)
}

// CreateExperience valida company/position_es/position_en y crea (201).
func (a *API) CreateExperience(w http.ResponseWriter, r *http.Request) {
	var item domain.Experience
	if err := readJSON(r, &item); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	item.Company = strings.TrimSpace(item.Company)
	item.PositionEs = strings.TrimSpace(item.PositionEs)
	item.PositionEn = strings.TrimSpace(item.PositionEn)
	if item.Company == "" || item.PositionEs == "" || item.PositionEn == "" {
		writeError(w, http.StatusBadRequest, "company, position_es and position_en are required")
		return
	}

	id, err := a.store.Experience.Create(r.Context(), item)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	writeJSON(w, http.StatusCreated, map[string]int64{"id": id})
}

// UpdateExperience modifica la experiencia {id}; 404 si no existe.
func (a *API) UpdateExperience(w http.ResponseWriter, r *http.Request) {
	id, err := idParam(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	var item domain.Experience
	if err := readJSON(r, &item); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	item.ID = id
	item.Company = strings.TrimSpace(item.Company)
	item.PositionEs = strings.TrimSpace(item.PositionEs)
	item.PositionEn = strings.TrimSpace(item.PositionEn)
	if item.Company == "" || item.PositionEs == "" || item.PositionEn == "" {
		writeError(w, http.StatusBadRequest, "company, position_es and position_en are required")
		return
	}

	affected, err := a.store.Experience.Update(r.Context(), item)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	if !affected {
		writeError(w, http.StatusNotFound, "experience not found")
		return
	}
	writeJSON(w, http.StatusOK, map[string]int64{"id": id})
}

// DeleteExperience elimina la experiencia {id}; 204 en éxito.
func (a *API) DeleteExperience(w http.ResponseWriter, r *http.Request) {
	id, err := idParam(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	affected, err := a.store.Experience.Delete(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	if !affected {
		writeError(w, http.StatusNotFound, "experience not found")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
