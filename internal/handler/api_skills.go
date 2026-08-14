// api_skills.go — handlers JSON de /api/v1/skills.
package handler

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/SalvucciFacundo/portfolio-go/internal/domain"
)

// ListSkills devuelve todos los skills ordenados por position.
func (a *API) ListSkills(w http.ResponseWriter, r *http.Request) {
	skills, err := a.store.Skill.List(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	if skills == nil {
		skills = []domain.Skill{}
	}
	writeJSON(w, http.StatusOK, skills)
}

// CreateSkill valida el nombre, chequea duplicados (409) y crea el skill (201).
func (a *API) CreateSkill(w http.ResponseWriter, r *http.Request) {
	var skill domain.Skill
	if err := readJSON(r, &skill); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	skill.Name = strings.TrimSpace(skill.Name)
	if skill.Name == "" {
		writeError(w, http.StatusBadRequest, "name is required")
		return
	}

	exists, err := a.store.Skill.ExistsByName(r.Context(), skill.Name)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	if exists {
		writeError(w, http.StatusConflict, "skill already exists")
		return
	}

	id, err := a.store.Skill.Create(r.Context(), skill)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	writeJSON(w, http.StatusCreated, map[string]int64{"id": id})
}

// UpdateSkill modifica el skill {id}; 404 si no existe.
func (a *API) UpdateSkill(w http.ResponseWriter, r *http.Request) {
	id, err := idParam(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	var skill domain.Skill
	if err := readJSON(r, &skill); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	skill.ID = id
	skill.Name = strings.TrimSpace(skill.Name)
	if skill.Name == "" {
		writeError(w, http.StatusBadRequest, "name is required")
		return
	}

	affected, err := a.store.Skill.Update(r.Context(), skill)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	if !affected {
		writeError(w, http.StatusNotFound, "skill not found")
		return
	}
	writeJSON(w, http.StatusOK, map[string]int64{"id": id})
}

// DeleteSkill elimina el skill {id}; 204 en éxito.
func (a *API) DeleteSkill(w http.ResponseWriter, r *http.Request) {
	id, err := idParam(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	affected, err := a.store.Skill.Delete(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	if !affected {
		writeError(w, http.StatusNotFound, "skill not found")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// UploadSkillIcon recibe multipart icon (imagen), lo convierte a WebP en local,
// lo sube a Cloudinary y persiste la URL en el skill {id}. 404 si el skill no
// existe.
func (a *API) UploadSkillIcon(w http.ResponseWriter, r *http.Request) {
	id, err := idParam(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	if _, err := a.store.Skill.GetByID(r.Context(), id); err != nil {
		if errNoRows(err) {
			writeError(w, http.StatusNotFound, "skill not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 10<<20)
	_, header, err := r.FormFile("icon")
	if err != nil {
		writeError(w, http.StatusBadRequest, "missing icon file")
		return
	}

	url, _, err := a.uploadFile(r, header, fmt.Sprintf("skill-icon-%d", id))
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	if affected, err := a.store.Skill.SetIcon(r.Context(), id, url); err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	} else if !affected {
		writeError(w, http.StatusNotFound, "skill not found")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"icon_url": url})
}
