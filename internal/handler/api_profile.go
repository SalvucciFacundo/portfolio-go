// api_profile.go — handlers JSON de /api/v1/profile y /api/v1/socials.
package handler

import (
	"errors"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"

	"github.com/jackc/pgx/v5"

	"github.com/SalvucciFacundo/portfolio-go/internal/domain"
)

// profileUpdate son los campos editables del profile vía PUT. Los punteros
// permiten updates parciales sin pisar avatar_url/resume_url (que se manejan
// por uploads).
type profileUpdate struct {
	Name       *string `json:"name"`
	RoleEs     *string `json:"role_es"`
	RoleEn     *string `json:"role_en"`
	HeadlineEs *string `json:"headline_es"`
	HeadlineEn *string `json:"headline_en"`
	SummaryEs  *string `json:"summary_es"`
	SummaryEn  *string `json:"summary_en"`
	Email      *string `json:"email"`
}

// GetProfile devuelve el perfil completo: campos base + socials embebidos. Los
// skills quedan vacíos (el front los consume desde /api/v1/skills).
func (a *API) GetProfile(w http.ResponseWriter, r *http.Request) {
	profile, err := a.store.Profile.Get(r.Context())
	if err != nil {
		if errNoRows(err) {
			writeError(w, http.StatusNotFound, "profile not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	socials, err := a.store.Profile.ListSocials(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	profile.Socials = socials
	emptyProfileSlices(&profile)

	writeJSON(w, http.StatusOK, profile)
}

// UpdateProfile actualiza los campos base del perfil (name, role, headline,
// summary, email). Los uploads (avatar/cv) no se tocan aquí.
func (a *API) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	var upd profileUpdate
	if err := readJSON(r, &upd); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	current, err := a.store.Profile.Get(r.Context())
	if err != nil {
		if errNoRows(err) {
			writeError(w, http.StatusNotFound, "profile not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	if upd.Name != nil {
		current.Name = *upd.Name
	}
	if upd.RoleEs != nil {
		current.RoleEs = *upd.RoleEs
	}
	if upd.RoleEn != nil {
		current.RoleEn = *upd.RoleEn
	}
	if upd.HeadlineEs != nil {
		current.HeadlineEs = *upd.HeadlineEs
	}
	if upd.HeadlineEn != nil {
		current.HeadlineEn = *upd.HeadlineEn
	}
	if upd.SummaryEs != nil {
		current.SummaryEs = *upd.SummaryEs
	}
	if upd.SummaryEn != nil {
		current.SummaryEn = *upd.SummaryEn
	}
	if upd.Email != nil {
		current.Email = *upd.Email
	}

	if err := a.store.Profile.Update(r.Context(), current); err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

// UploadAvatar recibe multipart avatar, lo convierte a WebP en local, lo sube
// a Cloudinary y persiste la URL devuelta.
func (a *API) UploadAvatar(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 10<<20)
	_, header, err := r.FormFile("avatar")
	if err != nil {
		writeError(w, http.StatusBadRequest, "missing avatar file")
		return
	}

	url, _, err := a.uploadFile(r, header, "avatar")
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := a.store.Profile.SetAvatar(r.Context(), url); err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"avatar_url": url})
}

// UploadCV recibe multipart cv (PDF), lo sube a Cloudinary como raw y persiste
// la URL devuelta + el nombre original del archivo.
func (a *API) UploadCV(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 10<<20)
	_, header, err := r.FormFile("cv")
	if err != nil {
		writeError(w, http.StatusBadRequest, "missing cv file")
		return
	}

	url, filename, err := a.uploadFile(r, header, "cv")
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := a.store.Profile.SetResume(r.Context(), url, filename); err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"resume_url": url, "resume_filename": filename})
}

// GetCV redirige (302) a la URL del CV. Si resume_url apunta a Cloudinary raw,
// se inserta el flag fl_attachment:<resume_filename url-encoded> justo después
// de "/raw/upload/" para que el navegador descargue el archivo con su nombre
// original (Content-Disposition: attachment). URLs legacy locales
// (static/uploads) no tienen ese segmento y se redirigen tal cual.
func (a *API) GetCV(w http.ResponseWriter, r *http.Request) {
	profile, err := a.store.Profile.Get(r.Context())
	if err != nil {
		if errNoRows(err) {
			writeError(w, http.StatusNotFound, "no resume set")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	if profile.ResumeURL == "" {
		writeError(w, http.StatusNotFound, "no resume set")
		return
	}
	target := profile.ResumeURL
	if profile.ResumeFilename != "" {
		target = cloudinaryDownloadURL(target, profile.ResumeFilename)
	}
	http.Redirect(w, r, target, http.StatusFound)
}

// cloudinaryDownloadURL inserta el flag fl_attachment:<filename> en una URL de
// entrega raw de Cloudinary. Es la forma robusta: no re-construye la URL desde
// el public_id (que ya queda embebido en resume_url tal como Cloudinary lo
// devolvió), solo ancla el flag con el nombre original guardado en la DB.
//
// El nombre del flag NO lleva extensión: Cloudinary la agrega según el formato
// del asset ("Mi CV 2026" → filename="Mi CV 2026.pdf").
func cloudinaryDownloadURL(rawURL, filename string) string {
	const marker = "/raw/upload/"
	i := strings.Index(rawURL, marker)
	if i < 0 {
		return rawURL
	}
	base := strings.TrimSuffix(filename, filepath.Ext(filename))
	flag := "fl_attachment:" + url.PathEscape(base)
	insertAt := i + len(marker)
	return rawURL[:insertAt] + flag + "/" + rawURL[insertAt:]
}

// UpdateSocials reemplaza todas las social links con el array enviado.
func (a *API) UpdateSocials(w http.ResponseWriter, r *http.Request) {
	var socials []domain.SocialLink
	if err := readJSON(r, &socials); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := a.store.Profile.SetSocials(r.Context(), socials); err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

// errNoRows reporta si err envuelve pgx.ErrNoRows.
func errNoRows(err error) bool {
	return errors.Is(err, pgx.ErrNoRows)
}
