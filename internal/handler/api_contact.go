// api_contact.go — handler JSON de POST /api/v1/contact.
package handler

import (
	"net/http"
	"strings"
)

type contactRequest struct {
	Email   string `json:"email"`
	Subject string `json:"subject"`
	Message string `json:"message"`
}

// Contact valida el mensaje y responde ok. El envío real (mailer) se agrega en
// una fase posterior.
func (a *API) Contact(w http.ResponseWriter, r *http.Request) {
	var req contactRequest
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if strings.TrimSpace(req.Email) == "" || strings.TrimSpace(req.Subject) == "" || strings.TrimSpace(req.Message) == "" {
		writeError(w, http.StatusBadRequest, "email, subject and message are required")
		return
	}

	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}
