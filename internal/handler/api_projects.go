// api_projects.go — handlers JSON de /api/v1/projects y sus imágenes.
package handler

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/SalvucciFacundo/portfolio-go/internal/domain"
)

// ListProjects devuelve los proyectos ordenados por position (sin screenshots).
func (a *API) ListProjects(w http.ResponseWriter, r *http.Request) {
	projects, err := a.store.Project.List(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	if projects == nil {
		projects = []domain.Project{}
	}
	writeJSON(w, http.StatusOK, projects)
}

// GetProject devuelve el proyecto {id} con sus screenshots; 404 si no existe.
func (a *API) GetProject(w http.ResponseWriter, r *http.Request) {
	id, err := idParam(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	project, err := a.store.Project.GetByID(r.Context(), id)
	if err != nil {
		if errNoRows(err) {
			writeError(w, http.StatusNotFound, "project not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	if project.Screenshots == nil {
		project.Screenshots = []domain.ProjectImage{}
	}
	writeJSON(w, http.StatusOK, project)
}

// CreateProject valida los títulos, chequea duplicado por title_en (409) y crea
// el proyecto (201). Los screenshots en el body son opcionales.
func (a *API) CreateProject(w http.ResponseWriter, r *http.Request) {
	var project domain.Project
	if err := readJSON(r, &project); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	project.TitleEs = strings.TrimSpace(project.TitleEs)
	project.TitleEn = strings.TrimSpace(project.TitleEn)
	if project.TitleEs == "" || project.TitleEn == "" {
		writeError(w, http.StatusBadRequest, "title_es and title_en are required")
		return
	}
	if project.Tags == nil {
		project.Tags = []string{}
	}
	project.Link = normalizeURL(project.Link)
	project.RepoLink = normalizeURL(project.RepoLink)
	project.Status = normalizeStatus(project.Status)

	exists, err := a.store.Project.ExistsByTitleEn(r.Context(), project.TitleEn)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	if exists {
		writeError(w, http.StatusConflict, "project already exists")
		return
	}

	id, err := a.store.Project.Create(r.Context(), project)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	writeJSON(w, http.StatusCreated, map[string]int64{"id": id})
}

// UpdateProject modifica el proyecto {id}; 404 si no existe.
func (a *API) UpdateProject(w http.ResponseWriter, r *http.Request) {
	id, err := idParam(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	var project domain.Project
	if err := readJSON(r, &project); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	project.ID = id
	project.TitleEs = strings.TrimSpace(project.TitleEs)
	project.TitleEn = strings.TrimSpace(project.TitleEn)
	if project.TitleEs == "" || project.TitleEn == "" {
		writeError(w, http.StatusBadRequest, "title_es and title_en are required")
		return
	}
	if project.Tags == nil {
		project.Tags = []string{}
	}
	project.Link = normalizeURL(project.Link)
	project.RepoLink = normalizeURL(project.RepoLink)
	project.Status = normalizeStatus(project.Status)

	affected, err := a.store.Project.Update(r.Context(), project)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	if !affected {
		writeError(w, http.StatusNotFound, "project not found")
		return
	}
	writeJSON(w, http.StatusOK, map[string]int64{"id": id})
}

// DeleteProject elimina el proyecto {id}; 204 en éxito.
func (a *API) DeleteProject(w http.ResponseWriter, r *http.Request) {
	id, err := idParam(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	affected, err := a.store.Project.Delete(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	if !affected {
		writeError(w, http.StatusNotFound, "project not found")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// UploadProjectImages recibe multipart screenshots[] (varios archivos), los
// convierte a WebP en local, los sube a Cloudinary y los agrega al proyecto
// (201 con las URLs).
func (a *API) UploadProjectImages(w http.ResponseWriter, r *http.Request) {
	projectID, err := idParam(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	if _, err := a.store.Project.GetByID(r.Context(), projectID); err != nil {
		if errNoRows(err) {
			writeError(w, http.StatusNotFound, "project not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 20<<20)
	if err := r.ParseMultipartForm(20 << 20); err != nil {
		writeError(w, http.StatusBadRequest, "request too large")
		return
	}
	files := r.MultipartForm.File["screenshots"]
	if len(files) == 0 {
		writeError(w, http.StatusBadRequest, "no screenshots uploaded")
		return
	}

	urls := make([]string, 0, len(files))
	folder := fmt.Sprintf("portfolio/projects/%d", projectID)
	for _, fh := range files {
		url, _, err := a.uploadFile(r, fh, folder, "screenshot")
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		urls = append(urls, url)
	}

	if err := a.store.Project.AddImages(r.Context(), projectID, urls); err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	writeJSON(w, http.StatusCreated, map[string][]string{"urls": urls})
}

// DeleteProjectImage elimina el screenshot {imageId} del proyecto; 204 en éxito.
func (a *API) DeleteProjectImage(w http.ResponseWriter, r *http.Request) {
	imageID, err := idParam(r, "imageId")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid image id")
		return
	}

	affected, err := a.store.Project.DeleteImage(r.Context(), imageID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	if !affected {
		writeError(w, http.StatusNotFound, "image not found")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// UploadProjectCover recibe multipart cover (imagen), lo convierte a WebP en
// local, lo sube a Cloudinary y persiste la URL en el proyecto {id}. 404 si el
// proyecto no existe.
func (a *API) UploadProjectCover(w http.ResponseWriter, r *http.Request) {
	id, err := idParam(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	if _, err := a.store.Project.GetByID(r.Context(), id); err != nil {
		if errNoRows(err) {
			writeError(w, http.StatusNotFound, "project not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 10<<20)
	_, header, err := r.FormFile("cover")
	if err != nil {
		writeError(w, http.StatusBadRequest, "missing cover file")
		return
	}

	url, _, err := a.uploadFile(r, header, fmt.Sprintf("portfolio/projects/%d", id), "cover")
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	if affected, err := a.store.Project.SetCover(r.Context(), id, url); err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	} else if !affected {
		writeError(w, http.StatusNotFound, "project not found")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"cover_url": url})
}

// normalizeURL asegura que un link guardado sea una URL absoluta. Si el valor
// no trae protocolo (ej. "www.miscanarios.com.ar"), le antepone https:// —
// de otro modo el navegador lo interpretaría como URL relativa al dominio
// actual (https://facundosalvucci.dev/www.miscanarios.com.ar). Vacío pasa igual.
func normalizeURL(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	if strings.HasPrefix(raw, "http://") || strings.HasPrefix(raw, "https://") {
		return raw
	}
	return "https://" + raw
}

// normalizeStatus valida el estado del proyecto. Valores permitidos:
// production | development | demo. Vacío o desconocido → development.
func normalizeStatus(s string) string {
	switch strings.TrimSpace(s) {
	case "production", "development", "demo":
		return strings.TrimSpace(s)
	default:
		return "development"
	}
}
