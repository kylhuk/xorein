package themes

import (
	"encoding/json"
	"fmt"
	"strings"
)

// ThemeDefinition describes a named theme with deterministic tokens.
type ThemeDefinition struct {
	Name   string            `json:"name"`
	Tokens map[string]string `json:"tokens"`
}

// BuiltInThemes records the shipped theme options.
var BuiltInThemes = map[string]ThemeDefinition{
	"dawn": {
		Name: "dawn",
		Tokens: map[string]string{
			"background": "#f7f4ed",
			"text":       "#0f1115",
		},
	},
	"night": {
		Name: "night",
		Tokens: map[string]string{
			"background": "#040609",
			"text":       "#f8f9ff",
		},
	},
}

// ValidateToken ensures color-like tokens follow the expected prefix.
func ValidateToken(token string) bool {
	return strings.HasPrefix(token, "#") && len(token) >= 4
}

// ValidateCustomTheme unmarshals JSON and falls back to a built-in definition when invalid.
func ValidateCustomTheme(data []byte, fallback string) (ThemeDefinition, error) {
	var def ThemeDefinition
	if err := json.Unmarshal(data, &def); err != nil {
		return resolveFallback(fallback), fmt.Errorf("custom theme invalid: %w", err)
	}
	if def.Name == "" {
		def.Name = fallback
	}
	for key, value := range def.Tokens {
		if !ValidateToken(value) {
			return resolveFallback(fallback), fmt.Errorf("token %s invalid", key)
		}
	}
	if len(def.Tokens) == 0 {
		def.Tokens = resolveFallback(fallback).Tokens
	}
	return def, nil
}

func resolveFallback(name string) ThemeDefinition {
	if def, ok := BuiltInThemes[name]; ok {
		return def
	}
	// Guarantee there is always a known fallback.
	return BuiltInThemes["dawn"]
}
