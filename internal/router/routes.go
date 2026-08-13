package router

import (
	"net/http"

	"github.com/SalvucciFacundo/portfolio-go/internal/handler"
)

func Register(mux *http.ServeMux) {
	mux.HandleFunc("GET /", handler.PageHandler)
	mux.HandleFunc("POST /contact", handler.ContactHandler)
	mux.HandleFunc("POST /admin/login", handler.LoginHandler)
	mux.HandleFunc("GET /admin/logout", handler.LogoutHandler)
	mux.HandleFunc("POST /admin/hero", handler.HeroUpdateHandler)
	mux.HandleFunc("POST /admin/avatar", handler.AvatarUpdateHandler)
	mux.HandleFunc("POST /admin/skills", handler.SkillsUpdateHandler)
	mux.HandleFunc("DELETE /admin/skills", handler.SkillsDeleteHandler)
	mux.HandleFunc("POST /admin/education", handler.EducationUpdateHandler)
	mux.HandleFunc("DELETE /admin/education", handler.EducationDeleteHandler)
	mux.HandleFunc("POST /admin/experience", handler.ExperienceUpdateHandler)
	mux.HandleFunc("DELETE /admin/experience", handler.ExperienceDeleteHandler)
	mux.HandleFunc("POST /admin/socials", handler.SocialsUpdateHandler)
	mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
}
