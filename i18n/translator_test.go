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
		{"français", "fr", "fr"},
		{"anglais", "en", "en"},
		{"espagnol", "es", "es"},
		{"allemand", "de", "de"},
		{"langue invalide", "invalid", "fr"}, // Français par défaut
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			translator := NewTranslator(tt.lang)
			if translator == nil {
				t.Fatal("NewTranslator ne devrait pas retourner nil")
			}
		})
	}
}

func TestGetAvailableLanguages(t *testing.T) {
	languages := GetAvailableLanguages()
	expected := []string{"fr", "en", "es", "de"}

	if len(languages) != len(expected) {
		t.Errorf("Nombre de langues incorrect: got %d, want %d", len(languages), len(expected))
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
			t.Errorf("Langue %s manquante dans la liste", lang)
		}
	}
}

func TestIsValidLanguage(t *testing.T) {
	tests := []struct {
		name     string
		lang     string
		expected bool
	}{
		{"français valide", "fr", true},
		{"anglais valide", "en", true},
		{"espagnol valide", "es", true},
		{"allemand valide", "de", true},
		{"langue invalide", "invalid", false},
		{"chaîne vide", "", false},
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

	// Test avec une clé simple
	result := translator.T("report_title")
	if result == "" {
		t.Error("La traduction ne devrait pas être vide")
	}

	// Test avec une clé inexistante (devrait retourner la clé elle-même)
	result = translator.T("nonexistent_key")
	if result != "nonexistent_key" {
		t.Errorf("Clé inexistante devrait retourner la clé elle-même, got: %s", result)
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
		t.Error("La traduction avec données ne devrait pas être vide")
	}

	// Vérifier que les arguments sont bien insérés
	if result == "lock_info_format" {
		t.Error("La traduction devrait inclure les arguments fournis")
	}
}

func TestTranslatorLanguageSpecific(t *testing.T) {
	// Test que les traductions sont différentes selon la langue
	frTranslator := NewTranslator("fr")
	enTranslator := NewTranslator("en")

	frTitle := frTranslator.T("report_title")
	enTitle := enTranslator.T("report_title")

	if frTitle == enTitle {
		t.Error("Les traductions français et anglais devraient être différentes")
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
		{"Aucune variable", "", "", "", "fr"},
		{"LANG=fr_FR.UTF-8", "fr_FR.UTF-8", "", "", "fr"},
		{"LANG=en_US.UTF-8", "en_US.UTF-8", "", "", "en"},
		{"LANG=es_ES.UTF-8", "es_ES.UTF-8", "", "", "es"},
		{"LANG=de_DE.UTF-8", "de_DE.UTF-8", "", "", "de"},
		{"LC_ALL=fr_FR.UTF-8", "", "fr_FR.UTF-8", "", "fr"},
		{"LC_ALL=en_US.UTF-8", "", "en_US.UTF-8", "", "en"},
		{"LC_MESSAGES=es_ES.UTF-8", "", "", "es_ES.UTF-8", "es"},
		{"LC_MESSAGES=de_DE.UTF-8", "", "", "de_DE.UTF-8", "de"},
		{"Non supporté", "it_IT.UTF-8", "", "", "fr"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Sauvegarder l'ancien environnement
			oldLang := os.Getenv("LANG")
			oldLcAll := os.Getenv("LC_ALL")
			oldLcMsg := os.Getenv("LC_MESSAGES")

			// Définir les variables d'environnement
			os.Setenv("LANG", tt.langEnv)
			os.Setenv("LC_ALL", tt.lcAll)
			os.Setenv("LC_MESSAGES", tt.lcMsg)

			lang := detectSystemLanguage()
			if lang != tt.expected {
				t.Errorf("detectSystemLanguage() = %s, want %s", lang, tt.expected)
			}

			// Restaurer l'environnement
			os.Setenv("LANG", oldLang)
			os.Setenv("LC_ALL", oldLcAll)
			os.Setenv("LC_MESSAGES", oldLcMsg)
		})
	}
}
