package layouts

import (
	"crypto/sha1"
	"encoding/hex"
	"os"
)

// assetVersion devuelve un hash corto del archivo estático (admin.js) para
// cache-busting: cuando el archivo cambia, el hash cambia y el navegador
// descarta la versión cacheada. Evita que los usuarios vean JS viejo.
func assetVersion() string {
	data, err := os.ReadFile("static/js/admin.js")
	if err != nil {
		return "0"
	}
	sum := sha1.Sum(data)
	return hex.EncodeToString(sum[:])[:8]
}
