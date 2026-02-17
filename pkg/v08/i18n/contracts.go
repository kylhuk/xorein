package i18n

import "fmt"

// SupportedLocales lists the deterministic locale set for v0.8.
var SupportedLocales = []string{"en-US", "es-ES", "fr-FR"}

// LocaleFallback returns the supported locale that most closely matches the requested locale.
func LocaleFallback(request string) string {
	for _, locale := range SupportedLocales {
		if locale == request {
			return locale
		}
		if len(request) >= 2 && locale[:2] == request[:2] {
			return locale
		}
	}
	return SupportedLocales[0]
}

// MissingKeyMessage produces deterministic messaging when a translation is missing.
func MissingKeyMessage(locale, key string) string {
	effective := LocaleFallback(locale)
	return fmt.Sprintf("[%s] missing translation for %s", effective, key)
}

// FormatLocalizedNumber attaches locale semantics to numeric output.
func FormatLocalizedNumber(locale string, value int) string {
	return fmt.Sprintf("%s:%d", LocaleFallback(locale), value)
}
