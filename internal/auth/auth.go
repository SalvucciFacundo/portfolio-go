package auth

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"sync"
	"time"
)

// Mock credentials — will be replaced by DB auth later.
const mockPassword = "admin123"

var (
	sessions = make(map[string]time.Time)
	mu       sync.RWMutex
)

// Login validates the password and returns a session token.
func Login(password string) (string, bool) {
	if password != mockPassword {
		return "", false
	}
	token := generateToken()
	mu.Lock()
	sessions[token] = time.Now().Add(24 * time.Hour)
	mu.Unlock()
	return token, true
}

// Logout invalidates a session token.
func Logout(token string) {
	mu.Lock()
	delete(sessions, token)
	mu.Unlock()
}

// IsAdmin checks whether the current request has a valid admin session.
func IsAdmin(r *http.Request) bool {
	cookie, err := r.Cookie("admin_session")
	if err != nil {
		return false
	}
	mu.RLock()
	exp, ok := sessions[cookie.Value]
	mu.RUnlock()
	if !ok || time.Now().After(exp) {
		return false
	}
	return true
}

// SetSessionCookie writes the session cookie to the response.
func SetSessionCookie(w http.ResponseWriter, token string) {
	http.SetCookie(w, &http.Cookie{
		Name:     "admin_session",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   86400,
	})
}

// ClearSessionCookie removes the session cookie.
func ClearSessionCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     "admin_session",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		MaxAge:   -1,
	})
}

func generateToken() string {
	b := make([]byte, 32)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
