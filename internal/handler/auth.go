// auth.go — handlers JSON de login/logout para /api/v1/auth.
//
// Las rutas se registran en la fase de handlers /api/v1 (router/routes.go).
package handler

import (
	"encoding/json"
	"net/http"

	"github.com/SalvucciFacundo/portfolio-go/internal/auth"
)

type loginRequest struct {
	Password string `json:"password"`
}

// LoginJSONHandler implementa POST /api/v1/auth/login: rate limit por IP
// remota, valida el password contra el Service y setea las cookies de sesión
// (HttpOnly) y CSRF (legible por JS).
func LoginJSONHandler(svc *auth.Service, limiter *auth.Limiter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeJSONError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}

		if !limiter.Allow(remoteIP(r)) {
			writeJSONError(w, http.StatusTooManyRequests, "too many requests")
			return
		}

		var req loginRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSONError(w, http.StatusBadRequest, "invalid request")
			return
		}

		token, err := svc.Login(r.Context(), req.Password, r.UserAgent())
		if err != nil {
			writeJSONError(w, http.StatusUnauthorized, "invalid credentials")
			return
		}

		csrf, err := auth.GenerateCSRFToken()
		if err != nil {
			writeJSONError(w, http.StatusInternalServerError, "internal error")
			return
		}
		setSessionCookie(w, token)
		setCSRFCookie(w, csrf)

		writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
	}
}

// LogoutJSONHandler implementa POST /api/v1/auth/logout: invalida la sesión en
// la DB y limpia ambas cookies.
func LogoutJSONHandler(svc *auth.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeJSONError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}

		if cookie, err := r.Cookie(adminSessionCookie); err == nil {
			_ = svc.Logout(r.Context(), cookie.Value)
		}
		clearSessionCookie(w)
		clearCSRFCookie(w)

		writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
	}
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeJSONError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}
