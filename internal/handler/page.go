package handler

import (
	"net/http"

	"github.com/SalvucciFacundo/portfolio-go/internal/auth"
	"github.com/SalvucciFacundo/portfolio-go/internal/data"
	"github.com/SalvucciFacundo/portfolio-go/internal/i18n"
	"github.com/SalvucciFacundo/portfolio-go/views/components"
	"github.com/SalvucciFacundo/portfolio-go/views/pages"
)

func PageHandler(w http.ResponseWriter, r *http.Request) {
	lang := "es"
	if v := r.URL.Query().Get("lang"); v == "en" {
		lang = "en"
	}

	profile := data.GetProfile()
	isAdmin := auth.IsAdmin(r)

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = pages.Home(profile, lang, isAdmin).Render(r.Context(), w)
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
