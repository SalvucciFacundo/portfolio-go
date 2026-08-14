// middleware.go — middleware HTTP de auth y CSRF para los endpoints /api/v1.
package handler

import (
	"net/http"

	"github.com/SalvucciFacundo/portfolio-go/internal/auth"
)

// RequireAuth rechaza con 401 cualquier request sin una sesión admin válida
// (cookie admin_session + Authenticate). Encadena con next.
func RequireAuth(svc *auth.Service, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(adminSessionCookie)
		if err != nil || !svc.Authenticate(r.Context(), cookie.Value, r.UserAgent()) {
			writeError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		next.ServeHTTP(w, r)
	})
}

// RequireCSRF valida X-CSRF-Token contra la cookie csrf_token en constant time
// para los métodos de mutación (POST/PUT/PATCH/DELETE). Los métodos seguros
// pasan sin verificación. 403 si el token no coincide.
func RequireCSRF(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete:
		default:
			next.ServeHTTP(w, r)
			return
		}

		cookie, err := r.Cookie(csrfCookie)
		if err != nil {
			writeError(w, http.StatusForbidden, "invalid csrf token")
			return
		}
		if !auth.ValidCSRFToken(r.Header.Get("X-CSRF-Token"), cookie.Value) {
			writeError(w, http.StatusForbidden, "invalid csrf token")
			return
		}
		next.ServeHTTP(w, r)
	})
}
