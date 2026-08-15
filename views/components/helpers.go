package components

import (
	"fmt"
	"strings"

	"github.com/SalvucciFacundo/portfolio-go/internal/domain"
)

// screenshotURLs extrae las URLs de los screenshots (optimizadas a 1000px)
// para los atributos de datos del front (data-images espera URLs separadas por
// coma).
func screenshotURLs(imgs []domain.ProjectImage) []string {
	urls := make([]string, 0, len(imgs))
	for _, img := range imgs {
		urls = append(urls, responsiveURL(img.URL, 1000))
	}
	return urls
}

func brandColor(key string) string {
	switch key {
	case "github":
		return "#181717"
	case "linkedin":
		return "#0A66C2"
	case "mail":
		return "#EA4335"
	default:
		return "var(--text-primary)"
	}
}

func brandStyle(key string) string {
	return "--brand:" + brandColor(key)
}

func activeLang(current, target string) string {
	if current == target {
		return "is-active"
	}
	return ""
}

func otherLangHref(lang string) string {
	if lang == "en" {
		return "/?lang=es"
	}
	return "/?lang=en"
}

func copyEmailClass(hasLabel bool) string {
	if hasLabel {
		return "copy-email copy-email--with-label grayscale-reveal"
	}
	return "copy-email grayscale-reveal"
}

// responsiveURL transforma una URL de Cloudinary para que entregue el asset
// redimensionado/optimizado al ancho deseado. Las imágenes se suben a tamaño
// completo (p.ej. 1920px) y sin esto el navegador descargaría todo el archivo
// para mostrarlo en 800px. f_auto + q_auto optimizan formato y calidad en la
// entrega (no consumen créditos de transformación de subida).
func responsiveURL(raw string, width int) string {
	const marker = "/image/upload/"
	i := strings.Index(raw, marker)
	if i < 0 {
		return raw
	}
	transform := fmt.Sprintf("w_%d,q_auto,f_auto", width)
	insertAt := i + len(marker)
	return raw[:insertAt] + transform + "/" + raw[insertAt:]
}
