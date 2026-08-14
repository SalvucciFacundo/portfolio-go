// api.go — núcleo de los handlers JSON /api/v1: estructura API, helpers de
// JSON y utilidades compartidas.
package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/SalvucciFacundo/portfolio-go/internal/adapters/db"
	"github.com/SalvucciFacundo/portfolio-go/internal/auth"
	"github.com/SalvucciFacundo/portfolio-go/internal/domain"
)

const uploadsDir = "static/uploads"

// API agrupa los handlers JSON del portafolio sobre el Store y el Service de
// auth inyectados.
type API struct {
	store *db.Store
	auth  *auth.Service
}

// NewAPI crea un API con las dependencias dadas.
func NewAPI(store *db.Store, svc *auth.Service) *API {
	return &API{store: store, auth: svc}
}

// writeError escribe una respuesta JSON de error: {"error":"msg"}.
func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

// readJSON decodifica el body JSON en dst rechazando campos desconocidos.
func readJSON(r *http.Request, dst any) error {
	r.Body = http.MaxBytesReader(nil, r.Body, 1<<20)
	defer r.Body.Close()

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(dst); err != nil {
		return fmt.Errorf("decode json: %w", err)
	}
	return nil
}

// idParam parsea un path value numérico de Go 1.22 ServeMux.
func idParam(r *http.Request, name string) (int64, error) {
	id, err := strconv.ParseInt(r.PathValue(name), 10, 64)
	if err != nil || id <= 0 {
		return 0, errors.New("invalid id")
	}
	return id, nil
}

var (
	imageExts = map[string]bool{".jpg": true, ".jpeg": true, ".png": true, ".webp": true, ".gif": true}
	cvExts    = map[string]bool{".pdf": true}
)

// saveUpload guarda un archivo multipart en static/uploads con nombre
// generado por el server (<entidad>-<unix>.<ext>) y devuelve su URL pública.
// Rechaza extensiones fuera del set allowed para evitar path traversal y tipos
// inesperados.
func saveUpload(file multipart.File, header *multipart.FileHeader, entity string, allowed map[string]bool) (string, error) {
	ext := strings.ToLower(filepath.Ext(header.Filename))
	if !allowed[ext] {
		return "", fmt.Errorf("file type %q not allowed", ext)
	}

	name := fmt.Sprintf("%s-%d%s", entity, time.Now().UnixNano(), ext)
	dst := filepath.Join(uploadsDir, name)
	if err := os.MkdirAll(uploadsDir, 0o755); err != nil {
		return "", fmt.Errorf("create uploads dir: %w", err)
	}

	out, err := os.Create(dst)
	if err != nil {
		return "", fmt.Errorf("create file: %w", err)
	}
	defer out.Close()

	if _, err := io.Copy(out, file); err != nil {
		return "", fmt.Errorf("write file: %w", err)
	}
	return "/static/uploads/" + name, nil
}

// emptyProfileSlices normaliza las listas anidadas del profile a slices vacíos
// ([] en JSON) en lugar de null cuando el repo no las carga.
func emptyProfileSlices(p *domain.Profile) {
	p.Skills = []domain.Skill{}
	p.Projects = []domain.Project{}
	p.Experience = []domain.Experience{}
	p.Education = []domain.Education{}
}
