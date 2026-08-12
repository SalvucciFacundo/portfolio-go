package i18n

// L returns the localized variant of a bilingual editable field.
// Defaults to Spanish when lang is not "en".
func L(es, en, lang string) string {
	if lang == "en" {
		return en
	}
	return es
}

var translations = map[string]map[string]string{
	"es": {
		"meta.description":   "Portafolio de Facundo Salvucci — Desarrollador Full Stack en Go.",
		"hero.greeting":      "Hola a todos, soy",
		"hero.matrix":        "backend / frontend / interfaz",
		"section.hero":       "Inicio",
		"section.skills":     "Skills",
		"section.projects":   "Proyectos",
		"section.experience": "Experiencia",
		"section.education":  "Educación",
		"section.contact":    "Contacto",
		"action.download_cv": "DESCARGAR_CV",
		"action.view_more":   "Ver más",
		"action.copy_email":  "Copiar email",
		"action.copied":      "¡Copiado!",
		"contact.cta":        "Escribime y hablemos de tu próximo proyecto.",
		"footer.cta":         "¿Listo para construir algo juntos?",
		"footer.rights":      "Todos los derechos reservados.",
		"theme.to_dark":      "Cambiar a tema oscuro",
		"theme.to_light":     "Cambiar a tema claro",
		"login.aria":         "Iniciar sesión",
		"lang.label":         "Idioma",
		"lang.switch_to":     "Cambiar a inglés",
		"sidebar.aria":       "Barra lateral",
	},
	"en": {
		"meta.description":   "Portfolio of Facundo Salvucci — Full Stack Developer in Go.",
		"hero.greeting":      "Hello everyone, I'm",
		"hero.matrix":        "backend / frontend / interface",
		"section.hero":       "Home",
		"section.skills":     "Skills",
		"section.projects":   "Projects",
		"section.experience": "Experience",
		"section.education":  "Education",
		"section.contact":    "Contact",
		"action.download_cv": "DOWNLOAD_CV",
		"action.view_more":   "View more",
		"action.copy_email":  "Copy email",
		"action.copied":      "Copied!",
		"contact.cta":        "Write to me and let's talk about your next project.",
		"footer.cta":         "Ready to build something together?",
		"footer.rights":      "All rights reserved.",
		"theme.to_dark":      "Switch to dark theme",
		"theme.to_light":     "Switch to light theme",
		"login.aria":         "Sign in",
		"lang.label":         "Language",
		"lang.switch_to":     "Switch to Spanish",
		"sidebar.aria":       "Sidebar",
	},
}

// T returns the fixed UI translation for a key in the given lang.
// Falls back to Spanish, then to the key itself.
func T(lang, key string) string {
	if m, ok := translations[lang]; ok {
		if v, ok := m[key]; ok {
			return v
		}
	}
	if v, ok := translations["es"][key]; ok {
		return v
	}
	return key
}
