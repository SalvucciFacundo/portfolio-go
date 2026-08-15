package components

import (
	"github.com/SalvucciFacundo/portfolio-go/internal/domain"
)

// screenshotURLs extrae las URLs de los screenshots para los atributos de datos
// del front (data-images espera URLs separadas por coma). Se usan las URLs
// originales: las imágenes del modal cargan con lazy al abrir el proyecto.
func screenshotURLs(imgs []domain.ProjectImage) []string {
	urls := make([]string, 0, len(imgs))
	for _, img := range imgs {
		urls = append(urls, img.URL)
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
