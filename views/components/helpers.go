package components

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
