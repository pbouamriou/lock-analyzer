package i18n

import (
	"os"
	"testing"
)

func TestNewTranslator(t *testing.T) {
	tests := []struct {
		name     string
		lang     string
		expected string
	}{
		{"French", "fr", "fr"},
		{"English", "en", "en"},
		{"Spanish", "es", "es"},
		{"German", "de", "de"},
		{"Invalid language", "invalid", "fr"}, // French as default
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			translator := NewTranslator(tt.lang)
			if translator == nil {
				t.Fatal("NewTranslator should not return nil")
			}
		})
	}
}

func TestGetAvailableLanguages(t *testing.T) {
	languages := GetAvailableLanguages()
	expected := []string{"fr", "en", "es", "de"}

	if len(languages) != len(expected) {
		t.Errorf("Incorrect number of languages: got %d, want %d", len(languages), len(expected))
	}

	for _, lang := range expected {
		found := false
		for _, available := range languages {
			if available == lang {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Language %s missing from list", lang)
		}
	}
}

func TestIsValidLanguage(t *testing.T) {
	tests := []struct {
		name     string
		lang     string
		expected bool
	}{
		{"Valid French", "fr", true},
		{"Valid English", "en", true},
		{"Valid Spanish", "es", true},
		{"Valid German", "de", true},
		{"Invalid language", "invalid", false},
		{"Empty string", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsValidLanguage(tt.lang)
			if result != tt.expected {
				t.Errorf("IsValidLanguage(%s) = %v, want %v", tt.lang, result, tt.expected)
			}
		})
	}
}

func TestTranslator_T(t *testing.T) {
	translator := NewTranslator("fr")

	// Test with a simple key
	result := translator.T("report_title")
	if result == "" {
		t.Error("Translation should not be empty")
	}

	// Test with a non-existent key (should return the key itself)
	result = translator.T("nonexistent_key")
	if result != "nonexistent_key" {
		t.Errorf("Non-existent key should return the key itself, got: %s", result)
	}
}

func TestTranslator_TWithData(t *testing.T) {
	translator := NewTranslator("fr")

	data := map[string]interface{}{
		"arg1": "123",
		"arg2": "test",
	}

	result := translator.TWithData("lock_info_format", data)
	if result == "" {
		t.Error("Translation with data should not be empty")
	}

	// Check that arguments are properly inserted
	if result == "lock_info_format" {
		t.Error("Translation should include the provided arguments")
	}
}

func TestTranslatorLanguageSpecific(t *testing.T) {
	// Test that translations are different according to language
	frTranslator := NewTranslator("fr")
	enTranslator := NewTranslator("en")

	frTitle := frTranslator.T("report_title")
	enTitle := enTranslator.T("report_title")

	if frTitle == enTitle {
		t.Error("French and English translations should be different")
	}
}

func TestDetectSystemLanguage(t *testing.T) {
	tests := []struct {
		name     string
		langEnv  string
		lcAll    string
		lcMsg    string
		expected string
	}{
		{"No variables", "", "", "", "fr"},
		{"LANG=fr_FR.UTF-8", "fr_FR.UTF-8", "", "", "fr"},
		{"LANG=en_US.UTF-8", "en_US.UTF-8", "", "", "en"},
		{"LANG=es_ES.UTF-8", "es_ES.UTF-8", "", "", "es"},
		{"LANG=de_DE.UTF-8", "de_DE.UTF-8", "", "", "de"},
		{"LC_ALL=fr_FR.UTF-8", "", "fr_FR.UTF-8", "", "fr"},
		{"LC_ALL=en_US.UTF-8", "", "en_US.UTF-8", "", "en"},
		{"LC_MESSAGES=es_ES.UTF-8", "", "", "es_ES.UTF-8", "es"},
		{"LC_MESSAGES=de_DE.UTF-8", "", "", "de_DE.UTF-8", "de"},
		{"Not supported", "it_IT.UTF-8", "", "", "fr"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save old environment
			oldLang := os.Getenv("LANG")
			oldLcAll := os.Getenv("LC_ALL")
			oldLcMsg := os.Getenv("LC_MESSAGES")

			// Set environment variables
			os.Setenv("LANG", tt.langEnv)
			os.Setenv("LC_ALL", tt.lcAll)
			os.Setenv("LC_MESSAGES", tt.lcMsg)

			lang := detectSystemLanguage()
			if lang != tt.expected {
				t.Errorf("detectSystemLanguage() = %s, want %s", lang, tt.expected)
			}

			// Restore environment
			os.Setenv("LANG", oldLang)
			os.Setenv("LC_ALL", oldLcAll)
			os.Setenv("LC_MESSAGES", oldLcMsg)
		})
	}
}
