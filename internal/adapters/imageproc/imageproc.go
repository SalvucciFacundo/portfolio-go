// Package imageproc performs local image format conversion with the cwebp
// binary. Per portfolio policy (logic_spec.md §10): images are NEVER resized
// (original dimensions are preserved), quality is -q 90, and flat PNGs/logos
// are encoded lossless. Cloudinary only stores what this package produces.
package imageproc

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// ConvertToWebP converts image bytes to WebP using the cwebp binary and
// returns the encoded bytes. It never resizes, uses -q 90 (with -lossless for
// PNG input so flat graphics/logos keep exact fidelity), and passes files that
// are already WebP through untouched.
func ConvertToWebP(src []byte, srcFilename string) ([]byte, error) {
	if isWebP(src) {
		return src, nil
	}
	if _, err := exec.LookPath("cwebp"); err != nil {
		return nil, fmt.Errorf("cwebp not found in PATH — install libwebp (apk add webp / apt install webp)")
	}

	ext := filepath.Ext(srcFilename)
	tmpIn, err := os.CreateTemp("", "imgproc-in-*"+ext)
	if err != nil {
		return nil, fmt.Errorf("create temp input: %w", err)
	}
	defer os.Remove(tmpIn.Name())
	if _, err := tmpIn.Write(src); err != nil {
		_ = tmpIn.Close()
		return nil, fmt.Errorf("write temp input: %w", err)
	}
	if err := tmpIn.Close(); err != nil {
		return nil, fmt.Errorf("close temp input: %w", err)
	}

	tmpOut, err := os.CreateTemp("", "imgproc-out-*.webp")
	if err != nil {
		return nil, fmt.Errorf("create temp output: %w", err)
	}
	outPath := tmpOut.Name()
	_ = tmpOut.Close()
	defer os.Remove(outPath)

	args := []string{"-q", "90", "-quiet"}
	if isPNG(src) {
		args = append(args, "-lossless")
	}
	args = append(args, tmpIn.Name(), "-o", outPath)

	if out, err := exec.Command("cwebp", args...).CombinedOutput(); err != nil {
		return nil, fmt.Errorf("cwebp convert: %w: %s", err, out)
	}

	result, err := os.ReadFile(outPath)
	if err != nil {
		return nil, fmt.Errorf("read webp output: %w", err)
	}
	return result, nil
}

// isPNG detects PNG by its full magic byte signature (0x89 'P' 'N' 'G' CR LF
// SUB LF). Magic bytes are preferred over the filename because the extension
// is client-controlled and unreliable.
func isPNG(data []byte) bool {
	return len(data) >= 8 &&
		data[0] == 0x89 && data[1] == 'P' && data[2] == 'N' && data[3] == 'G' &&
		data[4] == 0x0D && data[5] == 0x0A && data[6] == 0x1A && data[7] == 0x0A
}

// isWebP detects WebP by its RIFF....WEBP container header.
func isWebP(data []byte) bool {
	return len(data) >= 12 &&
		bytes.Equal(data[:4], []byte("RIFF")) &&
		bytes.Equal(data[8:12], []byte("WEBP"))
}
