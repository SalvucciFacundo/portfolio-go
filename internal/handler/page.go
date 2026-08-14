package handler

import (
	"net/http"

	"github.com/SalvucciFacundo/portfolio-go/internal/adapters/db"
	"github.com/SalvucciFacundo/portfolio-go/internal/domain"
	"github.com/SalvucciFacundo/portfolio-go/internal/i18n"
	"github.com/SalvucciFacundo/portfolio-go/views/components"
	"github.com/SalvucciFacundo/portfolio-go/views/pages"
)

// PageHandler compone el perfil público desde el Store real (DB) y renderiza
// GET /. Las listas vacías se normalizan a [] en vez de null. El render sigue
// usando ?lang= (es por defecto); isAdmin se resuelve con el bridge HTMX.
func PageHandler(store *db.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		lang := "es"
		if v := r.URL.Query().Get("lang"); v == "en" {
			lang = "en"
		}

		profile, err := store.Profile.Get(r.Context())
		if err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}

		profile.Socials, err = store.Profile.ListSocials(r.Context())
		if err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		profile.Skills, err = store.Skill.List(r.Context())
		if err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		profile.Projects, err = store.Project.List(r.Context())
		if err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		profile.Experience, err = store.Experience.List(r.Context())
		if err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		profile.Education, err = store.Education.List(r.Context())
		if err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		normalizeEmpty(&profile)

		isAdmin := IsAdmin(r)

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_ = pages.Home(profile, lang, isAdmin).Render(r.Context(), w)
	}
}

// normalizeEmpty deja las listas del perfil como slices vacíos (no null)
// cuando la DB no tiene datos.
func normalizeEmpty(p *domain.Profile) {
	if p.Socials == nil {
		p.Socials = []domain.SocialLink{}
	}
	if p.Skills == nil {
		p.Skills = []domain.Skill{}
	}
	if p.Projects == nil {
		p.Projects = []domain.Project{}
	}
	if p.Experience == nil {
		p.Experience = []domain.Experience{}
	}
	if p.Education == nil {
		p.Education = []domain.Education{}
	}
}

func ContactHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	lang := "es"
	if v := r.URL.Query().Get("lang"); v == "en" {
		lang = "en"
	}

	email := r.FormValue("email")
	subject := r.FormValue("subject")
	message := r.FormValue("message")

	// Validate inputs
	if email == "" || subject == "" || message == "" {
		errMsg := i18n.T(lang, "contact.error")
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_ = components.ContactForm(lang, email, subject, message, errMsg, false).Render(r.Context(), w)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = components.ContactForm(lang, "", "", "", "", true).Render(r.Context(), w)
}
