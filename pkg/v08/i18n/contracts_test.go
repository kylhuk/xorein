package i18n

import "testing"

func TestLocaleFallback(t *testing.T) {
	cases := []struct {
		request string
		want    string
	}{
		{"en-US", "en-US"},
		{"es-MX", "es-ES"},
		{"fr-CA", "fr-FR"},
		{"de-DE", "en-US"},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.request, func(t *testing.T) {
			t.Parallel()
			if got := LocaleFallback(tc.request); got != tc.want {
				t.Fatalf("expected %s, got %s", tc.want, got)
			}
		})
	}
}

func TestMissingKeyMessage(t *testing.T) {
	got := MissingKeyMessage("pt-BR", "welcome")
	if got != "[en-US] missing translation for welcome" {
		t.Fatalf("unexpected message %s", got)
	}
}

func TestFormatLocalizedNumber(t *testing.T) {
	if got := FormatLocalizedNumber("es-AR", 42); got != "es-ES:42" {
		t.Fatalf("expected fallback, got %s", got)
	}
	if got := FormatLocalizedNumber("fr-FR", 7); got != "fr-FR:7" {
		t.Fatalf("unexpected formatting %s", got)
	}
}
