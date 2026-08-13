package router

import (
	"net/http"

	"github.com/SalvucciFacundo/portfolio-go/internal/handler"
)

func Register(mux *http.ServeMux) {
	mux.HandleFunc("GET /", handler.PageHandler)
	mux.HandleFunc("POST /contact", handler.ContactHandler)
	mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
}
