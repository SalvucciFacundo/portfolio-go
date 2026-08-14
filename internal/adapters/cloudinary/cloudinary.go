// Package cloudinary implements the Cloudinary media-storage adapter.
//
// Per the portfolio policy (logic_spec.md §10) Cloudinary ONLY stores and
// delivers: image format conversion happens locally (internal/adapters/imageproc)
// and no delivery transformations are requested, so zero transformation credits
// are consumed. All public_ids are flat (no folders): "<entity>-<unix>.<ext>".
package cloudinary

import (
	"bytes"
	"context"
	"fmt"
	"net/url"
	"os"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
)

// Adapter wraps the Cloudinary client with credentials read from the
// environment. It is stateless and safe for concurrent use.
type Adapter struct {
	cld *cloudinary.Cloudinary
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
	return &Adapter{cld: cld}, nil
}

// UploadImage stores an already-encoded WebP image (format conversion is the
// caller's responsibility, see internal/adapters/imageproc) and returns its
// secure delivery URL. overwrite=true makes re-uploads with the same public_id
// idempotent.
func (a *Adapter) UploadImage(ctx context.Context, fileBytes []byte, publicID string) (string, error) {
	resp, err := a.cld.Upload.Upload(ctx, bytes.NewReader(fileBytes), uploader.UploadParams{
		PublicID:     publicID,
		ResourceType: "image",
		Overwrite:    api.Bool(true),
	})
	if err != nil {
		return "", fmt.Errorf("cloudinary: upload image %q: %w", publicID, err)
	}
	return resp.SecureURL, nil
}

// UploadRaw stores a raw file (PDFs). publicID must carry the file extension
// (e.g. "cv-1723000000.pdf") because Cloudinary requires it for raw delivery.
// Returns the secure delivery URL.
func (a *Adapter) UploadRaw(ctx context.Context, fileBytes []byte, publicID string) (string, error) {
	resp, err := a.cld.Upload.Upload(ctx, bytes.NewReader(fileBytes), uploader.UploadParams{
		PublicID:     publicID,
		ResourceType: "raw",
		Overwrite:    api.Bool(true),
	})
	if err != nil {
		return "", fmt.Errorf("cloudinary: upload raw %q: %w", publicID, err)
	}
	return resp.SecureURL, nil
}

// DownloadURL builds the canonical raw delivery URL that forces the browser to
// download the asset as an attachment:
//
//	https://res.cloudinary.com/<cloud>/raw/upload/fl_attachment:<name>/<public_id>
//
// GetCV does not use this builder directly: the original filename lives only in
// the DB (the public_id is "cv-<unix>.pdf"), so the handler inserts
// fl_attachment:<resume_filename> into the stored URL instead (see
// api_profile.go). This method is kept for programmatic construction from a
// known public_id.
func (a *Adapter) DownloadURL(rawPublicID string) string {
	name := url.PathEscape(rawPublicID)
	return fmt.Sprintf("https://res.cloudinary.com/%s/raw/upload/fl_attachment:%s/%s",
		a.cld.Config.Cloud.CloudName, name, rawPublicID)
}
