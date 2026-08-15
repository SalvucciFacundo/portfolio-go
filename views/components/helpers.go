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

// statusBadgeClass devuelve la clase CSS del badge de estado del proyecto.
func statusBadgeClass(status string) string {
	switch status {
	case "production":
		return "project-card__badge--production"
	case "demo":
		return "project-card__badge--demo"
	default:
		return "project-card__badge--development"
	}
}

// statusLabel devuelve la etiqueta visible del badge según el idioma.
func statusLabel(status, lang string) string {
	labels := map[string]map[string]string{
		"production":  {"es": "Producción", "en": "Production"},
		"development": {"es": "Desarrollo", "en": "Development"},
		"demo":        {"es": "Demo", "en": "Demo"},
	}
	if m, ok := labels[status]; ok {
		if v, ok := m[lang]; ok {
			return v
		}
		return m["es"]
	}
	return "Desarrollo"
}

func copyEmailClass(hasLabel bool) string {
	if hasLabel {
		return "copy-email copy-email--with-label grayscale-reveal"
	}
	return "copy-email grayscale-reveal"
}
