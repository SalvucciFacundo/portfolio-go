package handler

import (
	"net/http"

	"github.com/SalvucciFacundo/portfolio-go/internal/data"
	"github.com/SalvucciFacundo/portfolio-go/views/pages"
)

func PageHandler(w http.ResponseWriter, r *http.Request) {
	lang := "es"
	if v := r.URL.Query().Get("lang"); v == "en" {
		lang = "en"
	}

	profile := data.MockData()

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = pages.Home(profile, lang).Render(r.Context(), w)
}
