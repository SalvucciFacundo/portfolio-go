// auth_helpers.go — bridge temporal mientras se migran los handlers admin a /api/v1.
//
// El mock en memoria (internal/auth/auth.go, password admin123) fue eliminado y
// reemplazado por el Service real. Este archivo mantiene compilando los handlers
// HTMX de admin.go exponiendo la MISMA firma que el mock, pero delegando en el
// Service (sesiones hasheadas en PostgreSQL). Se reemplaza por completo en la
// migración a /api/v1 (handlers JSON + fetch + X-CSRF-Token).
package handler

import (
	"context"
	"net"
	"net/http"
	"os"

	"github.com/SalvucciFacundo/portfolio-go/internal/auth"
)

const (
	adminSessionCookie = "admin_session"
	csrfCookie         = "csrf_token"
	sessionMaxAge      = 86400 // 24h, coincide con sessionTTL del Service
)

// authSvc se inyecta desde cmd/api (runServe) vía SetupAuth. nil hasta entonces.
var authSvc *auth.Service

// SetupAuth inyecta el Service real de auth en los helpers HTMX.
func SetupAuth(svc *auth.Service) { authSvc = svc }

// IsAdmin reports whether the request carries a valid, unexpired admin session.
func IsAdmin(r *http.Request) bool {
	if authSvc == nil {
		return false
	}
	cookie, err := r.Cookie(adminSessionCookie)
	if err != nil {
		return false
	}
	return authSvc.Authenticate(r.Context(), cookie.Value, r.UserAgent())
}

// loginPassword valida el password contra el Service y, en caso de éxito, setea
// las cookies admin_session (HttpOnly) + csrf_token (legible por JS). Devuelve
// el token CSRF de la sesión.
func loginPassword(w http.ResponseWriter, r *http.Request, password string) (string, bool) {
	if authSvc == nil {
		return "", false
	}
	token, err := authSvc.Login(r.Context(), password, r.UserAgent())
	if err != nil {
		return "", false
	}
	csrf, err := auth.GenerateCSRFToken()
	if err != nil {
		return "", false
	}
	setSessionCookie(w, token)
	setCSRFCookie(w, csrf)
	return csrf, true
}

// Logout invalida la sesión asociada a token.
func Logout(token string) {
	if authSvc == nil {
		return
	}
	_ = authSvc.Logout(context.Background(), token)
}

// SetSessionCookie escribe la cookie de sesión. Mantiene la firma del mock
// eliminado para que los handlers HTMX sigan compilando.
func SetSessionCookie(w http.ResponseWriter, token string) {
	setSessionCookie(w, token)
}

// ClearSessionCookie elimina la cookie de sesión. Mantiene la firma del mock
// eliminado para que los handlers HTMX sigan compilando.
func ClearSessionCookie(w http.ResponseWriter) {
	clearSessionCookie(w)
}

// setSessionCookie escribe admin_session: HttpOnly, SameSite Strict, Secure en
// producción, MaxAge 24h.
func setSessionCookie(w http.ResponseWriter, token string) {
	http.SetCookie(w, &http.Cookie{
		Name:     adminSessionCookie,
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   isProduction(),
		SameSite: http.SameSiteStrictMode,
		MaxAge:   sessionMaxAge,
	})
}

// setCSRFCookie escribe csrf_token. NO es HttpOnly para que el JS del admin lo
// lea y lo mande en X-CSRF-Token.
func setCSRFCookie(w http.ResponseWriter, token string) {
	http.SetCookie(w, &http.Cookie{
		Name:     csrfCookie,
		Value:    token,
		Path:     "/",
		HttpOnly: false,
		Secure:   isProduction(),
		SameSite: http.SameSiteStrictMode,
		MaxAge:   sessionMaxAge,
	})
}

// clearSessionCookie elimina admin_session.
func clearSessionCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     adminSessionCookie,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   isProduction(),
		SameSite: http.SameSiteStrictMode,
		MaxAge:   -1,
	})
}

// clearCSRFCookie elimina csrf_token.
func clearCSRFCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     csrfCookie,
		Value:    "",
		Path:     "/",
		Secure:   isProduction(),
		SameSite: http.SameSiteStrictMode,
		MaxAge:   -1,
	})
}

// isProduction reports whether APP_ENV=production.
func isProduction() bool {
	return os.Getenv("APP_ENV") == "production"
}

// remoteIP devuelve la IP del cliente sin el puerto.
func remoteIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}
