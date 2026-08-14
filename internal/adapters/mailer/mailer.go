// mailer.go — adapter del servicio de email del formulario de contacto.
//
// Usa la API REST de Resend directamente (POST https://api.resend.com/emails)
// con Authorization: Bearer, sin el SDK oficial: la API es lo bastante simple
// como para no justificar una dependencia más.
package mailer

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"html"
	"io"
	"net/http"
	"os"
	"time"
)

const (
	resendEndpoint = "https://api.resend.com/emails"
	clientTimeout  = 15 * time.Second
)

// Mailer envía emails de contacto usando Resend. El remitente real es siempre
// m.from (dominio verificado en Resend); el email del visitante solo viaja en
// el cuerpo del mensaje, nunca como "from".
type Mailer struct {
	apiKey string
	from   string
	to     string
	client *http.Client
}

// New crea un Mailer leyendo RESEND_API_KEY, CONTACT_EMAIL y CONTACT_FROM del
// entorno. Error si falta cualquiera de las tres (sin destinatario o remitente
// el envío no tiene sentido).
func New() (*Mailer, error) {
	apiKey := os.Getenv("RESEND_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("mailer: RESEND_API_KEY no configurada")
	}
	to := os.Getenv("CONTACT_EMAIL")
	if to == "" {
		return nil, fmt.Errorf("mailer: CONTACT_EMAIL no configurada")
	}
	from := os.Getenv("CONTACT_FROM")
	if from == "" {
		return nil, fmt.Errorf("mailer: CONTACT_FROM no configurada")
	}

	return &Mailer{
		apiKey: apiKey,
		from:   from,
		to:     to,
		client: &http.Client{Timeout: clientTimeout},
	}, nil
}

// SendContact envía el mensaje del formulario al destinatario configurado.
// Devuelve error si Resend no responde 200. El senderEmail del visitante se
// incluye en el cuerpo (html), nunca en "from".
func (m *Mailer) SendContact(ctx context.Context, senderEmail, subject, message string) error {
	payload := struct {
		From    string   `json:"from"`
		To      []string `json:"to"`
		Subject string   `json:"subject"`
		HTML    string   `json:"html"`
	}{
		From:    m.from,
		To:      []string{m.to},
		Subject: "Contacto del portafolio: " + subject,
		HTML:    contactHTML(senderEmail, subject, message),
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("mailer: marshal payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, resendEndpoint, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("mailer: build request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+m.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := m.client.Do(req)
	if err != nil {
		return fmt.Errorf("mailer: send: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		raw, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return fmt.Errorf("mailer: resend status %d: %s", resp.StatusCode, string(raw))
	}
	return nil
}

// contactHTML arma un bloque HTML legible con el remitente, el asunto y el
// mensaje. Todo el contenido del visitante va escapado con html.EscapeString.
func contactHTML(senderEmail, subject, message string) string {
	from := html.EscapeString(senderEmail)
	subj := html.EscapeString(subject)
	msg := html.EscapeString(message)
	return fmt.Sprintf(`<!doctype html>
<html lang="es">
<head>
  <meta charset="utf-8">
  <title>Contacto del portafolio</title>
</head>
<body style="font-family:Arial,Helvetica,sans-serif;line-height:1.6;color:#333;max-width:640px;margin:0 auto;">
  <h2 style="border-bottom:2px solid #eee;padding-bottom:8px;">Nuevo mensaje del portafolio</h2>
  <p><strong>De:</strong> %s</p>
  <p><strong>Asunto:</strong> %s</p>
  <p><strong>Mensaje:</strong></p>
  <div style="background:#f6f6f6;padding:16px;border-radius:8px;white-space:pre-wrap;">%s</div>
</body>
</html>`, from, subj, msg)
}
