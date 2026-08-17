package router

import (
	"net/http"

	"github.com/SalvucciFacundo/portfolio-go/internal/adapters/db"
	"github.com/SalvucciFacundo/portfolio-go/internal/adapters/mailer"
	"github.com/SalvucciFacundo/portfolio-go/internal/auth"
	"github.com/SalvucciFacundo/portfolio-go/internal/handler"
)

// Deps agrupa las dependencias que main inyecta al router.
type Deps struct {
	Store          *db.Store
	Auth           *auth.Service
	Limiter        *auth.Limiter
	ContactLimiter *auth.Limiter
	Uploader       handler.Uploader
	Mailer         *mailer.Mailer
}

// Register registra las rutas HTMX y las JSON /api/v1. Los endpoints admin de
// mutación se componen con RequireAuth + RequireCSRF.
func Register(mux *http.ServeMux, deps Deps) {
	// ---- Rutas HTMX existentes ----
	mux.HandleFunc("GET /", handler.PageHandler(deps.Store))
	mux.HandleFunc("POST /contact", handler.ContactHandler(deps.Mailer, deps.ContactLimiter))
	mux.HandleFunc("POST /admin/login", handler.LoginHandler)
	mux.HandleFunc("GET /admin/logout", handler.LogoutHandler)
	mux.HandleFunc("POST /admin/hero", handler.HeroUpdateHandler)
	mux.HandleFunc("POST /admin/avatar", handler.AvatarUpdateHandler)
	mux.HandleFunc("POST /admin/skills", handler.SkillsUpdateHandler)
	mux.HandleFunc("DELETE /admin/skills", handler.SkillsDeleteHandler)
	mux.HandleFunc("POST /admin/projects", handler.ProjectsUpdateHandler)
	mux.HandleFunc("DELETE /admin/projects", handler.ProjectsDeleteHandler)
	mux.HandleFunc("POST /admin/education", handler.EducationUpdateHandler)
	mux.HandleFunc("DELETE /admin/education", handler.EducationDeleteHandler)
	mux.HandleFunc("POST /admin/experience", handler.ExperienceUpdateHandler)
	mux.HandleFunc("DELETE /admin/experience", handler.ExperienceDeleteHandler)
	mux.HandleFunc("POST /admin/socials", handler.SocialsUpdateHandler)
	mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	// ---- API JSON /api/v1 ----
	api := handler.NewAPI(deps.Store, deps.Auth, deps.Uploader, deps.Mailer, deps.ContactLimiter)

	// Rutas públicas
	mux.HandleFunc("GET /api/v1/profile/cv", api.GetCV)
	mux.HandleFunc("GET /api/v1/profile", api.GetProfile)
	mux.HandleFunc("GET /api/v1/skills", api.ListSkills)
	mux.HandleFunc("GET /api/v1/projects", api.ListProjects)
	mux.HandleFunc("GET /api/v1/projects/{id}", api.GetProject)
	mux.HandleFunc("GET /api/v1/experience", api.ListExperience)
	mux.HandleFunc("GET /api/v1/education", api.ListEducation)
	mux.HandleFunc("POST /api/v1/contact", api.Contact)
	mux.HandleFunc("POST /api/v1/auth/login", handler.LoginJSONHandler(deps.Auth, deps.Limiter))
	mux.Handle("POST /api/v1/auth/logout", handler.RequireCSRF(handler.LogoutJSONHandler(deps.Auth)))

	// Rutas admin: sesión válida + CSRF en toda mutación.
	admin := func(h http.HandlerFunc) http.Handler {
		return handler.RequireAuth(deps.Auth, handler.RequireCSRF(h))
	}

	mux.Handle("PUT /api/v1/profile", admin(api.UpdateProfile))
	mux.Handle("POST /api/v1/profile/avatar", admin(api.UploadAvatar))
	mux.Handle("POST /api/v1/profile/cv", admin(api.UploadCV))
	mux.Handle("PUT /api/v1/socials", admin(api.UpdateSocials))
	mux.Handle("POST /api/v1/skills", admin(api.CreateSkill))
	mux.Handle("PUT /api/v1/skills/{id}", admin(api.UpdateSkill))
	mux.Handle("DELETE /api/v1/skills/{id}", admin(api.DeleteSkill))
	mux.Handle("POST /api/v1/skills/{id}/icon", admin(api.UploadSkillIcon))
	mux.Handle("POST /api/v1/projects", admin(api.CreateProject))
	mux.Handle("PUT /api/v1/projects/{id}", admin(api.UpdateProject))
	mux.Handle("PUT /api/v1/projects/{id}/position", admin(api.ReorderProject))
	mux.Handle("DELETE /api/v1/projects/{id}", admin(api.DeleteProject))
	mux.Handle("POST /api/v1/projects/{id}/cover", admin(api.UploadProjectCover))
	mux.Handle("POST /api/v1/projects/{id}/images", admin(api.UploadProjectImages))
	mux.Handle("DELETE /api/v1/projects/{id}/images/{imageId}", admin(api.DeleteProjectImage))
	mux.Handle("POST /api/v1/experience", admin(api.CreateExperience))
	mux.Handle("PUT /api/v1/experience/{id}", admin(api.UpdateExperience))
	mux.Handle("DELETE /api/v1/experience/{id}", admin(api.DeleteExperience))
	mux.Handle("POST /api/v1/education", admin(api.CreateEducation))
	mux.Handle("PUT /api/v1/education/{id}", admin(api.UpdateEducation))
	mux.Handle("DELETE /api/v1/education/{id}", admin(api.DeleteEducation))
}
