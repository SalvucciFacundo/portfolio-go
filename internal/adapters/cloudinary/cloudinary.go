// Package cloudinary implements the Cloudinary media-storage adapter.
//
// Per the portfolio policy (logic_spec.md §10) Cloudinary ONLY stores and
// delivers: image format conversion happens locally (internal/adapters/imageproc)
// and no delivery transformations are requested, so zero transformation credits
// are consumed. All public_ids are flat (no folders): "<entity>-<unix>".
//
// Gotchas handled here:
//   - The cloudinary-go SDK v2.16.0 silently fails (returns an empty result,
//     no error) when uploading with resource_type=raw, and also when uploading
//     an in-memory bytes.Reader for images. We therefore: (a) write image bytes
//     to a temp file and upload by path through the SDK; (b) upload raw files
//     (PDFs) through the raw REST API with a signed request.
//   - Cloudinary appends the format extension to the delivery URL itself, so
//     image public_ids must NOT carry an extension ("avatar-1723…" not
//     "avatar-1723….webp"); raw public_ids MUST ("cv-1723….pdf").
//   - fl_attachment:<name> must NOT include the file extension — Cloudinary
//     appends it from the asset format.
package cloudinary

import (
	"bytes"
	"context"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
)

// Adapter wraps the Cloudinary client with credentials read from the
// environment. It is stateless and safe for concurrent use.
type Adapter struct {
	cld        *cloudinary.Cloudinary
	cloudName  string
	apiKey     string
	apiSecret  string
	httpClient *http.Client
}

// New reads the Cloudinary credentials from the environment and builds the
// client. It fails fast with a clear error when any required var is missing.
func New() (*Adapter, error) {
	cloudName := os.Getenv("CLOUDINARY_CLOUD_NAME")
	apiKey := os.Getenv("CLOUDINARY_API_KEY")
	apiSecret := os.Getenv("CLOUDINARY_API_SECRET")
	if cloudName == "" || apiKey == "" || apiSecret == "" {
		return nil, fmt.Errorf(
			"cloudinary: missing required env vars (CLOUDINARY_CLOUD_NAME, CLOUDINARY_API_KEY, CLOUDINARY_API_SECRET) — revisar .env")
	}
	cld, err := cloudinary.NewFromParams(cloudName, apiKey, apiSecret)
	if err != nil {
		return nil, fmt.Errorf("cloudinary: create client: %w", err)
	}
	return &Adapter{
		cld:        cld,
		cloudName:  cloudName,
		apiKey:     apiKey,
		apiSecret:  apiSecret,
		httpClient: &http.Client{Timeout: 60 * time.Second},
	}, nil
}

// UploadImage stores an already-encoded WebP image (format conversion is the
// caller's responsibility, see internal/adapters/imageproc) and returns its
// secure delivery URL. publicID must NOT include a file extension (Cloudinary
// appends the format to the URL). folder is the Media Library folder (e.g.
// "portfolio/avatar") — setting AssetFolder makes Cloudinary index the asset
// into a real folder instead of only prefixing the public_id. overwrite=true
// makes re-uploads with the same public_id idempotent.
func (a *Adapter) UploadImage(ctx context.Context, fileBytes []byte, folder, publicID string) (string, error) {
	url, err := a.uploadImageByPath(ctx, fileBytes, folder, publicID)
	if err != nil {
		return "", fmt.Errorf("cloudinary: upload image %q: %w", publicID, err)
	}
	return url, nil
}

// UploadRaw stores a raw file (PDFs) through the signed REST API (the SDK
// silently fails for resource_type=raw). publicID MUST carry the file
// extension (e.g. "cv-1723000000.pdf"). folder is the Media Library folder.
// Returns the secure delivery URL.
func (a *Adapter) UploadRaw(ctx context.Context, fileBytes []byte, folder, publicID string) (string, error) {
	url, err := a.uploadRawREST(ctx, fileBytes, folder, publicID)
	if err != nil {
		return "", fmt.Errorf("cloudinary: upload raw %q: %w", publicID, err)
	}
	return url, nil
}

