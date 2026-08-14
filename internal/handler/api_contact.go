// api_contact.go — handler JSON de POST /api/v1/contact.
package handler

import (
	"log"
	"net/http"
	"strings"
)

type contactRequest struct {
	Email   string `json:"email"`
	Subject string `json:"subject"`
	Message string `json:"message"`
}

// Contact valida el mensaje y lo envía vía el Mailer. Rate limit por IP (3/min,
// mismo límite que el form público). Errores de envío se loguean con detalle y
// la API responde 500 con un mensaje genérico.
func (a *API) Contact(w http.ResponseWriter, r *http.Request) {
	if !a.limiter.Allow(remoteIP(r)) {
		writeError(w, http.StatusTooManyRequests, "too many requests")
		return
	}

	var req contactRequest
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if strings.TrimSpace(req.Email) == "" || strings.TrimSpace(req.Subject) == "" || strings.TrimSpace(req.Message) == "" {
		writeError(w, http.StatusBadRequest, "email, subject and message are required")
		return
	}

	if err := a.mailer.SendContact(r.Context(), req.Email, req.Subject, req.Message); err != nil {
		log.Printf("contact api: enviar email: %v", err)
		writeError(w, http.StatusInternalServerError, "send failed")
		return
	}

	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}
