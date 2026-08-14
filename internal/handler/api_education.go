// api_education.go — handlers JSON de /api/v1/education.
package handler

import (
	"net/http"
	"strings"

	"github.com/SalvucciFacundo/portfolio-go/internal/domain"
)

// ListEducation devuelve toda la educación ordenada por position.
func (a *API) ListEducation(w http.ResponseWriter, r *http.Request) {
	items, err := a.store.Education.List(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	if items == nil {
		items = []domain.Education{}
	}
	writeJSON(w, http.StatusOK, items)
}

// CreateEducation valida title_es/title_en y crea (201).
func (a *API) CreateEducation(w http.ResponseWriter, r *http.Request) {
	var item domain.Education
	if err := readJSON(r, &item); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	item.TitleEs = strings.TrimSpace(item.TitleEs)
	item.TitleEn = strings.TrimSpace(item.TitleEn)
	if item.TitleEs == "" || item.TitleEn == "" {
		writeError(w, http.StatusBadRequest, "title_es and title_en are required")
		return
	}

	id, err := a.store.Education.Create(r.Context(), item)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	writeJSON(w, http.StatusCreated, map[string]int64{"id": id})
}

// UpdateEducation modifica la educación {id}; 404 si no existe.
func (a *API) UpdateEducation(w http.ResponseWriter, r *http.Request) {
	id, err := idParam(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	var item domain.Education
	if err := readJSON(r, &item); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	item.ID = id
	item.TitleEs = strings.TrimSpace(item.TitleEs)
	item.TitleEn = strings.TrimSpace(item.TitleEn)
	if item.TitleEs == "" || item.TitleEn == "" {
		writeError(w, http.StatusBadRequest, "title_es and title_en are required")
		return
	}

	affected, err := a.store.Education.Update(r.Context(), item)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	if !affected {
		writeError(w, http.StatusNotFound, "education not found")
		return
	}
	writeJSON(w, http.StatusOK, map[string]int64{"id": id})
}

// DeleteEducation elimina la educación {id}; 204 en éxito.
func (a *API) DeleteEducation(w http.ResponseWriter, r *http.Request) {
	id, err := idParam(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	affected, err := a.store.Education.Delete(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	if !affected {
		writeError(w, http.StatusNotFound, "education not found")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
