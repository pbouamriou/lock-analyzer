package i18n

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
)

// Translator gère les traductions pour les rapports
type Translator struct {
	bundle    *i18n.Bundle
	localizer *i18n.Localizer
}

// detectSystemLanguage tente de détecter la langue système à partir des variables d'environnement
func detectSystemLanguage() string {
	candidates := []string{
		os.Getenv("LANG"),
		os.Getenv("LC_ALL"),
		os.Getenv("LC_MESSAGES"),
	}
	for _, val := range candidates {
		if val == "" {
			continue
		}
		lang := strings.ToLower(val)
		if strings.Contains(lang, "fr") {
			return "fr"
		}
		if strings.Contains(lang, "en") {
			return "en"
		}
		if strings.Contains(lang, "es") {
			return "es"
		}
		if strings.Contains(lang, "de") {
			return "de"
		}
	}
	return "fr" // fallback
}

// NewTranslator crée un nouveau traducteur pour la langue spécifiée
func NewTranslator(lang string) *Translator {
	bundle := i18n.NewBundle(language.English)
	bundle.RegisterUnmarshalFunc("json", json.Unmarshal)

	// Déterminer la langue à utiliser
	chosenLang := lang
	if chosenLang == "" || !IsValidLanguage(chosenLang) {
		chosenLang = detectSystemLanguage()
	}
	if !IsValidLanguage(chosenLang) {
		chosenLang = "fr"
	}

	// Charger les fichiers de traduction
	// Essayer plusieurs chemins possibles pour le dossier locales
	possiblePaths := []string{
		"locales",          // Depuis le répertoire racine
		"../locales",       // Depuis un sous-répertoire
		"../../locales",    // Depuis un sous-sous-répertoire
		"../../../locales", // Depuis un sous-sous-sous-répertoire
	}

	var localesDir string
	for _, path := range possiblePaths {
		if _, err := os.Stat(path); err == nil {
			localesDir = path
			break
		}
	}

	if localesDir != "" {
		// Charger tous les fichiers .json dans le dossier locales
		files, err := filepath.Glob(filepath.Join(localesDir, "*.json"))
		if err == nil {
			for _, file := range files {
				fmt.Printf("Chargement du fichier de traduction: %s\n", file)
				bundle.MustLoadMessageFile(file)
			}
		} else {
			fmt.Printf("Erreur lors de la recherche des fichiers de traduction: %v\n", err)
		}
	} else {
		fmt.Printf("Dossier locales non trouvé dans les chemins: %v\n", possiblePaths)
	}

	localizer := i18n.NewLocalizer(bundle, chosenLang)

	return &Translator{
		bundle:    bundle,
		localizer: localizer,
	}
}

// T traduit une clé avec des arguments optionnels
func (t *Translator) T(key string, args ...interface{}) string {
	if len(args) == 0 {
		translation, err := t.localizer.Localize(&i18n.LocalizeConfig{
			MessageID: key,
		})
		if err != nil {
			return key
		}
		return translation
	}

	// Si des arguments sont fournis, les traiter comme des données de template
	templateData := make(map[string]interface{})
	for i, arg := range args {
		templateData[fmt.Sprintf("arg%d", i+1)] = arg
	}

	translation, err := t.localizer.Localize(&i18n.LocalizeConfig{
		MessageID:    key,
		TemplateData: templateData,
	})
	if err != nil {
		return key
	}
	return translation
}

// TWithData traduit une clé avec des données de template spécifiques
func (t *Translator) TWithData(key string, data map[string]interface{}) string {
	translation, err := t.localizer.Localize(&i18n.LocalizeConfig{
		MessageID:    key,
		TemplateData: data,
	})
	if err != nil {
		return key
	}
	return translation
}

// GetAvailableLanguages retourne la liste des langues disponibles
func GetAvailableLanguages() []string {
	return []string{"fr", "en", "es", "de"}
}

// IsValidLanguage vérifie si une langue est valide
func IsValidLanguage(lang string) bool {
	validLangs := GetAvailableLanguages()
	for _, validLang := range validLangs {
		if lang == validLang {
			return true
		}
	}
	return false
}
