// api.go — núcleo de los handlers JSON /api/v1: estructura API, helpers de
// JSON y utilidades compartidas.
package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/SalvucciFacundo/portfolio-go/internal/adapters/db"
	"github.com/SalvucciFacundo/portfolio-go/internal/adapters/imageproc"
	"github.com/SalvucciFacundo/portfolio-go/internal/adapters/mailer"
	"github.com/SalvucciFacundo/portfolio-go/internal/auth"
	"github.com/SalvucciFacundo/portfolio-go/internal/domain"
)

// Uploader es el puerto de salida del handler hacia el almacenamiento externo
// (Cloudinary). Desacopla los handlers del adapter concreto: el llamador decide
// la carpeta (Media Library) y el nombre, y recibe la URL de entrega.
type Uploader interface {
	UploadImage(ctx context.Context, data []byte, folder, publicID string) (string, error)
	UploadRaw(ctx context.Context, data []byte, folder, publicID string) (string, error)
}

// API agrupa los handlers JSON del portafolio sobre el Store, el Service de
// auth, el Uploader, el Mailer y el Limiter inyectados.
type API struct {
	store    *db.Store
	auth     *auth.Service
	uploader Uploader
	mailer   *mailer.Mailer
	limiter  *auth.Limiter
}

// NewAPI crea un API con las dependencias dadas.
func NewAPI(store *db.Store, svc *auth.Service, uploader Uploader, mailer *mailer.Mailer, limiter *auth.Limiter) *API {
	return &API{store: store, auth: svc, uploader: uploader, mailer: mailer, limiter: limiter}
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
	imageExts = map[string]bool{".jpg": true, ".jpeg": true, ".png": true, ".webp": true, ".gif": true, ".avif": true, ".bmp": true, ".tiff": true, ".tif": true}
	// vectorExts son archivos vectoriales (SVG) que se suben TAL CUAL a
	// Cloudinary sin conversión a WebP (cwebp no procesa vectores y convertirlos
	// perdería nitidez). El navegador los renderiza directamente.
	vectorExts = map[string]bool{".svg": true}
	cvExts     = map[string]bool{".pdf": true}
)

// uploadFile lee un archivo multipart a bytes y lo sube al Uploader: las
// imágenes bitmap se convierten a WebP en local (imageproc); los SVG y PDFs se
// suben tal cual. El public_id se arma como <folder>/<name>-<unix>.<ext> —
// Cloudinary crea las carpetas automáticamente. Devuelve la URL de entrega y
// el nombre original del archivo.
func (a *API) uploadFile(r *http.Request, fh *multipart.FileHeader, folder, name string) (url, filename string, err error) {
	if a.uploader == nil {
		return "", "", errors.New("uploader not configured")
	}

	filename = filepath.Base(fh.Filename)
	file, err := fh.Open()
	if err != nil {
		return "", "", fmt.Errorf("open upload: %w", err)
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return "", "", fmt.Errorf("read upload: %w", err)
	}

	ext := strings.ToLower(filepath.Ext(filename))
	switch {
	case cvExts[ext]:
		publicID := fmt.Sprintf("%s-%d.pdf", name, time.Now().Unix())
		url, err = a.uploader.UploadRaw(r.Context(), data, folder, publicID)
	case vectorExts[ext]:
		// SVG: subir tal cual, SIN conversión. public_id con extensión para
		// que Cloudinary lo sirva como image/svg+xml.
		publicID := fmt.Sprintf("%s-%d.svg", name, time.Now().Unix())
		url, err = a.uploader.UploadImage(r.Context(), data, folder, publicID)
	case imageExts[ext]:
		webp, convErr := imageproc.ConvertToWebP(data, filename)
		if convErr != nil {
			return "", "", convErr
		}
		// public_id SIN extensión: Cloudinary agrega el formato a la URL
		publicID := fmt.Sprintf("%s-%d", name, time.Now().Unix())
		url, err = a.uploader.UploadImage(r.Context(), webp, folder, publicID)
	default:
		return "", "", fmt.Errorf("file type %q not allowed", ext)
	}
	if err != nil {
		return "", "", fmt.Errorf("upload %s/%s: %w", folder, name, err)
	}
	return url, filename, nil
}

// emptyProfileSlices normaliza las listas anidadas del profile a slices vacíos
// ([] en JSON) en lugar de null cuando el repo no las carga.
func emptyProfileSlices(p *domain.Profile) {
	p.Skills = []domain.Skill{}
	p.Projects = []domain.Project{}
	p.Experience = []domain.Experience{}
	p.Education = []domain.Education{}
}