// uploadImageByPath writes the bytes to a temp file and uploads by path,
// working around the SDK bug where bytes.Reader uploads return empty results.
func (a *Adapter) uploadImageByPath(ctx context.Context, fileBytes []byte, folder, publicID string) (string, error) {
	tmp, err := os.CreateTemp("", "cld-upload-*")
	if err != nil {
		return "", fmt.Errorf("create temp file: %w", err)
	}
	defer os.Remove(tmp.Name())
	defer tmp.Close()

	if _, err := tmp.Write(fileBytes); err != nil {
		return "", fmt.Errorf("write temp file: %w", err)
	}
	if err := tmp.Sync(); err != nil {
		return "", fmt.Errorf("sync temp file: %w", err)
	}

	resp, err := a.cld.Upload.Upload(ctx, tmp.Name(), uploader.UploadParams{
		PublicID:     publicID,
		AssetFolder:  folder,
		ResourceType: "image",
		Overwrite:    api.Bool(true),
	})
	if err != nil {
		return "", err
	}
	if resp.SecureURL == "" {
		return "", fmt.Errorf("empty response from cloudinary for %q (resource_type=image)", publicID)
	}
	return resp.SecureURL, nil
}

// uploadRawREST sube un archivo raw (PDF) usando la API REST de Cloudinary con
// firma SHA-1. El SDK Go falla silenciosamente con resource_type=raw.
func (a *Adapter) uploadRawREST(ctx context.Context, fileBytes []byte, folder, publicID string) (string, error) {
	timestamp := fmt.Sprintf("%d", time.Now().Unix())

	// Firma: todos los campos del form (excepto api_key/file/resource_type,
	// que van en la URL) en orden alfabético + secret.
	signParams := map[string]string{
		"timestamp":    timestamp,
		"public_id":    publicID,
		"asset_folder": folder,
		"overwrite":    "true",
	}
	sign := buildSignature(signParams, a.apiSecret)

	endpoint := fmt.Sprintf("https://api.cloudinary.com/v1_1/%s/raw/upload", a.cloudName)

	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	for _, k := range []string{"timestamp", "overwrite", "resource_type", "public_id", "asset_folder"} {
		_ = mw.WriteField(k, signParams[k])
	}
	_ = mw.WriteField("api_key", a.apiKey)
	_ = mw.WriteField("signature", sign)

	part, err := mw.CreateFormFile("file", publicID)
	if err != nil {
		return "", fmt.Errorf("create form file: %w", err)
	}
	if _, err := part.Write(fileBytes); err != nil {
		return "", fmt.Errorf("write file part: %w", err)
	}
	if err := mw.Close(); err != nil {
		return "", fmt.Errorf("close multipart: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, &buf)
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", mw.FormDataContentType())

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return "", fmt.Errorf("read response: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("cloudinary raw upload http %d: %s", resp.StatusCode, truncate(string(body), 200))
	}

	var result struct {
		SecureURL string `json:"secure_url"`
		Error     *struct {
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("parse response: %w", err)
	}
	if result.Error != nil {
		return "", fmt.Errorf("cloudinary error: %s", result.Error.Message)
	}
	if result.SecureURL == "" {
		return "", fmt.Errorf("empty secure_url in cloudinary raw response")
	}
	return result.SecureURL, nil
}

// DownloadURL builds the canonical raw delivery URL that forces the browser to
// download the asset as an attachment with the given filename (WITHOUT
// extension — Cloudinary appends it from the asset format):
//
//	https://res.cloudinary.com/<cloud>/raw/upload/fl_attachment:<name>/<public_id>
//
// GetCV does not use this builder directly: the original filename lives only in
// the DB, so the handler inserts fl_attachment:<base filename> into the stored
// URL instead (see api_profile.go).
func (a *Adapter) DownloadURL(rawPublicID string) string {
	base := strings.TrimSuffix(filepath.Base(rawPublicID), filepath.Ext(rawPublicID))
	name := url.PathEscape(base)
	return fmt.Sprintf("https://res.cloudinary.com/%s/raw/upload/fl_attachment:%s/%s",
		a.cloudName, name, rawPublicID)
}

// buildSignature firma los parámetros de la API de Cloudinary: concatena
// key=value ordenados alfabéticamente + el api_secret, y aplica SHA-1.
func buildSignature(params map[string]string, secret string) string {
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var sb strings.Builder
	for _, k := range keys {
		sb.WriteString(k)
		sb.WriteString("=")
		sb.WriteString(params[k])
		sb.WriteString("&")
	}
	toSign := strings.TrimSuffix(sb.String(), "&") + secret

	h := sha1.Sum([]byte(toSign))
	return hex.EncodeToString(h[:])
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
